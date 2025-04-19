import Docker from "dockerode";

// This instance can be used for general operations once settings are loaded/applied
// For now, it might default to the standard connection
// TODO: Initialize this based on saved settings later
const defaultDocker = new Docker();

/**
 * Creates a Dockerode instance options object from a host string.
 * Handles unix sockets and TCP addresses.
 */
function getDockerOpts(host?: string): Docker.DockerOptions | undefined {
  if (!host) {
    return undefined; // Use default connection methods
  }
  if (host.startsWith("unix://")) {
    return { socketPath: host.substring(7) };
  }
  if (host.startsWith("tcp://")) {
    try {
      const url = new URL(host);
      return {
        host: url.hostname,
        port: url.port || 2375, // Default Docker TCP port
        // TODO: Add support for TLS options if needed (protocol: 'https')
      };
    } catch (e) {
      console.error("Invalid Docker TCP host URL:", host);
      return { host: "invalid-host" }; // Force failure
    }
  }
  // Assume it's a socket path if no protocol is specified
  return { socketPath: host };
}

/**
 * Gets basic Docker system information, optionally testing a specific host.
 * @param hostToTest - The Docker host string (e.g., "unix:///var/run/docker.sock" or "tcp://localhost:2375") to test. If undefined, uses the default connection.
 */
export async function getDockerInfo(hostToTest?: string) {
  try {
    // Create a specific instance for testing if hostToTest is provided
    const dockerInstance = hostToTest
      ? new Docker(getDockerOpts(hostToTest))
      : defaultDocker;
    return await dockerInstance.info();
  } catch (error: any) {
    console.error(
      `Error getting Docker info (Host: ${hostToTest || "default"}):`,
      error.message
    );
    // Make error more specific
    let message = `Failed to connect to Docker Engine`;
    if (hostToTest) {
      message += ` at ${hostToTest}`;
    }
    if (error.code === "ENOENT" || error.message.includes("ENOENT")) {
      message += `. Socket/Path not found.`;
    } else if (
      error.code === "ECONNREFUSED" ||
      error.message.includes("ECONNREFUSED")
    ) {
      message += `. Connection refused. Is Docker running?`;
    } else if (error.message) {
      message += `. Error: ${error.message}`;
    }
    throw new Error(message);
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
 * @param all - Whether to show all containers (including stopped). Defaults to true.
 */
// Add the return type annotation
export async function listContainers(all = true): Promise<ServiceContainer[]> {
  try {
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
    console.error("Error listing containers:", error);
    // Rethrow or handle as appropriate for your app's error strategy
    throw new Error("Failed to list Docker containers.");
  }
}

/**
 * Gets details for a specific container.
 * @param containerId - The ID of the container.
 */
export async function getContainer(containerId: string) {
  try {
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
    console.error(`Error getting container ${containerId}:`, error);
    if (error.statusCode === 404) {
      return null; // Container not found
    }
    throw new Error(`Failed to get container details for ${containerId}.`);
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
    console.error(`Error starting container ${containerId}:`, error);
    // Provide more context if possible (e.g., container already started)
    throw new Error(
      `Failed to start container ${containerId}. ${error.message || ""}`
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
    console.error(`Error stopping container ${containerId}:`, error);
    // Provide more context (e.g., container not running)
    throw new Error(
      `Failed to stop container ${containerId}. ${error.message || ""}`
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
    console.error(`Error restarting container ${containerId}:`, error);
    throw new Error(
      `Failed to restart container ${containerId}. ${error.message || ""}`
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
    console.error(`Error removing container ${containerId}:`, error);
    // Provide more context (e.g., conflict, container not found)
    if (error.statusCode === 404) {
      throw new Error(`Container ${containerId} not found.`);
    }
    if (error.statusCode === 409) {
      throw new Error(
        `Cannot remove running container ${containerId}. Stop it first or use force.`
      );
    }
    throw new Error(
      `Failed to remove container ${containerId}. ${error.message || ""}`
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

    // Set default options if not provided
    const logOptions = {
      tail: options.tail === "all" ? undefined : options.tail || 100, // Default to last 100 lines, undefined for 'all'
      stdout: options.stdout !== false, // Include stdout by default
      stderr: options.stderr !== false, // Include stderr by default
      follow: false, // Don't stream by default (won't work well with SSR)
      timestamps: true, // Include timestamps
      since: options.since || 0, // Get all logs by default
      until: options.until || undefined,
    };

    // Get the logs as a Buffer
    const logsBuffer = await container.logs(logOptions);

    // Convert the Buffer to a string and handle the format
    let logString = logsBuffer.toString();

    // Process the log string to handle Docker log format
    // Docker prefixes each line with 8 bytes: first byte is 01 (stdout) or 02 (stderr),
    // followed by 3 bytes of 0s, then 4 bytes for the size
    if (logOptions.stdout || logOptions.stderr) {
      // This is a simple approach - split by lines and remove Docker headers
      const lines = logString.split("\n");
      const processedLines = lines
        .map((line) => {
          // Skip empty lines
          if (!line) return "";
          // Remove the 8-byte header if present (check line length)
          if (line.length > 8) {
            return line.substring(8);
          }
          return line;
        })
        .filter(Boolean); // Remove empty lines

      logString = processedLines.join("\n");
    }

    return logString;
  } catch (error: any) {
    console.error(`Error getting logs for container ${containerId}:`, error);
    throw new Error(
      `Failed to get logs for container ${containerId}. ${error.message || ""}`
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
  // Add primary repo and tag for easier display
  repo: string;
  tag: string;
};

/**
 * Lists Docker images.
 */
// Add the return type annotation
export async function listImages(): Promise<ServiceImage[]> {
  try {
    const images = await defaultDocker.listImages({ all: false }); // Usually only show non-intermediate images

    // Function to parse repo and tag
    const parseRepoTag = (
      tag: string | undefined
    ): { repo: string; tag: string } => {
      if (!tag || tag === "<none>:<none>") {
        return { repo: "<none>", tag: "<none>" };
      }
      const parts = tag.split(":");
      if (parts.length === 1) {
        return { repo: parts[0], tag: "latest" }; // Assume latest if no tag
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
    console.error("Error listing images:", error);
    throw new Error("Failed to list Docker images.");
  }
}

// Define and export the type returned by listNetworks
export type ServiceNetwork = {
  id: string;
  name: string;
  driver: string;
  scope: string;
  subnet: string | null; // Extract the first subnet if available
  gateway: string | null; // Extract the first gateway if available
  created: string; // Dockerode returns date as string
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
        // Safely access the first IPAM config and its subnet/gateway
        subnet: net.IPAM?.Config?.[0]?.Subnet ?? null,
        gateway: net.IPAM?.Config?.[0]?.Gateway ?? null,
        created: net.Created, // Keep as string or parse if needed
      })
    );
  } catch (error: any) {
    console.error("Error listing networks:", error);
    throw new Error("Failed to list Docker networks.");
  }
}

// Define and export the type returned by listVolumes
export type ServiceVolume = {
  name: string;
  driver: string;
  scope: string; // Usually 'local' or 'global'
  mountpoint: string;
  labels: { [label: string]: string } | null;
};

/**
 * Lists Docker volumes.
 */
export async function listVolumes(): Promise<ServiceVolume[]> {
  try {
    // The listVolumes response structure is slightly different
    const volumeResponse = await defaultDocker.listVolumes();
    const volumes = volumeResponse.Volumes || []; // Access the Volumes array

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
    console.error("Error listing volumes:", error);
    throw new Error("Failed to list Docker volumes.");
  }
}

export default defaultDocker;
