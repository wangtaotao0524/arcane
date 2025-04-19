import { redirect } from "@sveltejs/kit";
import { createStack } from "$lib/services/compose";

/** @type {import('./$types').Actions} */
export const actions = {
  default: async ({ request }) => {
    const formData = await request.formData();

    const name = formData.get("name")?.toString() || "";
    const composeContent = formData.get("composeContent")?.toString() || "";

    if (!name) {
      return {
        success: false,
        error: "Stack name is required",
      };
    }

    if (!composeContent) {
      return {
        success: false,
        error: "Compose file content is required",
      };
    }

    try {
      const stack = await createStack(name, composeContent);
      throw redirect(303, `/stacks/${stack.id}`);
    } catch (err) {
      if (err instanceof Response) throw err; // Rethrow the redirect

      console.error("Error creating stack:", err);
      return {
        success: false,
        error: err instanceof Error ? err.message : "Failed to create stack",
      };
    }
  },
};
