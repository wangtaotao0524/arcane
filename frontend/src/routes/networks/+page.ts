import { environmentAPI } from '$lib/services/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	// try {
	// 	const networks = await environmentAPI.getNetworks();
	// 	return {
	// 		networks: networks || []
	// 	};
	// } catch (error) {
	// 	console.error('Error loading networks data:', error);
	// 	return {
	// 		networks: [],
	// 		error: error instanceof Error ? error.message : String(error)
	// 	};
	// }
};
