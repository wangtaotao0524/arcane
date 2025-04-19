<script lang="ts">
  import type { PageData } from "./$types";
  import DataTable from "$lib/components/data-table.svelte";
  import { columns } from "./columns"; // Import the column definitions
  import { Button } from "$lib/components/ui/button/index.js";
  import { Download, AlertCircle } from "@lucide/svelte";
  import * as Alert from "$lib/components/ui/alert/index.js";

  let { data }: { data: PageData } = $props();
  const { images, error } = data;

  function pullImage() {
    // TODO: Implement pull image modal/logic
    alert("Implement pull image functionality");
  }
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Docker Images</h1>
  <Button variant="default" onclick={pullImage}>
    <Download class="w-4 h-4 mr-2" />
    Pull Image
  </Button>
</div>

{#if error}
  <Alert.Root variant="destructive" class="mb-6">
    <AlertCircle class="h-4 w-4" />
    <Alert.Title>Error Loading Images</Alert.Title>
    <Alert.Description>{error}</Alert.Description>
  </Alert.Root>
{/if}

{#if images}
  <DataTable data={images} {columns} />
{:else if !error}
  <!-- Show loading or no data state if needed, DataTable might handle empty state -->
  <p class="text-center text-muted-foreground">
    Loading images or no images found.
  </p>
{/if}
