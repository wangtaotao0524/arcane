<script lang="ts" module>
	import { sidebarItems } from '$lib/config/sidebar-config';
</script>

<script lang="ts">
	import SidebarItemGroup from './sidebar-itemgroup.svelte';
	import SidebarUser from './sidebar-user.svelte';
	import SidebarEnvSwitcher from './sidebar-env-switcher.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { useSidebar } from '$lib/components/ui/sidebar/index.js';
	import type { ComponentProps } from 'svelte';
	import type { User } from '$lib/types/user.type';
	import type { AppVersionInformation } from '$lib/types/application-configuration';
	import SidebarLogo from './sidebar-logo.svelte';
	import SidebarUpdatebanner from './sidebar-updatebanner.svelte';

	let {
		ref = $bindable(null),
		collapsible = 'icon',
		user,
		versionInformation,
		...restProps
	}: ComponentProps<typeof Sidebar.Root> & {
		versionInformation: AppVersionInformation;
		user?: User | null;
	} = $props();

	const sidebar = useSidebar();

	const isCollapsed = $derived(sidebar.state === 'collapsed');
</script>

<Sidebar.Root {collapsible} {...restProps}>
	<Sidebar.Header>
		<SidebarLogo {isCollapsed} {versionInformation} />
		<SidebarEnvSwitcher />
	</Sidebar.Header>
	<Sidebar.Content>
		<SidebarItemGroup label="Management" items={sidebarItems.managementItems} />
		<SidebarItemGroup label="Customization" items={sidebarItems.customizationItems} />
		<SidebarItemGroup label="Environments" items={sidebarItems.environmentItems} />
		<SidebarItemGroup label="Administration" items={sidebarItems.settingsItems} />
	</Sidebar.Content>
	<Sidebar.Footer>
		<SidebarUpdatebanner
			{isCollapsed}
			{versionInformation}
			updateAvailable={versionInformation.updateAvailable}
		/>
		{#if user}
			<SidebarUser {isCollapsed} {user} />
		{/if}
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
