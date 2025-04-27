import type { PageServerLoad, Actions } from './$types';
import { getImage, removeImage } from '$lib/services/docker/image-service'; // Import getImage and removeImage
import { error, fail, redirect } from '@sveltejs/kit';
import { NotFoundError, ConflictError, DockerApiError } from '$lib/types/errors';

export const load: PageServerLoad = async ({ params }) => {
	const imageId = params.imageId; // Use imageId to match the route parameter

	try {
		const image = await getImage(imageId);

		// Optionally, you could fetch related containers here if needed
		// const containersUsingImage = await getContainersUsingImage(imageId);

		return {
			image // Return the detailed image info
		};
	} catch (err: any) {
		console.error(`Failed to load image ${imageId}:`, err);
		if (err instanceof NotFoundError) {
			// Use instanceof for custom errors
			error(404, { message: err.message });
		} else {
			error(err.status || 500, {
				message: err.message || `Failed to load image details for "${imageId}".`
			});
		}
	}
};

// Add action for removing the image from its details page
export const actions: Actions = {
	remove: async ({ params, url }) => {
		const imageId = params.imageId;
		const force = url.searchParams.get('force') === 'true'; // Check for force parameter
		try {
			await removeImage(imageId, force);
			// Redirect back to the main images list after successful removal
			redirect(303, '/images');
		} catch (err: any) {
			// Handle specific errors from removeImage
			// Note: image-service removeImage needs updating to throw custom errors
			if (err instanceof NotFoundError || err instanceof ConflictError || err instanceof DockerApiError) {
				return fail(err.status || 409, { error: err.message });
			}
			// Fallback for unexpected errors or generic errors from service
			return fail(500, { error: err.message || 'An unexpected error occurred during removal.' });
		}
	}
};
