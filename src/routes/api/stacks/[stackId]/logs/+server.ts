import { getDockerClient } from '$lib/services/docker/core';
import type { RequestHandler } from '@sveltejs/kit';
import type { Readable } from 'stream';

interface StackLogStreamSource extends UnderlyingDefaultSource<Uint8Array> {
	logStreams?: Map<string, Readable>;
}

export const GET: RequestHandler = async ({ params }) => {
	if (!params.stackId) {
		return new Response('Stack ID is required', { status: 400 });
	}

	const docker = await getDockerClient();

	const stream = new ReadableStream<Uint8Array>({
		async start(controller) {
			try {
				const containers = await docker.listContainers({
					all: true,
					filters: { label: [`com.docker.compose.project=${params.stackId}`] }
				});

				const logStreams = new Map<string, Readable>();
				let isControllerClosed = false;

				const cleanup = () => {
					isControllerClosed = true;
					logStreams.forEach((stream, containerId) => {
						if (stream && stream.destroy) {
							stream.destroy();
						}
					});
					logStreams.clear();
				};

				for (const containerInfo of containers) {
					const container = docker.getContainer(containerInfo.Id);
					const serviceName = containerInfo.Labels['com.docker.compose.service'] || containerInfo.Names[0]?.replace('/', '') || 'unknown';

					try {
						const logStream = (await container.logs({
							follow: true,
							stdout: true,
							stderr: true,
							timestamps: true
						})) as Readable;

						logStreams.set(containerInfo.Id, logStream);

						logStream.on('data', (chunk) => {
							try {
								if (!isControllerClosed) {
									const logData = chunk.toString();
									const lines = logData.split('\n');

									for (const line of lines) {
										if (line.length > 8) {
											// Remove Docker's 8-byte header
											const cleanLine = line.substring(8);
											if (cleanLine.trim()) {
												// Format each line with service prefix and ensure it ends with newline
												const formattedLine = `[${serviceName}] ${cleanLine}\n`;
												controller.enqueue(new TextEncoder().encode(`data: ${formattedLine}\n\n`));
											}
										}
									}
								}
							} catch (error: any) {
								if (error.code === 'ERR_INVALID_STATE') {
									cleanup();
								} else {
									console.error(`Error streaming logs for ${serviceName}:`, error);
								}
							}
						});

						logStream.on('end', () => {
							logStreams.delete(containerInfo.Id);
							if (logStreams.size === 0 && !isControllerClosed) {
								controller.close();
								isControllerClosed = true;
							}
						});

						logStream.on('error', (err) => {
							console.error(`Docker logs stream error for ${serviceName}:`, err);
							logStreams.delete(containerInfo.Id);
						});
					} catch (error) {
						console.error(`Failed to start log stream for ${serviceName}:`, error);
					}
				}

				(this as StackLogStreamSource).logStreams = logStreams;

				if (containers.length === 0) {
					controller.enqueue(new TextEncoder().encode(`data: No containers found in this stack\n\n`));
				}
			} catch (error) {
				console.error('Error starting stack logs stream:', error);
				controller.error(error);
			}
		},

		cancel() {
			console.log('Client disconnected from stack logs stream');
			const self = this as StackLogStreamSource;
			if (self.logStreams) {
				self.logStreams.forEach((stream) => {
					if (stream && stream.destroy) {
						stream.destroy();
					}
				});
			}
		}
	} as StackLogStreamSource);

	return new Response(stream, {
		headers: {
			'Content-Type': 'text/event-stream',
			'Cache-Control': 'no-cache',
			Connection: 'keep-alive'
		}
	});
};
