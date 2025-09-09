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
	import userStore from '$lib/stores/user-store';
	import { m } from '$lib/paraglide/messages';

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

	let storeUser: User | null = $state(null);
	$effect(() => {
		const unsub = userStore.subscribe((u) => (storeUser = u));
		return unsub;
	});
	const effectiveUser = $derived(user ?? storeUser);

	const isCollapsed = $derived(sidebar.state === 'collapsed');
	const isAdmin = $derived(!!effectiveUser?.roles?.includes('admin'));
</script>

<Sidebar.Root {collapsible} {...restProps}>
	<Sidebar.Header>
		<SidebarLogo {isCollapsed} {versionInformation} />
		<SidebarEnvSwitcher {isAdmin} />
	</Sidebar.Header>
	<Sidebar.Content>
		<SidebarItemGroup label={m.sidebar_management()} items={sidebarItems.managementItems} />
		<SidebarItemGroup label={m.sidebar_customization()} items={sidebarItems.customizationItems} />
		{#if isAdmin}
			<SidebarItemGroup label={m.sidebar_environments()} items={sidebarItems.environmentItems} />
			<SidebarItemGroup label={m.sidebar_administration()} items={sidebarItems.settingsItems} />
		{/if}
	</Sidebar.Content>
	<Sidebar.Footer>
		<SidebarUpdatebanner {isCollapsed} {versionInformation} updateAvailable={versionInformation.updateAvailable} />
		{#if effectiveUser}
			<SidebarUser {isCollapsed} user={effectiveUser} />
		{/if}
	</Sidebar.Footer>
	<Sidebar.Rail />
</Sidebar.Root>
