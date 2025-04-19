import type { PageServerLoad } from "./$types";
import { listImages } from "$lib/services/docker-service";
import type { ServiceImage } from "$lib/services/docker-service";

type ImageData = {
  images: ServiceImage[];
  error?: string;
};

export const load: PageServerLoad = async (): Promise<ImageData> => {
  try {
    const images = await listImages();
    return {
      images,
    };
  } catch (err: any) {
    console.error("Failed to load images:", err);
    return {
      images: [],
      error: err.message || "Failed to connect to Docker or list images.",
    };
  }
};
