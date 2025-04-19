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

/**
 * Lists Docker containers.
 * @param all - Whether to show all containers (including stopped). Defaults to true.
 */
export async function listContainers(all = true) {
  try {
    const containers = await defaultDocker.listContainers({ all });
    // Map to a simpler format if desired, or return raw data
    return containers.map((c) => ({
      id: c.Id,
      names: c.Names,
      name: c.Names[0]?.substring(1) || c.Id.substring(0, 12), // Extract a primary name
      image: c.Image,
      imageId: c.ImageID,
      command: c.Command,
      created: c.Created,
      state: c.State, // 'created', 'running', 'paused', 'restarting', 'removing', 'exited', 'dead'
      status: c.Status,
      ports: c.Ports,
      mounts: c.Mounts,
    }));
  } catch (error) {
    console.error("Error listing containers:", error);
    // Rethrow or return an empty array/error indicator
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

// Add more functions here as needed (e.g., startContainer, stopContainer, removeContainer, listImages, etc.)

export default defaultDocker;
