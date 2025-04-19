import { error, redirect } from "@sveltejs/kit";
import {
  getStack,
  updateStack,
  startStack,
  stopStack,
  restartStack,
  removeStack,
} from "$lib/services/compose";

/** @type {import('./$types').PageServerLoad} */
export async function load({ params }) {
  const { stackId } = params;

  try {
    const stack = await getStack(stackId);
    return { stack };
  } catch (err) {
    throw error(404, `Stack not found: ${err.message}`);
  }
}

/** @type {import('./$types').Actions} */
export const actions = {
  update: async ({ params, request }) => {
    const { stackId } = params;
    const formData = await request.formData();

    const name = formData.get("name")?.toString() || "";
    const composeContent = formData.get("composeContent")?.toString() || "";

    try {
      await updateStack(stackId, { name, composeContent });
      return {
        success: true,
        message: "Stack updated successfully",
      };
    } catch (err) {
      console.error("Error updating stack:", err);
      return {
        success: false,
        error: err.message || "Failed to update stack",
      };
    }
  },

  start: async ({ params }) => {
    const { stackId } = params;

    try {
      await startStack(stackId);
      return {
        success: true,
        message: "Stack started successfully",
      };
    } catch (err) {
      console.error("Error starting stack:", err);
      return {
        success: false,
        error: err.message || "Failed to start stack",
      };
    }
  },

  stop: async ({ params }) => {
    const { stackId } = params;

    try {
      await stopStack(stackId);
      return {
        success: true,
        message: "Stack stopped successfully",
      };
    } catch (err) {
      console.error("Error stopping stack:", err);
      return {
        success: false,
        error: err.message || "Failed to stop stack",
      };
    }
  },

  restart: async ({ params }) => {
    const { stackId } = params;

    try {
      await restartStack(stackId);
      return {
        success: true,
        message: "Stack restarted successfully",
      };
    } catch (err) {
      console.error("Error restarting stack:", err);
      return {
        success: false,
        error: err.message || "Failed to restart stack",
      };
    }
  },

  remove: async ({ params }) => {
    const { stackId } = params;

    try {
      await removeStack(stackId);
      throw redirect(303, "/stacks");
    } catch (err) {
      if (err instanceof Response) throw err; // Rethrow the redirect

      console.error("Error removing stack:", err);
      return {
        success: false,
        error: err.message || "Failed to remove stack",
      };
    }
  },
};
