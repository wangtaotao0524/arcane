import Docker from "dockerode";
import fs from "fs"; // Use synchronous fs for initial load
import path from "path";

// Define Settings Type (can be shared with +page.server.ts if moved to a types file)
type SettingsData = {
  dockerHost: string;
  autoRefresh: boolean;
  refreshInterval: number;
  darkMode: boolean;
};

// --- Load Initial Settings ---
const SETTINGS_FILE = path.resolve("./app-settings.json");
const DEFAULT_SETTINGS: SettingsData = {
  dockerHost: "unix:///var/run/docker.sock",
  autoRefresh: true,
  refreshInterval: 10,
  darkMode: true,
};

function loadInitialSettings(): SettingsData {
  try {
    // Use synchronous read for initial module load simplicity
    if (fs.existsSync(SETTINGS_FILE)) {
      const data = fs.readFileSync(SETTINGS_FILE, "utf-8");
      const settings = JSON.parse(data);
      console.log("Docker Service: Loaded initial settings:", settings);
      return { ...DEFAULT_SETTINGS, ...settings }; // Merge with defaults
    } else {
      console.log("Docker Service: Settings file not found, using defaults.");
      return DEFAULT_SETTINGS;
    }
  } catch (error) {
    console.error("Docker Service: Error loading initial settings:", error);
    return DEFAULT_SETTINGS; // Fallback to defaults
  }
}

const initialSettings = loadInitialSettings();
// --- End Load Initial Settings ---

/**
 * Creates Dockerode options from a host string.
 */
function getDockerOpts(host?: string): Docker.DockerOptions {
  const targetHost = host || initialSettings.dockerHost; // Use provided host or loaded setting
  const options: Docker.DockerOptions = {};

  if (targetHost.startsWith("unix://")) {
    options.socketPath = targetHost.substring(7);
  } else if (
    targetHost.startsWith("tcp://") ||
    targetHost.startsWith("http://") ||
    targetHost.startsWith("https://")
  ) {
    try {
      const url = new URL(targetHost);
      options.host = url.hostname;
      options.port = url.port || (url.protocol === "https:" ? "2376" : "2375");
      options.protocol = url.protocol.slice(0, -1) as "https" | "http";
      // TODO: Add TLS options if needed based on protocol or settings
    } catch (e) {
      console.error("Invalid Docker host URL format:", targetHost, e);
      // Return options that will likely fail connection, signaling an issue
      return { host: "invalid-host-format" };
    }
  } else {
    // Assume it's a socket path if no protocol
    options.socketPath = targetHost;
  }
  return options;
}

// --- Initialize the shared defaultDocker instance ---
// Use let so it can be reassigned by updateDefaultDockerInstance
let defaultDocker = new Docker(getDockerOpts());
console.log(
  `Docker Service: Initialized defaultDocker with options:`,
  getDockerOpts()
);
// --- End Initialization ---

/**
 * Updates the shared default Docker instance used by the service.
 * @param host - The new Docker host string.
 */
export function updateDefaultDockerInstance(host: string) {
  try {
    const newOptions = getDockerOpts(host);
    // Re-assign the module-level defaultDocker variable
    defaultDocker = new Docker(newOptions);
    // Update the settings object in memory (optional, but keeps it consistent)
    initialSettings.dockerHost = host;
    console.log(
      "Docker Service: Updated defaultDocker instance for host:",
      host
    );
    console.log("Docker Service: New options:", newOptions);
  } catch (e) {
    console.error(
      `Docker Service: Failed to update default Docker instance for host: ${host}`,
      e
    );
    // Optionally revert to a known good state or throw
  }
}

/**
 * Gets Docker system information.
 * Uses the *shared* defaultDocker instance unless a specific host is passed for testing.
 * @param host - Optional Docker host URL to connect to for *testing*. If not provided, uses the default shared instance.
 */
export async function getDockerInfo(host?: string): Promise<Docker.Info> {
  // Use a temporary instance ONLY if a specific host is provided for testing
  const dockerInstance = host ? new Docker(getDockerOpts(host)) : defaultDocker;
  const targetHost = host || initialSettings.dockerHost; // For logging

  try {
    const info = await dockerInstance.info();
    return info;
  } catch (error: any) {
    console.error(
      `Docker Service: Error getting Docker info for host "${targetHost}":`,
      error.message || error
    );
    throw new Error(
      `Failed to connect to Docker Engine at "${targetHost}". ${
        error.message || ""
      }`
    );
  }
}

// Define and export the type returned by listContainers
export type ServiceContainer = {
  id: string;
  names: string[];
  name: string; // Your derived name
  image: string;
  imageId: string;
  command: string;
  created: number;
  state: string; // 'created', 'running', 'paused', 'restarting', 'removing', 'exited', 'dead'
  status: string; // e.g., "Up 2 hours"
  ports: Docker.Port[];
};

