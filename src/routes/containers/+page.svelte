<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";

  export let data;
  const { containers } = data;
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Containers</h1>
  <button class="bg-primary text-primary-foreground px-4 py-2 rounded"
    >New Container</button
  >
</div>

<Card.Root>
  <Card.Header>
    <Card.Title>Container List</Card.Title>
    <Card.Description>Manage your Docker containers</Card.Description>
  </Card.Header>
  <Card.Content>
    <table class="w-full">
      <thead>
        <tr class="border-b">
          <th class="text-left py-2">Name</th>
          <th class="text-left py-2">Image</th>
          <th class="text-left py-2">Status</th>
          <th class="text-left py-2">Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each containers as container}
          <tr class="border-b">
            <td class="py-2">
              <a
                href="/containers/{container.id}"
                class="text-primary hover:underline"
              >
                {container.name}
              </a>
            </td>
            <td>{container.image}</td>
            <td>
              <span
                class={container.status === "running"
                  ? "text-green-600"
                  : "text-gray-500"}
              >
                {container.status}
              </span>
            </td>
            <td class="flex gap-2 py-2">
              {#if container.status === "running"}
                <button
                  class="px-2 py-1 bg-secondary text-secondary-foreground text-xs rounded"
                  >Stop</button
                >
                <button
                  class="px-2 py-1 bg-secondary text-secondary-foreground text-xs rounded"
                  >Restart</button
                >
              {:else}
                <button
                  class="px-2 py-1 bg-primary text-primary-foreground text-xs rounded"
                  >Start</button
                >
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </Card.Content>
</Card.Root>
