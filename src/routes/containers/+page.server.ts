import type { PageServerLoad } from "./$types";
import { listContainers } from "$lib/services/docker-service"; // Import the service function

export const load: PageServerLoad = async () => {
  try {
    // Call the listContainers function from your service
    const containers = await listContainers(true); // Get all containers (running and stopped)

    // The listContainers function already maps the data, so we can return it directly
    return {
      containers,
    };
  } catch (error: any) {
    console.error("Failed to load containers:", error);
    // Return an error state or an empty array so the page can handle it
    return {
      containers: [],
      error: error.message || "Failed to connect to Docker or list containers.",
    };
  }
};
