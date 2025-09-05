<script lang="ts">
	import { cn } from '$lib/utils';

	type Variant =
		| 'red'
		| 'purple'
		| 'green'
		| 'blue'
		| 'gray'
		| 'amber'
		| 'pink'
		| 'indigo'
		| 'cyan'
		| 'lime'
		| 'emerald'
		| 'teal'
		| 'sky'
		| 'violet'
		| 'fuchsia'
		| 'rose'
		| 'orange';

	type MinWidth = 'none' | '16' | '20' | '24' | '28';

	let {
		text,
		variant = 'gray',
		minWidth = '20',
		class: className = '',
		...restProps
	} = $props<{
		text: string;
		variant?: Variant;
		minWidth?: MinWidth;
		class?: string;
	}>();

	const minWidthClasses: Record<MinWidth, string> = {
		none: '',
		'16': 'min-w-16', // 4rem
		'20': 'min-w-20', // 5rem (default)
		'24': 'min-w-24', // 6rem
		'28': 'min-w-28' // 7rem
	};

	const variantStyles: Record<Variant, string> = {
		red: 'bg-red-500/25 text-red-800 border border-red-700/30 dark:bg-red-500/30 dark:text-red-200 dark:border-red-300/30',
		purple:
			'bg-purple-500/25 text-purple-800 border border-purple-700/30 dark:bg-purple-500/30 dark:text-purple-200 dark:border-purple-300/30',
		green:
			'bg-green-500/30 text-green-800 border border-green-700/30 dark:bg-green-500/35 dark:text-green-200 dark:border-green-300/30',
		blue: 'bg-blue-500/25 text-blue-800 border border-blue-700/30 dark:bg-blue-500/30 dark:text-blue-200 dark:border-blue-300/30',
		gray: 'bg-gray-400/25 text-gray-800 border border-gray-700/30 dark:bg-gray-400/30 dark:text-gray-200 dark:border-gray-300/30',
		amber:
			'bg-amber-500/30 text-amber-800 border border-amber-700/30 dark:bg-amber-500/35 dark:text-amber-200 dark:border-amber-300/30',
		pink: 'bg-pink-500/25 text-pink-800 border border-pink-700/30 dark:bg-pink-500/30 dark:text-pink-200 dark:border-pink-300/30',
		indigo:
			'bg-indigo-500/25 text-indigo-800 border border-indigo-700/30 dark:bg-indigo-500/30 dark:text-indigo-200 dark:border-indigo-300/30',
		cyan: 'bg-cyan-500/25 text-cyan-800 border border-cyan-700/30 dark:bg-cyan-500/30 dark:text-cyan-200 dark:border-cyan-300/30',
		lime: 'bg-lime-500/25 text-lime-800 border border-lime-700/30 dark:bg-lime-500/30 dark:text-lime-200 dark:border-lime-300/30',
		emerald:
			'bg-emerald-500/30 text-emerald-800 border border-emerald-700/30 dark:bg-emerald-500/35 dark:text-emerald-200 dark:border-emerald-300/30',
		teal: 'bg-teal-500/25 text-teal-800 border border-teal-700/30 dark:bg-teal-500/30 dark:text-teal-200 dark:border-teal-300/30',
		sky: 'bg-sky-500/25 text-sky-800 border border-sky-700/30 dark:bg-sky-500/30 dark:text-sky-200 dark:border-sky-300/30',
		violet:
			'bg-violet-500/25 text-violet-800 border border-violet-700/30 dark:bg-violet-500/30 dark:text-violet-200 dark:border-violet-300/30',
		fuchsia:
			'bg-fuchsia-500/25 text-fuchsia-800 border border-fuchsia-700/30 dark:bg-fuchsia-500/30 dark:text-fuchsia-200 dark:border-fuchsia-300/30',
		rose: 'bg-rose-500/25 text-rose-800 border border-rose-700/30 dark:bg-rose-500/30 dark:text-rose-200 dark:border-rose-300/30',
		orange:
			'bg-orange-500/25 text-orange-800 border border-orange-700/30 dark:bg-orange-500/30 dark:text-orange-200 dark:border-orange-300/30'
	};

	const resolvedMinWidth = minWidth as MinWidth;
</script>

<span
	class={cn(
		'inline-flex items-center justify-center rounded-full px-2.5 py-0.5 text-center text-xs font-medium whitespace-nowrap',
		minWidthClasses[resolvedMinWidth],
		variantStyles[variant as Variant],
		className
	)}
	{...restProps}
>
	{text}
</span>
