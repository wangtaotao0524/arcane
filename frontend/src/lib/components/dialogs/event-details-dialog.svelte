<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import InfoIcon from '@lucide/svelte/icons/info';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
	import XCircleIcon from '@lucide/svelte/icons/x-circle';
	import UserIcon from '@lucide/svelte/icons/user';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import ServerIcon from '@lucide/svelte/icons/server';
	import TagIcon from '@lucide/svelte/icons/tag';
	import type { Event } from '$lib/types/event.type';
	import { toast } from 'svelte-sonner';

	let {
		open = $bindable(),
		event
	}: {
		open: boolean;
		event: Event | null;
	} = $props();

	function formatDate(ts?: string) {
		if (!ts) return '-';
		try {
			return new Date(ts).toLocaleString();
		} catch {
			return ts;
		}
	}

	async function copy(text?: string) {
		if (!text) return;
		try {
			await navigator.clipboard.writeText(text);
			toggleToast('Copied to clipboard');
		} catch {
			toggleToast('Failed to copy', true);
		}
	}
	function toggleToast(msg: string, isErr = false) {
		isErr ? toast.error(msg) : toast.success(msg);
	}

	const hasMetadata = $derived(!!event?.metadata && Object.keys(event.metadata ?? {}).length > 0);
	const jsonPretty = $derived(() => JSON.stringify(event ?? {}, null, 2));
	const metadataPretty = $derived(() => JSON.stringify(event?.metadata ?? {}, null, 2));

	function sevColor(sev?: string) {
		switch (sev) {
			case 'success':
				return 'text-emerald-600 dark:text-emerald-400';
			case 'warning':
				return 'text-amber-600 dark:text-amber-400';
			case 'error':
				return 'text-red-600 dark:text-red-400';
			default:
				return 'text-blue-600 dark:text-blue-400';
		}
	}
	function sevBadge(sev?: string) {
		switch (sev) {
			case 'success':
				return 'bg-emerald-500/15 text-emerald-600 border-emerald-500/30 dark:text-emerald-300';
			case 'warning':
				return 'bg-amber-500/15 text-amber-700 border-amber-500/30 dark:text-amber-300';
			case 'error':
				return 'bg-red-500/15 text-red-600 border-red-500/30 dark:text-red-300';
			default:
				return 'bg-blue-500/15 text-blue-700 border-blue-500/30 dark:text-blue-300';
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="bg-card flex max-h-[90vh] flex-col rounded-xl border shadow-xl sm:max-w-[980px]">
		<!-- Header -->
		<Dialog.Header class="flex-shrink-0 border-b p-4">
			<div class="flex items-start gap-3">
				<div class="mt-0.5">
					{#if event?.severity === 'success'}
						<CheckCircle2Icon class={sevColor(event?.severity) + ' size-6'} />
					{:else if event?.severity === 'warning'}
						<TriangleAlertIcon class={sevColor(event?.severity) + ' size-6'} />
					{:else if event?.severity === 'error'}
						<XCircleIcon class={sevColor(event?.severity) + ' size-6'} />
					{:else}
						<InfoIcon class={sevColor(event?.severity) + ' size-6'} />
					{/if}
				</div>
				<div class="min-w-0 flex-1">
					<Dialog.Title class="text-xl font-semibold leading-tight">
						{event?.title || 'Event Details'}
					</Dialog.Title>
					{#if event?.description}
						<Dialog.Description class="text-muted-foreground mt-1 text-sm">
							{event.description}
						</Dialog.Description>
					{/if}
					<div class="mt-3 flex flex-wrap items-center gap-2">
						<Badge class={'border ' + sevBadge(event?.severity)}>
							{event?.severity ?? 'info'}
						</Badge>
						<Badge variant="outline" class="gap-1">
							<TagIcon class="size-3" />
							{event?.type ?? '-'}
						</Badge>
						{#if event?.environmentId}
							<Badge variant="outline" class="gap-1">
								<ServerIcon class="size-3" />
								env: {event.environmentId}
							</Badge>
						{/if}
						{#if event?.timestamp}
							<span class="text-muted-foreground inline-flex items-center gap-1 text-xs">
								<ClockIcon class="size-3" />
								{formatDate(event.timestamp)}
							</span>
						{/if}
					</div>
				</div>
			</div>
		</Dialog.Header>

		<div class="min-h-0 flex-1 space-y-4 overflow-y-auto p-4">
			<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				<div class="rounded-lg border p-3">
					<div class="text-muted-foreground text-xs">Resource ID</div>
					<div class="mt-1 flex items-center justify-between gap-2">
						<div class="break-all text-sm">{event?.resourceType || '-'}</div>
						<Button variant="ghost" size="icon" class="size-7" onclick={() => copy(event?.resourceType)} title="Copy resource id">
							<CopyIcon class="size-4" />
						</Button>
					</div>
				</div>

				<div class="rounded-lg border p-3">
					<div class="text-muted-foreground text-xs">Resource Name</div>
					<div class="mt-1 flex items-center justify-between gap-2">
						<div class="break-all text-sm">{event?.resourceName || '-'}</div>
						<Button
							variant="ghost"
							size="icon"
							class="size-7"
							onclick={() => copy(event?.resourceName)}
							title="Copy resource name"
						>
							<CopyIcon class="size-4" />
						</Button>
					</div>
				</div>

				<div class="rounded-lg border p-3">
					<div class="text-muted-foreground text-xs">User</div>
					<div class="mt-1 flex items-center gap-2 text-sm">
						<UserIcon class="text-muted-foreground size-4" />
						{event?.username || '-'}
					</div>
				</div>
			</div>

			<div class="rounded-lg border">
				<div class="flex items-center justify-between border-b px-3 py-2">
					<h3 class="text-sm font-medium">Metadata</h3>
					<Button
						variant="outline"
						size="sm"
						onclick={() => copy(metadataPretty())}
						disabled={!hasMetadata}
						title="Copy metadata JSON"
					>
						<CopyIcon class="mr-2 size-3" />
						Copy JSON
					</Button>
				</div>
				{#if hasMetadata}
					<pre class="bg-muted/40 max-h-[40vh] overflow-auto p-3 text-xs leading-relaxed">
<code class="font-mono">{metadataPretty()}</code>
</pre>
				{:else}
					<div class="text-muted-foreground p-3 text-xs">No metadata provided.</div>
				{/if}
			</div>

			<div class="rounded-lg border">
				<div class="flex items-center justify-between border-b px-3 py-2">
					<h3 class="text-sm font-medium">Raw Event</h3>
					<Button variant="outline" size="sm" onclick={() => copy(jsonPretty())} title="Copy full event JSON">
						<CopyIcon class="mr-2 size-3" />
						Copy JSON
					</Button>
				</div>
				<pre class="bg-muted/40 max-h-[40vh] overflow-auto p-3 text-xs leading-relaxed">
<code class="font-mono">{jsonPretty()}</code>
</pre>
			</div>
		</div>

		<Dialog.Footer class="flex flex-shrink-0 items-center justify-end gap-2 border-t p-3">
			<Button variant="outline" onclick={() => (open = false)}>Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<style>
	/* Lock background scroll while dialog is open; no animations/effects */
	:global(html:has([data-slot='dialog-overlay'][data-state='open'])) {
		overflow: hidden;
	}
	:global(body:has([data-slot='dialog-overlay'][data-state='open'])) {
		overflow: hidden;
	}
</style>
