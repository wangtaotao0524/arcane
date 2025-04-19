<script lang="ts">
  import type { PageData } from "./$types";
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import * as Table from "$lib/components/ui/table/index.js";
  import StatusBadge from "$lib/components/docker/StatusBadge.svelte";
  import { AlertCircle } from "@lucide/svelte";
  import * as Alert from "$lib/components/ui/alert/index.js";

  let { data }: { data: PageData } = $props();
  const { dockerInfo, containers, images, error } = data;

  // Helper to format bytes
  function formatBytes(bytes: number, decimals = 2): string {
    if (!+bytes) return "0 Bytes";
    const k = 1024;
    const dm = decimals < 0 ? 0 : decimals;
    const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
  }
</script>

<h1 class="text-2xl font-bold mb-6">Dashboard</h1>

{#if error}
  <Alert.Root variant="destructive" class="mb-6">
    <AlertCircle class="h-4 w-4" />
    <Alert.Title>Error Loading Data</Alert.Title>
    <Alert.Description>{error}</Alert.Description>
  </Alert.Root>
{/if}

<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
  <Card.Root>
    <Card.Header>
      <Card.Title>Docker Info</Card.Title>
      <Card.Description>System overview</Card.Description>
    </Card.Header>
    <Card.Content>
      {#if dockerInfo}
        <ul class="text-sm space-y-1">
          <li><strong>Version:</strong> {dockerInfo.ServerVersion || "N/A"}</li>
          <li><strong>OS:</strong> {dockerInfo.OperatingSystem || "N/A"}</li>
          <li>
            <strong>Architecture:</strong>
            {dockerInfo.Architecture || "N/A"}
          </li>
          <li>
            <strong>Total Memory:</strong>
            {formatBytes(dockerInfo.MemTotal || 0)}
          </li>
          <li><strong>Containers:</strong> {dockerInfo.Containers || 0}</li>
          <li><strong>Images:</strong> {dockerInfo.Images || 0}</li>
        </ul>
      {:else}
        <p class="text-sm text-muted-foreground">Could not load Docker info.</p>
      {/if}
    </Card.Content>
  </Card.Root>

  <Card.Root class="md:col-span-2">
    <Card.Header>
      <Card.Title>Quick Actions</Card.Title>
      <Card.Description>Manage all containers</Card.Description>
    </Card.Header>
    <Card.Content>
      <div class="flex gap-2 flex-wrap">
        <!-- TODO: Implement actions -->
        <Button disabled={!dockerInfo}>Start All</Button>
        <Button variant="secondary" disabled={!dockerInfo}>Stop All</Button>
        <Button variant="destructive" disabled={!dockerInfo}
          >Prune System</Button
        >
      </div>
      <p class="text-xs text-muted-foreground mt-2">
        Note: Actions not yet implemented.
      </p>
    </Card.Content>
  </Card.Root>
</div>

<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
  <Card.Root>
    <Card.Header>
      <Card.Title>Containers ({containers.length})</Card.Title>
      <Card.Description>Active and stopped containers</Card.Description>
    </Card.Header>
    <Card.Content>
      {#if containers.length > 0}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>Name</Table.Head>
              <Table.Head>Image</Table.Head>
              <Table.Head>State</Table.Head>
              <Table.Head>Status</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each containers.slice(0, 5) as c (c.id)}
              <Table.Row>
                <Table.Cell class="font-medium">{c.name}</Table.Cell>
                <Table.Cell class="text-xs truncate max-w-xs"
                  >{c.image}</Table.Cell
                >
                <Table.Cell>
                  <StatusBadge state={c.state} />
                </Table.Cell>
                <Table.Cell class="text-xs">{c.status}</Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
        {#if containers.length > 5}
          <p class="text-xs text-muted-foreground mt-2">
            Showing first 5 containers.
          </p>
        {/if}
      {:else if !error}
        <p class="text-sm text-muted-foreground">No containers found.</p>
      {/if}
    </Card.Content>
    <Card.Footer>
      <Button
        variant="outline"
        size="sm"
        href="/containers"
        disabled={!dockerInfo}>View All Containers</Button
      >
    </Card.Footer>
  </Card.Root>

  <Card.Root>
    <Card.Header>
      <Card.Title>Images ({images.length})</Card.Title>
      <Card.Description>Available Docker images</Card.Description>
    </Card.Header>
    <Card.Content>
      {#if images.length > 0}
        <Table.Root>
          <Table.Header>
            <Table.Row>
              <Table.Head>Repository</Table.Head>
              <Table.Head>Tag</Table.Head>
              <Table.Head>Size</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {#each images.slice(0, 5) as img (img.id)}
              <Table.Row>
                <Table.Cell class="font-medium truncate max-w-xs"
                  >{img.repo}</Table.Cell
                >
                <Table.Cell>{img.tag}</Table.Cell>
                <Table.Cell>{formatBytes(img.size)}</Table.Cell>
              </Table.Row>
            {/each}
          </Table.Body>
        </Table.Root>
        {#if images.length > 5}
          <p class="text-xs text-muted-foreground mt-2">
            Showing first 5 images.
          </p>
        {/if}
      {:else if !error}
        <p class="text-sm text-muted-foreground">No images found.</p>
      {/if}
    </Card.Content>
    <Card.Footer>
      <Button variant="outline" size="sm" href="/images" disabled={!dockerInfo}
        >View All Images</Button
      >
    </Card.Footer>
  </Card.Root>
</div>
