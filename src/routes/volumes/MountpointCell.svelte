<script lang="ts">
	interface Props {
		mountpoint: string;
	}

	let { mountpoint }: Props = $props();

	const MAX_LENGTH = 50;
	let displayValue = $derived(mountpoint.length > MAX_LENGTH ? mountpoint.substring(0, MAX_LENGTH - 3) + '...' : mountpoint);
</script>

<span class="font-mono text-xs" title={mountpoint}>
	{displayValue}
</span>
