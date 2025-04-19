<script lang="ts">
  import type { PageData, ActionData } from "./$types";
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import StatusBadge from "$lib/components/docker/StatusBadge.svelte";
  import { ArrowLeft, Loader2, AlertCircle, RefreshCw } from "@lucide/svelte";
  import * as Breadcrumb from "$lib/components/ui/breadcrumb/index.js";
  import { enhance } from "$app/forms";
  import { invalidateAll } from "$app/navigation";
  import * as Alert from "$lib/components/ui/alert/index.js";
  import { Separator } from "$lib/components/ui/separator/index.js";

  let { data, form }: { data: PageData; form: ActionData } = $props();
  let { container, logs } = $derived(data);

  let starting = $state(false);
  let stopping = $state(false);
  let restarting = $state(false);
  let removing = $state(false);

  function formatDate(dateString: string | undefined | null): string {
    if (!dateString) return "Unknown";
    try {
      return new Date(dateString).toLocaleString();
    } catch (e) {
      return "Invalid Date";
    }
  }

  // Function to format logs with some basic highlighting
  function formatLogLine(line: string): string {
    if (
      line.includes("ERROR") ||
      line.includes("FATAL") ||
      line.includes("WARN")
    ) {
      return `<span class="text-red-400">${line}</span>`;
    }
    if (line.includes("INFO")) {
      return `<span class="text-blue-400">${line}</span>`;
    }
    return line;
  }

  // Format all logs
  let formattedLogs = $derived(
    logs ? logs.split("\n").map(formatLogLine).join("\n") : ""
  );

  // Set up auto-scroll for logs
  let logsContainer = $state<HTMLDivElement | undefined>(undefined);

  $effect(() => {
    starting = false;
    stopping = false;
    restarting = false;
    removing = false;
  });

  $effect(() => {
    // Scroll to bottom whenever logs change
    if (logsContainer && logs) {
      logsContainer.scrollTop = logsContainer.scrollHeight;
    }
  });
</script>

