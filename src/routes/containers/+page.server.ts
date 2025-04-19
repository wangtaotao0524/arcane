import type { PageServerLoad } from "./$types";

// This will eventually call your Docker API
export const load: PageServerLoad = async ({ fetch }) => {
  // In production, this would call your API endpoint
  // const response = await fetch('/api/docker/containers');
  // const containers = await response.json();

  // For now, return mock data
  return {
    containers: [
      { id: "1a2b3c", name: "web", status: "running", image: "nginx:latest" },
      { id: "4d5e6f", name: "db", status: "exited", image: "postgres:15" },
      { id: "7g8h9i", name: "cache", status: "running", image: "redis:alpine" },
    ],
  };
};
