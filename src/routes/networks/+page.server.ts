import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ fetch }) => {
  // Mock data until API is ready
  const networks = [
    {
      name: "bridge",
      driver: "bridge",
      scope: "local",
      subnet: "172.17.0.0/16",
    },
    {
      name: "host",
      driver: "host",
      scope: "local",
      subnet: null,
    },
    {
      name: "web_backend",
      driver: "bridge",
      scope: "local",
      subnet: "172.18.0.0/16",
    },
  ];

  return { networks };
};
