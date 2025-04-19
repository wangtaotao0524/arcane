<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Plus, AlertCircle } from "@lucide/svelte";
  import DataTable from "$lib/components/data-table.svelte";
  import { columns } from "./columns";
  import type { PageData } from "./$types";
  import * as Alert from "$lib/components/ui/alert/index.js";

  let { data }: { data: PageData } = $props();
  const { volumes, error } = data;

  function createVolume() {
    // TODO: Implement create volume functionality
    alert("Implement create volume functionality");
  }
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Volumes</h1>
  <Button variant="default" onclick={createVolume}>
    <Plus class="w-4 h-4 mr-2" />
    Create Volume
  </Button>
</div>

{#if error}
  <Alert.Root variant="destructive" class="mb-6">
    <AlertCircle class="h-4 w-4" />
    <Alert.Title>Error Loading Volumes</Alert.Title>
    <Alert.Description>{error}</Alert.Description>
  </Alert.Root>
{/if}

<Card.Root>
  <Card.Header>
    <Card.Title>Docker Volumes</Card.Title>
    <Card.Description>Manage persistent data storage</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if volumes && volumes.length > 0}
      <DataTable data={volumes} {columns} />
    {:else if !error}
      <p class="text-center text-muted-foreground py-4">No volumes found.</p>
    {/if}
  </Card.Content>
</Card.Root>
