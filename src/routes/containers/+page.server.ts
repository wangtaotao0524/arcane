import type { PageServerLoad } from "./$types";
import type { Container } from "$lib/types/container";

// This will eventually call your Docker API
export const load: PageServerLoad = async ({ fetch }) => {
  // In production, this would call your API endpoint
  // const response = await fetch('/api/docker/containers');
  // const containers = await response.json();

  const containers: Container[] = [
    {
      id: "1a2b3c-11111",
      name: "1a2b3c",
      status: "running",
      image: "nginx:latest",
    },
    {
      id: "1a2b3c-22222",
      name: "1a2b3c",
      status: "running",
      image: "nginx:latest",
    },
    {
      id: "1a2b3c-3333",
      name: "1a2b3c",
      status: "running",
      image: "nginx:latest",
    },
  ];

  return {
    containers,
  };
};
