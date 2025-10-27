## Why
Users currently need to manually SSH into environment nodes using external tools, which disrupts their workflow within Arcane. Adding integrated SSH connection capabilities will provide a seamless experience for managing Docker environments directly from the platform interface.

## What Changes
- Add web-based SSH terminal interface for environment nodes
- Implement SSH credential management and connection configuration
- Add terminal session persistence and management
- Provide SSH connection status indicators

## Impact
- Affected specs: `ssh-connections`, `environment-management`, `user-interface`
- Affected code: New SSH service, terminal components, environment detail views
- Dependencies: WebSocket connections, xterm.js integration, SSH key management