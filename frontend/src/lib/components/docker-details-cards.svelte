<script lang="ts">
	import GradientFrameCard from '$lib/components/gradient-frame-card.svelte';
	import BoxIcon from '@lucide/svelte/icons/box';
	import HardDriveIcon from '@lucide/svelte/icons/hard-drive';
	import DockerIcon from '$lib/icons/docker-icon.svelte';
	import type { Component } from 'svelte';
	import type { IconProps } from '@lucide/svelte';
	import { m } from '$lib/paraglide/messages';

	const DockerIconTyped = DockerIcon as unknown as Component<IconProps, {}, ''>;

	type DockerInfo =
		| {
				version?: string;
				os?: string;
				architecture?: string;
		  }
		| null
		| undefined;

	let {
		title = m.dashboard_docker_details_title(),
		isLoadingDockerInfo = false,
		isLoadingStats = false,
		isLoadingImages = false,
		dockerInfo,
		totalContainers,
		stoppedContainers,
		containersRunning,
		imagesTotal,
		class: className
	}: {
		title?: string;
		isLoadingDockerInfo?: boolean;
		isLoadingStats?: boolean;
		isLoadingImages?: boolean;
		dockerInfo: DockerInfo;
		totalContainers: number;
		stoppedContainers: number;
		containersRunning: number;
		imagesTotal: number;
		class?: string;
	} = $props();
</script>

<section class={className}>
	<h2 class="mb-4 text-lg font-semibold tracking-tight">{title}</h2>

	<div class="grid grid-cols-1 gap-4 md:grid-cols-3">
		<GradientFrameCard
			class="w-full"
			title={m.docker_engine_title()}
			icon={DockerIconTyped}
			color="blue"
			loading={isLoadingDockerInfo}
			rightBadge={dockerInfo?.version ?? null}
			subtitle={!isLoadingDockerInfo
				? m.docker_engine_subtitle({
						os: dockerInfo?.os ?? m.common_unknown(),
						arch: dockerInfo?.architecture ?? m.common_unknown()
					})
				: null}
		/>

		<GradientFrameCard
			class="w-full"
			title={m.containers_title()}
			icon={BoxIcon}
			color="emerald"
			loading={isLoadingStats}
			rightBadge={!isLoadingStats ? m.docker_containers_running_badge({ count: containersRunning }) : null}
		>
			{#if !isLoadingStats}
				<div class="text-muted-foreground mt-1 flex flex-wrap items-center gap-2 text-xs">
					<span class="bg-muted rounded-md px-1.5 py-0.5">{m.docker_containers_total_badge({ count: totalContainers })}</span>
					<span class="bg-muted rounded-md px-1.5 py-0.5">{m.docker_containers_stopped_badge({ count: stoppedContainers })}</span>
				</div>
			{/if}
		</GradientFrameCard>

		<GradientFrameCard
			class="w-full"
			title={m.images_title()}
			icon={HardDriveIcon}
			color="purple"
			loading={isLoadingImages}
			rightBadge={!isLoadingImages ? imagesTotal : null}
			subtitle={!isLoadingImages ? m.docker_images_subtitle() : null}
		/>
	</div>
</section>
