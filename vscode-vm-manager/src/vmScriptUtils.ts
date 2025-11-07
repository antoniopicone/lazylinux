import * as vscode from 'vscode';
import { spawn, exec } from 'child_process';
import { promisify } from 'util';
import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';

const execAsync = promisify(exec);

export interface VmInfo {
    name: string;
    status: string;
    arch: string;
    image: string;
    hostname: string;
    ipAddress: string;
    username: string;
    password: string;
    sshPort: string;
    uptime?: string;
}

export interface VmCreateOptions {
    name?: string;
    arch?: 'arm64' | 'amd64';
    user?: string;
    pass?: string;
    memory?: string;
    cpus?: string;
    disk?: string;
    image?: string;
}

export class VmScriptUtils {
    private getScriptPath(): string {
        const config = vscode.workspace.getConfiguration('vmManager');
        return config.get<string>('scriptPath') || './vm';
    }

    // Sanitize VM name: replace underscores with hyphens (matches vm script behavior)
    private sanitizeName(name: string): string {
        return name.replace(/_/g, '-');
    }

    private async executeCommand(args: string[]): Promise<string> {
        const scriptPath = this.getScriptPath();
        const command = `${scriptPath} ${args.join(' ')}`;

        try {
            const { stdout, stderr } = await execAsync(command);
            // Don't treat all stderr as errors - some commands output info to stderr
            // Only fail if the command itself failed (non-zero exit)
            return stdout.trim();
        } catch (error: any) {
            // Check if the error is due to script not being found
            if (error.message.includes('command not found') ||
                error.message.includes('No such file or directory') ||
                error.code === 'ENOENT') {
                throw new Error('VM_SCRIPT_NOT_FOUND');
            }

            // Check if it's a sudo/permission error
            if (error.message.includes('cannot be stopped without sudo') ||
                error.message.includes('NOPASSWD: /bin/kill') ||
                error.message.includes('Failed to send stop signal')) {
                throw new Error('VM_REQUIRES_SUDO');
            }

            // Include both stdout and stderr in error for better debugging
            const errorMsg = error.stderr || error.stdout || error.message || 'Command execution failed';
            throw new Error(errorMsg);
        }
    }

    async checkScriptAvailability(): Promise<boolean> {
        try {
            await this.executeCommand(['--version']);
            return true;
        } catch (error: any) {
            return false;
        }
    }

    async listVms(): Promise<VmInfo[]> {
        try {
            return await this.loadVmsFromDisk();
        } catch (error) {
            console.error('Failed to list VMs:', error);
            return [];
        }
    }

    private async loadVmsFromDisk(): Promise<VmInfo[]> {
        const vms: VmInfo[] = [];
        const vmDir = path.join(os.homedir(), '.vm', 'vms');

        // Check if VM directory exists
        if (!fs.existsSync(vmDir)) {
            return [];
        }

        // Read all VM directories
        const vmDirs = fs.readdirSync(vmDir);

        for (const vmName of vmDirs) {
            const vmPath = path.join(vmDir, vmName);
            const infoPath = path.join(vmPath, 'info.json');
            const pidPath = path.join(vmPath, 'qemu.pid');

            // Skip if not a directory or info.json doesn't exist
            if (!fs.statSync(vmPath).isDirectory() || !fs.existsSync(infoPath)) {
                continue;
            }

            try {
                // Read and parse info.json
                const infoData = fs.readFileSync(infoPath, 'utf8');
                const info = JSON.parse(infoData);

                const name = info.name || vmName;
                const arch = info.arch || 'unknown';
                const image = info.image || 'unknown';
                const username = info.username || '-';
                const password = info.password || '-';
                const netType = info.net_type || 'bridge';

                // Get SSH info
                let hostname = '-';
                let sshPort = '22';

                if (info.ssh) {
                    if (netType === 'bridge') {
                        // For bridge mode, use mDNS hostname
                        hostname = `${name}.local`;
                        sshPort = '22';
                    } else {
                        // For port forwarding mode
                        hostname = info.ssh.host || '127.0.0.1';
                        sshPort = String(info.ssh.port || '22');
                    }
                } else {
                    hostname = `${name}.local`;
                }

                // Check if VM is running
                let isRunning = false;
                if (fs.existsSync(pidPath)) {
                    try {
                        const pidString = fs.readFileSync(pidPath, 'utf8').trim();
                        if (pidString) {
                            const pid = parseInt(pidString, 10);
                            if (!isNaN(pid)) {
                                // Check if process is alive using ps command. It might fail if the process is gone.
                                try {
                                    const { stdout } = await execAsync(`ps -p ${pid}`);
                                    isRunning = stdout.includes(String(pid));
                                } catch {
                                    isRunning = false;
                                }
                            }
                        }
                    } catch {
                        isRunning = false;
                    }
                }

                // Fallback for bridge mode or stale PID files: check with pgrep
                if (!isRunning) {
                    try {
                        const { stdout } = await execAsync(`pgrep -f "qemu.*-name ${name}"`);
                        isRunning = stdout.trim().length > 0;
                    } catch {
                        isRunning = false;
                    }
                }

                const status = isRunning ? 'RUNNING' : 'STOPPED';

                // Get IP address if VM is running
                let ipAddress = '-';
                if (isRunning) {
                    if (netType === 'bridge') {
                        // First, try to read IP from info.json (cached by CLI script)
                        if (info.ssh && info.ssh.host && info.ssh.host !== 'dhcp-assigned' && info.ssh.host !== '127.0.0.1') {
                            ipAddress = info.ssh.host;
                        } else {
                            // Fallback: try to resolve mDNS hostname
                            try {
                                const { stdout } = await execAsync(`ping -c 1 -t 1 ${hostname} 2>/dev/null | grep -oE '\\([0-9.]+\\)' | tr -d '()'`);
                                const resolvedIP = stdout.trim();
                                if (resolvedIP && resolvedIP.match(/^\d+\.\d+\.\d+\.\d+$/)) {
                                    ipAddress = resolvedIP;
                                }
                            } catch {
                                // If ping fails, try using arp to find IP
                                try {
                                    const { stdout } = await execAsync(`arp -a | grep -i "${hostname}" | grep -oE '\\([0-9.]+\\)' | tr -d '()'`);
                                    const arpIP = stdout.trim();
                                    if (arpIP && arpIP.match(/^\d+\.\d+\.\d+\.\d+$/)) {
                                        ipAddress = arpIP;
                                    }
                                } catch {
                                    ipAddress = '-';
                                }
                            }
                        }
                    } else {
                        // For port forwarding mode, use localhost
                        ipAddress = '127.0.0.1';
                    }
                }

                vms.push({
                    name,
                    status,
                    arch,
                    image,
                    hostname,
                    ipAddress,
                    username,
                    password,
                    sshPort,
                    uptime: isRunning ? 'Active' : undefined
                });
            } catch (error) {
                console.error(`Failed to parse VM info for ${vmName}:`, error);
                continue;
            }
        }

        // Sort by name
        return vms.sort((a, b) => a.name.localeCompare(b.name));
    }

