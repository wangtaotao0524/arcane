import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

// Here you would integrate with Docker API or SDK
export const GET: RequestHandler = async () => {
  // In production implementation:
  // const docker = new Docker();
  // const containers = await docker.listContainers({all: true});

  // For now, return mock data
  return json([
    { id: "1a2b3c", name: "web", status: "running", image: "nginx:latest" },
    { id: "4d5e6f", name: "db", status: "exited", image: "postgres:15" },
  ]);
};

export const POST: RequestHandler = async ({ request }) => {
  const data = await request.json();
  // Create a new container
  // const docker = new Docker();
  // const container = await docker.createContainer(data);

  return json({ id: "new_container_id", name: data.name });
};
