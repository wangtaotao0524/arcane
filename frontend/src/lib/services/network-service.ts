import BaseAPIService from './api-service';
import { environmentStore } from '$lib/stores/environment.store.svelte';
import type { NetworkSummaryDto, NetworkUsageCounts, NetworkCreateRequest } from '$lib/types/network.type';
import type { SearchPaginationSortRequest, Paginated } from '$lib/types/pagination.type';
import type { NetworkCreateOptions } from 'dockerode';
import { transformPaginationParams } from '$lib/utils/params.util';

export class NetworkService extends BaseAPIService {
	async getNetworks(options?: SearchPaginationSortRequest): Promise<Paginated<NetworkSummaryDto>> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		const params = transformPaginationParams(options);
		const res = await this.api.get(`/environments/${envId}/networks`, { params });
		return res.data;
	}

	async getNetworkUsageCounts(): Promise<NetworkUsageCounts> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		const res = await this.api.get(`/environments/${envId}/networks/counts`);
		return res.data.data;
	}

	async getNetwork(networkId: string): Promise<any> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(this.api.get(`/environments/${envId}/networks/${networkId}`));
	}

	async createNetwork(options: NetworkCreateOptions): Promise<any> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		const request: NetworkCreateRequest = {
			name: options.Name || '',
			options: {
				Driver: options.Driver,
				CheckDuplicate: options.CheckDuplicate,
				Internal: options.Internal,
				Attachable: options.Attachable,
				Ingress: options.Ingress,
				IPAM: options.IPAM,
				EnableIPv6: options.EnableIPv6,
				Options: options.Options,
				Labels: options.Labels
			}
		};
		return this.handleResponse(this.api.post(`/environments/${envId}/networks`, request));
	}

	async deleteNetwork(networkId: string): Promise<any> {
		const envId = await environmentStore.getCurrentEnvironmentId();
		return this.handleResponse(this.api.delete(`/environments/${envId}/networks/${networkId}`));
	}
}

export const networkService = new NetworkService();
