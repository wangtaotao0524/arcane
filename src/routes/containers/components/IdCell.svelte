<script lang="ts">
	let { id }: { id: string } = $props();
	let shortId = $derived(id.substring(0, 12));
</script>

<span class="font-mono text-xs">{shortId}</span>
