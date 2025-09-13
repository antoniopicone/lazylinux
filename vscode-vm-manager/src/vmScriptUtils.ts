import * as vscode from 'vscode';
import { spawn, exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

export interface VmInfo {
    name: string;
    status: string;
    arch: string;
    image: string;
    hostname: string;
    credentials: string;
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

    private async executeCommand(args: string[]): Promise<string> {
        const scriptPath = this.getScriptPath();
        const command = `${scriptPath} ${args.join(' ')}`;
        
        try {
            const { stdout, stderr } = await execAsync(command);
            if (stderr) {
                throw new Error(stderr);
            }
            return stdout.trim();
        } catch (error: any) {
            throw new Error(error.message || 'Command execution failed');
        }
    }

    async listVms(): Promise<VmInfo[]> {
        try {
            const output = await this.executeCommand(['list']);
            return this.parseVmList(output);
        } catch (error) {
            console.error('Failed to list VMs:', error);
            return [];
        }
    }

    private parseVmList(output: string): VmInfo[] {
        const lines = output.split('\n').filter(line => line.trim());
        const vms: VmInfo[] = [];

        // Skip header lines and empty lines
        let dataStarted = false;
        for (const line of lines) {
            if (line.includes('----')) {
                dataStarted = true;
                continue;
            }
            if (!dataStarted || line.startsWith('Total VMs:') || line === 'No VMs found') {
                continue;
            }

            // Parse VM line: NAME STATUS ARCH IMAGE HOSTNAME CREDENTIALS
            const parts = line.split(/\s+/);
            if (parts.length >= 6) {
                const name = parts[0];
                const status = parts[1];
                const arch = parts[2];
                const image = parts[3];
                const hostname = parts[4];
                const credentials = parts.slice(5).join(' ');

                vms.push({
                    name,
                    status,
                    arch: arch === '-' ? 'unknown' : arch,
                    image: image === '-' ? 'unknown' : image,
                    hostname: hostname === '-' ? name + '.local' : hostname,
                    credentials: credentials === '-' ? 'unknown' : credentials,
                    uptime: status === 'RUNNING' ? 'Active' : undefined
                });
            }
        }

        return vms;
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
    }

    async startVm(name: string): Promise<void> {
        await this.executeCommand(['start', name]);
    }

    async stopVm(name: string): Promise<void> {
        await this.executeCommand(['stop', name]);
    }

    async deleteVm(name: string): Promise<void> {
        await this.executeCommand(['delete', name, '--force']);
    }

    async getVmIp(name: string): Promise<string> {
        const output = await this.executeCommand(['ip', name]);
        
        // Extract IP from output
        const ipMatch = output.match(/Found IP: (\d+\.\d+\.\d+\.\d+)/);
        if (ipMatch) {
            return ipMatch[1];
        }

        // Check for port forwarding
        const portMatch = output.match(/ssh user@127\.0\.0\.1 -p (\d+)/);
        if (portMatch) {
            return `127.0.0.1:${portMatch[1]}`;
        }

        throw new Error('Could not determine VM IP');
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