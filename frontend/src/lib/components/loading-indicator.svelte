<script lang="ts">
	let {
		active = false,
		delay = 80,
		minDuration = 350,
		thickness = 'h-2',
		class: className = ''
	}: {
		active?: boolean;
		delay?: number;
		minDuration?: number;
		thickness?: string;
		class?: string;
	} = $props();

	let visible = $state(false);
	let startedAt = 0;
	let showTO: any;
	let hideTO: any;

	function clearTimers() {
		clearTimeout(showTO);
		clearTimeout(hideTO);
	}

	$effect(() => {
		clearTimers();
		if (active) {
			showTO = setTimeout(() => {
				visible = true;
				startedAt = Date.now();
			}, delay);
		} else {
			const elapsed = Date.now() - startedAt;
			const remaining = Math.max(0, minDuration - elapsed);
			hideTO = setTimeout(() => (visible = false), remaining);
		}
		return clearTimers;
	});
</script>

{#if visible}
	<div
		class={`pointer-events-none fixed inset-x-0 top-0 z-[9999] ${thickness} ${className}`}
		role="progressbar"
		aria-busy="true"
	>
		<div class="relative h-full w-full overflow-hidden">
			<div class="absolute inset-0 bg-primary/25"></div>

			<span class="bar absolute inset-y-0 w-1/3"></span>

			<span class="peg absolute right-0 top-0 h-full w-3"></span>
		</div>
	</div>
{/if}

<style>
	:root {
		--arcane-primary: hsl(var(--primary));
	}

	.bar {
		background: var(--arcane-primary);
		animation: arcane-slide 0.9s cubic-bezier(0.4, 0, 0.2, 1) infinite;
		box-shadow:
			0 0 12px var(--arcane-primary),
			0 0 2px var(--arcane-primary);
		opacity: 0.95;
	}

	.peg {
		background: linear-gradient(
			to right,
			color-mix(in oklab, var(--arcane-primary) 0%, transparent),
			color-mix(in oklab, var(--arcane-primary) 90%, transparent)
		);
		filter: blur(2px);
	}

	@keyframes arcane-slide {
		0% {
			transform: translateX(-120%);
		}
		100% {
			transform: translateX(520%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.bar {
			animation-duration: 2s;
		}
	}
</style>
