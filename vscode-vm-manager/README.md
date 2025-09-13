# VM Manager - VS Code Extension

A VS Code extension for managing Linux VMs created with the `vm` script. This extension provides a visual interface to view, create, start, stop, and manage virtual machines directly from VS Code.

## Features

- **VM Tree View**: See all your VMs in the Explorer panel with their current status
- **VM Management**: Create, start, stop, and delete VMs through the UI
- **SSH Integration**: Connect to running VMs with a single click
- **Status Indicators**: Visual icons show VM status (running/stopped)
- **Context Menus**: Right-click VMs for quick actions

## Installation

1. Make sure you have the `vm` script available in your PATH or configure the path in settings
2. Install the extension from the VSIX file or compile from source
3. Open VS Code and look for the "VM Manager" section in the Explorer panel

## Configuration

Set the path to your `vm` script in VS Code settings:

```json
{
    "vmManager.scriptPath": "./vm"
}
```

## Usage

### Viewing VMs
- VMs appear in the Explorer panel under "VM Manager"
- Green icons indicate running VMs, red icons indicate stopped VMs
- Hover over a VM to see detailed information

### Creating VMs
- Click the "+" icon in the VM Manager header
- Enter a VM name or leave empty for auto-generation

### Managing VMs
- Right-click on a VM to see available actions:
  - **Start**: Start a stopped VM
  - **Stop**: Stop a running VM  
  - **SSH**: Open terminal and connect to VM
  - **Get IP**: Show VM IP address
  - **Delete**: Remove the VM (with confirmation)

### Quick Actions
- Use the refresh button to update the VM list
- Running VMs show their hostname/port information

## Requirements

- VS Code 1.74.0 or higher
- The `vm` script must be accessible from VS Code's working directory
- Node.js for development/compilation

## Development

```bash
npm install
npm run compile
npm run watch  # for development with auto-compilation
```

## Extension Structure

- `src/extension.ts`: Main extension entry point and command registration
- `src/vmTreeProvider.ts`: Tree view provider for displaying VMs
- `src/vmScriptUtils.ts`: Utilities for executing vm script commands