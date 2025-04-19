<script lang="ts">
  import type { PageData } from "./$types";
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import StatusBadge from "$lib/components/docker/StatusBadge.svelte";
  import { ArrowLeft } from "@lucide/svelte";
  import * as Breadcrumb from "$lib/components/ui/breadcrumb/index.js";

  let { data }: { data: PageData } = $props();
  const { container } = data;

  function formatDate(dateString: string | undefined | null): string {
    if (!dateString) return "Unknown";
    try {
      return new Date(dateString).toLocaleString();
    } catch (e) {
      return "Invalid Date";
    }
  }
</script>

<div
  class="flex flex-col sm:flex-row justify-between items-start sm:items-center mb-6 gap-4"
>
  <Breadcrumb.Root>
    <Breadcrumb.List>
      <Breadcrumb.Item>
        <Breadcrumb.Link href="/">Dashboard</Breadcrumb.Link>
      </Breadcrumb.Item>
      <Breadcrumb.Separator />
      <Breadcrumb.Item>
        <Breadcrumb.Link href="/containers">Containers</Breadcrumb.Link>
      </Breadcrumb.Item>
      <Breadcrumb.Separator />
      <Breadcrumb.Item>
        <Breadcrumb.Page>{container.name}</Breadcrumb.Page>
      </Breadcrumb.Item>
    </Breadcrumb.List>
  </Breadcrumb.Root>

  {#if container}
    <div class="flex gap-2 flex-wrap">
      {#if container.state?.Running}
        <Button variant="secondary">Stop</Button>
        <Button variant="secondary">Restart</Button>
      {:else if container.state?.Status === "exited" || container.state?.Status === "created"}
        <Button variant="default">Start</Button>
      {/if}
      <Button variant="destructive">Remove</Button>
    </div>
  {/if}
</div>

{#if container}
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
    <Card.Root>
      <Card.Header class="pb-2">
        <Card.Title class="text-sm font-medium text-muted-foreground"
          >State</Card.Title
        >
      </Card.Header>
      <Card.Content>
        <StatusBadge state={container.state?.Status || "unknown"} />
      </Card.Content>
    </Card.Root>

    <Card.Root>
      <Card.Header class="pb-2">
        <Card.Title class="text-sm font-medium text-muted-foreground"
          >Image</Card.Title
        >
      </Card.Header>
      <Card.Content class="text-sm font-medium break-all">
        {container.config?.Image || "N/A"}
      </Card.Content>
    </Card.Root>

    <Card.Root>
      <Card.Header class="pb-2">
        <Card.Title class="text-sm font-medium text-muted-foreground"
          >Created</Card.Title
        >
      </Card.Header>
      <Card.Content class="text-sm font-medium">
        {formatDate(container.created)}
      </Card.Content>
    </Card.Root>

    <Card.Root>
      <Card.Header class="pb-2">
        <Card.Title class="text-sm font-medium text-muted-foreground"
          >IP Address</Card.Title
        >
      </Card.Header>
      <Card.Content class="text-sm font-medium">
        {container.networkSettings?.IPAddress || "N/A"}
      </Card.Content>
    </Card.Root>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
    <Card.Root class="lg:col-span-8">
      <Card.Header>
        <Card.Title>Logs</Card.Title>
        <Card.Description>Recent container output</Card.Description>
      </Card.Header>
      <Card.Content>
        <div
          class="bg-muted text-muted-foreground p-4 rounded font-mono text-xs h-64 overflow-y-auto"
        >
          <p>Log fetching not implemented yet.</p>
        </div>
      </Card.Content>
      <Card.Footer>
        <Button
          variant="link"
          class="p-0 h-auto"
          href="/containers/{container.id}/logs"
        >
          View full logs
        </Button>
      </Card.Footer>
    </Card.Root>

    <Card.Root class="lg:col-span-4">
      <Card.Header>
        <Card.Title>Stats</Card.Title>
        <Card.Description>Resource usage</Card.Description>
      </Card.Header>
      <Card.Content>
        <p class="text-sm text-muted-foreground">
          Stats fetching not implemented yet.
        </p>
        <ul>
          <li class="mb-2">
            <div class="text-muted-foreground text-sm">CPU</div>
            <div class="w-full bg-secondary rounded-full h-2.5 mt-1">
              <div
                class="bg-primary h-2.5 rounded-full"
                style="width: 0%"
              ></div>
            </div>
          </li>
          <li class="mb-2">
            <div class="text-muted-foreground text-sm">Memory</div>
            <div class="w-full bg-secondary rounded-full h-2.5 mt-1">
              <div
                class="bg-primary h-2.5 rounded-full"
                style="width: 0%"
              ></div>
            </div>
          </li>
        </ul>
      </Card.Content>
      <Card.Footer>
        <Button
          variant="link"
          class="p-0 h-auto"
          href="/containers/{container.id}/stats"
        >
          View detailed stats
        </Button>
      </Card.Footer>
    </Card.Root>
  </div>
{:else}
  <p class="text-center text-muted-foreground">
    Could not load container data.
  </p>
{/if}
