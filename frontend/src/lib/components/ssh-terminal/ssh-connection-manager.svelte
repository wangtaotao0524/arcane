<script lang="ts">
	import { onMount } from 'svelte';
	import SSHTerminal from './ssh-terminal.svelte';
	import sshService, { type SSHConnection, type SSHConnectRequest } from '$lib/services/ssh-service';

	interface SSHConnectionManagerProps {
		environmentId: string;
	}

	export let environmentId: string;

	let connections: SSHConnection[] = [];
	let activeSessionId: string | null = null;
	let showConnectionForm = false;
	let connectionForm: SSHConnectRequest = {
		environmentId,
		host: '',
		port: 22,
		username: '',
		password: '',
		privateKey: ''
	};
	let error: string | null = null;
	let loading = false;

	onMount(() => {
		loadConnections();
	});

	async function loadConnections() {
		try {
			connections = await sshService.listConnections();
		} catch (err) {
			console.error('Failed to load SSH connections:', err);
		}
	}

	async function connect() {
		loading = true;
		error = null;

		try {
			const connection = await sshService.connect(connectionForm);
			connections.push(connection);
			activeSessionId = connection.id;
			showConnectionForm = false;
			resetForm();
		} catch (err: any) {
			error = err.message || 'Failed to establish SSH connection';
		} finally {
			loading = false;
		}
	}

	async function disconnect(sessionId: string) {
		try {
			await sshService.disconnect(sessionId);
			connections = connections.filter(conn => conn.id !== sessionId);
			if (activeSessionId === sessionId) {
				activeSessionId = null;
			}
		} catch (err) {
			console.error('Failed to disconnect SSH connection:', err);
		}
	}

	function resetForm() {
		connectionForm = {
			environmentId,
			host: '',
			port: 22,
			username: '',
			password: '',
			privateKey: ''
		};
	}

	function openConnectionForm() {
		showConnectionForm = true;
		resetForm();
	}

	function closeConnectionForm() {
		showConnectionForm = false;
		resetForm();
	}
</script>

