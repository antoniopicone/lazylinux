import * as vscode from 'vscode';
import { VmTreeProvider } from './vmTreeProvider';
import { VmScriptUtils, VmCreateOptions } from './vmScriptUtils';

async function handleScriptNotFound(context: vscode.ExtensionContext): Promise<void> {
    const config = vscode.workspace.getConfiguration('vmManager');
    const currentPath = config.get<string>('scriptPath') || '/usr/local/bin/vm';

    const action = await vscode.window.showErrorMessage(
        `VM script not found at '${currentPath}'. Please configure the correct path or download the script.`,
        'Configure Path',
        'Download Instructions',
        'Dismiss'
    );

    if (action === 'Configure Path') {
        const newPath = await vscode.window.showInputBox({
            prompt: 'Enter the full path to the vm script',
            value: currentPath,
            placeHolder: '/path/to/vm',
            validateInput: (value) => {
                if (!value || value.trim() === '') {
                    return 'Path cannot be empty';
                }
                return null;
            }
        });

        if (newPath) {
            await config.update('vmManager.scriptPath', newPath, vscode.ConfigurationTarget.Global);
            vscode.window.showInformationMessage(`VM script path updated to: ${newPath}`);
        }
    } else if (action === 'Download Instructions') {
        const downloadUrl = 'https://github.com/antoniopicone/lazylinux';
        const viewRepo = await vscode.window.showInformationMessage(
            'LazyLinux VM Manager requires the vm script.\n\n' +
            'Installation steps:\n' +
            '1. Clone the repository: git clone https://github.com/antoniopicone/lazylinux\n' +
            '2. Copy or symlink the vm script to /usr/local/bin/vm\n' +
            '3. Make it executable: chmod +x /usr/local/bin/vm\n' +
            '4. Or configure a custom path in settings',
            'Open Repository',
            'Open Settings'
        );

        if (viewRepo === 'Open Repository') {
            vscode.env.openExternal(vscode.Uri.parse(downloadUrl));
        } else if (viewRepo === 'Open Settings') {
            vscode.commands.executeCommand('workbench.action.openSettings', 'vmManager.scriptPath');
        }
    }
}

