<script lang="ts">
  import StatusBadge from "./status-badge.svelte";

  export let name: string;
  export let inUse: boolean | undefined;
</script>

<div class="flex items-center gap-2">
  <span class="truncate">{name}</span>
  {#if !inUse}
    <StatusBadge text="Unused" variant="amber" />
  {/if}
</div>
