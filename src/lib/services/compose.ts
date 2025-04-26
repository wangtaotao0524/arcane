import { promises as fs } from 'fs';
import { join } from 'path';
import DockerodeCompose from 'dockerode-compose';
import { getDockerClient } from './docker-service';
import { nanoid } from 'nanoid';
import type { Stack, StackMeta, StackService, StackUpdate } from '$lib/types/stack';
import yaml from 'js-yaml';
import { getSettings, ensureStacksDirectory } from './settings-service';

let STACKS_DIR = '';

/**
 * Initialize the compose service with settings
 * This should be called at app startup
 */
export async function initComposeService(): Promise<void> {
	try {
		const settings = await getSettings();
		STACKS_DIR = settings.stacksDirectory;
		console.log(`Stacks directory initialized: ${STACKS_DIR}`);

		await ensureStacksDir();
	} catch (err) {
		console.error('Error initializing compose service:', err);
	}
}

/**
 * Update the stacks directory path - used when settings are updated
 * @param {string} directory New stacks directory path
 */
export function updateStacksDirectory(directory: string): void {
	if (directory) {
		STACKS_DIR = directory;
		console.log(`Stacks directory updated to: ${STACKS_DIR}`);
	}
}

/**
 * Ensure stacks directory exists and return the path
 */
async function ensureStacksDir(): Promise<string> {
	try {
		if (!STACKS_DIR) {
			STACKS_DIR = await ensureStacksDirectory();
		} else {
			await fs.mkdir(STACKS_DIR, { recursive: true });
		}
		return STACKS_DIR;
	} catch (err) {
		console.error('Error creating stacks directory:', err);
		throw new Error('Failed to create stacks storage directory');
	}
}

/**
 * Get stack directory path
 * @param {string} stackId
 */
async function getStackDir(stackId: string): Promise<string> {
	const stacksDir = await ensureStacksDir();
	return join(stacksDir, stackId);
}

/**
 * Get compose file path
 * @param {string} stackId
 */
async function getComposeFilePath(stackId: string): Promise<string> {
	const stackDir = await getStackDir(stackId);
	return join(stackDir, 'docker-compose.yml');
}

/**
 * Get stack metadata file path
 * @param {string} stackId
 */
async function getStackMetaPath(stackId: string): Promise<string> {
	const stackDir = await getStackDir(stackId);
	return join(stackDir, 'meta.json');
}

/**
 * Initialize a DockerodeCompose instance for a stack
 * @param {string} stackId
 */
async function getComposeInstance(stackId: string): Promise<DockerodeCompose> {
	const docker = getDockerClient();
	const composePath = await getComposeFilePath(stackId);
	return new DockerodeCompose(docker, composePath, stackId);
}

/**
 * Get the services and their status for a specific stack
 * @param {string} stackId Stack ID
 * @param {string} composeContent Compose file content
 */
async function getStackServices(stackId: string, composeContent: string): Promise<StackService[]> {
	const docker = getDockerClient();

	try {
		const composeData = yaml.load(composeContent) as any;
		if (!composeData || !composeData.services) {
			return [];
		}

		const serviceNames = Object.keys(composeData.services);

		const containers = await docker.listContainers({ all: true });

		const stackPrefix = `${stackId}_`;
		const stackContainers = containers.filter((container) => {
			const names = container.Names || [];
			return names.some((name) => name.startsWith(`/${stackPrefix}`));
		});

		const services: StackService[] = [];

		for (const containerData of stackContainers) {
			let containerName = containerData.Names?.[0] || '';
			containerName = containerName.substring(1);

			let serviceName = '';
			for (const name of serviceNames) {
				if (containerName.startsWith(`${stackId}_${name}_`) || containerName === `${stackId}_${name}`) {
					serviceName = name;
					break;
				}
			}

			if (!serviceName) {
				serviceName = containerName;
			}

			const service: StackService = {
				id: containerData.Id,
				name: serviceName,
				state: {
					Running: containerData.State === 'running',
					Status: containerData.State,
					ExitCode: 0 
				}
			};

			services.push(service);
		}

		for (const name of serviceNames) {
			if (!services.some((s) => s.name === name)) {
				services.push({
					id: '',
					name: name,
					state: {
						Running: false,
						Status: 'not created',
						ExitCode: 0
					}
				});
			}
		}

		return services;
	} catch (err) {
		console.error(`Error getting services for stack ${stackId}:`, err);
		return [];
	}
}

