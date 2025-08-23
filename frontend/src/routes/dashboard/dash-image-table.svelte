<script lang="ts">
	import ArcaneTable from '$lib/components/arcane-table.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ArrowRight, HardDrive, Loader2 } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table';
	import StatusBadge from '$lib/components/badges/status-badge.svelte';
	import type { SearchPaginationSortRequest, Paginated } from '$lib/types/pagination.type';
	import type { ImageSummaryDto } from '$lib/types/image.type';
	import { environmentAPI } from '$lib/services/api';
	import { formatBytes } from '$lib/utils/bytes.util';

	let {
		images = $bindable(),
		isLoading
	}: {
		images: Paginated<ImageSummaryDto>;
		isLoading: boolean;
	} = $props();

	let requestOptions = $state<SearchPaginationSortRequest>({
		pagination: { page: 1, limit: 5 },
		sort: { column: 'size', direction: 'desc' }
	});

	let selectedIds = $state<string[]>([]);

	// Size display removed on dashboard; keep sorting by size to show largest but do not render sizes
</script>

<Card.Root class="relative flex flex-col rounded-lg border shadow-sm">
	<Card.Header class="px-6 pb-3 pt-5">
		<div class="flex items-center justify-between">
			<div>
				<Card.Title>
					<a class="font-medium hover:underline" href="/images">Images</a>
				</Card.Title>
				<Card.Description class="pb-2">Top 5 Largest Images</Card.Description>
			</div>
			<div class="flex items-center gap-3">
				<Button variant="ghost" size="sm" href="/images" disabled={isLoading}>
					View All
					<ArrowRight class="ml-2 size-4" />
				</Button>
			</div>
		</div>
	</Card.Header>
	<Card.Content class="flex-1 p-0">
		{#if images.data.length > 0}
			<div class="flex h-full flex-col">
				<ArcaneTable
					items={images}
					bind:requestOptions
					bind:selectedIds
					onRefresh={async (options) => (images = await environmentAPI.getImages(options))}
					withoutSearch={true}
					selectionDisabled={true}
					withoutPagination={true}
					columns={[
						{ label: 'Name', sortColumn: 'repoTags' },
						{ label: 'Status', sortColumn: 'inUse' },
						{ label: 'Tag', sortColumn: 'tag' },
						{ label: 'Size', sortColumn: 'size' }
					]}
					filterPlaceholder="Search images..."
					noResultsMessage="No images found"
				>
					{#snippet rows({ item })}
						<Table.Cell class="py-3 md:py-3.5">
							<div class="flex items-center gap-2">
								<div class="flex flex-1 items-center">
									<a class="shrink truncate font-medium hover:underline" href="/images/{item.id}/">
										{#if item.repoTags && item.repoTags.length > 0 && item.repoTags[0] !== '<none>:<none>'}
											{item.repoTags[0].split(':')[0]}
										{:else}
											<span class="text-muted-foreground italic">Untagged</span>
										{/if}
									</a>
								</div>
							</div>
						</Table.Cell>
						<Table.Cell class="py-3 md:py-3.5">
							{#if item.inUse}
								<StatusBadge text="In Use" variant="green" />
							{:else}
								<StatusBadge text="Unused" variant="amber" />
							{/if}
						</Table.Cell>
						<Table.Cell class="py-3 md:py-3.5">
							{#if item.repoTags && item.repoTags.length > 0 && item.repoTags[0] !== '<none>:<none>'}
								{item.repoTags[0].split(':')[1] || 'latest'}
							{:else}
								<span class="text-muted-foreground italic">&lt;none&gt;</span>
							{/if}
						</Table.Cell>
						<Table.Cell class="py-3 md:py-3.5">{formatBytes(item.size)}</Table.Cell>
					{/snippet}
				</ArcaneTable>
				{#if images.data.length > 5}
					<div class="bg-muted/40 text-muted-foreground border-t px-6 py-2 text-xs">
						Showing 5 of {images.pagination.totalItems} images
					</div>
				{/if}
			</div>
		{:else}
			<div class="flex flex-col items-center justify-center px-6 py-10 text-center">
				<HardDrive class="text-muted-foreground mb-2 size-8 opacity-40" />
				<p class="text-muted-foreground text-sm">No images found</p>
				<p class="text-muted-foreground mt-1 text-xs">
					Pull images using Docker CLI or another tool
				</p>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