<div class="ssh-connection-manager">
	<div class="manager-header">
		<h3>SSH Connections</h3>
		<button on:click={openConnectionForm} class="btn btn-primary">
			New Connection
		</button>
	</div>

	{#if showConnectionForm}
		<div class="connection-form">
			<h4>Create SSH Connection</h4>
			
			<div class="form-group">
				<label for="host">Host</label>
				<input 
					id="host"
					type="text" 
					bind:value={connectionForm.host}
					placeholder="e.g., 192.168.1.100"
				/>
			</div>

			<div class="form-group">
				<label for="port">Port</label>
				<input 
					id="port"
					type="number" 
					bind:value={connectionForm.port}
					min="1" 
					max="65535"
				/>
			</div>

			<div class="form-group">
				<label for="username">Username</label>
				<input 
					id="username"
					type="text" 
					bind:value={connectionForm.username}
					placeholder="e.g., root"
				/>
			</div>

			<div class="form-group">
				<label for="password">Password</label>
				<input 
					id="password"
					type="password" 
					bind:value={connectionForm.password}
					placeholder="Password (optional if using private key)"
				/>
			</div>

			<div class="form-group">
				<label for="privateKey">Private Key</label>
				<textarea 
					id="privateKey"
					bind:value={connectionForm.privateKey}
					placeholder="Paste private key content (optional if using password)"
					rows="4"
				></textarea>
			</div>

			{#if error}
				<div class="error-message">{error}</div>
			{/if}

			<div class="form-actions">
				<button 
					on:click={connect} 
					disabled={loading}
					class="btn btn-primary"
				>
					{loading ? 'Connecting...' : 'Connect'}
				</button>
				<button 
					on:click={closeConnectionForm} 
					class="btn btn-outline"
				>
					Cancel
				</button>
			</div>
		</div>
	{/if}

	<div class="connections-list">
		{#if connections.length === 0}
			<div class="no-connections">
				<p>No active SSH connections</p>
				<p class="text-muted">Create a new connection to start managing your environment nodes</p>
			</div>
		{:else}
				<div class="connection-tabs">
					{#each connections as connection (connection.id)}
						<div class="tab-container {activeSessionId === connection.id ? 'active' : ''}">
							<button 
								class="tab"
								on:click={() => activeSessionId = connection.id}
							>
								{connection.username}@{connection.host}:{connection.port}
								<span class="status {connection.status}"></span>
							</button>
							<button 
								on:click={() => disconnect(connection.id)}
								class="btn btn-sm btn-danger"
							>
								Disconnect
							</button>
						</div>
					{/each}
				</div>

			{#each connections as connection (connection.id)}
				{#if activeSessionId === connection.id}
					<div class="terminal-panel">
						<SSHTerminal 
							sessionId={connection.id}
							host={connection.host}
							username={connection.username}
						/>
					</div>
				{/if}
			{/each}
		{/if}
	</div>
</div>

<style>
	.ssh-connection-manager {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.manager-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px;
		background: #f8f9fa;
		border-radius: 8px;
	}

	.connection-form {
		padding: 20px;
		background: white;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
	}

	.form-group {
		margin-bottom: 16px;
	}

	.form-group label {
		display: block;
		margin-bottom: 4px;
		font-weight: 600;
		color: #374151;
	}

	.form-group input,
	.form-group textarea {
		width: 100%;
		padding: 8px 12px;
		border: 1px solid #d1d5db;
		border-radius: 4px;
		font-size: 14px;
	}

	.form-actions {
		display: flex;
		gap: 8px;
		margin-top: 20px;
	}

	.connections-list {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.connection-tabs {
		display: flex;
		gap: 4px;
		border-bottom: 1px solid #e0e0e0;
		padding: 0 16px;
	}

	.tab-container {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 0;
		border-bottom: 2px solid transparent;
		transition: all 0.2s;
	}

	.tab-container.active {
		border-bottom-color: #3b82f6;
	}

	.tab {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 8px 16px;
		background: transparent;
		border: none;
		cursor: pointer;
		font-size: 14px;
		color: inherit;
		transition: all 0.2s;
	}

	.tab-container.active .tab {
		color: #3b82f6;
	}

	.tab:hover {
		background: #f8f9fa;
	}

	.status {
		width: 8px;
		height: 8px;
		border-radius: 50%;
	}

	.status.connected {
		background: #10b981;
	}

	.status.error {
		background: #ef4444;
	}

	.terminal-panel {
		flex: 1;
		min-height: 400px;
	}

	.no-connections {
		text-align: center;
		padding: 40px 20px;
		color: #6b7280;
	}

	.text-muted {
		color: #9ca3af;
		font-size: 14px;
	}

	.error-message {
		padding: 12px;
		background: #fee2e2;
		border: 1px solid #fecaca;
		border-radius: 4px;
		color: #dc2626;
		font-size: 14px;
		margin-bottom: 16px;
	}

	.btn {
		padding: 8px 16px;
		border: 1px solid;
		border-radius: 4px;
		cursor: pointer;
		font-size: 14px;
		transition: all 0.2s;
	}

	.btn-primary {
		background: #3b82f6;
		color: white;
		border-color: #3b82f6;
	}

	.btn-primary:hover {
		background: #2563eb;
		border-color: #2563eb;
	}

	.btn-outline {
		background: transparent;
		color: #6b7280;
		border-color: #d1d5db;
	}

	.btn-outline:hover {
		background: #f8f9fa;
	}

	.btn-danger {
		background: #ef4444;
		color: white;
		border-color: #ef4444;
		padding: 4px 8px;
		font-size: 12px;
	}

	.btn-danger:hover {
		background: #dc2626;
		border-color: #dc2626;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}
</style>