"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.VmTreeProvider = void 0;
const vscode = require("vscode");
class VmTreeProvider {
    constructor(scriptUtils) {
        this.scriptUtils = scriptUtils;
        this._onDidChangeTreeData = new vscode.EventEmitter();
        this.onDidChangeTreeData = this._onDidChangeTreeData.event;
    }
    refresh() {
        this._onDidChangeTreeData.fire();
    }
    getTreeItem(element) {
        return element;
    }
    async getChildren(element) {
        if (!element) {
            // Root level - return all VMs
            try {
                const vms = await this.scriptUtils.listVms();
                return vms.map(vm => new VmTreeItem(vm));
            }
            catch (error) {
                vscode.window.showErrorMessage(`Failed to list VMs: ${error}`);
                return [];
            }
        }
        else if (element instanceof VmTreeItem) {
            // VM selected - return detail items
            return this.getVmDetails(element.vm);
        }
        return [];
    }
    getVmDetails(vm) {
        const hostname = vm.hostname === '-' ? vm.name + '.local' : vm.hostname;
        const credentials = vm.credentials === '-' ? 'unknown' : vm.credentials;
        const details = [
            { label: 'Status', value: vm.status },
            { label: 'Hostname', value: hostname }
        ];
        // Add username and password as separate fields
        if (credentials !== 'unknown') {
            const { username, password } = this.parseCredentials(credentials);
            details.push({ label: 'Username', value: username });
            details.push({ label: 'Password', value: this.hidePassword(password) });
        }
        else {
            details.push({ label: 'Username', value: 'unknown' });
            details.push({ label: 'Password', value: 'unknown' });
        }
        // if (vm.uptime) {
        //     details.splice(1, 0, { label: 'Uptime', value: vm.uptime });
        // }
        return details.map(detail => {
            const copyable = detail.label === 'Hostname' || detail.label === 'Username' || detail.label === 'Password';
            let actualValue = detail.value;
            // For password, use the actual password value for copying, not the hidden version
            if (detail.label === 'Password' && credentials !== 'unknown') {
                const { password } = this.parseCredentials(credentials);
                actualValue = password;
            }
            return new VmDetailTreeItem(detail.label, detail.value, copyable, actualValue);
        });
    }
    parseCredentials(credentials) {
        if (credentials === 'unknown') {
            return { username: 'unknown', password: 'unknown' };
        }
        // Parse "username / password" format
        const parts = credentials.split(' / ');
        if (parts.length === 2) {
            return { username: parts[0], password: parts[1] };
        }
        // Fallback if format is different
        return { username: credentials, password: 'unknown' };
    }
    hidePassword(password) {
        if (password === 'unknown')
            return password;
        return '•'.repeat(password.length);
    }
}
exports.VmTreeProvider = VmTreeProvider;
class VmTreeItem extends vscode.TreeItem {
    constructor(vm) {
        super(vm.name, vscode.TreeItemCollapsibleState.Collapsed);
        this.vm = vm;
        this.tooltip = `${vm.name} (${vm.status})`;
        this.description = this.getDescription();
        this.iconPath = this.getIcon();
        this.contextValue = vm.status === 'RUNNING' ? 'vm-running' : 'vm-stopped';
    }
    get name() {
        return this.vm.name;
    }
    get hostname() {
        return this.vm.hostname;
    }
    get credentials() {
        return this.vm.credentials;
    }
    getDescription() {
        const image = this.vm.image === '-' ? 'unknown' : this.vm.image;
        const arch = this.vm.arch === '-' ? 'unknown' : this.vm.arch;
        return `${image} - ${arch}`;
    }
    getIcon() {
        switch (this.vm.status) {
            case 'RUNNING':
                return new vscode.ThemeIcon('vm', new vscode.ThemeColor('charts.green'));
            case 'STOPPED':
                return new vscode.ThemeIcon('vm', new vscode.ThemeColor('charts.red'));
            default:
                return new vscode.ThemeIcon('vm', new vscode.ThemeColor('charts.yellow'));
        }
    }
}
class VmDetailTreeItem extends vscode.TreeItem {
    constructor(detailLabel, value, copyable = false, actualValue = value) {
        super(`${detailLabel}: ${value}`, vscode.TreeItemCollapsibleState.None);
        this.detailLabel = detailLabel;
        this.copyable = copyable;
        this.actualValue = actualValue;
        this.iconPath = copyable ? new vscode.ThemeIcon('copy') : new vscode.ThemeIcon('info');
        this.contextValue = copyable ? 'vm-detail-copyable' : 'vm-detail';
        this.tooltip = copyable ? `Click to copy ${detailLabel.toLowerCase()}` : `${detailLabel}: ${value}`;
        if (copyable) {
            this.command = {
                command: 'vmManager.copyDetail',
                title: 'Copy',
                arguments: [this.actualValue, this.detailLabel]
            };
        }
    }
}
//# sourceMappingURL=vmTreeProvider.js.map