/**
 * Load all compose stacks
 * @returns {Promise<Array<Stack>>} List of stacks
 */
export async function loadComposeStacks(): Promise<Stack[]> {
	const stacksDir = await ensureStacksDir();

	try {
		const stackDirs = await fs.readdir(stacksDir);
		const stacks: Stack[] = [];

		for (const dir of stackDirs) {
			try {
				const metaPath = await getStackMetaPath(dir);
				const composePath = await getComposeFilePath(dir);

				const [metaContent, composeContent] = await Promise.all([fs.readFile(metaPath, 'utf8'), fs.readFile(composePath, 'utf8')]);

				const meta = JSON.parse(metaContent) as StackMeta;

				const services = await getStackServices(dir, composeContent);

				const serviceCount = services.length;
				const runningCount = services.filter((s) => s.state?.Running).length;

				let status: Stack['status'] = 'stopped';
				if (runningCount === serviceCount && serviceCount > 0) {
					status = 'running';
				} else if (runningCount > 0) {
					status = 'partially running';
				}

				stacks.push({
					id: dir,
					name: meta.name,
					serviceCount,
					runningCount,
					status,
					createdAt: meta.createdAt,
					updatedAt: meta.updatedAt
				});
			} catch (err) {
				console.warn(`Error loading stack ${dir}:`, err);
			}
		}

		return stacks;
	} catch (err) {
		console.error('Error loading stacks:', err);
		throw new Error('Failed to load compose stacks');
	}
}

/**
 * Get stack by ID
 * @param {string} stackId
 */
export async function getStack(stackId: string): Promise<Stack> {
	try {
		const metaPath = await getStackMetaPath(stackId);
		const composePath = await getComposeFilePath(stackId);

		const [metaContent, composeContent] = await Promise.all([fs.readFile(metaPath, 'utf8'), fs.readFile(composePath, 'utf8')]);

		const meta = JSON.parse(metaContent) as StackMeta;

		const services = await getStackServices(stackId, composeContent);

		const serviceCount = services.length;
		const runningCount = services.filter((s) => s.state?.Running).length;

		let status: Stack['status'] = 'stopped';
		if (runningCount === serviceCount && serviceCount > 0) {
			status = 'running';
		} else if (runningCount > 0) {
			status = 'partially running';
		}

		return {
			id: stackId,
			name: meta.name,
			services,
			serviceCount,
			runningCount,
			status,
			createdAt: meta.createdAt,
			updatedAt: meta.updatedAt,
			composeContent
		};
	} catch (err) {
		console.error(`Error getting stack ${stackId}:`, err);
		throw new Error(`Stack not found or cannot be accessed`);
	}
}

/**
 * Create a new stack
 * @param {string} name Stack name
 * @param {string} composeContent Compose file content
 */
export async function createStack(name: string, composeContent: string): Promise<Stack> {
	const stackId = nanoid();
	const stackDir = await getStackDir(stackId); 
	const composePath = join(stackDir, 'docker-compose.yml');
	const metaPath = join(stackDir, 'meta.json');

	const meta: StackMeta = {
		name,
		createdAt: new Date().toISOString(),
		updatedAt: new Date().toISOString()
	};

	try {
		await fs.mkdir(stackDir, { recursive: true });
		await Promise.all([fs.writeFile(composePath, composeContent, 'utf8'), fs.writeFile(metaPath, JSON.stringify(meta, null, 2), 'utf8')]);

		let serviceCount = 0;
		try {
			const composeData = yaml.load(composeContent) as any;
			if (composeData && composeData.services) {
				serviceCount = Object.keys(composeData.services).length;
			}
		} catch (parseErr) {
			console.warn(`Could not parse compose file during creation for stack ${stackId}:`, parseErr);
		}

		return {
			id: stackId, 
			name: meta.name,
			serviceCount: serviceCount, 
			runningCount: 0,
			status: 'stopped',
			createdAt: meta.createdAt,
			updatedAt: meta.updatedAt,
			composeContent: composeContent 
		};
	} catch (err) {
		console.error('Error creating stack:', err);
		try {
			await fs.rm(stackDir, { recursive: true, force: true });
		} catch (cleanupErr) {
			console.error(`Failed to cleanup partially created stack directory ${stackDir}:`, cleanupErr);
		}
		throw new Error('Failed to create stack files');
	}
}