export function activate(context: vscode.ExtensionContext) {
    const scriptUtils = new VmScriptUtils();
    const vmTreeProvider = new VmTreeProvider(scriptUtils);

    // Register tree data provider
    vscode.window.createTreeView('vmManagerView', {
        treeDataProvider: vmTreeProvider,
        showCollapseAll: false
    });

    // Auto-refresh VM list every 10 seconds
    const refreshInterval = setInterval(() => {
        vmTreeProvider.refresh();
    }, 10000);

    // Clean up interval on deactivation
    context.subscriptions.push({
        dispose: () => clearInterval(refreshInterval)
    });

    // Register commands
    const commands = [
        vscode.commands.registerCommand('vmManager.refresh', () => {
            vmTreeProvider.refresh();
        }),

        vscode.commands.registerCommand('vmManager.createVm', async () => {
            try {
                // Step 1: VM Name
                const name = await vscode.window.showInputBox({
                    prompt: 'Step 1/5: Enter VM name (leave empty for auto-generated)',
                    placeHolder: 'my-vm'
                });

                if (name === undefined) return; // User cancelled

                // Step 2: Image Selection
                const availableImages = await scriptUtils.getAvailableImages();
                const selectedImage = await vscode.window.showQuickPick(availableImages, {
                    placeHolder: 'Step 2/5: Select base image',
                    ignoreFocusOut: true
                });

                if (!selectedImage) return; // User cancelled

                // Step 3: Architecture Selection
                const availableArchs = await scriptUtils.getAvailableArchitectures();
                const archItems = availableArchs.map(arch => ({
                    label: arch,
                    description: arch === 'arm64' ? '(Default - Native on Apple Silicon)' : '(Emulated on Apple Silicon)',
                    value: arch
                }));

                const selectedArchItem = await vscode.window.showQuickPick(archItems, {
                    placeHolder: 'Step 3/5: Select architecture',
                    ignoreFocusOut: true
                });

                if (!selectedArchItem) return; // User cancelled

                // Step 4: Username
                const username = await vscode.window.showInputBox({
                    prompt: 'Step 4/5: Enter username',
                    value: 'user01',
                    placeHolder: 'user01'
                });

                if (username === undefined) return; // User cancelled

                // Step 5: Password
                const password = await vscode.window.showInputBox({
                    prompt: 'Step 5/5: Enter password (leave blank for auto-generated password)',
                    placeHolder: 'Leave blank for auto-generated password',
                    password: true
                });

                if (password === undefined) return; // User cancelled

                // Create VM with collected options
                const options: VmCreateOptions = {
                    name: name || undefined,
                    arch: selectedArchItem.value as 'arm64' | 'amd64',
                    user: username,
                    pass: password || undefined,
                    image: selectedImage
                };

                const vmDisplayName = name || 'auto-generated VM';

                await vscode.window.withProgress({
                    location: vscode.ProgressLocation.Notification,
                    title: `Creating VM '${vmDisplayName}'`,
                    cancellable: false
                }, async (progress, token) => {
                    progress.report({ increment: 0, message: "Downloading image and preparing VM..." });

                    await scriptUtils.createVm(options);

                    progress.report({ increment: 100, message: "VM created successfully!" });
                });

                vmTreeProvider.refresh();

                const archDisplay = selectedArchItem.value;
                const userDisplay = username;
                const passDisplay = password ? 'custom password' : 'auto-generated password';

                vscode.window.showInformationMessage(
                    `VM '${vmDisplayName}' created successfully!\n` +
                    `Image: ${selectedImage}, Architecture: ${archDisplay}\n` +
                    `User: ${userDisplay}, Password: ${passDisplay}`
                );
            } catch (error: any) {
                if (error.message === 'VM_SCRIPT_NOT_FOUND') {
                    await handleScriptNotFound(context);
                } else {
                    vscode.window.showErrorMessage(`Failed to create VM: ${error}`);
                }
            }
        }),

        vscode.commands.registerCommand('vmManager.startVm', async (vm) => {
            try {
                await vscode.window.withProgress({
                    location: vscode.ProgressLocation.Notification,
                    title: `Starting VM '${vm.name}'`,
                    cancellable: false
                }, async (progress, token) => {
                    progress.report({ increment: 0, message: "Initializing VM startup..." });

                    await scriptUtils.startVm(vm.name);

                    progress.report({ increment: 100, message: "VM started successfully!" });
                });

                vmTreeProvider.refresh();
                vscode.window.showInformationMessage(`VM ${vm.name} started successfully`);
            } catch (error: any) {
                if (error.message === 'VM_SCRIPT_NOT_FOUND') {
                    await handleScriptNotFound(context);
                } else {
                    vscode.window.showErrorMessage(`Failed to start VM: ${error}`);
                }
            }
        }),

        vscode.commands.registerCommand('vmManager.stopVm', async (vm) => {
            try {
                await vscode.window.withProgress({
                    location: vscode.ProgressLocation.Notification,
                    title: `Stopping VM '${vm.name}'`,
                    cancellable: false
                }, async (progress, token) => {
                    progress.report({ increment: 0, message: "Gracefully shutting down VM..." });

                    await scriptUtils.stopVm(vm.name);

                    progress.report({ increment: 100, message: "VM stopped successfully!" });
                });

                vmTreeProvider.refresh();
                vscode.window.showInformationMessage(`VM ${vm.name} stopped successfully`);
            } catch (error: any) {
                if (error.message === 'VM_SCRIPT_NOT_FOUND') {
                    await handleScriptNotFound(context);
                } else {
                    vscode.window.showErrorMessage(`Failed to stop VM: ${error}`);
                }
            }
        }),

        vscode.commands.registerCommand('vmManager.deleteVm', async (vm) => {
            const confirm = await vscode.window.showWarningMessage(
                `Are you sure you want to delete VM '${vm.name}'? This action cannot be undone.`,
                'Delete',
                'Cancel'
            );

            if (confirm === 'Delete') {
                try {
                    await vscode.window.withProgress({
                        location: vscode.ProgressLocation.Notification,
                        title: `Deleting VM '${vm.name}'`,
                        cancellable: false
                    }, async (progress, token) => {
                        progress.report({ increment: 0, message: "Stopping VM and removing files..." });

                        await scriptUtils.deleteVm(vm.name);

                        progress.report({ increment: 100, message: "VM deleted successfully!" });
                    });

                    vmTreeProvider.refresh();
                    vscode.window.showInformationMessage(`VM ${vm.name} deleted successfully`);
                } catch (error: any) {
                    if (error.message === 'VM_SCRIPT_NOT_FOUND') {
                        await handleScriptNotFound(context);
                    } else {
                        vscode.window.showErrorMessage(`Failed to delete VM: ${error}`);
                    }
                }
            }
        }),

        vscode.commands.registerCommand('vmManager.sshVm', async (vm) => {
            try {
                const terminal = vscode.window.createTerminal(`SSH: ${vm.name}`);

                // Get username from VM properties
                const username = vm.username || vm.vm?.username || 'user01';

                // Use IP address if available, otherwise fallback to hostname
                const sshTarget = vm.ipAddress && vm.ipAddress !== '-' ? vm.ipAddress : vm.hostname;

                // Get SSH port from VM properties
                const sshPort = vm.vm?.sshPort || '22';

                // Build SSH command with port if needed
                let sshCommand = `ssh ${username}@${sshTarget}`;
                if (sshPort !== '22') {
                    sshCommand += ` -p ${sshPort}`;
                }

                terminal.sendText(sshCommand);
                terminal.show();
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to SSH to VM: ${error}`);
            }
        }),

        vscode.commands.registerCommand('vmManager.getVmIp', async (vm) => {
            try {
                const ip = await scriptUtils.getVmIp(vm.name);
                if (ip) {
                    vscode.window.showInformationMessage(`VM ${vm.name} IP: ${ip}`);
                } else {
                    vscode.window.showWarningMessage(`Could not determine IP for VM ${vm.name}`);
                }
            } catch (error: any) {
                if (error.message === 'VM_SCRIPT_NOT_FOUND') {
                    await handleScriptNotFound(context);
                } else {
                    vscode.window.showErrorMessage(`Failed to get VM IP: ${error}`);
                }
            }
        }),

        vscode.commands.registerCommand('vmManager.copyDetail', async (value: string, label: string) => {
            try {
                await vscode.env.clipboard.writeText(value);
                vscode.window.showInformationMessage(`${label} copied to clipboard`);
            } catch (error) {
                vscode.window.showErrorMessage(`Failed to copy ${label.toLowerCase()}: ${error}`);
            }
        })
    ];

    context.subscriptions.push(...commands);
}

export function deactivate() {}