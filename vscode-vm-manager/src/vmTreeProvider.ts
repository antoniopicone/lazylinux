import * as vscode from 'vscode';
import { VmScriptUtils, VmInfo } from './vmScriptUtils';

interface VmDetailItem {
    label: string;
    value: string;
}

export class VmTreeProvider implements vscode.TreeDataProvider<VmTreeItem | VmDetailTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<VmTreeItem | VmDetailTreeItem | undefined | null | void> = new vscode.EventEmitter<VmTreeItem | VmDetailTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<VmTreeItem | VmDetailTreeItem | undefined | null | void> = this._onDidChangeTreeData.event;

    constructor(private scriptUtils: VmScriptUtils) {}

    refresh(): void {
        this._onDidChangeTreeData.fire();
    }

    getTreeItem(element: VmTreeItem | VmDetailTreeItem): vscode.TreeItem {
        return element;
    }

    async getChildren(element?: VmTreeItem | VmDetailTreeItem): Promise<(VmTreeItem | VmDetailTreeItem)[]> {
        if (!element) {
            // Root level - return all VMs
            try {
                const vms = await this.scriptUtils.listVms();
                return vms.map(vm => new VmTreeItem(vm));
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to list VMs: ${error}`);
                return [];
            }
        } else if (element instanceof VmTreeItem) {
            // VM selected - return detail items
            return this.getVmDetails(element.vm);
        }
        return [];
    }

    private getVmDetails(vm: VmInfo): VmDetailTreeItem[] {
        const hostname = vm.hostname === '-' ? vm.name + '.local' : vm.hostname;
        const username = vm.username || 'unknown';
        const password = vm.password || 'unknown';
        const ipAddress = vm.ipAddress || '-';

        const details: VmDetailItem[] = [
            { label: 'Status', value: vm.status },
            { label: 'IP Address', value: ipAddress },
            { label: 'Hostname', value: hostname },
            { label: 'Username', value: username },
            { label: 'Password', value: this.hidePassword(password) }
        ];

        // Add SSH port if not standard port 22
        if (vm.sshPort && vm.sshPort !== '22') {
            details.push({ label: 'SSH Port', value: vm.sshPort });
        }

        // if (vm.uptime) {
        //     details.splice(1, 0, { label: 'Uptime', value: vm.uptime });
        // }

        return details.map(detail => {
            const copyable = detail.label === 'IP Address' || detail.label === 'Hostname' || detail.label === 'Username' || detail.label === 'Password';
            let actualValue = detail.value;

            // For password, use the actual password value for copying, not the hidden version
            if (detail.label === 'Password') {
                actualValue = password;
            }

            return new VmDetailTreeItem(detail.label, detail.value, copyable, actualValue);
        });
    }

    private hidePassword(password: string): string {
        if (password === 'unknown') return password;
        return '•'.repeat(password.length);
    }
}

class VmTreeItem extends vscode.TreeItem {
    constructor(public readonly vm: VmInfo) {
        super(vm.name, vscode.TreeItemCollapsibleState.Collapsed);
        
        this.tooltip = `${vm.name} (${vm.status})`;
        this.description = this.getDescription();
        this.iconPath = this.getIcon();
        this.contextValue = vm.status === 'RUNNING' ? 'vm-running' : 'vm-stopped';
    }

    get name(): string {
        return this.vm.name;
    }

    get hostname(): string {
        return this.vm.hostname;
    }

    get ipAddress(): string {
        return this.vm.ipAddress;
    }

    get username(): string {
        return this.vm.username;
    }

    get password(): string {
        return this.vm.password;
    }

    private getDescription(): string {
        const image = this.vm.image === '-' ? 'unknown' : this.vm.image;
        const arch = this.vm.arch === '-' ? 'unknown' : this.vm.arch;
        return `${image} - ${arch}`;
    }

    private getIcon(): vscode.ThemeIcon {
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
    constructor(
        public readonly detailLabel: string,
        value: string,
        public readonly copyable: boolean = false,
        public readonly actualValue: string = value
    ) {
        super(`${detailLabel}: ${value}`, vscode.TreeItemCollapsibleState.None);
        
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