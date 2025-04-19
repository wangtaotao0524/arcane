import { error } from "@sveltejs/kit";
import { loadComposeStacks } from "$lib/services/compose";

/** @type {import('./$types').PageServerLoad} */
export async function load() {
  try {
    const stacks = await loadComposeStacks();
    return {
      stacks,
    };
  } catch (err) {
    console.error("Failed to load stacks:", err);
    return {
      stacks: [],
      error: "Failed to load Docker Compose stacks: " + err.message,
    };
  }
}
