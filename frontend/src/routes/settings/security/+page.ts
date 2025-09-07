import { settingsAPI } from '$lib/services/api';

export const load = async () => {
	const [settings, oidcStatus] = await Promise.all([
		settingsAPI.getSettings(),
		settingsAPI.getOidcStatus().catch(() => ({
			envForced: false,
			envConfigured: false
		}))
	]);

	return {
		settings,
		oidcStatus
	};
};
