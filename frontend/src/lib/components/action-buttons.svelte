<script lang="ts">
	import { openConfirmDialog } from './confirm-dialog';
	import { goto, invalidateAll } from '$app/navigation';
	import { toast } from 'svelte-sonner';
	import { environmentAPI } from '$lib/services/api';
	import { tryCatch } from '$lib/utils/try-catch';
	import { handleApiResultWithCallbacks } from '$lib/utils/api.util';
	import { ArcaneButton } from '$lib/components/arcane-button/index.js';
	import ProgressPopover from '$lib/components/progress-popover.svelte';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import { environmentStore, LOCAL_DOCKER_ENVIRONMENT_ID } from '$lib/stores/environment.store';
	import { get } from 'svelte/store';

	type TargetType = 'container' | 'stack';

	type LoadingStates = {
		start?: boolean;
		stop?: boolean;
		restart?: boolean;
		pull?: boolean;
		deploy?: boolean;
		redeploy?: boolean;
		remove?: boolean;
		validating?: boolean;
	};

	let {
		id,
		type = 'container',
		itemState = 'stopped',
		loading = {},
		onActionComplete = $bindable(() => {})
	}: {
		id: string;
		type?: TargetType;
		itemState?: string;
		loading: LoadingStates;
		onActionComplete?: () => void;
	} = $props();

	let isLoading = $state({
		start: false,
		stop: false,
		restart: false,
		remove: false,
		pulling: false,
		redeploy: false,
		validating: false
	});

	let pullPopoverOpen = $state(false);
	let pullProgress = $state(0);
	let pullStatusText = $state('');
	let pullError = $state('');
	let layerProgress = $state<Record<string, { current: number; total: number; status: string }>>({});

	const isRunning = $derived(itemState === 'running' || (type === 'stack' && itemState === 'partially running'));

	$effect(() => {
		isLoading.start = loading.start ?? false;
		isLoading.stop = loading.stop ?? false;
		isLoading.pulling = loading.pull ?? false;
		isLoading.remove = loading.remove ?? false;
		isLoading.restart = loading.restart ?? false;
		isLoading.redeploy = loading.redeploy ?? false;
		isLoading.validating = loading.validating ?? false;
	});

	function resetPullState() {
		pullProgress = 0;
		pullStatusText = '';
		pullError = '';
		layerProgress = {};
	}

	function calculateOverallProgress() {
		let totalCurrentBytes = 0;
		let totalExpectedBytes = 0;
		let activeLayers = 0;

		for (const id in layerProgress) {
			const layer = layerProgress[id];
			if (layer.total > 0) {
				totalCurrentBytes += layer.current;
				totalExpectedBytes += layer.total;
				activeLayers++;
			}
		}

		if (totalExpectedBytes > 0) {
			pullProgress = (totalCurrentBytes / totalExpectedBytes) * 100;
		} else if (activeLayers > 0 && totalCurrentBytes > 0) {
			pullProgress = 5;
		} else if (Object.keys(layerProgress).length > 0 && activeLayers === 0) {
			const allDone = Object.values(layerProgress).every(
				(l) => l.status && (l.status.toLowerCase().includes('pull complete') || l.status.toLowerCase().includes('already exists'))
			);
			if (allDone) pullProgress = 100;
		}
	}

	function buildPullApiUrl(): string {
		const envId = getCurrentEnvironmentId();
		if (envId === LOCAL_DOCKER_ENVIRONMENT_ID) {
			return `/api/stacks/${id}/pull`;
		}
		return `/api/environments/${envId}/stacks/${id}/pull`;
	}

	function getCurrentEnvironmentId(): string {
		const env = get(environmentStore.selected);
		return env?.id || LOCAL_DOCKER_ENVIRONMENT_ID;
	}

	function confirmAction(action: string) {
		if (action === 'remove') {
			openConfirmDialog({
				title: `Confirm ${type === 'stack' ? 'Destroy' : 'Removal'}`,
				message: `Are you sure you want to ${type === 'stack' ? 'destroy' : 'remove'} this ${type}? This action is DESTRUCTIVE and cannot be undone.`,
				confirm: {
					label: type === 'stack' ? 'Destroy' : 'Remove',
					destructive: true,
					action: async (checkboxStates) => {
						const removeFiles = checkboxStates['removeFiles'] === true;
						const removeVolumes = checkboxStates['removeVolumes'] === true;

						isLoading.remove = true;
						handleApiResultWithCallbacks({
							result: await tryCatch(
								type === 'container'
									? environmentAPI.deleteContainer(id)
									: environmentAPI.destroyProject(id, removeVolumes, removeFiles)
							),
							message: `Failed to ${type === 'stack' ? 'Destroy' : 'Remove'} ${type}`,
							setLoadingState: (value) => (isLoading.remove = value),
							onSuccess: async () => {
								toast.success(
									`${type.charAt(0).toUpperCase() + type.slice(1)} ${type === 'stack' ? 'Destroyed' : 'Removed'} Successfully`
								);
								await invalidateAll();
								goto(`${type === 'stack' ? '/compose' : 'containers'}`);
							}
						});
					}
				},
				checkboxes: [
					{ id: 'removeFiles', label: 'Remove project files', initialState: false },
					{
						id: 'removeVolumes',
						label: 'Remove volumes (Warning: Data will be lost)',
						initialState: false
					}
				]
			});
		} else if (action === 'redeploy') {
			openConfirmDialog({
				title: `Confirm Redeploy`,
				message: `Are you sure you want to redeploy this stack? This will STOP, PULL, and START the Project.`,
				confirm: {
					label: 'Redeploy',
					action: async () => {
						isLoading.redeploy = true;
						handleApiResultWithCallbacks({
							result: await tryCatch(environmentAPI.redeployProject(id)),
							message: `Failed to Redeploy ${type}`,
							setLoadingState: (value) => (isLoading.redeploy = value),
							onSuccess: async () => {
								toast.success(`${type.charAt(0).toUpperCase() + type.slice(1)} Redeployed Successfully`);
								await invalidateAll();
							}
						});
					}
				}
			});
		}
	}

	async function handleStart() {
		isLoading.start = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(type === 'container' ? environmentAPI.startContainer(id) : environmentAPI.startProject(id)),
			message: `Failed to Start ${type}`,
			setLoadingState: (value) => (isLoading.start = value),
			onSuccess: async () => {
				toast.success(`${type.charAt(0).toUpperCase() + type.slice(1)} Started Successfully`);
				await invalidateAll();
			}
		});
	}

	async function handleDeploy() {
		isLoading.start = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(environmentAPI.startProject(id)),
			message: `Failed to Start ${type}`,
			setLoadingState: (value) => (isLoading.start = value),
			onSuccess: async () => {
				toast.success(`${type.charAt(0).toUpperCase() + type.slice(1)} Started Successfully`);
				await invalidateAll();
			}
		});
	}

	async function handleStop() {
		isLoading.stop = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(type === 'container' ? environmentAPI.stopContainer(id) : environmentAPI.stopProject(id)),
			message: `Failed to Stop ${type}`,
			setLoadingState: (value) => (isLoading.stop = value),
			onSuccess: async () => {
				toast.success(`${type.charAt(0).toUpperCase() + type.slice(1)} Stopped Successfully`);
				await invalidateAll();
			}
		});
	}

	async function handleRestart() {
		isLoading.restart = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(type === 'container' ? environmentAPI.restartContainer(id) : environmentAPI.restartProject(id)),
			message: `Failed to Restart ${type}`,
			setLoadingState: (value) => (isLoading.restart = value),
			onSuccess: async () => {
				toast.success(`${type.charAt(0).toUpperCase() + type.slice(1)} Restarted Successfully`);
				await invalidateAll();
			}
		});
	}

	async function handlePull() {
		if (type === 'container') {
			// Use existing API for containers
			isLoading.pulling = true;
			handleApiResultWithCallbacks({
				result: await tryCatch(environmentAPI.pullContainerImage(id)),
				message: 'Failed to Pull Image(s)',
				setLoadingState: (value) => (isLoading.pulling = value),
				onSuccess: async () => {
					toast.success('Image(s) Pulled Successfully.');
					await invalidateAll();
				}
			});
		} else {
			await handleStackPull();
		}
	}

	async function handleStackPull() {
		resetPullState();
		isLoading.pulling = true;
		pullPopoverOpen = true;
		pullStatusText = 'Initiating pull...';

		let wasSuccessful = false;

		try {
			const response = await fetch(buildPullApiUrl(), {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				}
			});

			if (!response.ok || !response.body) {
				const errorData = await response.json().catch(() => ({
					error: 'Failed to pull images. Server returned an error.'
				}));
				const errorMessage =
					typeof errorData.error === 'string' ? errorData.error : errorData.message || `HTTP error ${response.status}`;
				throw new Error(errorMessage);
			}

			const reader = response.body.getReader();
			const decoder = new TextDecoder();
			let buffer = '';

			while (true) {
				const { done, value } = await reader.read();
				if (done) {
					pullStatusText = 'Processing final layers...';
					break;
				}

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() || '';

				for (const line of lines) {
					if (line.trim() === '') continue;
					try {
						const data = JSON.parse(line);

						if (data.error) {
							console.error('Error in stream:', data.error);
							pullError = typeof data.error === 'string' ? data.error : data.error.message || 'An error occurred during pull.';
							pullStatusText = `Error: ${pullError}`;
							continue;
						}

						pullStatusText = data.status || pullStatusText;
						if (data.id) {
							const currentLayer = layerProgress[data.id] || { current: 0, total: 0, status: '' };
							currentLayer.status = data.status || currentLayer.status;

							if (data.progressDetail) {
								currentLayer.current = data.progressDetail.current || currentLayer.current;
								currentLayer.total = data.progressDetail.total || currentLayer.total;
							}
							layerProgress[data.id] = currentLayer;
						}
						calculateOverallProgress();
					} catch (e: any) {
						console.warn('Failed to parse stream line or process data:', line, e);
					}
				}
			}

			calculateOverallProgress();
			if (!pullError && pullProgress < 100) {
				const allLayersCompleteOrExisting = Object.values(layerProgress).every(
					(l) =>
						l.status &&
						(l.status.toLowerCase().includes('complete') ||
							l.status.toLowerCase().includes('already exists') ||
							l.status.toLowerCase().includes('downloaded newer image'))
				);
				if (allLayersCompleteOrExisting && Object.keys(layerProgress).length > 0) {
					pullProgress = 100;
				}
			}

			if (pullError) {
				throw new Error(pullError);
			}

			wasSuccessful = true;
			pullProgress = 100;
			pullStatusText = 'Images pulled successfully.';
			toast.success('Images pulled successfully!');
			await invalidateAll();

			setTimeout(() => {
				pullPopoverOpen = false;
				isLoading.pulling = false;
				resetPullState();
			}, 2000);
		} catch (error: any) {
			console.error('Pull images error:', error);
			const message = error.message || 'An unexpected error occurred while pulling images.';
			pullError = message;
			pullStatusText = `Failed: ${message}`;
			toast.error(message);
		} finally {
			if (!wasSuccessful) {
				isLoading.pulling = false;
			}
		}
	}
