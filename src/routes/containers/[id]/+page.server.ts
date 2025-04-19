import type { PageServerLoad } from "./$types";
import { getContainer } from "$lib/services/docker-service";
import { error } from "@sveltejs/kit";

export const load: PageServerLoad = async ({ params }) => {
  const containerId = params.id;

  try {
    const container = await getContainer(containerId);

    if (!container) {
      error(404, {
        message: `Container with ID "${containerId}" not found.`,
      });
    }

    return {
      container,
    };
  } catch (err: any) {
    console.error(`Failed to load container ${containerId}:`, err);
    error(500, {
      message:
        err.message || `Failed to load container details for "${containerId}".`,
    });
  }
};
