import type { PageServerLoad } from "./$types";
import { listVolumes } from "$lib/services/docker-service";
import type { ServiceVolume } from "$lib/services/docker-service";

type VolumePageData = {
  volumes: ServiceVolume[];
  error?: string;
};

export const load: PageServerLoad = async (): Promise<VolumePageData> => {
  try {
    const volumes = await listVolumes();
    return {
      volumes,
    };
  } catch (err: any) {
    console.error("Failed to load volumes:", err);
    return {
      volumes: [],
      error: err.message || "Failed to connect to Docker or list volumes.",
    };
  }
};
