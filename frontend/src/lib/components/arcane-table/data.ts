import BadgeCheckIcon from '@lucide/svelte/icons/badge-check';
import BadgeXIcon from '@lucide/svelte/icons/badge-x';
import CircleFadingArrowUp from '@lucide/svelte/icons/circle-fading-arrow-up';
import CircleCheck from '@lucide/svelte/icons/circle-check';
import { m } from '$lib/paraglide/messages';

export const usageFilters = [
	{
		value: true,
		label: m.common_in_use(),
		icon: BadgeCheckIcon
	},
	{
		value: false,
		label: m.common_unused(),
		icon: CircleCheck
	}
];

export const imageUpdateFilters = [
	{
		value: true,
		label: m.images_has_updates(),
		icon: CircleFadingArrowUp
	},
	{
		value: false,
		label: m.images_no_updates(),
		icon: BadgeXIcon
	}
];
