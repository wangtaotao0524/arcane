import { environmentAPI } from '$lib/services/api';
import type { SearchPaginationSortRequest } from '$lib/types/pagination.type';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	const projectRequestOptions: SearchPaginationSortRequest = {
		pagination: {
			page: 1,
			limit: 10
		},
		sort: {
			column: 'name',
			direction: 'asc' as const
		}
	};

	const projects = await environmentAPI.getProjects(projectRequestOptions);

	return { projects, projectRequestOptions };
};