/**
 * Lists Docker containers.
 * Uses the shared defaultDocker instance.
 */
export async function listContainers(all = true): Promise<ServiceContainer[]> {
  try {
    // Uses the potentially updated defaultDocker
    const containers = await defaultDocker.listContainers({ all });
    // Ensure the mapping matches the ServiceContainer type
    return containers.map(
      (c): ServiceContainer => ({
        id: c.Id,
        names: c.Names,
        name: c.Names[0]?.substring(1) || c.Id.substring(0, 12),
        image: c.Image,
        imageId: c.ImageID,
        command: c.Command,
        created: c.Created,
        state: c.State,
        status: c.Status,
        ports: c.Ports,
      })
    );
  } catch (error: any) {
    console.error("Docker Service: Error listing containers:", error);
    throw new Error(
      `Failed to list Docker containers using host "${initialSettings.dockerHost}".`
    );
  }
}

/**
 * Gets details for a specific container.
 * Uses the shared defaultDocker instance.
 */
export async function getContainer(containerId: string) {
  try {
    // Uses the potentially updated defaultDocker
    const container = defaultDocker.getContainer(containerId);
    const inspectData = await container.inspect();
    // Return relevant details
    return {
      id: inspectData.Id,
      name: inspectData.Name.substring(1),
      created: inspectData.Created,
      path: inspectData.Path,
      args: inspectData.Args,
      state: inspectData.State, // More detailed state object
      image: inspectData.Image,
      config: inspectData.Config,
      networkSettings: inspectData.NetworkSettings,
      mounts: inspectData.Mounts,
      // Add more fields as needed
    };
  } catch (error: any) {
    console.error(
      `Docker Service: Error getting container ${containerId}:`,
      error
    );
    if (error.statusCode === 404) {
      return null; // Container not found
    }
    throw new Error(
      `Failed to get container details for ${containerId} using host "${initialSettings.dockerHost}".`
    );
  }
}

/**
 * Starts a specific container.
 * @param containerId - The ID of the container.
 */
export async function startContainer(containerId: string): Promise<void> {
  try {
    const container = defaultDocker.getContainer(containerId);
    await container.start();
  } catch (error: any) {
    console.error(
      `Docker Service: Error starting container ${containerId}:`,
      error
    );
    throw new Error(
      `Failed to start container ${containerId} using host "${
        initialSettings.dockerHost
      }". ${error.message || ""}`
    );
  }
}

/**
 * Stops a specific container.
 * @param containerId - The ID of the container.
 */
export async function stopContainer(containerId: string): Promise<void> {
  try {
    const container = defaultDocker.getContainer(containerId);
    await container.stop();
  } catch (error: any) {
    console.error(
      `Docker Service: Error stopping container ${containerId}:`,
      error
    );
    throw new Error(
      `Failed to stop container ${containerId} using host "${
        initialSettings.dockerHost
      }". ${error.message || ""}`
    );
  }
}

/**
 * Restarts a specific container.
 * @param containerId - The ID of the container.
 */
export async function restartContainer(containerId: string): Promise<void> {
  try {
    const container = defaultDocker.getContainer(containerId);
    await container.restart();
  } catch (error: any) {
    console.error(
      `Docker Service: Error restarting container ${containerId}:`,
      error
    );
    throw new Error(
      `Failed to restart container ${containerId} using host "${
        initialSettings.dockerHost
      }". ${error.message || ""}`
    );
  }
}

/**
 * Removes a specific container.
 * @param containerId - The ID of the container.
 * @param force - Force removal even if running (default: false).
 */
export async function removeContainer(
  containerId: string,
  force = false
): Promise<void> {
  try {
    const container = defaultDocker.getContainer(containerId);
    await container.remove({ force });
  } catch (error: any) {
    console.error(
      `Docker Service: Error removing container ${containerId}:`,
      error
    );
    if (error.statusCode === 404) {
      throw new Error(`Container ${containerId} not found.`);
    }
    if (error.statusCode === 409) {
      throw new Error(
        `Cannot remove running container ${containerId}. Stop it first or use force.`
      );
    }
    throw new Error(
      `Failed to remove container ${containerId} using host "${
        initialSettings.dockerHost
      }". ${error.message || ""}`
    );
  }
}

/**
 * Gets logs for a specific container.
 * @param containerId - The ID of the container.
 * @param options - Optional parameters for log retrieval
 */
