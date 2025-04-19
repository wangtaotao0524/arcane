<script lang="ts">
  import { enhance } from "$app/forms";
  import { Button } from "$lib/components/ui/button/index.js";
  import {
    Play,
    CircleStop,
    Trash2,
    Settings,
    RefreshCw,
  } from "@lucide/svelte";

  interface Props {
    id: string;
    status: string;
  }

  let { id, status }: Props = $props();

  const isRunning = status === "running" || status === "partially running";
</script>

<div class="flex items-center gap-2 justify-end">
  {#if isRunning}
    <form method="POST" action="/stacks/{id}?/stop" use:enhance>
      <Button type="submit" size="icon" variant="outline" title="Stop Stack">
        <CircleStop class="h-4 w-4" />
      </Button>
    </form>
  {:else}
    <form method="POST" action="/stacks/{id}?/start" use:enhance>
      <Button type="submit" size="icon" variant="outline" title="Start Stack">
        <Play class="h-4 w-4" />
      </Button>
    </form>
  {/if}

  <form method="POST" action="/stacks/{id}?/restart" use:enhance>
    <Button
      type="submit"
      size="icon"
      variant="outline"
      disabled={!isRunning}
      title="Restart Stack"
    >
      <RefreshCw class="h-4 w-4" />
    </Button>
  </form>

  <a href="/stacks/{id}">
    <Button size="icon" variant="outline" title="Edit Stack">
      <Settings class="h-4 w-4" />
    </Button>
  </a>

  <form method="POST" action="/stacks/{id}?/remove" use:enhance>
    <Button
      type="submit"
      size="icon"
      variant="destructive"
      title="Remove Stack"
    >
      <Trash2 class="h-4 w-4" />
    </Button>
  </form>
</div>
