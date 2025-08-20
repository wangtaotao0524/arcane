<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import {
		CircleCheck,
		CircleFadingArrowUp,
		CircleArrowUp,
		Loader2,
		AlertTriangle
	} from '@lucide/svelte';
	import { imageUpdateAPI } from '$lib/services/api';
	import { toast } from 'svelte-sonner';
	import type { ImageUpdateData } from '$lib/services/api/image-update-api-service';

	interface Props {
		updateInfo?: ImageUpdateData | undefined;
		isLoadingInBackground?: boolean;
		imageId: string;
		repo?: string;
		tag?: string;
		onUpdated?: (data: ImageUpdateData) => void;
	}

	let {
		updateInfo = undefined,
		isLoadingInBackground = false,
		imageId,
		repo,
		tag,
		onUpdated
	}: Props = $props();

	let localUpdateInfo = $state<ImageUpdateData | undefined>(updateInfo);
	let isChecking = $state(false);

	$effect(() => {
		localUpdateInfo = updateInfo;
	});

	const canCheckUpdate = $derived(repo && tag && repo !== '<none>' && tag !== '<none>');

	const displayCurrentVersion = $derived(() => {
		if (localUpdateInfo?.currentVersion && localUpdateInfo.currentVersion.trim() !== '') {
			return localUpdateInfo.currentVersion;
		}
		return tag || 'Unknown';
	});

	const displayLatestVersion = $derived(() => {
		if (localUpdateInfo?.latestVersion && localUpdateInfo.latestVersion.trim() !== '') {
			return localUpdateInfo.latestVersion;
		}
		if (localUpdateInfo?.updateType === 'digest' && localUpdateInfo?.latestDigest) {
			return localUpdateInfo.latestDigest.slice(7, 19) + '...';
		}
		return null;
	});

	async function checkImageUpdate() {
		if (!canCheckUpdate || isChecking) return;

		isChecking = true;
		try {
			const result = await imageUpdateAPI.checkImageUpdateByID(imageId);
			if (result && !result.error) {
				localUpdateInfo = result;
				onUpdated?.(result); // optional: lets a parent sync if desired
				toast.success('Update check completed');
			} else {
				toast.error(result?.error || 'Update check failed');
			}
		} catch (error) {
			console.error('Error checking update:', error);
			toast.error('Failed to check for updates');
		} finally {
			isChecking = false;
		}
	}

	function getUpdatePriority(u: ImageUpdateData): {
		level: string;
		color: string;
		description: string;
	} {
		if (!u.hasUpdate)
			return { level: 'None', color: 'text-green-500', description: 'Image is up to date' };
		if (u.updateType === 'digest')
			return {
				level: 'Security/Bug Fix',
				color: 'text-blue-500',
				description: 'Digest update - likely security or bug fixes'
			};
		if (u.updateType === 'tag') {
			let description = 'Update available';
			if (u.latestVersion) description = `Update to ${u.latestVersion} available`;
			return { level: 'Version Update', color: 'text-yellow-500', description };
		}
		return { level: 'Unknown', color: 'text-gray-500', description: 'Update type unknown' };
	}
</script>

