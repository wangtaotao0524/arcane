import type { PageLoad } from './$types';
import { environmentAPI } from '$lib/services/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
	const { volumeName } = params;

	try {
		const volumeBase = await environmentAPI.getVolume(volumeName);
		const volumeUsage = await environmentAPI.getVolumeUsage(volumeName);

		const volume = {
			...volumeBase,
			inUse: volumeUsage,
			containers: volumeUsage.containers
		};

		return {
			volume
		};
	} catch (err: any) {
		console.error('Failed to load volume:', err);
		if (err.status === 404) {
			throw err;
		}
		throw error(500, err.message || 'Failed to load volume details');
	}
};
