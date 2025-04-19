<script lang="ts" generics="TData, TValue">
  import {
    type ColumnDef,
    getCoreRowModel,
    getPaginationRowModel, // Import pagination row model
  } from "@tanstack/table-core";
  import {
    createSvelteTable,
    FlexRender,
  } from "$lib/components/ui/data-table/index.js";
  import * as Table from "$lib/components/ui/table/index.js";
  import * as Pagination from "$lib/components/ui/pagination/index.js"; // Import Pagination components

  type DataTableProps<TData, TValue> = {
    columns: ColumnDef<TData, TValue>[];
    data: TData[];
    pageSize?: number;
  };

  let {
    data,
    columns,
    pageSize = 10,
  }: DataTableProps<TData, TValue> = $props();

  // Pagination state
  let pageIndex = $state(0);
  let currentPage = $state(1); // Use 1-based index for Pagination component

  // Sync state between table (0-based) and pagination component (1-based)
  $effect(() => {
    // When pagination component page changes, update table state
    pageIndex = currentPage - 1;
  });

  $effect(() => {
    // When table page index changes (e.g., due to filtering reducing pages), update pagination component
    const tablePageIndex = table.getState().pagination.pageIndex;
    if (tablePageIndex !== pageIndex) {
      currentPage = tablePageIndex + 1;
    }
  });

  const table = createSvelteTable({
    get data() {
      return data;
    },
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(), // Enable pagination
    state: {
      // Manage pagination state within the table instance
      get pagination() {
        return {
          pageIndex,
          pageSize,
        };
      },
    },
  });

  const pageCount = $derived(table.getPageCount());
  const rowCount = $derived(table.getRowCount());
</script>

<div class="space-y-4">
  <div class="rounded-md border">
    <Table.Root>
      <Table.Header>
        {#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
          <Table.Row>
            {#each headerGroup.headers as header (header.id)}
              <Table.Head>
                {#if !header.isPlaceholder}
                  <FlexRender
                    content={header.column.columnDef.header}
                    context={header.getContext()}
                  />
                {/if}
              </Table.Head>
            {/each}
          </Table.Row>
        {/each}
      </Table.Header>
      <Table.Body>
        {#each table.getRowModel().rows as row (row.id)}
          <Table.Row data-state={row.getIsSelected() && "selected"}>
            {#each row.getVisibleCells() as cell (cell.id)}
              <Table.Cell>
                <FlexRender
                  content={cell.column.columnDef.cell}
                  context={cell.getContext()}
                />
              </Table.Cell>
            {/each}
          </Table.Row>
        {:else}
          <Table.Row>
            <Table.Cell colspan={columns.length} class="h-24 text-center">
              No results.
            </Table.Cell>
          </Table.Row>
        {/each}
      </Table.Body>
    </Table.Root>
  </div>

  <!-- Add Pagination Component -->
  {#if pageCount > 1}
    <Pagination.Root
      count={rowCount}
      perPage={pageSize}
      bind:page={currentPage}
    >
      {#snippet children({ pages, currentPage })}
        <Pagination.Content>
          <Pagination.Item>
            <Pagination.PrevButton />
          </Pagination.Item>
          {#each pages as page (page.key)}
            {#if page.type === "ellipsis"}
              <Pagination.Item>
                <Pagination.Ellipsis />
              </Pagination.Item>
            {:else}
              <Pagination.Item>
                <Pagination.Link {page} isActive={currentPage === page.value}>
                  {page.value}
                </Pagination.Link>
              </Pagination.Item>
            {/if}
          {/each}
          <Pagination.Item>
            <Pagination.NextButton />
          </Pagination.Item>
        </Pagination.Content>
      {/snippet}
    </Pagination.Root>
  {/if}
</div>
