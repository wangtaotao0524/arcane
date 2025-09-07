import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { settingsAPI } from '$lib/services/api';

export const load: PageLoad = async () => {
	const [settings, oidcStatus] = await Promise.all([
		settingsAPI.getSettings(),
		settingsAPI.getOidcStatus().catch(() => ({
			envForced: false,
			envConfigured: false
		}))
	]);

	if (settings.onboardingCompleted) {
		throw redirect(302, '/dashboard');
	}

	return { settings, oidcStatus };
};
