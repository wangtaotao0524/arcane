import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ fetch }) => {
  // Mock data until API is ready
  const volumes = [
    {
      name: "postgres_data",
      driver: "local",
      mountpoint: "/var/lib/docker/volumes/postgres_data/_data",
    },
    {
      name: "nginx_config",
      driver: "local",
      mountpoint: "/var/lib/docker/volumes/nginx_config/_data",
    },
    {
      name: "redis_data",
      driver: "local",
      mountpoint: "/var/lib/docker/volumes/redis_data/_data",
    },
  ];

  return { volumes };
};
