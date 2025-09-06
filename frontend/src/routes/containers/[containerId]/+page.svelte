<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import CircleAlertIcon from '@lucide/svelte/icons/alert-circle';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import HardDriveIcon from '@lucide/svelte/icons/hard-drive';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import NetworkIcon from '@lucide/svelte/icons/network';
	import TerminalIcon from '@lucide/svelte/icons/terminal';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import ActivityIcon from '@lucide/svelte/icons/activity';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import DatabaseIcon from '@lucide/svelte/icons/database';
	import { invalidateAll } from '$app/navigation';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import ActionButtons from '$lib/components/action-buttons.svelte';
	import { format } from 'date-fns';
	import bytes from 'bytes';
	import type Docker from 'dockerode';
	import StatusBadge from '$lib/components/badges/status-badge.svelte';
	import { onDestroy } from 'svelte';
	import LogViewer from '$lib/components/log-viewer.svelte';
	import Meter from '$lib/components/ui/meter/meter.svelte';
	import { browser } from '$app/environment';
	import Separator from '$lib/components/ui/separator/separator.svelte';
	import type { ContainerDetailsDto, NetworkSettingsDto } from '$lib/types/container.type';

	interface NetworkConfig {
		IPAddress?: string;
		IPPrefixLen?: number;
		Gateway?: string;
		MacAddress?: string;
		Aliases?: string[];
		Links?: string[];
		[key: string]: any;
	}

	function ensureNetworkConfig(config: any): NetworkConfig {
		return config as NetworkConfig;
	}

	let { data } = $props();
	// Use new DTO types for the page data
	let container = $derived(data?.container as ContainerDetailsDto);
	let stats = $derived((data?.stats ?? null) as Docker.ContainerStats | null);

	let starting = $state(false);
	let stopping = $state(false);
	let restarting = $state(false);
	let removing = $state(false);
	let isRefreshing = $state(false);

	let activeSection = $state<string>('overview');
	let autoScrollLogs = $state(true);
	let isStreaming = $state(false);

	let logViewer = $state<LogViewer>();
	let statsEventSource: EventSource | null = $state(null);
	let showFloatingHeader = $state(false);

	const cleanContainerName = (name: string | undefined): string => {
		if (!name) return 'Container Details';
		return name.replace(/^\/+/, '');
	};

	const containerDisplayName = $derived(cleanContainerName(container?.name));

	function startStatsStream() {
		if (statsEventSource || !container?.id || !container.state?.running) return;

		try {
			const url = `/api/containers/${container.id}/stats/stream`;
			const eventSource = new EventSource(url);
			statsEventSource = eventSource;

			eventSource.onmessage = (event) => {
				if (!event.data) return;

				try {
					const statsData = JSON.parse(event.data);

					if (statsData.removed) {
						invalidateAll();
						return;
					}

					stats = statsData;
				} catch (err) {
					console.error('Error parsing stats data:', err);
				}
			};

			eventSource.onerror = (err) => {
				console.error('Stats EventSource error:', err);
				eventSource.close();
				statsEventSource = null;
			};
		} catch (error) {
			console.error('Failed to connect to stats stream:', error);
		}
	}

	function closeStatsStream() {
		if (statsEventSource) {
			statsEventSource.close();
			statsEventSource = null;
		}
	}

	$effect(() => {
		if (activeSection === 'stats' && container?.state?.running) {
			startStatsStream();
		} else if (statsEventSource) {
			closeStatsStream();
		}
	});

	onDestroy(() => {
		closeStatsStream();
	});

	const calculateCPUPercent = (statsData: Docker.ContainerStats | null): number => {
		if (!statsData || !statsData.cpu_stats || !statsData.precpu_stats) {
			return 0;
		}

		const cpuDelta = statsData.cpu_stats.cpu_usage.total_usage - (statsData.precpu_stats.cpu_usage?.total_usage || 0);
		const systemDelta = statsData.cpu_stats.system_cpu_usage - (statsData.precpu_stats.system_cpu_usage || 0);
		const numberCPUs = statsData.cpu_stats.online_cpus || statsData.cpu_stats.cpu_usage?.percpu_usage?.length || 1;

		if (systemDelta > 0 && cpuDelta > 0) {
			const cpuPercent = (cpuDelta / systemDelta) * numberCPUs * 100.0;
			return Math.min(Math.max(cpuPercent, 0), 100 * numberCPUs);
		}
		return 0;
	};

	const cpuUsagePercent = $derived(calculateCPUPercent(stats));
	const memoryUsageBytes = $derived(stats?.memory_stats?.usage || 0);
	const memoryLimitBytes = $derived(stats?.memory_stats?.limit || 0);
	const memoryUsageFormatted = $derived(bytes.format(memoryUsageBytes));
	const memoryLimitFormatted = $derived(bytes.format(memoryLimitBytes));
	const memoryUsagePercent = $derived(memoryLimitBytes > 0 ? (memoryUsageBytes / memoryLimitBytes) * 100 : 0);

	const getPrimaryIpAddress = (networkSettings: NetworkSettingsDto | undefined | null): string => {
		if (!networkSettings?.networks) return 'N/A';

		for (const networkName in networkSettings.networks) {
			const net = networkSettings.networks[networkName];
			if (net?.ipAddress) return net.ipAddress;
		}
		return 'N/A';
	};

	const primaryIpAddress = $derived(getPrimaryIpAddress(container?.networkSettings));

	$effect(() => {
		starting = false;
		stopping = false;
		restarting = false;
		removing = false;
	});

	async function refreshData() {
		isRefreshing = true;
		await invalidateAll();
		setTimeout(() => {
			isRefreshing = false;
		}, 500);
	}

	function handleLogStart() {
		isStreaming = true;
	}

	function handleLogStop() {
		isStreaming = false;
	}

	function handleLogClear() {
		invalidateAll();
	}

	function handleToggleAutoScroll() {}

	const hasEnvVars = $derived(!!(container?.config?.env && container.config.env.length > 0));
	const hasPorts = $derived(!!(container?.ports && container.ports.length > 0));
	const hasLabels = $derived(!!(container?.labels && Object.keys(container.labels).length > 0));
	const showConfiguration = $derived(hasEnvVars || hasPorts || hasLabels);

	const hasNetworks = $derived(
		!!(container?.networkSettings?.networks && Object.keys(container.networkSettings.networks).length > 0)
	);
	const hasMounts = $derived(!!(container?.mounts && container.mounts.length > 0));
	const showStats = $derived(!!(container?.state?.running && stats));

	const navigationSections = [
		{ id: 'overview', label: 'Overview', icon: HardDriveIcon },
		{ id: 'stats', label: 'Metrics', icon: ActivityIcon },
		{ id: 'logs', label: 'Logs', icon: FileTextIcon },
		{ id: 'config', label: 'Configuration', icon: SettingsIcon },
		{ id: 'network', label: 'Networks', icon: NetworkIcon },
		{ id: 'storage', label: 'Storage', icon: DatabaseIcon }
	] as const;

	const visibleNavigationSections = $derived(
		navigationSections.filter((s) => {
			if (s.id === 'stats') return showStats;
			if (s.id === 'config') return showConfiguration;
			if (s.id === 'network') return hasNetworks;
			if (s.id === 'storage') return hasMounts;
			return true;
		})
	);

	$effect(() => {
		if (!visibleNavigationSections.some((s) => s.id === activeSection)) {
			activeSection = visibleNavigationSections[0]?.id ?? 'overview';
		}
	});

	type SectionId = (typeof navigationSections)[number]['id'];

	function scrollToSection(sectionId: SectionId) {
		activeSection = sectionId;
		const element = document.getElementById(sectionId);
		if (element) {
			element.scrollIntoView({ behavior: 'smooth', block: 'start' });
		}
	}

	$effect(() => {
		if (browser) {
			const onScroll = () => {
				showFloatingHeader = window.scrollY > 100;
			};
			window.addEventListener('scroll', onScroll);
			return () => window.removeEventListener('scroll', onScroll);
		}
	});

	function parseDockerDate(input: string | Date | undefined | null): Date | null {
		if (!input) return null;
		if (input instanceof Date) return isNaN(input.getTime()) ? null : input;

		const s = String(input).trim();
		if (!s || s.startsWith('0001-01-01')) return null;

		const m = s.match(/^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(\.\d+)?Z$/);
		let normalized = s;
		if (m) {
			const base = m[1];
			const frac = m[2] ? m[2].slice(1) : '';
			const ms = frac ? '.' + frac.slice(0, 3).padEnd(3, '0') : '';
			normalized = `${base}${ms}Z`;
		}

		const d = new Date(normalized);
		return isNaN(d.getTime()) ? null : d;
	}

	function formatDockerDate(input: string | Date | undefined | null, fmt = 'PP p'): string {
		const d = parseDockerDate(input);
		return d ? format(d, fmt) : 'N/A';
	}
