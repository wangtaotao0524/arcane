import BaseAPIService from './api-service';
import type { AutoUpdateCheck, AutoUpdateResult, AutoUpdateRecord, AutoUpdateStatus } from '$lib/types/auto-update.type';

export default class AutoUpdateAPIService extends BaseAPIService {
	async checkForUpdates(options?: AutoUpdateCheck): Promise<AutoUpdateResult> {
		const response = await this.api.post('/updater/run', options || {});
		return response.data.data;
	}

	async getStatus(): Promise<AutoUpdateStatus> {
		const response = await this.api.get('/updater/status');
		return response.data.data;
	}

	async getUpdateHistory(limit?: number): Promise<AutoUpdateRecord[]> {
		const params = limit ? { limit } : {};
		const response = await this.api.get('/updater/history', { params });
		return response.data.data;
	}
}
