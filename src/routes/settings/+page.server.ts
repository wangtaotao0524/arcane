import type { PageServerLoad, Actions } from "./$types";
import { fail } from "@sveltejs/kit";
import { updateDockerConnection } from "$lib/services/docker-service";

// Define a type for settings
type SettingsData = {
  dockerHost: string;
  autoRefresh: boolean;
  refreshInterval: number;
  darkMode: boolean;
};

// --- Persistence Layer (Example using simple file - replace with your actual storage) ---
import fs from "fs/promises";
import path from "path";
const SETTINGS_FILE = path.resolve("./app-settings.json"); // Store settings in project root (adjust path as needed)

async function loadSettingsFromFile(): Promise<SettingsData> {
  const defaults: SettingsData = {
    dockerHost: "unix:///var/run/docker.sock", // Default Docker host
    autoRefresh: true,
    refreshInterval: 10,
    darkMode: true,
  };
  try {
    const data = await fs.readFile(SETTINGS_FILE, "utf-8");
    return JSON.parse(data);
  } catch (error: any) {
    if (error.code === "ENOENT") {
      console.log("Settings file not found, using defaults.");
      // Optionally save defaults if file doesn't exist
      // await saveSettingsToFile(defaults);
      return defaults;
    }
    console.error("Error loading settings:", error);
    return defaults; // Fallback to defaults on other errors
  }
}

async function saveSettingsToFile(settings: SettingsData): Promise<void> {
  try {
    await fs.writeFile(
      SETTINGS_FILE,
      JSON.stringify(settings, null, 2),
      "utf-8"
    );
    console.log("Settings saved to:", SETTINGS_FILE);
  } catch (error) {
    console.error("Error saving settings:", error);
    throw new Error("Failed to save settings."); // Propagate error
  }
}
// --- End Persistence Layer ---

export const load: PageServerLoad = async () => {
  // *** Load actual saved settings here ***
  const currentSettings = await loadSettingsFromFile();
  console.log("Loaded settings:", currentSettings);
  return { settings: currentSettings };
};

export const actions: Actions = {
  default: async ({ request }) => {
    const formData = await request.formData();

    const dockerHost = formData.get("dockerHost") as string;
    const autoRefresh = formData.get("autoRefresh") === "on";
    const refreshIntervalStr = formData.get("refreshInterval") as string;
    const darkMode = formData.get("darkMode") === "on";

    if (!dockerHost) {
      return fail(400, {
        error: "Docker host cannot be empty.",
        values: Object.fromEntries(formData), // Return submitted values on error
      });
    }

    let refreshInterval = parseInt(refreshIntervalStr, 10);
    if (isNaN(refreshInterval) || refreshInterval < 5 || refreshInterval > 60) {
      // Return fail to show validation error in UI
      return fail(400, {
        error: "Refresh interval must be between 5 and 60 seconds.",
        values: Object.fromEntries(formData),
      });
    }

    const updatedSettings: SettingsData = {
      dockerHost,
      autoRefresh,
      refreshInterval,
      darkMode,
    };

    try {
      // *** Persist the updated settings ***
      await saveSettingsToFile(updatedSettings);

      // *** Optional: Update the running server's Docker instance immediately ***
      updateDockerConnection(updatedSettings.dockerHost);

      // Return success and the *updated* settings
      // This updates the `form.settings` prop in the page component
      return { success: true, settings: updatedSettings };
    } catch (error: any) {
      return fail(500, {
        error: error.message || "Failed to save settings.",
        values: Object.fromEntries(formData),
      });
    }
  },
};
