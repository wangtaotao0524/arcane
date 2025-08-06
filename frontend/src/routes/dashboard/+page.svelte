<script lang="ts">
	import type { PageData } from './$types';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import {
		AlertCircle,
		Box,
		HardDrive,
		RefreshCw,
		Loader2,
		Monitor,
		PlayCircle,
		StopCircle,
		Trash2,
		MemoryStick,
		Cpu,
		Container
	} from '@lucide/svelte';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { capitalizeFirstLetter } from '$lib/utils/string.utils';
	import { formatBytes } from '$lib/utils/bytes.util';
	import { toast } from 'svelte-sonner';
	import PruneConfirmationDialog from '$lib/components/dialogs/prune-confirmation-dialog.svelte';
	import { handleApiResultWithCallbacks } from '$lib/utils/api.util';
	import { tryCatch } from '$lib/utils/try-catch';
	import { systemAPI, environmentAPI, settingsAPI } from '$lib/services/api';
	import { openConfirmDialog } from '$lib/components/confirm-dialog';
	import { onMount } from 'svelte';
	import type { PruneType } from '$lib/types/actions.type';
	import DropdownCard from '$lib/components/dropdown-card.svelte';
	import MeterMetric from '$lib/components/meter-metric.svelte';
	import DockerIcon from '$lib/icons/docker-icon.svelte';
	import type { SystemStats } from '$lib/models/system-stats';
	import type { ContainerInfo } from '$lib/models/container-info';
	import type { SearchPaginationSortRequest } from '$lib/types/pagination.type';
	import DashboardContainerTable from './dash-container-table.svelte';
	import DashboardImageTable from './dash-image-table.svelte';

	let { data }: { data: PageData } = $props();

	let dashboardStates = $state({
		dockerInfo: data.dockerInfo,
		containers: Array.isArray(data.containers) ? data.containers : data.containers?.data || [],
		images: Array.isArray(data.images) ? data.images : data.images?.data || [],
		settings: data.settings,
		systemStats: null as SystemStats | null,
		error: data.error || null,
		isPruneDialogOpen: false
	});

	let isLoading = $state({
		starting: false,
		stopping: false,
		refreshing: false,
		pruning: false,
		loadingStats: true,
		loadingContainers: false,
		loadingDockerInfo: false,
		loadingImages: false
	});

	let liveSystemStats = $state(null as SystemStats | null);
	let statsInterval: NodeJS.Timeout | null = null;
	let hasInitialStatsLoaded = $state(false);

	let historicalData = $state({
		cpu: [] as Array<{ date: Date; value: number }>,
		memory: [] as Array<{ date: Date; value: number }>,
		disk: [] as Array<{ date: Date; value: number }>,
		containers: [] as Array<{ date: Date; value: number }>
	});

	const runningContainers = $derived(
		dashboardStates.containers?.filter((c: ContainerInfo) => c.State === 'running').length ?? 0
	);

	const stoppedContainers = $derived(
		dashboardStates.containers?.filter((c: ContainerInfo) => c.State === 'exited').length ?? 0
	);

	const totalImageSize = $derived(
		dashboardStates.images?.reduce((sum, image) => sum + (image.Size || 0), 0) ?? 0
	);

	const currentStats = $derived(dashboardStates.systemStats || liveSystemStats);

	function addToHistoricalData(stats: SystemStats) {
		const now = new Date();
		const maxPoints = 20;

		if (stats.cpuUsage !== undefined) {
			historicalData.cpu.push({ date: now, value: stats.cpuUsage });
			if (historicalData.cpu.length > maxPoints) {
				historicalData.cpu = historicalData.cpu.slice(-maxPoints);
			}
		}

		if (stats.memoryUsage !== undefined && stats.memoryTotal !== undefined) {
			const memoryPercent = (stats.memoryUsage / stats.memoryTotal) * 100;
			historicalData.memory.push({ date: now, value: memoryPercent });
			if (historicalData.memory.length > maxPoints) {
				historicalData.memory = historicalData.memory.slice(-maxPoints);
			}
		}

		if (stats.diskUsage !== undefined && stats.diskTotal !== undefined) {
			const diskPercent = (stats.diskUsage / stats.diskTotal) * 100;
			historicalData.disk.push({ date: now, value: diskPercent });
			if (historicalData.disk.length > maxPoints) {
				historicalData.disk = historicalData.disk.slice(-maxPoints);
			}
		}

		historicalData.containers.push({ date: now, value: runningContainers });
		if (historicalData.containers.length > maxPoints) {
			historicalData.containers = historicalData.containers.slice(-maxPoints);
		}
	}

	function getContainerDisplayName(container: ContainerInfo): string {
		if (container.Names && container.Names.length > 0) {
			return container.Names[0].startsWith('/')
				? container.Names[0].substring(1)
				: container.Names[0];
		}
		return container.Id?.substring(0, 12) || 'Unknown';
	}

	async function onContainerRefresh(options: SearchPaginationSortRequest) {
		const response = await environmentAPI.getContainers(
			options.pagination,
			options.sort,
			options.search,
			options.filters
		);

		if (Array.isArray(response)) {
			dashboardStates.containers = response;
			return {
				data: response,
				pagination: {
					totalPages: 1,
					totalItems: response.length,
					currentPage: options.pagination?.page || 1,
					itemsPerPage: response.length
				}
			};
		} else {
			dashboardStates.containers = response.data || [];
			return {
				data: response.data || [],
				pagination: response.pagination
			};
		}
	}

	async function onImageRefresh(options: SearchPaginationSortRequest) {
		const response = await environmentAPI.getImages(
			options.pagination,
			options.sort,
			options.search,
			options.filters
		);

		if (Array.isArray(response)) {
			dashboardStates.images = response;
			return {
				data: response,
				pagination: {
					totalPages: 1,
					totalItems: response.length,
					currentPage: options.pagination?.page || 1,
					itemsPerPage: response.length
				}
			};
		} else {
			dashboardStates.images = response.data || [];
			return {
				data: response.data || [],
				pagination: response.pagination
			};
		}
	}

	async function fetchLiveSystemStats() {
		if (!hasInitialStatsLoaded) {
			isLoading.loadingStats = true;
		}

		try {
			const response = await systemAPI.getStats();

			let stats: SystemStats | null = null;

			if (response && typeof response === 'object') {
				if ('success' in response && response.success && 'data' in response) {
					stats = response.data as SystemStats;
				} else if ('cpuUsage' in response) {
					stats = response as SystemStats;
				}
			}

			if (stats) {
				liveSystemStats = stats;
				dashboardStates.systemStats = stats;
				addToHistoricalData(stats);
			} else {
				console.warn('Invalid system stats response format:', response);
			}
		} catch (error) {
			console.error('Failed to fetch live system stats:', error);
		} finally {
			if (!hasInitialStatsLoaded) {
				isLoading.loadingStats = false;
				hasInitialStatsLoaded = true;
			}
		}
	}

	async function refreshData() {
		if (isLoading.refreshing) return;
		isLoading.refreshing = true;

		// Set all loading states to true at start
		isLoading.loadingDockerInfo = true;
		isLoading.loadingContainers = true;
		isLoading.loadingImages = true;

		const [dockerInfoResult, containersResult, imagesResult, settingsResult] =
			await Promise.allSettled([
				tryCatch(systemAPI.getDockerInfo()),
				tryCatch(
					environmentAPI.getContainers(
						data.containerRequestOptions.pagination,
						data.containerRequestOptions.sort,
						data.containerRequestOptions.search,
						data.containerRequestOptions.filters
					)
				),
				tryCatch(
					environmentAPI.getImages(
						data.imageRequestOptions.pagination,
						data.imageRequestOptions.sort,
						data.imageRequestOptions.search,
						data.imageRequestOptions.filters
					)
				),
				tryCatch(settingsAPI.getSettings())
			]);

		// Handle results and clear individual loading states
		if (dockerInfoResult.status === 'fulfilled' && !dockerInfoResult.value.error) {
			dashboardStates.dockerInfo = dockerInfoResult.value.data;
		}
		isLoading.loadingDockerInfo = false;

		if (containersResult.status === 'fulfilled' && !containersResult.value.error) {
			const containerData = containersResult.value.data;
			dashboardStates.containers = Array.isArray(containerData)
				? containerData
				: containerData?.data || [];
		}
		isLoading.loadingContainers = false;

		if (imagesResult.status === 'fulfilled' && !imagesResult.value.error) {
			const imageData = imagesResult.value.data;
			dashboardStates.images = Array.isArray(imageData) ? imageData : imageData?.data || [];
		}
		isLoading.loadingImages = false;

		if (settingsResult.status === 'fulfilled' && !settingsResult.value.error) {
			dashboardStates.settings = settingsResult.value.data;
		}

		await fetchLiveSystemStats();
		isLoading.refreshing = false;
	}

	onMount(() => {
		let mounted = true;

		fetchLiveSystemStats();

		if (!statsInterval) {
			statsInterval = setInterval(() => {
				if (mounted) {
					fetchLiveSystemStats();
				}
			}, 5000);
		}

		return () => {
			mounted = false;
			if (statsInterval) {
				clearInterval(statsInterval);
				statsInterval = null;
			}
		};
	});

	async function handleStartAll() {
		if (isLoading.starting || !dashboardStates.dockerInfo || stoppedContainers === 0) return;
		isLoading.starting = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(systemAPI.startAllStoppedContainers()),
			message: 'Failed to Start All Containers',
			setLoadingState: (value) => (isLoading.starting = value),
			onSuccess: async () => {
				toast.success('All Containers Started Successfully.');
				await refreshData();
			}
		});
	}

	async function handleStopAll() {
		if (isLoading.stopping || !dashboardStates.dockerInfo || runningContainers === 0) return;
		openConfirmDialog({
			title: 'Stop All Containers',
			message: 'Are you sure you want to stop all running containers?',
			confirm: {
				label: 'Confirm',
				destructive: false,
				action: async () => {
					handleApiResultWithCallbacks({
						result: await tryCatch(systemAPI.stopAllContainers()),
						message: 'Failed to Stop All Running Containers',
						setLoadingState: (value) => (isLoading.stopping = value),
						onSuccess: async () => {
							toast.success('All Containers Stopped Successfully.');
							await refreshData();
						}
					});
				}
			}
		});
	}

	async function confirmPrune(selectedTypes: PruneType[]) {
		if (isLoading.pruning || selectedTypes.length === 0) return;
		isLoading.pruning = true;

		const pruneOptions = {
			containers: selectedTypes.includes('containers'),
			images: selectedTypes.includes('images'),
			volumes: selectedTypes.includes('volumes'),
			networks: selectedTypes.includes('networks'),
			dangling: dashboardStates.settings?.dockerPruneMode === 'dangling'
		};

		handleApiResultWithCallbacks({
			result: await tryCatch(systemAPI.pruneAll(pruneOptions)),
			message: `Failed to Prune ${selectedTypes.join(', ')}`,
			setLoadingState: (value) => (isLoading.pruning = value),
			onSuccess: async () => {
				dashboardStates.isPruneDialogOpen = false;
				const formattedTypes = selectedTypes.map((type) => capitalizeFirstLetter(type)).join(', ');
				toast.success(
					`${formattedTypes} ${selectedTypes.length > 1 ? 'were' : 'was'} pruned successfully.`
				);
				await refreshData();
			}
		});
	}