    async createVm(options: VmCreateOptions = {}): Promise<void> {
        const args = ['create'];

        if (options.name) {
            args.push('--name', options.name);
        }
        if (options.arch) {
            args.push('--arch', options.arch);
        }
        if (options.user) {
            args.push('--user', options.user);
        }
        if (options.pass) {
            args.push('--pass', options.pass);
        }
        if (options.memory) {
            args.push('--memory', options.memory);
        }
        if (options.cpus) {
            args.push('--cpus', options.cpus);
        }
        if (options.disk) {
            args.push('--disk', options.disk);
        }

        await this.executeCommand(args);

        // Wait a moment for the VM to fully initialize and get its IP
        // The CLI script extracts IP from console.log after cloud-init completes
        if (options.name) {
            await this.waitForVmIp(options.name, 10);
        }
    }

    async startVm(name: string): Promise<void> {
        await this.executeCommand(['start', name]);

        // Wait a moment for the VM to get its IP
        await this.waitForVmIp(name, 10);
    }

    async stopVm(name: string): Promise<void> {
        await this.executeCommand(['stop', name]);
    }

    async deleteVm(name: string): Promise<void> {
        await this.executeCommand(['delete', name, '--force']);
    }

    async getVmIp(name: string): Promise<string> {
        const output = await this.executeCommand(['ip', name]);

        // Extract IP from output - new format: "VM 'name' IP Address: 192.168.105.20"
        const ipMatch = output.match(/IP Address: (\d+\.\d+\.\d+\.\d+)/);
        if (ipMatch) {
            return ipMatch[1];
        }

        // Check for port forwarding - format: "ssh user@127.0.0.1 -p PORT"
        const portMatch = output.match(/ssh.*@127\.0\.0\.1.*-p\s+(\d+)/);
        if (portMatch) {
            return `127.0.0.1:${portMatch[1]}`;
        }

        throw new Error('Could not determine VM IP');
    }

    private async waitForVmIp(name: string, maxRetries: number = 10): Promise<string | null> {
        // Wait for the VM info.json to be updated with IP address
        // The CLI script updates this after cloud-init completes
        const vmDir = path.join(os.homedir(), '.vm', 'vms', name);
        const infoPath = path.join(vmDir, 'info.json');

        for (let i = 0; i < maxRetries; i++) {
            try {
                if (fs.existsSync(infoPath)) {
                    const infoData = fs.readFileSync(infoPath, 'utf8');
                    const info = JSON.parse(infoData);

                    // Check if IP is available in info.json
                    if (info.ssh && info.ssh.host && info.ssh.host !== 'dhcp-assigned' && info.ssh.host !== '127.0.0.1') {
                        return info.ssh.host;
                    }

                    // For bridge mode, also try to get IP via the CLI command
                    if (info.net_type === 'bridge') {
                        try {
                            const ip = await this.getVmIp(name);
                            if (ip && !ip.includes('not running') && !ip.includes('Could not')) {
                                return ip;
                            }
                        } catch {
                            // IP not available yet, continue waiting
                        }
                    }
                }
            } catch (error) {
                // Continue retrying
            }

            // Wait 2 seconds before next retry
            await new Promise(resolve => setTimeout(resolve, 2000));
        }

        return null;
    }

    async getAvailableImages(): Promise<string[]> {
        // Based on the VM script help, these are commonly available images
        // This could be enhanced to dynamically fetch from the script if it supports listing images
        return [
            'debian13',
            'debian12'
        ];
    }

    async getAvailableArchitectures(): Promise<string[]> {
        return ['arm64', 'amd64'];
    }
}