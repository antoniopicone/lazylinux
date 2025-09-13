"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.VmScriptUtils = void 0;
const vscode = require("vscode");
const child_process_1 = require("child_process");
const util_1 = require("util");
const execAsync = (0, util_1.promisify)(child_process_1.exec);
class VmScriptUtils {
    getScriptPath() {
        const config = vscode.workspace.getConfiguration('vmManager');
        return config.get('scriptPath') || './vm';
    }
    async executeCommand(args) {
        const scriptPath = this.getScriptPath();
        const command = `${scriptPath} ${args.join(' ')}`;
        try {
            const { stdout, stderr } = await execAsync(command);
            if (stderr) {
                throw new Error(stderr);
            }
            return stdout.trim();
        }
        catch (error) {
            throw new Error(error.message || 'Command execution failed');
        }
    }
    async listVms() {
        try {
            const output = await this.executeCommand(['list']);
            return this.parseVmList(output);
        }
        catch (error) {
            console.error('Failed to list VMs:', error);
            return [];
        }
    }
    parseVmList(output) {
        const lines = output.split('\n').filter(line => line.trim());
        const vms = [];
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
    async createVm(options = {}) {
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
    async startVm(name) {
        await this.executeCommand(['start', name]);
    }
    async stopVm(name) {
        await this.executeCommand(['stop', name]);
    }
    async deleteVm(name) {
        await this.executeCommand(['delete', name, '--force']);
    }
    async getVmIp(name) {
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
    async getAvailableImages() {
        // Based on the VM script help, these are commonly available images
        // This could be enhanced to dynamically fetch from the script if it supports listing images
        return [
            'debian13',
            'debian12',
            'ubuntu22',
            'ubuntu20',
            'alpine'
        ];
    }
    async getAvailableArchitectures() {
        return ['arm64', 'amd64'];
    }
}
exports.VmScriptUtils = VmScriptUtils;
//# sourceMappingURL=vmScriptUtils.js.map