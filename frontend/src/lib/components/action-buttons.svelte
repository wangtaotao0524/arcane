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
	import { m } from '$lib/paraglide/messages';

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
				title: type === 'stack' ? m.compose_destroy() : m.action_confirm_removal_title(),
				message:
					type === 'stack'
						? m.action_confirm_destroy_message({ type: m.common_project() })
						: m.action_confirm_removal_message({ type: m.common_container() }),
				confirm: {
					label: type === 'stack' ? m.compose_destroy() : m.common_remove(),
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
							message: m.action_failed_generic({
								action: type === 'stack' ? m.compose_destroy() : m.common_remove(),
								type: type
							}),
							setLoadingState: (value) => (isLoading.remove = value),
							onSuccess: async () => {
								toast.success(
									type === 'stack'
										? m.action_destroyed_success({ type: m.common_stack() })
										: m.action_removed_success({ type: m.common_container() })
								);
								await invalidateAll();
								goto(`${type === 'stack' ? '/compose' : 'containers'}`);
							}
						});
					}
				},
				checkboxes: [
					{ id: 'removeFiles', label: m.confirm_remove_project_files(), initialState: false },
					{
						id: 'removeVolumes',
						label: m.confirm_remove_volumes_warning(),
						initialState: false
					}
				]
			});
		} else if (action === 'redeploy') {
			openConfirmDialog({
				title: m.action_confirm_redeploy_title(),
				message: m.action_confirm_redeploy_message(),
				confirm: {
					label: m.action_redeploy(),
					action: async () => {
						isLoading.redeploy = true;
						handleApiResultWithCallbacks({
							result: await tryCatch(environmentAPI.redeployProject(id)),
							message: m.action_failed_generic({ action: m.action_redeploy(), type }),
							setLoadingState: (value) => (isLoading.redeploy = value),
							onSuccess: async () => {
								toast.success(m.action_redeploy_success({ type: type }));
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
			message: m.action_failed_generic({ action: m.common_start(), type }),
			setLoadingState: (value) => (isLoading.start = value),
			onSuccess: async () => {
				toast.success(m.action_started_success({ type }));
				await invalidateAll();
			}
		});
	}

	async function handleDeploy() {
		isLoading.start = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(environmentAPI.startProject(id)),
			message: m.action_failed_generic({ action: m.common_start(), type }),
			setLoadingState: (value) => (isLoading.start = value),
			onSuccess: async () => {
				toast.success(m.action_started_success({ type }));
				await invalidateAll();
			}
		});
	}

	async function handleStop() {
		isLoading.stop = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(type === 'container' ? environmentAPI.stopContainer(id) : environmentAPI.stopProject(id)),
			message: m.action_failed_generic({ action: m.common_stop(), type }),
			setLoadingState: (value) => (isLoading.stop = value),
			onSuccess: async () => {
				toast.success(m.action_stopped_success({ type }));
				await invalidateAll();
			}
		});
	}

	async function handleRestart() {
		isLoading.restart = true;
		handleApiResultWithCallbacks({
			result: await tryCatch(type === 'container' ? environmentAPI.restartContainer(id) : environmentAPI.restartProject(id)),
			message: m.action_failed_generic({ action: m.common_restart(), type }),
			setLoadingState: (value) => (isLoading.restart = value),
			onSuccess: async () => {
				toast.success(m.action_restarted_success({ type }));
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
				message: m.images_pull_failed(),
				setLoadingState: (value) => (isLoading.pulling = value),
				onSuccess: async () => {
					toast.success(m.images_pulled_success());
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
		pullStatusText = m.images_pull_initiating();

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
					error: m.images_pull_server_error()
				}));
				const errorMessage =
					typeof errorData.error === 'string'
						? errorData.error
						: errorData.message || `${m.images_pull_server_error()}: HTTP ${response.status}`;
				throw new Error(errorMessage);
			}

			const reader = response.body.getReader();
			const decoder = new TextDecoder();
			let buffer = '';

			while (true) {
				const { done, value } = await reader.read();
				if (done) {
					pullStatusText = m.images_pull_processing_final_layers();
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
							pullError = typeof data.error === 'string' ? data.error : data.error.message || m.images_pull_stream_error();
							pullStatusText = m.images_pull_failed_with_error({ error: pullError });
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
			pullStatusText = m.images_pulled_success();
			toast.success(m.images_pulled_success());
			await invalidateAll();

			setTimeout(() => {
				pullPopoverOpen = false;
				isLoading.pulling = false;
				resetPullState();
			}, 2000);
		} catch (error: any) {
			console.error('Pull images error:', error);
			const message = error.message || m.images_pull_failed();
			pullError = message;
			pullStatusText = m.images_pull_failed_with_error({ error: message });
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
			customLabel={m.action_down ? m.action_down() : 'Down'}
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
				title={m.progress_pulling_images()}
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
			customLabel={type === 'stack' ? m.compose_destroy() : m.common_remove()}
			action="remove"
			onclick={() => confirmAction('remove')}
			loading={isLoading.remove}
		/>
	{/if}
</div>
