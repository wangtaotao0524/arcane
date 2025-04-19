<!-- filepath: /Users/kylemendell/dev/ofkm/arcane/src/routes/networks/SubnetCell.svelte -->
<script lang="ts">
  export let subnet: string | null;
</script>

{#if subnet}
  <span class="font-mono text-xs">{subnet}</span>
{:else}
  <span>N/A</span>
{/if}
