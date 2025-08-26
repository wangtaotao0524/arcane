import axios, { type AxiosResponse } from 'axios';

abstract class BaseAPIService {
	api = axios.create({
		baseURL: '/api',
		withCredentials: true
	});

	constructor() {
		if (typeof process !== 'undefined' && process?.env?.DEV_BACKEND_URL) {
			this.api.defaults.baseURL = process.env.DEV_BACKEND_URL;
		}

		this.api.interceptors.response.use(
			(response) => response,
			(error) => {
				const status = error?.response?.status;
				if (status === 401 && typeof window !== 'undefined') {
					const reqUrl: string = error?.config?.url ?? '';
					// Skip auth endpoints and public endpoints to avoid loops
					const isAuthApi =
						reqUrl.startsWith('/auth') ||
						reqUrl.startsWith('/oidc') ||
						reqUrl.startsWith('/settings/public');
					const pathname = window.location.pathname || '/';
					const isOnAuthPage = pathname.startsWith('/auth');

					if (!isAuthApi && !isOnAuthPage) {
						const redirectTo = encodeURIComponent(pathname);
						// Hard replace to avoid history pollution
						window.location.replace(`/auth/login?redirect=${redirectTo}`);
						// Stop further promise chain
						return new Promise(() => {});
					}
				}
				return Promise.reject(error);
			}
		);
	}

	protected async handleResponse<T>(promise: Promise<AxiosResponse>): Promise<T> {
		const response = await promise;
		const payload = response.data;
		const extracted =
			payload &&
			typeof payload === 'object' &&
			'data' in payload &&
			(payload as any).data !== undefined
				? (payload as any).data
				: payload;
		return extracted as T;
	}
}

export default BaseAPIService;
