import type { PageServerLoad, Actions } from "./$types";
import { fail } from "@sveltejs/kit";

// Define a type for settings (optional but good practice)
type SettingsData = {
  dockerHost: string;
  autoRefresh: boolean;
  refreshInterval: number;
  darkMode: boolean;
};

export const load: PageServerLoad = async () => {
  // TODO: Load actual saved settings here instead of defaults
  const currentSettings: SettingsData = {
    dockerHost: "unix:///var/run/docker.sock",
    autoRefresh: true,
    refreshInterval: 10,
    darkMode: true,
  };
  // Return the raw settings data
  return { settings: currentSettings };
};

export const actions: Actions = {
  default: async ({ request }) => {
    const formData = await request.formData();

    // Manually get and parse form data
    const dockerHost = formData.get("dockerHost") as string;
    const autoRefresh = formData.get("autoRefresh") === "on"; // Checkbox/Switch value is 'on' or null
    const refreshIntervalStr = formData.get("refreshInterval") as string;
    const darkMode = formData.get("darkMode") === "on";

    // Basic validation (optional)
    if (!dockerHost) {
      return fail(400, {
        error: "Docker host cannot be empty.",
        values: Object.fromEntries(formData),
      });
    }

    let refreshInterval = parseInt(refreshIntervalStr, 10);
    if (isNaN(refreshInterval) || refreshInterval < 5 || refreshInterval > 60) {
      refreshInterval = 10; // Default or handle error
      // Optionally return fail:
      // return fail(400, { error: "Invalid refresh interval.", values: Object.fromEntries(formData) });
    }

    const updatedSettings: SettingsData = {
      dockerHost,
      autoRefresh,
      refreshInterval,
      darkMode,
    };

    // TODO: Persist the updated settings
    console.log("Saving settings:", updatedSettings);

    // You can return the updated settings or a success message
    // Returning the settings allows the page to update if needed
    return { success: true, settings: updatedSettings };
  },
};
