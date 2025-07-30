<script lang="ts">
	import Meter from '$lib/components/ui/meter/meter.svelte';
	import type { Icon as IconType } from '@lucide/svelte';
	import { Loader2 } from '@lucide/svelte';

	interface Props {
		title: string;
		description?: string;
		currentValue?: number;
		unit?: string;
		formatValue?: (value: number) => string;
		maxValue?: number;
		icon: typeof IconType;
		loading?: boolean;
	}

	let {
		title,
		description,
		currentValue,
		unit = '',
		formatValue = (v) => `${v.toFixed(1)}${unit}`,
		maxValue = 100,
		icon,
		loading = false
	}: Props = $props();

	const percentage = $derived(
		currentValue !== undefined && !loading ? (currentValue / maxValue) * 100 : 0
	);
	const Icon = icon;
</script>

<div
	class="group relative overflow-hidden rounded-xl border border-gray-200/50 bg-white/95 shadow-2xl shadow-black/10 backdrop-blur-sm transition-all duration-300 hover:border-gray-300/60 hover:shadow-3xl hover:shadow-black/15 dark:border-gray-800/50 dark:bg-gray-950/95 dark:shadow-black/30 dark:hover:border-gray-700/60"
>
	<!-- Header Section with Gradient -->
	<div
		class="bg-gradient-to-br from-gray-50 to-slate-50/30 p-4 dark:from-gray-900/20 dark:to-slate-900/10"
	>
		<div class="flex items-center gap-3">
			<div
				class="flex h-10 w-10 items-center justify-center rounded-full bg-gradient-to-br from-primary to-primary/80 shadow-lg shadow-primary/25"
			>
				{#if loading}
					<Loader2 class="size-5 animate-spin text-white" />
				{:else}
					<Icon class="size-5 text-white" />
				{/if}
			</div>
			<div class="flex-1 min-w-0">
				<div class="text-sm font-semibold text-gray-900 dark:text-gray-100">{title}</div>
				{#if description}
					<div class="text-xs text-gray-700/80 dark:text-gray-300/80">{description}</div>
				{/if}
			</div>
		</div>
	</div>

	<!-- Content Section -->
	<div class="bg-white/90 p-4 dark:bg-gray-950/90">
		<div class="space-y-3">
			<div class="text-center">
				{#if loading}
					<div class="h-8 w-16 bg-muted animate-pulse rounded mx-auto"></div>
				{:else}
					<div class="text-2xl font-bold text-gray-900 dark:text-gray-100">
						{currentValue !== undefined ? formatValue(currentValue) : '--'}
					</div>
				{/if}
			</div>

			<div class="flex justify-center">
				<div class="w-full max-w-[120px]">
					{#if loading}
						<div class="h-2 w-full bg-muted animate-pulse rounded"></div>
					{:else}
						<Meter value={percentage} class="h-2 bg-gray-200 dark:bg-gray-800" />
					{/if}
				</div>
			</div>

			<div class="rounded-lg bg-primary/10 p-3 dark:bg-primary/20">
				{#if loading}
					<div class="h-3 w-20 bg-muted animate-pulse rounded mx-auto"></div>
				{:else}
					<div
						class="text-xs text-center font-medium leading-relaxed text-primary dark:text-primary"
					>
						{percentage.toFixed(1)}% of resources being used
					</div>
				{/if}
			</div>
		</div>
	</div>
</div>
