<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Plus, AlertCircle } from "@lucide/svelte";
  import DataTable from "$lib/components/data-table.svelte";
  import { columns } from "./columns";
  import type { PageData } from "./$types";
  import * as Alert from "$lib/components/ui/alert/index.js";

  let { data }: { data: PageData } = $props();
  const { networks, error } = data;

  function createNetwork() {
    alert("Implement create network functionality");
  }
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Networks</h1>
  <Button variant="default" onclick={createNetwork}>
    <Plus class="w-4 h-4 mr-2" />
    Create Network
  </Button>
</div>

{#if error}
  <Alert.Root variant="destructive" class="mb-6">
    <AlertCircle class="h-4 w-4" />
    <Alert.Title>Error Loading Networks</Alert.Title>
    <Alert.Description>{error}</Alert.Description>
  </Alert.Root>
{/if}

<Card.Root>
  <Card.Header>
    <Card.Title>Docker Networks</Card.Title>
    <Card.Description>Manage container communication</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if networks && networks.length > 0}
      <DataTable data={networks} {columns} />
    {:else if !error}
      <p class="text-center text-muted-foreground py-4">No networks found.</p>
    {/if}
  </Card.Content>
</Card.Root>