</script>

<div class="flex items-center gap-2">
	{#if !isRunning}
		<ArcaneButton
			action={type === 'container' ? 'start' : 'deploy'}
			onclick={type === 'container' ? () => handleStart() : () => handleDeploy()}
			loading={isLoading.start}
		/>
	{:else}
		<ArcaneButton
			customLabel={type === 'stack' ? 'Down' : 'Stop'}
			action="stop"
			onclick={() => handleStop()}
			loading={isLoading.stop}
		/>
		<ArcaneButton action="restart" onclick={() => handleRestart()} loading={isLoading.restart} />
	{/if}

	{#if type === 'container'}
		<ArcaneButton action="remove" onclick={() => confirmAction('remove')} loading={isLoading.remove} />
	{:else}
		<ArcaneButton action="redeploy" onclick={() => confirmAction('redeploy')} loading={isLoading.redeploy} />

		{#if type === 'stack'}
			<ProgressPopover
				bind:open={pullPopoverOpen}
				bind:progress={pullProgress}
				title="Pulling Images"
				statusText={pullStatusText}
				error={pullError}
				loading={isLoading.pulling}
				icon={DownloadIcon}
			>
				<ArcaneButton action="pull" onclick={() => handlePull()} loading={isLoading.pulling} />
			</ProgressPopover>
		{:else}
			<ArcaneButton action="pull" onclick={() => handlePull()} loading={isLoading.pulling} />
		{/if}

		<ArcaneButton
			customLabel={type === 'stack' ? 'Destroy' : 'Remove'}
			action="remove"
			onclick={() => confirmAction('remove')}
			loading={isLoading.remove}
		/>
	{/if}
</div>
