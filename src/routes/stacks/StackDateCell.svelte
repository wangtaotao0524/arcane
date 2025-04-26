<script lang="ts">
	interface Props {
		date: string;
	}

	let { date }: Props = $props();

	// Format the date with a nice relative format
	function formatDate(date: string): string {
		const dateObj = new Date(date);
		const now = new Date();
		const diff = now.getTime() - dateObj.getTime();
		const diffDays = Math.floor(diff / (1000 * 3600 * 24));

		if (diffDays === 0) {
			return 'Today, ' + dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
		} else if (diffDays === 1) {
			return 'Yesterday';
		} else if (diffDays < 7) {
			return `${diffDays} days ago`;
		} else {
			return dateObj.toLocaleDateString();
		}
	}
</script>

<span title={new Date(date).toLocaleString()}>
	{formatDate(date)}
</span>
