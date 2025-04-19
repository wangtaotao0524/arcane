import type { PageServerLoad } from "./$types";

// This will be replaced with actual Docker API calls later
export const load: PageServerLoad = async ({ fetch }) => {
  // Mock data until API is ready
  const images = [
    {
      id: "sha256:abc123",
      repo: "nginx",
      tag: "latest",
      size: "133MB",
      created: "2 days ago",
    },
    {
      id: "sha256:def456",
      repo: "postgres",
      tag: "15",
      size: "350MB",
      created: "1 week ago",
    },
    {
      id: "sha256:ghi789",
      repo: "redis",
      tag: "alpine",
      size: "28MB",
      created: "3 days ago",
    },
  ];

  return { images };
};
