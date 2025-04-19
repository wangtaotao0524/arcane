<script lang="ts">
  import DataTable from "$lib/components/data-table.svelte";
  import Button from "$lib/components/ui/button/button.svelte";
  import * as Card from "$lib/components/ui/card/index.js";
  import { columns } from "./columns.js";
  import { Plus } from "@lucide/svelte";

  let { data } = $props();
  const { containers } = data;
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Containers</h1>
  <Button variant="default">
    <Plus class="w-4 h-4" />
    Create Container
  </Button>
</div>

<Card.Root>
  <Card.Header>
    <Card.Title>Container List</Card.Title>
    <Card.Description>Manage your Docker containers</Card.Description>
  </Card.Header>
  <Card.Content>
    <DataTable data={containers} {columns} />
  </Card.Content>
</Card.Root>