/**
 * Update an existing stack
 * @param {string} stackId Stack ID
 * @param {StackUpdate} updates Updates to apply
 */
export async function updateStack(stackId: string, updates: StackUpdate): Promise<Stack> {
	const metaPath = await getStackMetaPath(stackId);
	const composePath = await getComposeFilePath(stackId);

	try {
		const metaContent = await fs.readFile(metaPath, 'utf8');
		const meta = JSON.parse(metaContent) as StackMeta;

		const updatedMeta: StackMeta = {
			...meta,
			name: updates.name || meta.name,
			updatedAt: new Date().toISOString()
		};

		const promises = [fs.writeFile(metaPath, JSON.stringify(updatedMeta, null, 2), 'utf8')];

		if (updates.composeContent) {
			promises.push(fs.writeFile(composePath, updates.composeContent, 'utf8'));
		}

		await Promise.all(promises);

		const composeContent = updates.composeContent || (await fs.readFile(composePath, 'utf8'));
		const services = await getStackServices(stackId, composeContent);

		const serviceCount = services.length;
		const runningCount = services.filter((s) => s.state?.Running).length;

		let status: Stack['status'] = 'stopped';
		if (runningCount === serviceCount && serviceCount > 0) {
			status = 'running';
		} else if (runningCount > 0) {
			status = 'partially running';
		}

		return {
			id: stackId,
			name: updatedMeta.name,
			serviceCount,
			runningCount,
			status,
			createdAt: updatedMeta.createdAt,
			updatedAt: updatedMeta.updatedAt
		};
	} catch (err) {
		console.error(`Error updating stack ${stackId}:`, err);
		throw new Error('Failed to update stack');
	}
}

/**
 * Start a stack
 * @param {string} stackId Stack ID
 */
export async function startStack(stackId: string): Promise<boolean> {
	try {
		const compose = await getComposeInstance(stackId);
		await compose.pull();
		await compose.up();
		return true;
	} catch (err: unknown) {
		console.error(`Error starting stack ${stackId}:`, err);
		const errorMessage = err instanceof Error ? err.message : String(err);
		throw new Error(`Failed to start stack: ${errorMessage}`);
	}
}

/**
 * Stop a stack
 * @param {string} stackId Stack ID
 */
export async function stopStack(stackId: string): Promise<boolean> {
	try {
		const compose = await getComposeInstance(stackId);
		await compose.down();
		return true;
	} catch (err: unknown) {
		console.error(`Error stopping stack ${stackId}:`, err);
		const errorMessage = err instanceof Error ? err.message : String(err);
		throw new Error(`Failed to stop stack: ${errorMessage}`);
	}
}

/**
 * Restart a stack
 * @param {string} stackId Stack ID
 */
export async function restartStack(stackId: string): Promise<boolean> {
	try {
		const compose = await getComposeInstance(stackId);
		await compose.down();
		await compose.up();
		return true;
	} catch (err: unknown) {
		console.error(`Error restarting stack ${stackId}:`, err);
		const errorMessage = err instanceof Error ? err.message : String(err);
		throw new Error(`Failed to restart stack: ${errorMessage}`);
	}
}

/**
 * Remove a stack
 * @param {string} stackId Stack ID
 */
export async function removeStack(stackId: string): Promise<boolean> {
	try {
		const compose = await getComposeInstance(stackId);
		await compose.down();

		const stackDir = await getStackDir(stackId);
		await fs.rm(stackDir, { recursive: true, force: true });

		return true;
	} catch (err: unknown) {
		console.error(`Error removing stack ${stackId}:`, err);
		const errorMessage = err instanceof Error ? err.message : String(err);
		throw new Error(`Failed to remove stack: ${errorMessage}`);
	}
}

