import BaseAPIService from './api-service';

export interface SSHConnection {
	id: string;
	environmentId: string;
	host: string;
	port: number;
	username: string;
	status: 'connected' | 'disconnected' | 'error';
	createdAt: string;
}

export interface SSHConnectRequest {
	environmentId: string;
	host: string;
	port: number;
	username: string;
	password?: string;
	privateKey?: string;
}

class SSHService extends BaseAPIService {
	async connect(request: SSHConnectRequest): Promise<SSHConnection> {
		return this.handleResponse(
			this.api.post('/ssh/connect', request)
		);
	}

	async listConnections(): Promise<SSHConnection[]> {
		return this.handleResponse(
			this.api.get('/ssh/connections')
		);
	}

	async getConnectionStatus(sessionId: string): Promise<SSHConnection> {
		return this.handleResponse(
			this.api.get(`/ssh/connections/${sessionId}/status`)
		);
	}

	async disconnect(sessionId: string): Promise<void> {
		return this.handleResponse(
			this.api.delete(`/ssh/connections/${sessionId}`)
		);
	}
}

export default new SSHService();