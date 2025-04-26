<script lang="ts">
	interface Props {
		id: string;
		name: string;
	}

	let { id, name }: Props = $props();
</script>

<a href="/stacks/{id}" class="font-medium hover:underline">
	{name}
</a>