</script>

<div class="space-y-8">
	<!-- Header -->
	<div class="flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
		<div class="space-y-1">
			<h1 class="text-3xl font-bold tracking-tight">Dashboard</h1>
			<p class="text-muted-foreground max-w-2xl text-sm">Overview of your Container Environment</p>
		</div>
		<Button
			variant="outline"
			size="sm"
			class="arcane-button-restart h-9"
			onclick={refreshData}
			disabled={isLoading.refreshing ||
				isLoading.starting ||
				isLoading.stopping ||
				isLoading.pruning}
		>
			{#if isLoading.refreshing}
				<Loader2 class="mr-2 size-4 animate-spin" />
			{:else}
				<RefreshCw class="mr-2 size-4" />
			{/if}
			Refresh
		</Button>
	</div>

	{#if dashboardStates.error}
		<Alert.Root variant="destructive">
			<AlertCircle class="mr-2 size-4" />
			<Alert.Title>Connection Error</Alert.Title>
			<Alert.Description>
				{dashboardStates.error} Please check your Docker connection in
				<a href="/settings" class="underline">Settings</a>.
			</Alert.Description>
		</Alert.Root>
	{/if}

	<!-- Quick Actions -->
	<section>
		<h2 class="mb-3 text-lg font-semibold tracking-tight">Quick Actions</h2>
		{#if isLoading.loadingDockerInfo}
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
				{#each Array(3) as _}
					<div class="bg-card flex items-center rounded-lg border p-3">
						<div class="bg-muted mr-3 size-6 animate-pulse rounded-full"></div>
						<div class="flex-1">
							<div class="bg-muted mb-1 h-3 w-16 animate-pulse rounded"></div>
							<div class="bg-muted h-2 w-12 animate-pulse rounded"></div>
						</div>
					</div>
				{/each}
			</div>
		{:else}
			<div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
				<button
					class="group bg-card flex items-center rounded-lg border p-3 shadow-sm ring-offset-background transition-all hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
					disabled={!dashboardStates.dockerInfo ||
						stoppedContainers === 0 ||
						isLoading.starting ||
						isLoading.stopping ||
						isLoading.pruning}
					onclick={handleStartAll}
				>
					<div
						class="mr-3 flex size-6 items-center justify-center rounded-full bg-green-500/10 transition-colors group-hover:bg-green-500/20"
					>
						{#if isLoading.starting}
							<Loader2 class="size-3 animate-spin text-green-500" />
						{:else}
							<PlayCircle class="size-3 text-green-500" />
						{/if}
					</div>
					<div class="flex-1 text-left">
						<div class="text-sm font-medium">Start All Stopped</div>
						<div class="text-muted-foreground text-xs">{stoppedContainers} containers</div>
					</div>
				</button>

				<button
					class="group bg-card flex items-center rounded-lg border p-3 shadow-sm ring-offset-background transition-all hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
					disabled={!dashboardStates.dockerInfo ||
						runningContainers === 0 ||
						isLoading.starting ||
						isLoading.stopping ||
						isLoading.pruning}
					onclick={handleStopAll}
				>
					<div
						class="mr-3 flex size-6 items-center justify-center rounded-full bg-blue-500/10 transition-colors group-hover:bg-blue-500/20"
					>
						{#if isLoading.stopping}
							<Loader2 class="size-3 animate-spin text-blue-500" />
						{:else}
							<StopCircle class="size-3 text-blue-500" />
						{/if}
					</div>
					<div class="flex-1 text-left">
						<div class="text-sm font-medium">Stop All Running</div>
						<div class="text-muted-foreground text-xs">{runningContainers} containers</div>
					</div>
				</button>

				<button
					class="group bg-card hover:border-destructive/50 disabled:hover:border-border flex items-center rounded-lg border p-3 shadow-sm ring-offset-background transition-all hover:shadow-md focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-60"
					disabled={!dashboardStates.dockerInfo ||
						isLoading.starting ||
						isLoading.stopping ||
						isLoading.pruning}
					onclick={() => (dashboardStates.isPruneDialogOpen = true)}
				>
					<div
						class="mr-3 flex size-6 items-center justify-center rounded-full bg-red-500/10 transition-colors group-hover:bg-red-500/20"
					>
						{#if isLoading.pruning}
							<Loader2 class="size-3 animate-spin text-red-500" />
						{:else}
							<Trash2 class="size-3 text-red-500" />
						{/if}
					</div>
					<div class="flex-1 text-left">
						<div class="text-sm font-medium">Prune System</div>
						<div class="text-muted-foreground text-xs">Clean unused resources</div>
					</div>
				</button>
			</div>
		{/if}
	</section>

	<!-- System Overview -->
	<section>
		<DropdownCard
			id="system-overview"
			title="System Overview"
			description="Hardware and Docker engine information"
			icon={Monitor}
			defaultExpanded={true}
		>
			<div class="grid grid-cols-2 gap-4 md:grid-cols-4">
				<MeterMetric
					title="Running Containers"
					icon={Container}
					description="Active containers"
					currentValue={isLoading.loadingContainers ? undefined : runningContainers}
					formatValue={(v) => v.toString()}
					maxValue={dashboardStates.containers?.length || 10}
					unit="containers"
					loading={isLoading.loadingContainers}
				/>

				<MeterMetric
					title="CPU Usage"
					icon={Cpu}
					description="Processor utilization"
					currentValue={isLoading.loadingStats || !hasInitialStatsLoaded
						? undefined
						: currentStats?.cpuUsage}
					unit="%"
					maxValue={100}
					loading={isLoading.loadingStats || !hasInitialStatsLoaded}
				/>

				<MeterMetric
					title="Memory Usage"
					icon={MemoryStick}
					description="RAM utilization"
					currentValue={isLoading.loadingStats || !hasInitialStatsLoaded
						? undefined
						: currentStats?.memoryUsage !== undefined && currentStats?.memoryTotal !== undefined
							? (currentStats.memoryUsage / currentStats.memoryTotal) * 100
							: undefined}
					unit="%"
					formatValue={(v) => `${v.toFixed(1)}`}
					maxValue={100}
					loading={isLoading.loadingStats || !hasInitialStatsLoaded}
				/>

				<MeterMetric
					icon={HardDrive}
					title="Disk Usage"
					description="Storage utilization"
					currentValue={isLoading.loadingStats || !hasInitialStatsLoaded
						? undefined
						: currentStats?.diskUsage !== undefined && currentStats?.diskTotal !== undefined
							? (currentStats.diskUsage / currentStats.diskTotal) * 100
							: undefined}
					unit="%"
					formatValue={(v) => `${v.toFixed(1)}`}
					maxValue={100}
					loading={isLoading.loadingStats || !hasInitialStatsLoaded}
				/>
			</div>

			<!-- Always show these cards but with loading states -->
			<div class="mt-6 grid grid-cols-1 gap-4 md:grid-cols-3">
				<Card.Root class="rounded-lg border shadow-sm">
					<Card.Content class="p-4">
						<div class="flex items-center gap-3">
							<div class="rounded-lg bg-blue-500/10 p-2">
								{#if isLoading.loadingDockerInfo}
									<Loader2 class="size-5 text-blue-500 animate-spin" />
								{:else}
									<DockerIcon class="size-5 text-blue-500" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								{#if isLoading.loadingDockerInfo}
									<div class="space-y-2">
										<div class="h-4 w-24 bg-muted animate-pulse rounded"></div>
										<div class="h-3 w-32 bg-muted animate-pulse rounded"></div>
									</div>
								{:else}
									<p class="text-sm font-medium">Docker Engine</p>
									<p class="text-xs text-muted-foreground">
										{dashboardStates.dockerInfo?.version || 'Unknown'} •
										{dashboardStates.dockerInfo?.os || 'Unknown'}
									</p>
								{/if}
							</div>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="rounded-lg border shadow-sm">
					<Card.Content class="p-4">
						<div class="flex items-center gap-3">
							<div class="rounded-lg bg-green-500/10 p-2">
								{#if isLoading.loadingContainers}
									<Loader2 class="size-5 text-green-500 animate-spin" />
								{:else}
									<Box class="size-5 text-green-500" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								{#if isLoading.loadingContainers}
									<div class="space-y-2">
										<div class="h-4 w-28 bg-muted animate-pulse rounded"></div>
										<div class="h-3 w-20 bg-muted animate-pulse rounded"></div>
									</div>
								{:else}
									<p class="text-sm font-medium">Total Containers</p>
									<p class="text-xs text-muted-foreground">
										{dashboardStates.containers?.length || 0} total •
										{stoppedContainers} stopped
									</p>
								{/if}
							</div>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root class="rounded-lg border shadow-sm">
					<Card.Content class="p-4">
						<div class="flex items-center gap-3">
							<div class="rounded-lg bg-purple-500/10 p-2">
								{#if isLoading.loadingImages}
									<Loader2 class="size-5 text-purple-500 animate-spin" />
								{:else}
									<HardDrive class="size-5 text-purple-500" />
								{/if}
							</div>
							<div class="flex-1 min-w-0">
								{#if isLoading.loadingImages}
									<div class="space-y-2">
										<div class="h-4 w-24 bg-muted animate-pulse rounded"></div>
										<div class="h-3 w-28 bg-muted animate-pulse rounded"></div>
									</div>
								{:else}
									<p class="text-sm font-medium">Docker Images</p>
									<p class="text-xs text-muted-foreground">
										{dashboardStates.dockerInfo?.images || 0} images •
										{formatBytes(totalImageSize)}
									</p>
								{/if}
							</div>
						</div>
					</Card.Content>
				</Card.Root>
			</div>
		</DropdownCard>
	</section>

	<!-- Resources -->
	<section>
		<h2 class="mb-4 text-lg font-semibold tracking-tight">Resources</h2>
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<DashboardContainerTable
				containers={dashboardStates.containers || []}
				isLoading={isLoading.loadingContainers}
				onRefresh={onContainerRefresh}
				{getContainerDisplayName}
			/>

			<DashboardImageTable
				images={dashboardStates.images}
				isLoading={isLoading.loadingImages}
				onRefresh={onImageRefresh}
			/>
		</div>
	</section>

	<PruneConfirmationDialog
		bind:open={dashboardStates.isPruneDialogOpen}
		isPruning={isLoading.pruning}
		imagePruneMode={(dashboardStates.settings?.dockerPruneMode as 'dangling' | 'all') || 'dangling'}
		onConfirm={confirmPrune}
		onCancel={() => (dashboardStates.isPruneDialogOpen = false)}
	/>
</div>
