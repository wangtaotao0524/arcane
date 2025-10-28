<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { Terminal } from '@xterm/xterm';
	import { FitAddon } from '@xterm/addon-fit';
	import '@xterm/xterm/css/xterm.css';

	interface SSHTerminalProps {
		sessionId: string;
		host: string;
		username: string;
		rows?: number;
		cols?: number;
	}

	export let sessionId: string;
	export let host: string;
	export let username: string;
	export let rows = 30;
	export let cols = 120;

	let terminalContainer: HTMLDivElement;
	let terminal: Terminal;
	let fitAddon: FitAddon;
	let ws: WebSocket | null = null;

	onMount(() => {
		// Initialize terminal
		terminal = new Terminal({
			rows,
			cols,
			fontSize: 14,
			fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
			cursorBlink: true,
			theme: {
				background: '#1a1a1a',
				foreground: '#f0f0f0',
				cursor: '#ffffff',
				selection: '#404040'
			}
		});

		// Add fit addon
		fitAddon = new FitAddon();
		terminal.loadAddon(fitAddon);

		// Mount terminal
		terminal.open(terminalContainer);
		fitAddon.fit();

		// Connect to SSH terminal WebSocket
		connectWebSocket();

		// Handle window resize
		const resizeObserver = new ResizeObserver(() => {
			fitAddon.fit();
		});
		resizeObserver.observe(terminalContainer);

		return () => {
			resizeObserver.disconnect();
			disconnect();
		};
	});

	function connectWebSocket() {
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const hostname = window.location.hostname;
		const port = window.location.port ? `:${window.location.port}` : '';
		
		const wsUrl = `${protocol}//${hostname}${port}/api/ssh/terminal/${sessionId}`;
		
		ws = new WebSocket(wsUrl);

		ws.onopen = () => {
			terminal.write(`\x1b[1;32mConnected to ${username}@${host}\x1b[0m\r\n`);
			terminal.write(`\x1b[1;37mPress Ctrl+D to disconnect\x1b[0m\r\n\r\n`);
		};

		ws.onmessage = (event) => {
			terminal.write(event.data);
		};

		ws.onclose = () => {
			terminal.write(`\r\n\x1b[1;31mConnection closed\x1b[0m\r\n`);
		};

		ws.onerror = (error) => {
			terminal.write(`\r\n\x1b[1;31mConnection error: ${error}\x1b[0m\r\n`);
		};

		// Handle terminal input
		terminal.onData((data) => {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(data);
			}
		});

		// Handle terminal resize
		terminal.onResize(({ rows, cols }) => {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(JSON.stringify({ type: 'resize', rows, cols }));
			}
		});

		// Handle special keys
		terminal.attachCustomKeyEventHandler((event) => {
			if (event.ctrlKey && event.key === 'd') {
				disconnect();
				event.preventDefault();
				return false;
			}
			return true;
		});
	}

	function disconnect() {
		if (ws) {
			ws.close();
			ws = null;
		}
		if (terminal) {
			terminal.dispose();
		}
	}

	onDestroy(() => {
		disconnect();
	});

	// Export functions for parent component
	export function fitTerminal() {
		fitAddon?.fit();
	}

	export function clearTerminal() {
		terminal?.clear();
	}

	export function writeToTerminal(data: string) {
		terminal?.write(data);
	}
</script>

<div class="ssh-terminal">
	<div class="terminal-header">
		<span class="connection-info">
			{username}@{host} - Session: {sessionId.slice(0, 8)}
		</span>
		<div class="terminal-controls">
			<button on:click={clearTerminal} class="btn btn-sm btn-outline">Clear</button>
			<button on:click={fitTerminal} class="btn btn-sm btn-outline">Fit</button>
			<button on:click={disconnect} class="btn btn-sm btn-danger">Disconnect</button>
		</div>
	</div>
	<div class="terminal-container" bind:this={terminalContainer}></div>
</div>

<style>
	.ssh-terminal {
		display: flex;
		flex-direction: column;
		height: 100%;
		background: #1a1a1a;
		border-radius: 8px;
		overflow: hidden;
	}

	.terminal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 12px 16px;
		background: #2a2a2a;
		border-bottom: 1px solid #404040;
		color: #f0f0f0;
		font-family: 'SF Mono', Monaco, Menlo, monospace;
		font-size: 12px;
	}

	.connection-info {
		font-weight: 600;
	}

	.terminal-controls {
		display: flex;
		gap: 8px;
	}

	.terminal-container {
		flex: 1;
		padding: 16px;
		min-height: 0;
	}

	.btn {
		padding: 4px 8px;
		border: 1px solid #404040;
		border-radius: 4px;
		background: transparent;
		color: #f0f0f0;
		cursor: pointer;
		font-size: 12px;
		transition: all 0.2s;
	}

	.btn:hover {
		background: #404040;
	}

	.btn-danger {
		border-color: #dc2626;
		color: #dc2626;
	}

	.btn-danger:hover {
		background: #dc2626;
		color: white;
	}
</style>