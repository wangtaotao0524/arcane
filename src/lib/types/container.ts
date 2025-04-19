export type Container = {
  name: string;
  id: string;
  image: string;
  status: "running" | "stopped" | "failed" | "healthy" | "unhealthy";
};