{#if localUpdateInfo}
	{@const priority = getUpdatePriority(localUpdateInfo)}
	<Tooltip.Provider>
		<Tooltip.Root>
			<Tooltip.Trigger>
				<span class="mr-2 inline-flex size-4 items-center justify-center align-middle">
					{#if !localUpdateInfo.hasUpdate}
						<CircleCheck class="size-4 text-green-500" />
					{:else if localUpdateInfo.updateType === 'digest'}
						<CircleArrowUp class="size-4 text-blue-500" />
					{:else}
						<CircleFadingArrowUp class="size-4 text-yellow-500" />
					{/if}
				</span>
			</Tooltip.Trigger>
			<Tooltip.Content
				side="right"
				class="tooltip-with-arrow relative max-w-[280px] rounded-xl border border-gray-200/50 bg-white/95 p-0 shadow-2xl shadow-black/10 backdrop-blur-sm dark:border-gray-800/50 dark:bg-gray-950/95 dark:shadow-black/30"
				align="center"
			>
				<div class="overflow-hidden rounded-xl">
					{#if !localUpdateInfo.hasUpdate}
						<div class="p-3 text-xs">Running {displayCurrentVersion()}</div>
					{:else if localUpdateInfo.updateType === 'digest'}
						<div class="p-3 text-xs">
							<div>Current: {displayCurrentVersion()}</div>
							{#if displayLatestVersion()}<div>Latest digest: {displayLatestVersion()}</div>{/if}
							<div class="mt-1 text-blue-600 dark:text-blue-300">{priority.description}</div>
						</div>
					{:else}
						<div class="p-3 text-xs">
							<div>Current: {displayCurrentVersion()}</div>
							{#if displayLatestVersion()}<div>Latest: {displayLatestVersion()}</div>{/if}
							<div class="mt-1 text-amber-600 dark:text-amber-300">{priority.description}</div>
						</div>
					{/if}

					{#if canCheckUpdate}
						<div
							class="border-t border-gray-200/50 bg-gray-50/50 p-3 dark:border-gray-800/50 dark:bg-gray-900/50"
						>
							<button
								onclick={checkImageUpdate}
								disabled={isChecking}
								class="group flex w-full items-center justify-center gap-2 rounded-lg bg-white/80 px-3 py-2 text-xs font-medium text-gray-700 shadow-sm transition-all hover:bg-white hover:shadow-md disabled:cursor-not-allowed disabled:opacity-50 dark:bg-gray-800/80 dark:text-gray-300 dark:hover:bg-gray-800"
							>
								{#if isChecking}
									<Loader2 class="size-3 animate-spin" />
									Checking...
								{:else}
									Re-check Updates
								{/if}
							</button>
						</div>
					{/if}
				</div>
			</Tooltip.Content>
		</Tooltip.Root>
	</Tooltip.Provider>
{:else if isLoadingInBackground || isChecking}
	<Tooltip.Provider>
		<Tooltip.Root>
			<Tooltip.Trigger>
				<span class="mr-2 inline-flex size-4 items-center justify-center align-middle">
					<Loader2 class="size-4 animate-spin text-blue-400" />
				</span>
			</Tooltip.Trigger>
		</Tooltip.Root>
	</Tooltip.Provider>
{:else}
	<Tooltip.Provider>
		<Tooltip.Root>
			<Tooltip.Trigger>
				<span class="mr-2 inline-flex size-4 items-center justify-center">
					{#if canCheckUpdate}
						<button
							onclick={checkImageUpdate}
							disabled={isChecking}
							class="group flex h-4 w-4 items-center justify-center rounded-full border-2 border-dashed border-gray-400 transition-colors hover:border-blue-400 hover:bg-blue-50 disabled:cursor-not-allowed dark:hover:bg-blue-950"
						>
							{#if isChecking}
								<Loader2 class="h-2 w-2 animate-spin text-blue-400" />
							{:else}
								<div
									class="h-1.5 w-1.5 rounded-full bg-gray-400 transition-colors group-hover:bg-blue-400"
								></div>
							{/if}
						</button>
					{:else}
						<div
							class="flex h-4 w-4 items-center justify-center rounded-full border-2 border-dashed border-gray-400 opacity-30"
						>
							<div class="h-1.5 w-1.5 rounded-full bg-gray-400"></div>
						</div>
					{/if}
				</span>
			</Tooltip.Trigger>
			<Tooltip.Content
				side="right"
				class="tooltip-with-arrow relative max-w-[240px] rounded-xl border border-gray-200/50 bg-white/95 p-0 shadow-2xl shadow-black/10 backdrop-blur-sm dark:border-gray-800/50 dark:bg-gray-950/95 dark:shadow-black/30"
				align="center"
			>
				<div class="overflow-hidden rounded-xl">
					<div class="p-3 text-xs flex items-center gap-2">
						<AlertTriangle class="size-4" />
						Status unknown. Click to check for updates.
					</div>
				</div>
			</Tooltip.Content>
		</Tooltip.Root>
	</Tooltip.Provider>
{/if}