<div class="container mx-auto pb-8">
  <!-- Breadcrumb Navigation -->
  <div
    class="flex flex-col sm:flex-row justify-between items-start sm:items-center mt-4 mb-6 gap-4"
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
          <Breadcrumb.Page>{container?.name || "Loading..."}</Breadcrumb.Page>
        </Breadcrumb.Item>
      </Breadcrumb.List>
    </Breadcrumb.Root>

    {#if container}
      <div class="flex gap-2 flex-wrap">
        {#if container.state?.Running}
          <form
            method="POST"
            action="?/stop"
            use:enhance={() => {
              stopping = true;
              return async ({ update }) => {
                await update({ reset: false });
              };
            }}
          >
            <Button
              type="submit"
              variant="secondary"
              disabled={stopping}
              size="sm"
              class="font-medium"
            >
              {#if stopping}
                <Loader2 class="w-4 h-4 mr-2 animate-spin" />
              {/if}
              Stop
            </Button>
          </form>
          <form
            method="POST"
            action="?/restart"
            use:enhance={() => {
              restarting = true;
              return async ({ update }) => {
                await update({ reset: false });
              };
            }}
          >
            <Button
              type="submit"
              variant="secondary"
              disabled={restarting}
              size="sm"
              class="font-medium"
            >
              {#if restarting}
                <Loader2 class="w-4 h-4 mr-2 animate-spin" />
              {/if}
              Restart
            </Button>
          </form>
        {:else if container.state?.Status === "exited" || container.state?.Status === "created"}
          <form
            method="POST"
            action="?/start"
            use:enhance={() => {
              starting = true;
              return async ({ update }) => {
                await update({ reset: false });
              };
            }}
          >
            <Button
              type="submit"
              variant="default"
              disabled={starting}
              size="sm"
              class="font-medium"
            >
              {#if starting}
                <Loader2 class="w-4 h-4 mr-2 animate-spin" />
              {/if}
              Start
            </Button>
          </form>
        {/if}
        <form
          method="POST"
          action="?/remove"
          use:enhance={() => {
            if (
              !confirm(
                `Are you sure you want to remove container "${container?.name}"?`
              )
            ) {
              return;
            }
            removing = true;
            return async ({ update }) => {
              await update({ reset: false });
            };
          }}
        >
          <Button
            type="submit"
            variant="destructive"
            disabled={removing}
            size="sm"
            class="font-medium"
          >
            {#if removing}
              <Loader2 class="w-4 h-4 mr-2 animate-spin" />
            {/if}
            Remove
          </Button>
        </form>
      </div>
    {/if}
  </div>

  <!-- Error Alert -->
  {#if form?.error}
    <Alert.Root variant="destructive" class="mb-6">
      <AlertCircle class="h-4 w-4 mr-2" />
      <Alert.Title>Action Failed</Alert.Title>
      <Alert.Description>{form.error}</Alert.Description>
    </Alert.Root>
  {/if}

  {#if container}
    <!-- Container Details Section -->
    <div class="mb-6">
      <h2 class="text-xl font-bold mb-4">Container Details</h2>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card.Root class="overflow-hidden">
          <Card.Header class="pb-2 bg-muted/40">
            <Card.Title class="text-sm font-medium">State</Card.Title>
          </Card.Header>
          <Card.Content class="pt-4">
            <StatusBadge state={container.state?.Status ?? "unknown"} />
          </Card.Content>
        </Card.Root>

        <Card.Root class="overflow-hidden">
          <Card.Header class="pb-2 bg-muted/40">
            <Card.Title class="text-sm font-medium">Image</Card.Title>
          </Card.Header>
          <Card.Content class="pt-4 text-sm font-medium break-all">
            <div class="truncate" title={container.config?.Image || "N/A"}>
              {container.config?.Image || "N/A"}
            </div>
          </Card.Content>
        </Card.Root>

        <Card.Root class="overflow-hidden">
          <Card.Header class="pb-2 bg-muted/40">
            <Card.Title class="text-sm font-medium">Created</Card.Title>
          </Card.Header>
          <Card.Content class="pt-4 text-sm font-medium">
            {formatDate(container.created)}
          </Card.Content>
        </Card.Root>

        <Card.Root class="overflow-hidden">
          <Card.Header class="pb-2 bg-muted/40">
            <Card.Title class="text-sm font-medium">IP Address</Card.Title>
          </Card.Header>
          <Card.Content class="pt-4 text-sm font-medium">
            {container.networkSettings?.IPAddress || "N/A"}
          </Card.Content>
        </Card.Root>
      </div>
    </div>

    <Separator class="my-6" />

    <!-- Logs and Stats Section -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6">
      <!-- Logs Card -->
      <Card.Root class="lg:col-span-8 border shadow-sm">
        <Card.Header>
          <div
            class="flex flex-col sm:flex-row justify-between items-start sm:items-center w-full gap-4"
          >
            <div>
              <Card.Title class="text-lg font-semibold"
                >Container Logs</Card.Title
              >
              <Card.Description
                >Recent output from the container</Card.Description
              >
            </div>
          </div>
        </Card.Header>

        <Card.Content>
          <div
            class="bg-muted/50 text-foreground p-4 rounded-md font-mono text-xs h-80 overflow-y-auto whitespace-pre-wrap border"
            bind:this={logsContainer}
            id="logs-container"
          >
            {#if logs}
              {@html formattedLogs}
            {:else}
              <p class="text-muted-foreground italic">
                No logs available. The container may not have started yet.
              </p>
            {/if}
          </div>
        </Card.Content>

        <Card.Footer class="flex justify-between border-t pt-4">
          <Button
            variant="outline"
            size="sm"
            href="/containers/{container.id}/logs"
            class="text-xs"
          >
            View full logs
          </Button>
          <Button
            variant="outline"
            size="sm"
            onclick={() => invalidateAll()}
            class="text-xs"
          >
            <RefreshCw class="w-3.5 h-3.5 mr-1" /> Refresh
          </Button>
        </Card.Footer>
      </Card.Root>

      <!-- Stats Card -->
      <Card.Root class="lg:col-span-4 border shadow-sm">
        <Card.Header>
          <Card.Title class="text-lg font-semibold">Resource Usage</Card.Title>
          <Card.Description>Container performance metrics</Card.Description>
        </Card.Header>

        <Card.Content>
          <div class="space-y-4">
            <div>
              <div class="flex justify-between mb-1">
                <span class="text-sm font-medium">CPU Usage</span>
                <span class="text-sm text-muted-foreground">0%</span>
              </div>
              <div class="w-full bg-secondary rounded-full h-2">
                <div
                  class="bg-primary h-2 rounded-full"
                  style="width: 0%"
                ></div>
              </div>
            </div>

            <div>
              <div class="flex justify-between mb-1">
                <span class="text-sm font-medium">Memory Usage</span>
                <span class="text-sm text-muted-foreground">0 / 0 MB</span>
              </div>
              <div class="w-full bg-secondary rounded-full h-2">
                <div
                  class="bg-primary h-2 rounded-full"
                  style="width: 0%"
                ></div>
              </div>
            </div>

            <div class="pt-2 text-xs italic text-muted-foreground">
              Stats fetching not implemented yet.
            </div>
          </div>
        </Card.Content>

        <Card.Footer class="border-t pt-4">
          <Button
            variant="outline"
            size="sm"
            href="/containers/{container.id}/stats"
            class="w-full text-xs"
          >
            View detailed stats
          </Button>
        </Card.Footer>
      </Card.Root>
    </div>
  {:else}
    <div class="flex items-center justify-center h-48 border rounded-lg">
      <p class="text-center text-muted-foreground">
        Could not load container data.
      </p>
    </div>
  {/if}
</div>
