<script lang="ts">
  import CustomBadge from "$lib/components/badges/custom-badge.svelte";

  export let name: string;
  export let inUse: boolean | undefined;
</script>

<div class="flex items-center gap-2">
  <span class="truncate">{name}</span>
  {#if !inUse}
    <CustomBadge variant="status" text="Unused" className="ml-2" />
  {/if}
</div>