</script>

<div class="bg-background min-h-screen">
	{#if container}
		<div
			class="bg-background/95 sticky top-0 z-20 border-b backdrop-blur transition-all duration-300"
			style="opacity: {showFloatingHeader ? 0 : 1}; pointer-events: {showFloatingHeader ? 'none' : 'auto'};"
		>
			<div class="max-w-full px-4 py-3">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<Button variant="ghost" size="sm" href="/containers">
							<ArrowLeftIcon class="mr-2 size-4" />
							Back
						</Button>
						<div class="bg-border h-4 w-px"></div>
						<div class="flex items-center gap-2">
							<h1 class="max-w-[300px] truncate text-lg font-semibold" title={containerDisplayName}>
								{containerDisplayName}
							</h1>
							{#if container?.state}
								<StatusBadge
									variant={container.state.status === 'running' ? 'green' : container.state.status === 'exited' ? 'red' : 'amber'}
									text={container.state.status}
								/>
							{/if}
						</div>
					</div>

					<div class="flex items-center gap-2">
						{#if container}
							<ActionButtons
								id={container.id}
								type="container"
								itemState={container.state?.running ? 'running' : 'stopped'}
								loading={{ start: starting, stop: stopping, restart: restarting, remove: removing }}
							/>
						{/if}
					</div>
				</div>
			</div>
		</div>

		{#if showFloatingHeader}
			<div class="fixed left-1/2 top-4 z-30 -translate-x-1/2 transition-all duration-300 ease-in-out">
				<div class="bg-background/90 border-border/50 rounded-lg border px-4 py-3 shadow-xl backdrop-blur-xl">
					<div class="flex items-center gap-4">
						<div class="flex items-center gap-2">
							<h2 class="max-w-[150px] truncate text-sm font-medium" title={containerDisplayName}>
								{containerDisplayName}
							</h2>
							{#if container?.state}
								<StatusBadge
									variant={container.state.status === 'running' ? 'green' : container.state.status === 'exited' ? 'red' : 'amber'}
									text={container.state.status}
									class="text-xs"
								/>
							{/if}
						</div>
						<div class="bg-border h-4 w-px"></div>
						<ActionButtons
							id={container.id}
							type="container"
							itemState={container.state?.running ? 'running' : 'stopped'}
							loading={{ start: starting, stop: stopping, restart: restarting, remove: removing }}
						/>
					</div>
				</div>
			</div>
		{/if}

		<div class="flex min-h-0 overflow-hidden">
			<div class="bg-background/50 w-16 shrink-0 border-r">
				<div class="sticky top-16 p-2">
					<nav class="space-y-1">
						{#each visibleNavigationSections as section}
							{@const IconComponent = section.icon}
							<button
								onclick={() => scrollToSection(section.id)}
								class="relative flex w-full items-center justify-center rounded-md p-3 text-sm font-medium transition-colors
                                    {activeSection === section.id
									? 'bg-primary/10 text-primary border-primary/20 border'
									: 'text-muted-foreground hover:text-foreground hover:bg-muted/50'}"
								title={section.label}
							>
								<IconComponent class="size-4" />
							</button>
						{/each}
					</nav>
				</div>
			</div>

			<div class="min-w-0 flex-1 overflow-hidden">
				<div class="max-w-none p-6">
					<div class="space-y-8">
						<section id="overview" class="scroll-mt-20">
							<h2 class="mb-6 flex items-center gap-2 text-xl font-semibold">
								<HardDriveIcon class="size-5" />
								Overview
							</h2>

							<div class="mb-6">
								<Card.Root class="rounded-lg border shadow-sm">
									<Card.Header class="pb-4">
										<Card.Title>Container Details</Card.Title>
										<Card.Description class="text-muted-foreground text-sm">
											Identity, runtime, and system metadata
										</Card.Description>
									</Card.Header>

									<Card.Content class="space-y-6">
										<div class="space-y-4">
											<div class="flex items-center gap-3">
												<div class="rounded bg-blue-50 p-2 dark:bg-blue-950/20">
													<HardDriveIcon class="size-4 text-blue-600" />
												</div>
												<div class="min-w-0 flex-1">
													<div class="text-muted-foreground text-sm">Image</div>
													<div class="truncate font-medium" title={container.image}>
														{container.image || 'N/A'}
													</div>
												</div>
											</div>

											<div class="flex items-center gap-3">
												<div class="rounded bg-green-50 p-2 dark:bg-green-950/20">
													<ClockIcon class="size-4 text-green-600" />
												</div>
												<div class="min-w-0 flex-1">
													<div class="text-muted-foreground text-sm">Created</div>
													<div class="font-medium" title={formatDockerDate(container?.created)}>
														{formatDockerDate(container?.created)}
													</div>
												</div>
											</div>

											<div class="flex items-center gap-3">
												<div class="rounded bg-purple-50 p-2 dark:bg-purple-950/20">
													<NetworkIcon class="size-4 text-purple-600" />
												</div>
												<div class="min-w-0 flex-1">
													<div class="text-muted-foreground text-sm">IP Address</div>
													<div class="font-medium">{primaryIpAddress}</div>
												</div>
											</div>

											<div class="flex items-center gap-3">
												<div class="rounded bg-amber-50 p-2 dark:bg-amber-950/20">
													<TerminalIcon class="size-4 text-amber-600" />
												</div>
												<div class="min-w-0 flex-1">
													<div class="text-muted-foreground text-sm">Command</div>
													<div class="truncate font-medium" title={container.config?.cmd?.join(' ')}>
														{container.config?.cmd?.join(' ') || 'N/A'}
													</div>
												</div>
											</div>
										</div>

										<Separator />

										<div class="space-y-3">
											<h4 class="text-sm font-semibold tracking-tight">System</h4>

											<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
												<div class="space-y-1">
													<div class="text-muted-foreground text-xs">Container ID</div>
													<div class="bg-muted/50 max-w-full truncate rounded px-2 py-1.5 font-mono text-xs">
														{container.id}
													</div>
												</div>

												{#if container.config?.workingDir}
													<div class="space-y-1">
														<div class="text-muted-foreground text-xs">Working Directory</div>
														<div class="bg-muted/50 max-w-full truncate rounded px-2 py-1.5 font-mono text-xs">
															{container.config.workingDir}
														</div>
													</div>
												{/if}

												{#if container.config?.user}
													<div class="space-y-1">
														<div class="text-muted-foreground text-xs">User</div>
														<div class="bg-muted/50 inline-flex rounded px-2 py-1.5 font-mono text-xs">
															{container.config.user}
														</div>
													</div>
												{/if}

												{#if container.state?.health}
													<div class="space-y-1 sm:col-span-2">
														<div class="text-muted-foreground text-xs">Health</div>
														<div class="flex flex-wrap items-center gap-3">
															<StatusBadge
																variant={container.state.health.status === 'healthy'
																	? 'green'
																	: container.state.health.status === 'unhealthy'
																		? 'red'
																		: 'amber'}
																text={container.state.health.status}
															/>
															{#if container.state.health.log && container.state.health.log.length > 0}
																{@const first = container.state.health.log[0]}
																{@const lastCheck = (first?.Start ?? first?.start) as string | undefined}
																{#if lastCheck}
																	<span class="text-muted-foreground text-xs">
																		Last check: {formatDockerDate(lastCheck)}
																	</span>
																{/if}
															{/if}
														</div>
													</div>
												{/if}
											</div>
										</div>
									</Card.Content>
								</Card.Root>
							</div>
						</section>

						{#if showStats}
							<section id="stats" class="scroll-mt-20">
								<h2 class="mb-6 flex items-center gap-2 text-xl font-semibold">
									<ActivityIcon class="size-5" />
									Resource Metrics
								</h2>

								<Card.Root class="rounded-lg border shadow-sm">
									<Card.Content class="p-6">
										{#if stats && container.state?.running}
											<div class="grid grid-cols-1 gap-8 lg:grid-cols-2">
												<div class="space-y-6">
													<Meter
														label="CPU Usage"
														valueLabel="{cpuUsagePercent.toFixed(2)}%"
														value={cpuUsagePercent}
														max={100}
														variant={cpuUsagePercent > 80 ? 'destructive' : cpuUsagePercent > 60 ? 'warning' : 'default'}
														size="lg"
													/>

													<Meter
														label="Memory Usage"
														valueLabel="{memoryUsageFormatted} / {memoryLimitFormatted} ({memoryUsagePercent.toFixed(1)}%)"
														value={memoryUsagePercent}
														max={100}
														variant={memoryUsagePercent > 80 ? 'destructive' : memoryUsagePercent > 60 ? 'warning' : 'default'}
														size="lg"
													/>
												</div>

												<div class="space-y-6">
													<div>
														<h4 class="mb-4 flex items-center gap-2 font-medium">
															<NetworkIcon class="size-4" /> Network I/O
														</h4>
														<div class="grid grid-cols-2 gap-4">
															<div class="bg-muted/30 rounded p-4">
																<div class="text-muted-foreground text-sm">Received</div>
																<div class="mt-1 font-medium">
																	{bytes.format(stats.networks?.eth0?.rx_bytes || 0)}
																</div>
															</div>
															<div class="bg-muted/30 rounded p-4">
																<div class="text-muted-foreground text-sm">Transmitted</div>
																<div class="mt-1 font-medium">
																	{bytes.format(stats.networks?.eth0?.tx_bytes || 0)}
																</div>
															</div>
														</div>
													</div>

													{#if stats.blkio_stats && stats.blkio_stats.io_service_bytes_recursive && stats.blkio_stats.io_service_bytes_recursive.length > 0}
														<div>
															<h4 class="mb-4 font-medium">Block I/O</h4>
															<div class="grid grid-cols-2 gap-4">
																<div class="bg-muted/30 rounded p-4">
																	<div class="text-muted-foreground text-sm">Read</div>
																	<div class="mt-1 font-medium">
																		{bytes.format(
																			stats.blkio_stats.io_service_bytes_recursive
																				.filter((item) => item.op === 'Read')
																				.reduce((acc, item) => acc + item.value, 0)
																		)}
																	</div>
																</div>
																<div class="bg-muted/30 rounded p-4">
																	<div class="text-muted-foreground text-sm">Write</div>
																	<div class="mt-1 font-medium">
																		{bytes.format(
																			stats.blkio_stats.io_service_bytes_recursive
																				.filter((item) => item.op === 'Write')
																				.reduce((acc, item) => acc + item.value, 0)
																		)}
																	</div>
																</div>
															</div>
														</div>
													{/if}
												</div>
											</div>

											{#if stats.pids_stats && stats.pids_stats.current !== undefined}
												<div class="mt-6 border-t pt-6">
													<div class="text-sm">
														<span class="text-muted-foreground">Process count:</span>
														<span class="ml-2 font-medium">{stats.pids_stats.current}</span>
													</div>
												</div>
											{/if}
										{:else if !container.state?.running}
											<div class="text-muted-foreground py-12 text-center">Container is not running. Stats unavailable.</div>
										{:else}
											<div class="text-muted-foreground py-12 text-center">Loading stats...</div>
										{/if}
									</Card.Content>
								</Card.Root>
							</section>
						{/if}

						<section id="logs" class="scroll-mt-20">
							<div class="mb-6 flex items-center justify-between">
								<h2 class="flex items-center gap-2 text-xl font-semibold">
									<FileTextIcon class="size-5" />
									Container Logs
								</h2>

								<div class="flex items-center gap-3">
									<label class="flex items-center gap-2">
										<input type="checkbox" bind:checked={autoScrollLogs} class="size-4" />
										Auto-scroll
									</label>
									<Button variant="outline" size="sm" onclick={() => logViewer?.clearLogs()}>Clear</Button>
									{#if isStreaming}
										<div class="flex items-center gap-2">
											<div class="size-2 animate-pulse rounded-full bg-green-500"></div>
											<span class="text-sm font-medium text-green-600">Live</span>
										</div>
										<Button variant="outline" size="sm" onclick={() => logViewer?.stopLogStream()}>Stop</Button>
									{:else}
										<Button variant="outline" size="sm" onclick={() => logViewer?.startLogStream()} disabled={!container?.id}>
											Start
										</Button>
									{/if}
								</div>
							</div>

							<Card.Root class="rounded-lg border shadow-sm">
								<Card.Content class="p-0">
									<LogViewer
										bind:this={logViewer}
										bind:autoScroll={autoScrollLogs}
										type="container"
										containerId={container?.id}
										maxLines={500}
										showTimestamps={true}
										height="400px"
										onStart={handleLogStart}
										onStop={handleLogStop}
										onClear={handleLogClear}
										onToggleAutoScroll={handleToggleAutoScroll}
									/>
								</Card.Content>
							</Card.Root>
						</section>

						{#if showConfiguration}
							<section id="config" class="scroll-mt-20">
								<h2 class="mb-6 flex items-center gap-2 text-xl font-semibold">
									<SettingsIcon class="size-5" />
									Configuration
								</h2>

								<Card.Root class="rounded-lg border shadow-sm">
									<Card.Header class="pb-4">
										<Card.Title>Environment, Ports & Labels</Card.Title>
										<Card.Description class="text-muted-foreground text-sm">
											Runtime configuration and metadata for this container
										</Card.Description>
									</Card.Header>

									<Card.Content class="space-y-8">
										{#if hasEnvVars}
											<div>
												<h3 class="mb-3 text-sm font-semibold tracking-tight">Environment Variables</h3>

												{#if container.config?.env && container.config.env.length > 0}
													<ul class="divide-border/60 divide-y">
														{#each container.config.env as env, index (index)}
															{#if env.includes('=')}
																{@const [key, ...valueParts] = env.split('=')}
																{@const value = valueParts.join('=')}
																<li class="px-4 py-2.5">
																	<div class="flex min-w-0 items-center gap-3">
																		<Badge variant="secondary">
																			{key}:
																		</Badge>
																		<span class="truncate font-semibold" title={value}>{value}</span>
																	</div>
																</li>
															{:else}
																<li class="px-4 py-2.5">
																	<div class="flex min-w-0 items-center gap-3">
																		<Badge variant="secondary">ENV:</Badge>
																		<span class="truncate font-semibold" title={env}>{env}</span>
																	</div>
																</li>
															{/if}
														{/each}
													</ul>
												{:else}
													<div class="text-muted-foreground py-8 text-center">No environment variables</div>
												{/if}
											</div>
										{/if}

										{#if hasEnvVars && (hasPorts || hasLabels)}
											<Separator />
										{/if}

										{#if hasPorts}
											<div>
												<h3 class="mb-3 text-sm font-semibold tracking-tight">Port Mappings</h3>

												<ul class="divide-border/60 divide-y">
													{#each container.ports as p, idx (idx)}
														<li class="px-4 py-2.5">
															<div class="flex min-w-0 items-center gap-3">
																<Badge variant="secondary">
																	{p.privatePort}/{p.type}
																</Badge>
																{#if p.publicPort}
																	<span class="truncate font-semibold">
																		{p.ip || '0.0.0.0'}:{p.publicPort} → {p.privatePort}/{p.type}
																	</span>
																{:else}
																	<span class="text-muted-foreground font-semibold">Not published</span>
																{/if}
															</div>
														</li>
													{/each}
												</ul>
											</div>
										{/if}

										{#if hasPorts && hasLabels}
											<Separator />
										{/if}

										{#if hasLabels}
											<div>
												<h3 class="mb-3 text-sm font-semibold tracking-tight">Labels</h3>

												{#if container.labels && Object.keys(container.labels).length > 0}
													<ul class="divide-border/60 divide-y">
														{#each Object.entries(container.labels) as [key, value] (key)}
															<li class="px-4 py-2.5">
																<div class="flex min-w-0 items-center gap-3">
																	<Badge variant="secondary">
																		{key}:
																	</Badge>
																	<span class="truncate font-semibold" title={value?.toString()}>
																		{value?.toString() || ''}
																	</span>
																</div>
															</li>
														{/each}
													</ul>
												{:else}
													<div class="text-muted-foreground py-8 text-center">No labels defined</div>
												{/if}
											</div>
										{/if}
									</Card.Content>
								</Card.Root>
							</section>
						{/if}

						{#if hasNetworks}
							<section id="network" class="scroll-mt-20">
								<h2 class="mb-6 flex items-center gap-2 text-xl font-semibold">
									<NetworkIcon class="size-5" />
									Networks
								</h2>

								<Card.Root class="rounded-lg border shadow-sm">
									<Card.Content class="p-6">
										{#if container.networkSettings?.networks && Object.keys(container.networkSettings.networks).length > 0}
											<div class="space-y-6">
												{#each Object.entries(container.networkSettings.networks) as [networkName, rawNetworkConfig] (networkName)}
													{@const networkConfig = ensureNetworkConfig({
														IPAddress: rawNetworkConfig.ipAddress,
														IPPrefixLen: rawNetworkConfig.ipPrefixLen,
														Gateway: rawNetworkConfig.gateway,
														MacAddress: rawNetworkConfig.macAddress,
														Aliases: rawNetworkConfig.aliases
													})}
													<div class="rounded border p-4">
														<div class="mb-4 font-medium">{networkName}</div>
														<div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
															<div>
																<div class="text-muted-foreground text-sm">IP Address</div>
																<div class="font-mono">{rawNetworkConfig.ipAddress || 'N/A'}</div>
															</div>
															<div>
																<div class="text-muted-foreground text-sm">Gateway</div>
																<div class="font-mono">{rawNetworkConfig.gateway || 'N/A'}</div>
															</div>
															<div>
																<div class="text-muted-foreground text-sm">MAC Address</div>
																<div class="font-mono">{rawNetworkConfig.macAddress || 'N/A'}</div>
															</div>
															<div>
																<div class="text-muted-foreground text-sm">Subnet</div>
																<div class="font-mono">
																	{rawNetworkConfig.ipPrefixLen
																		? `${rawNetworkConfig.ipAddress}/${rawNetworkConfig.ipPrefixLen}`
																		: 'N/A'}
																</div>
															</div>
															{#if rawNetworkConfig.aliases && rawNetworkConfig.aliases.length > 0}
																<div class="col-span-2">
																	<div class="text-muted-foreground text-sm">Aliases</div>
																	<div class="font-mono">{rawNetworkConfig.aliases.join(', ')}</div>
																</div>
															{/if}
														</div>
													</div>
												{/each}
											</div>
										{:else}
											<div class="text-muted-foreground py-12 text-center">No networks connected</div>
										{/if}
									</Card.Content>
								</Card.Root>
							</section>
						{/if}

						{#if hasMounts}
							<section id="storage" class="scroll-mt-20">
								<h2 class="mb-6 flex items-center gap-2 text-xl font-semibold">
									<DatabaseIcon class="size-5" />
									Storage & Mounts
								</h2>

								<Card.Root class="rounded-lg border shadow-sm">
									<Card.Content class="p-6">
										{#if container.mounts && container.mounts.length > 0}
											<div class="space-y-4">
												{#each container.mounts as mount (mount.destination)}
													<div class="overflow-hidden rounded border">
														<div class="bg-muted/20 flex items-center justify-between p-4">
															<div class="flex items-center gap-3">
																<div
																	class="rounded p-2 {mount.type === 'volume'
																		? 'bg-purple-100 dark:bg-purple-950'
																		: mount.type === 'bind'
																			? 'bg-blue-100 dark:bg-blue-950'
																			: 'bg-amber-100 dark:bg-amber-950'}"
																>
																	{#if mount.type === 'volume'}
																		<DatabaseIcon class="size-4 text-purple-600" />
																	{:else if mount.type === 'bind'}
																		<HardDriveIcon class="size-4 text-blue-600" />
																	{:else}
																		<TerminalIcon class="size-4 text-amber-600" />
																	{/if}
																</div>
																<div>
																	<div class="font-medium">
																		{mount.type === 'tmpfs'
																			? 'Temporary filesystem'
																			: mount.type === 'volume'
																				? mount.name || 'Docker volume'
																				: 'Host directory'}
																	</div>
																	<div class="text-muted-foreground text-sm">
																		{mount.type} mount {mount.rw ? '(read-write)' : '(read-only)'}
																	</div>
																</div>
															</div>
															<Badge variant={mount.rw ? 'outline' : 'secondary'}>
																{mount.rw ? 'RW' : 'RO'}
															</Badge>
														</div>
														<div class="space-y-3 p-4">
															<div class="flex">
																<span class="text-muted-foreground w-24 font-medium">Container:</span>
																<span class="bg-muted/50 flex-1 rounded px-2 py-1 font-mono">{mount.destination}</span>
															</div>
															<div class="flex">
																<span class="text-muted-foreground w-24 font-medium">
																	{mount.type === 'volume' ? 'Volume:' : mount.type === 'bind' ? 'Host:' : 'Source:'}
																</span>
																<span class="bg-muted/50 flex-1 rounded px-2 py-1 font-mono">{mount.source}</span>
															</div>
														</div>
													</div>
												{/each}
											</div>
										{:else}
											<div class="py-12 text-center">
												<div class="bg-muted/50 mx-auto mb-4 flex size-16 items-center justify-center rounded-full">
													<DatabaseIcon class="text-muted-foreground size-6" />
												</div>
												<div class="text-muted-foreground">No volumes or mounts configured</div>
											</div>
										{/if}
									</Card.Content>
								</Card.Root>
							</section>
						{/if}
					</div>
				</div>
			</div>
		</div>
	{:else}
		<div class="flex min-h-screen items-center justify-center">
			<div class="text-center">
				<div class="bg-muted/50 mb-6 inline-flex rounded-full p-6">
					<CircleAlertIcon class="text-muted-foreground size-10" />
				</div>
				<h2 class="mb-3 text-2xl font-medium">Container Not Found</h2>
				<p class="text-muted-foreground mb-8 max-w-md text-center">
					Could not load container data. It may have been removed or the Docker engine is not accessible.
				</p>
				<div class="flex justify-center gap-4">
					<Button variant="outline" href="/containers">
						<ArrowLeftIcon class="mr-2 size-4" />
						Back to Containers
					</Button>
					<Button variant="default" onclick={refreshData}>
						<RefreshCwIcon class="mr-2 size-4" />
						Retry
					</Button>
				</div>
			</div>
		</div>
	{/if}
</div>
