<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Image, Trash2, Download } from "@lucide/svelte";
  import ImageRow from "$lib/components/docker/ImageRow.svelte";

  export let data;
  const { images } = data;
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Docker Images</h1>
  <Button variant="default">
    <Download class="w-4 h-4 mr-2" />
    Pull Image
  </Button>
</div>

<Card.Root>
  <Card.Header>
    <Card.Title>Images</Card.Title>
    <Card.Description>Docker images available on this host</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b">
            <th class="text-left py-2">Repository</th>
            <th class="text-left py-2">Tag</th>
            <th class="text-left py-2">Image ID</th>
            <th class="text-left py-2">Size</th>
            <th class="text-left py-2">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each images as image}
            <ImageRow {image} />
          {/each}
        </tbody>
      </table>
    </div>
  </Card.Content>
</Card.Root>
