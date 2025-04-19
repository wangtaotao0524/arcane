<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";

  let { data } = $props();
  const { container } = data;
</script>

<div class="flex justify-between items-center mb-6">
  <div>
    <a
      href="/containers"
      class="text-muted-foreground hover:underline mb-2 inline-block"
    >
      &larr; Back to containers
    </a>
    <h1 class="text-2xl font-bold">{container.name}</h1>
  </div>

  <div class="flex gap-2">
    {#if container.status === "running"}
      <button class="bg-secondary text-secondary-foreground px-3 py-1 rounded"
        >Stop</button
      >
      <button class="bg-secondary text-secondary-foreground px-3 py-1 rounded"
        >Restart</button
      >
    {:else}
      <button class="bg-primary text-primary-foreground px-3 py-1 rounded"
        >Start</button
      >
    {/if}
    <button class="bg-destructive text-destructive-foreground px-3 py-1 rounded"
      >Remove</button
    >
  </div>
</div>

<div class="grid grid-cols-1 md:grid-cols-4 gap-6 mb-6">
  <Card.Root>
    <Card.Header>
      <Card.Title>Status</Card.Title>
    </Card.Header>
    <Card.Content>
      <span
        class={container.status === "running"
          ? "text-green-600"
          : "text-gray-500"}
      >
        {container.status}
      </span>
    </Card.Content>
  </Card.Root>

  <Card.Root>
    <Card.Header>
      <Card.Title>Image</Card.Title>
    </Card.Header>
    <Card.Content>
      {container.image}
    </Card.Content>
  </Card.Root>

  <Card.Root>
    <Card.Header>
      <Card.Title>Created</Card.Title>
    </Card.Header>
    <Card.Content>
      {container.created || "Unknown"}
    </Card.Content>
  </Card.Root>

  <Card.Root>
    <Card.Header>
      <Card.Title>ID</Card.Title>
    </Card.Header>
    <Card.Content class="font-mono text-sm">
      {container.id}
    </Card.Content>
  </Card.Root>
</div>

<div class="grid grid-cols-1 md:grid-cols-12 gap-6">
  <Card.Root class="md:col-span-8">
    <Card.Header>
      <Card.Title>Logs</Card.Title>
      <Card.Description>Recent container output</Card.Description>
    </Card.Header>
    <Card.Content>
      <div
        class="bg-black text-white p-4 rounded font-mono text-sm h-64 overflow-y-auto"
      >
        <p>Starting container...</p>
        <p>Container ready</p>
        <!-- Logs content would go here -->
      </div>
    </Card.Content>
    <Card.Footer>
      <a
        href="/containers/{container.id}/logs"
        class="text-primary hover:underline">View full logs</a
      >
    </Card.Footer>
  </Card.Root>

  <Card.Root class="md:col-span-4">
    <Card.Header>
      <Card.Title>Stats</Card.Title>
      <Card.Description>Resource usage</Card.Description>
    </Card.Header>
    <Card.Content>
      <ul>
        <li class="mb-2">
          <div class="text-muted-foreground text-sm">CPU</div>
          <div class="w-full bg-secondary rounded-full h-2.5 mt-1">
            <div class="bg-primary h-2.5 rounded-full" style="width: 25%"></div>
          </div>
        </li>
        <li class="mb-2">
          <div class="text-muted-foreground text-sm">Memory</div>
          <div class="w-full bg-secondary rounded-full h-2.5 mt-1">
            <div class="bg-primary h-2.5 rounded-full" style="width: 40%"></div>
          </div>
        </li>
      </ul>
    </Card.Content>
    <Card.Footer>
      <a
        href="/containers/{container.id}/stats"
        class="text-primary hover:underline">View detailed stats</a
      >
    </Card.Footer>
  </Card.Root>
</div>