/**
 * Discover existing Docker Compose stacks by analyzing container labels
 * @returns {Promise<Stack[]>} External stacks discovered
 */
export async function discoverExternalStacks(): Promise<Stack[]> {
	try {
		const docker = getDockerClient();
		const containers = await docker.listContainers({ all: true });

		const composeProjectLabel = 'com.docker.compose.project';
		const composeServiceLabel = 'com.docker.compose.service';

		const projectMap: Record<string, any[]> = {};

		containers.forEach((container) => {
			const labels = container.Labels || {};
			const projectName = labels[composeProjectLabel];

			if (projectName) {
				if (!projectMap[projectName]) {
					projectMap[projectName] = [];
				}

				projectMap[projectName].push({
					id: container.Id,
					name: labels[composeServiceLabel] || container.Names[0]?.substring(1),
					state: {
						Running: container.State === 'running',
						Status: container.State,
						ExitCode: 0
					}
				});
			}
		});

		const externalStacks: Stack[] = [];

		for (const [projectName, services] of Object.entries(projectMap)) {
			const stackDir = await getStackDir(projectName);
			try {
				await fs.access(stackDir);
				// Stack is managed by Arcane, skip it
				continue;
			} catch {
				// Stack is not managed by Arcane, include it
			}

			const serviceCount = services.length;
			const runningCount = services.filter((s) => s.state.Running).length;

			let status: Stack['status'] = 'stopped';
			if (runningCount === serviceCount && serviceCount > 0) {
				status = 'running';
			} else if (runningCount > 0) {
				status = 'partially running';
			}

			externalStacks.push({
				id: projectName,
				name: projectName,
				services,
				serviceCount,
				runningCount,
				status,
				isExternal: true,
				createdAt: new Date().toISOString(),
				updatedAt: new Date().toISOString()
			});
		}

		return externalStacks;
	} catch (err) {
		console.error('Error discovering external stacks:', err);
		return [];
	}
}

/**
 * Import an external stack into Arcane
 *
 * @param stackId The ID of the external stack to import
 * @returns The newly imported stack object
 */
export async function importExternalStack(stackId: string): Promise<Stack> {
	// 1. First, check if we can find the stack by ID
	const docker = getDockerClient();
	const containers = await docker.listContainers({ all: true });

	// Filter containers that belong to this stack
	const stackContainers = containers.filter((container) => {
		const labels = container.Labels || {};
		return labels['com.docker.compose.project'] === stackId;
	});

	if (stackContainers.length === 0) {
		throw new Error(`No containers found for stack '${stackId}'`);
	}

	// 2. Try to locate the compose file (if available)
	let composeFilePath = '';
	const container = stackContainers[0];
	const labels = container.Labels || {};

	if (labels['com.docker.compose.project.config_files']) {
		composeFilePath = labels['com.docker.compose.project.config_files'];
	}

	// 3. Read the compose file if available, or create a new one
	let composeContent = '';

	if (composeFilePath) {
		try {
			composeContent = await fs.readFile(composeFilePath, 'utf8');
		} catch (err) {
			console.warn(`Couldn't read compose file at ${composeFilePath}:`, err);
			// Will generate a new one below
		}
	}

	// 4. If we couldn't read the compose file, generate one based on container inspection
	if (!composeContent) {
		// Create a basic compose file from container inspection
		const services: Record<string, any> = {};

		for (const container of stackContainers) {
			const containerLabels = container.Labels || {};
			const serviceName = containerLabels['com.docker.compose.service'] || container.Names[0]?.replace(`/${stackId}_`, '').replace('_1', '') || `service_${container.Id.substring(0, 8)}`;

			// Inspect the container to get more details
			const containerDetails = await docker.getContainer(container.Id).inspect();

			services[serviceName] = {
				image: container.Image
				// Add other properties based on containerDetails
			};
		}

		// Generate the compose file content
		composeContent = `# Generated compose file for imported stack: ${stackId}
# This was automatically generated by Arcane from an external stack
# You may need to adjust it manually for correct operation

version: '3'
services:
${yaml.dump({ services }).substring(10)}`; // Remove the services: line
	}

	// 5. Create a new stack in Arcane's managed stacks
	return await createStack(stackId, composeContent);
}