export async function getContainerLogs(
  containerId: string,
  options: {
    tail?: number | "all"; // Number of lines to show from the end of logs (use undefined for all logs)
    since?: number; // Timestamp (in seconds) to filter logs since
    until?: number; // Timestamp (in seconds) to filter logs until
    follow?: boolean; // Stream logs (not typically used in SSR)
    stdout?: boolean; // Include stdout
    stderr?: boolean; // Include stderr
  } = {}
): Promise<string> {
  try {
    const container = defaultDocker.getContainer(containerId);

    const logOptions = {
      tail: options.tail === "all" ? undefined : options.tail || 100,
      stdout: options.stdout !== false,
      stderr: options.stderr !== false,
      follow: false,
      timestamps: true,
      since: options.since || 0,
      until: options.until || undefined,
    };

    const logsBuffer = await container.logs(logOptions);
    let logString = logsBuffer.toString();

    if (logOptions.stdout || logOptions.stderr) {
      const lines = logString.split("\n");
      const processedLines = lines
        .map((line) => {
          if (!line) return "";
          if (line.length > 8) {
            return line.substring(8);
          }
          return line;
        })
        .filter(Boolean);

      logString = processedLines.join("\n");
    }

    return logString;
  } catch (error: any) {
    console.error(
      `Docker Service: Error getting logs for container ${containerId}:`,
      error
    );
    throw new Error(
      `Failed to get logs for container ${containerId} using host "${
        initialSettings.dockerHost
      }". ${error.message || ""}`
    );
  }
}

// Define and export the type returned by listImages
export type ServiceImage = {
  id: string;
  repoTags: string[] | undefined;
  repoDigests: string[] | undefined;
  created: number;
  size: number;
  virtualSize: number;
  labels: { [label: string]: string } | undefined;
  repo: string;
  tag: string;
};

/**
 * Lists Docker images.
 */
export async function listImages(): Promise<ServiceImage[]> {
  try {
    const images = await defaultDocker.listImages({ all: false });

    const parseRepoTag = (
      tag: string | undefined
    ): { repo: string; tag: string } => {
      if (!tag || tag === "<none>:<none>") {
        return { repo: "<none>", tag: "<none>" };
      }
      const parts = tag.split(":");
      if (parts.length === 1) {
        return { repo: parts[0], tag: "latest" };
      }
      const tagPart = parts.pop() || "latest";
      const repoPart = parts.join(":");
      return { repo: repoPart, tag: tagPart };
    };

    return images.map((img): ServiceImage => {
      const { repo, tag } = parseRepoTag(img.RepoTags?.[0]);
      return {
        id: img.Id,
        repoTags: img.RepoTags,
        repoDigests: img.RepoDigests,
        created: img.Created,
        size: img.Size,
        virtualSize: img.VirtualSize,
        labels: img.Labels,
        repo: repo,
        tag: tag,
      };
    });
  } catch (error: any) {
    console.error("Docker Service: Error listing images:", error);
    throw new Error(
      `Failed to list Docker images using host "${initialSettings.dockerHost}".`
    );
  }
}

// Define and export the type returned by listNetworks
export type ServiceNetwork = {
  id: string;
  name: string;
  driver: string;
  scope: string;
  subnet: string | null;
  gateway: string | null;
  created: string;
};

/**
 * Lists Docker networks.
 */
export async function listNetworks(): Promise<ServiceNetwork[]> {
  try {
    const networks = await defaultDocker.listNetworks();
    return networks.map(
      (net): ServiceNetwork => ({
        id: net.Id,
        name: net.Name,
        driver: net.Driver,
        scope: net.Scope,
        subnet: net.IPAM?.Config?.[0]?.Subnet ?? null,
        gateway: net.IPAM?.Config?.[0]?.Gateway ?? null,
        created: net.Created,
      })
    );
  } catch (error: any) {
    console.error("Docker Service: Error listing networks:", error);
    throw new Error(
      `Failed to list Docker networks using host "${initialSettings.dockerHost}".`
    );
  }
}

// Define and export the type returned by listVolumes
export type ServiceVolume = {
  name: string;
  driver: string;
  scope: string;
  mountpoint: string;
  labels: { [label: string]: string } | null;
};

/**
 * Lists Docker volumes.
 */
export async function listVolumes(): Promise<ServiceVolume[]> {
  try {
    const volumeResponse = await defaultDocker.listVolumes();
    const volumes = volumeResponse.Volumes || [];

    return volumes.map(
      (vol): ServiceVolume => ({
        name: vol.Name,
        driver: vol.Driver,
        scope: vol.Scope,
        mountpoint: vol.Mountpoint,
        labels: vol.Labels,
      })
    );
  } catch (error: any) {
    console.error("Docker Service: Error listing volumes:", error);
    throw new Error(
      `Failed to list Docker volumes using host "${initialSettings.dockerHost}".`
    );
  }
}

export default defaultDocker;
