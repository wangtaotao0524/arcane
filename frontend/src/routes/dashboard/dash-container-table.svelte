<script lang="ts">
	import ArcaneTable from '$lib/components/arcane-table.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ArrowRight, Box, Loader2 } from '@lucide/svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table';
	import StatusBadge from '$lib/components/badges/status-badge.svelte';
	import { statusVariantMap } from '$lib/types/statuses';
	import { capitalizeFirstLetter, truncateString } from '$lib/utils/string.utils';
	import type { SearchPaginationSortRequest, Paginated } from '$lib/types/pagination.type';
	import type { ContainerSummaryDto } from '$lib/types/container.type';
	import { environmentAPI } from '$lib/services/api';

	let {
		containers = $bindable(),
		isLoading
	}: {
		containers: Paginated<ContainerSummaryDto>;
		isLoading: boolean;
	} = $props();

	let selectedIds = $state<string[]>([]);

	let requestOptions = $state<SearchPaginationSortRequest>({
		pagination: { page: 1, limit: 5 },
		sort: { column: 'created', direction: 'desc' }
	});
</script>

<Card.Root class="relative flex flex-col rounded-lg border shadow-sm">
	<Card.Header class="px-6 pb-3 pt-5">
		<div class="flex items-center justify-between">
			<div>
				<Card.Title>
					<a class="font-medium hover:underline" href="/containers">Containers</a>
				</Card.Title>
				<Card.Description class="pb-2">Recent containers</Card.Description>
			</div>
			<Button variant="ghost" size="sm" href="/containers" disabled={isLoading}>
				View All
				<ArrowRight class="ml-2 size-4" />
			</Button>
		</div>
	</Card.Header>
	<Card.Content class="flex-1 p-0">
		{#if containers.data.length > 0}
			<div class="flex h-full flex-col">
				<ArcaneTable
					items={containers}
					bind:requestOptions
					bind:selectedIds
					onRefresh={async (options) => (containers = await environmentAPI.getContainers(options))}
					withoutSearch={true}
					withoutPagination={true}
					selectionDisabled={true}
					columns={[
						{ label: 'Name', sortColumn: 'name' },
						{ label: 'Image', sortColumn: 'image' },
						{ label: 'State', sortColumn: 'state' },
						{ label: 'Status', sortColumn: 'status' }
					]}
					filterPlaceholder="Search containers..."
					noResultsMessage="No containers found"
				>
					{#snippet rows({ item })}
						<Table.Cell class="py-3 md:py-3.5">
							<a class="font-medium hover:underline" href="/containers/{item.id}/">
								{#if item.names && item.names.length > 0}
									{item.names[0].startsWith('/') ? item.names[0].substring(1) : item.names[0]}
								{:else}
									{item.id.substring(0, 12)}
								{/if}
							</a>
						</Table.Cell>
						<Table.Cell>
							<span class="text-sm">{item.image}</span>
						</Table.Cell>
						<Table.Cell class="py-3 md:py-3.5">
							{@const stateVariant = statusVariantMap[item.state.toLowerCase()] || 'gray'}
							<StatusBadge variant={stateVariant} text={capitalizeFirstLetter(item.state)} />
						</Table.Cell>
						<Table.Cell class="py-3 md:py-3.5">{item.status}</Table.Cell>
					{/snippet}
				</ArcaneTable>
				{#if containers.data.length > 5}
					<div class="bg-muted/40 text-muted-foreground border-t px-6 py-2 text-xs">
						Showing 5 of {containers.pagination.totalItems} containers
					</div>
				{/if}
			</div>
		{:else}
			<div class="flex flex-col items-center justify-center px-6 py-10 text-center">
				<Box class="text-muted-foreground mb-2 size-8 opacity-40" />
				<p class="text-muted-foreground text-sm">No containers found</p>
				<p class="text-muted-foreground mt-1 text-xs">
					Use Docker CLI or another tool to create containers
				</p>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
