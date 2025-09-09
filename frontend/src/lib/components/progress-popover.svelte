<script lang="ts">
	import { cn } from '$lib/utils.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Popover as PopoverPrimitive } from 'bits-ui';
	import { Progress } from '$lib/components/ui/progress/index.js';
	import DownloadIcon from '@lucide/svelte/icons/download';
	import type { Icon as IconType } from '@lucide/svelte';
	import { type Snippet } from 'svelte';
	import { m } from '$lib/paraglide/messages';

	let {
		open = $bindable(false),
		title = m.progress_title(),
		subtitle = m.progress_subtitle(),
		progress = $bindable(0),
		statusText = '',
		error = '',
		loading = false,
		align = 'center' as 'start' | 'center' | 'end',
		sideOffset = 4,
		class: className = '',
		icon,
		iconClass = 'text-primary size-4',
		preventCloseWhileLoading = true,
		children
	} = $props<{
		open?: boolean;
		title?: string;
		subtitle?: string;
		progress?: number;
		statusText?: string;
		error?: string;
		loading?: boolean;
		align?: 'start' | 'center' | 'end';
		sideOffset?: number;
		class?: string;
		icon?: typeof IconType;
		iconClass?: string;
		preventCloseWhileLoading?: boolean;
		children: Snippet;
	}>();

	function handleOpenChange(next: boolean) {
		if (preventCloseWhileLoading && !next && loading) {
			open = true;
			return;
		}
		open = next;
	}

	const percent = $derived(Math.round(progress));
	const displaySubtitle = $derived(
		error ? (subtitle ?? m.common_unknown()) : progress === 100 ? m.progress_pull_completed() : (subtitle ?? m.common_unknown())
	);
</script>

<Popover.Root bind:open onOpenChange={handleOpenChange}>
	<Popover.Trigger>
		{@render children()}
	</Popover.Trigger>

	<Popover.Content class={cn('bg-popover w-80 rounded-md border p-4 shadow-md', className)} {align} {sideOffset}>
		<div class="space-y-3">
			<div class="flex items-center gap-3">
				<div class="bg-primary/10 flex size-8 shrink-0 items-center justify-center rounded-full">
					{#if icon}
						{@const Icon = icon}
						<Icon class={iconClass} />
					{:else}
						<DownloadIcon class={iconClass} />
					{/if}
				</div>
				<div>
					<h4 class="text-sm font-semibold">{title ?? m.common_unknown()}</h4>
					<p class="text-muted-foreground text-xs">{displaySubtitle}</p>
				</div>
			</div>

			{#if error}
				<div class="rounded-md bg-red-50 p-3 dark:bg-red-950/20">
					<p class="text-destructive text-sm">{error}</p>
				</div>
			{:else}
				<div class="space-y-2">
					<div class="flex justify-between text-xs">
						<span class="text-muted-foreground truncate pr-2">{statusText || m.progress_preparing()}</span>
						<span class="text-muted-foreground shrink-0">{percent}%</span>
					</div>
					<Progress value={progress} max={100} class="h-2 w-full" />
					{#if loading}
						<p class="text-muted-foreground text-xs">{m.progress_in_progress_note()}</p>
					{/if}
				</div>
			{/if}

			{#if !loading && progress === 100 && !error}
				<div class="rounded-md bg-green-50 p-3 dark:bg-green-950/20">
					<p class="text-sm text-green-600 dark:text-green-400">{m.progress_completed_success()}</p>
				</div>
			{/if}
		</div>

		<PopoverPrimitive.Arrow class="fill-background stroke-border" />
	</Popover.Content>
</Popover.Root>
