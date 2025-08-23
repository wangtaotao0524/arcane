import { environmentAPI, systemAPI } from '$lib/services/api';
import type { SearchPaginationSortRequest } from '$lib/types/pagination.type';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const imageRequestOptions: SearchPaginationSortRequest = {
		pagination: {
			page: 1,
			limit: 20
		},
		sort: {
			column: 'created',
			direction: 'desc' as const
		}
	};

	const totalSize = await environmentAPI.getTotalImageSize();
	const images = await environmentAPI.getImages(imageRequestOptions);

	return { totalSize, images, imageRequestOptions };
};
