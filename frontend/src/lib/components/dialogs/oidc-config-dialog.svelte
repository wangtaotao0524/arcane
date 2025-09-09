<script lang="ts">
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import KeyIcon from '@lucide/svelte/icons/key';
	import TriangleAlertIcon from '@lucide/svelte/icons/triangle-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import CopyIcon from '@lucide/svelte/icons/copy';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { toast } from 'svelte-sonner';
	import type { Settings } from '$lib/types/settings.type';
	import type { OidcStatusInfo } from '$lib/types/settings.type';
	import { m } from '$lib/paraglide/messages';

	interface OidcForm {
		clientId: string;
		clientSecret: string;
		issuerUrl: string;
		scopes: string;
		adminClaim?: string;
		adminValue?: string;
	}

	let {
		open = $bindable(),
		currentSettings,
		oidcStatus,
		oidcForm = $bindable(),
		onSave
	}: {
		open: boolean;
		currentSettings: Settings;
		oidcStatus: OidcStatusInfo;
		oidcForm: OidcForm;
		onSave: () => void;
	} = $props();

	let isOidcViewMode = $derived(oidcStatus.envForced && oidcStatus.envConfigured);
	let redirectUri = $derived(`${globalThis?.location?.origin ?? ''}/auth/oidc/callback`);

	let parsedOidcConfig = $derived(() => {
		if (!currentSettings.authOidcConfig) {
			return {
				clientId: '',
				clientSecret: '',
				issuerUrl: '',
				scopes: 'openid email profile',
				adminClaim: '',
				adminValue: ''
			};
		}

		try {
			return JSON.parse(currentSettings.authOidcConfig);
		} catch (error) {
			console.warn('Failed to parse OIDC config:', error);
			return {
				clientId: '',
				clientSecret: '',
				issuerUrl: '',
				scopes: 'openid email profile',
				adminClaim: '',
				adminValue: ''
			};
		}
	});

	const statusItems = $derived([
		{ label: m.oidc_status_env_forced(), value: oidcStatus.envForced, hint: m.oidc_status_env_forced_hint() },
		{ label: m.oidc_status_env_configured(), value: oidcStatus.envConfigured, hint: m.oidc_status_env_configured_hint() }
	]);

	function chipClass(v: boolean) {
		return v
			? 'bg-emerald-500/10 text-emerald-500 ring-1 ring-emerald-500/20'
			: 'bg-muted text-muted-foreground ring-1 ring-border';
	}

	async function copyToClipboard(text: string) {
		try {
			await navigator.clipboard.writeText(text);
			toast.success(m.common_copied());
		} catch {
			toast.error(m.common_copy_failed());
		}
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content class="flex max-h-[90vh] flex-col sm:max-w-[860px]">
		<Dialog.Header class="flex-shrink-0 space-y-3">
			<Dialog.Title class="flex items-center gap-2">
				<span class="bg-primary/10 text-primary ring-primary/20 flex size-8 items-center justify-center rounded-md ring-1">
					<KeyIcon class="size-4" />
				</span>
				{#if isOidcViewMode}
					{m.security_oidc_configured_forced_view()}
				{:else if oidcStatus.envForced && !oidcStatus.envConfigured}
					{m.security_server_forces_oidc_missing_env()}
				{:else}
					{m.security_manage_oidc_config()}
				{/if}
			</Dialog.Title>
			<Dialog.Description class="text-sm leading-relaxed">
				{#if isOidcViewMode}
					{m.oidc_viewmode_description()}
				{:else if oidcStatus.envForced && !oidcStatus.envConfigured}
					{m.security_server_forces_oidc_missing_env()}
				{:else}
					{m.oidc_configure_description()}
				{/if}
			</Dialog.Description>
		</Dialog.Header>

		<div class="min-h-0 flex-1 overflow-y-auto py-4">
			<div class="grid gap-6 lg:grid-cols-[1fr_18rem]">
				<!-- Main column -->
				<div class="space-y-6">
					{#if isOidcViewMode}
						<div class="bg-card/50 rounded-lg border p-4">
							<h3 class="text-base font-semibold">{m.oidc_server_env_config_title()}</h3>
							<p class="text-muted-foreground mb-4 mt-1 text-sm">{m.oidc_server_env_config_description()}</p>

							{#if currentSettings.authOidcEnabled}
								<div class="divide-border/60 divide-y text-sm">
									<div class="flex items-center justify-between py-2">
										<span class="font-medium">{m.oidc_client_id_label()}</span>
										<code class="bg-muted rounded px-2 py-1 text-xs">{parsedOidcConfig().clientId || m.common_unknown()}</code>
									</div>
									<div class="flex items-center justify-between py-2">
										<span class="font-medium">{m.oidc_issuer_url_label()}</span>
										<code class="bg-muted rounded px-2 py-1 text-xs">{parsedOidcConfig().issuerUrl || m.common_unknown()}</code>
									</div>
									<div class="flex items-center justify-between py-2">
										<span class="font-medium">{m.oidc_client_secret_label()}</span>
										<span class="text-muted-foreground text-xs italic">{m.oidc_client_secret_hidden()}</span>
									</div>
									<div class="flex items-center justify-between py-2">
										<code class="bg-muted rounded px-2 py-1 text-xs">{parsedOidcConfig().scopes || 'openid email profile'}</code>
										<span class="font-medium">{m.oidc_scopes_label()}</span>

										<code class="bg-muted rounded px-2 py-1 text-xs"
											>{parsedOidcConfig().scopes || m.oidc_scopes_placeholder()}</code
										>
									</div>
									<div class="flex items-center justify-between py-2">
										<span class="font-medium">{m.oidc_admin_claim_label()}</span>

										<code class="bg-muted rounded px-2 py-1 text-xs">{parsedOidcConfig().adminClaim || m.common_unknown()}</code>
									</div>
									<div class="flex items-center justify-between py-2">
										<span class="font-medium">{m.oidc_admin_value_label()}</span>

										<code class="bg-muted rounded px-2 py-1 text-xs">{parsedOidcConfig().adminValue || m.common_unknown()}</code>
									</div>
								</div>
							{:else}
								<div class="py-6 text-center">
									<TriangleAlertIcon class="text-destructive mx-auto mb-2 size-8" />
									<p class="text-destructive text-sm font-medium">{m.oidc_env_config_missing()}</p>
								</div>
							{/if}

							<div class="bg-muted/40 mt-4 rounded-md p-3">
								<p class="text-muted-foreground text-xs">{m.oidc_env_changes_note()}</p>
							</div>
						</div>
					{:else}
						<!-- Editable form -->
						<div class="bg-card/50 rounded-lg border p-4">
							<h3 class="text-base font-semibold">{m.oidc_basic_configuration_title()}</h3>
							<p class="text-muted-foreground mb-4 mt-1 text-sm">{m.oidc_basic_description()}</p>

							<div class="space-y-4">
								<div class="space-y-2">
									<Label for="oidcClientId" class="text-sm font-medium">{m.oidc_client_id_label()}</Label>
									<Input
										id="oidcClientId"
										bind:value={oidcForm.clientId}
										placeholder={m.oidc_client_id_placeholder()}
										class="font-mono text-sm"
									/>
								</div>

								<div class="space-y-2">
									<Label for="oidcClientSecret" class="text-sm font-medium">{m.oidc_client_secret_label()}</Label>
									<Input
										id="oidcClientSecret"
										type="password"
										bind:value={oidcForm.clientSecret}
										placeholder={m.oidc_client_secret_placeholder()}
										class="font-mono text-sm"
									/>
								</div>

								<div class="space-y-2">
									<Label for="oidcIssuerUrl" class="text-sm font-medium">{m.oidc_issuer_url_label()}</Label>
									<Input
										id="oidcIssuerUrl"
										bind:value={oidcForm.issuerUrl}
										placeholder={m.oidc_issuer_url_placeholder()}
										class="font-mono text-sm"
									/>
									<p class="text-muted-foreground text-xs">{m.oidc_issuer_url_description()}</p>
								</div>

								<div class="space-y-2">
									<Label for="oidcScopes" class="text-sm font-medium">{m.oidc_scopes_label()}</Label>
									<Input
										id="oidcScopes"
										bind:value={oidcForm.scopes}
										placeholder={m.oidc_scopes_placeholder()}
										class="font-mono text-sm"
									/>
								</div>

								<div class="pt-2">
									<h4 class="text-sm font-semibold">{m.oidc_admin_role_mapping_title()}</h4>
									<p class="text-muted-foreground mb-3 text-xs">{m.oidc_admin_role_mapping_description()}</p>
									<div class="grid gap-3 sm:grid-cols-2">
										<div class="space-y-2">
											<Label for="oidcAdminClaim" class="text-sm font-medium">{m.oidc_admin_claim_label()}</Label>
											<Input
												id="oidcAdminClaim"
												bind:value={oidcForm.adminClaim}
												placeholder={m.oidc_admin_claim_placeholder()}
												class="font-mono text-sm"
											/>
										</div>
										<div class="space-y-2">
											<Label for="oidcAdminValue" class="text-sm font-medium">{m.oidc_admin_value_label()}</Label>
											<Input
												id="oidcAdminValue"
												bind:value={oidcForm.adminValue}
												placeholder={m.oidc_admin_value_placeholder()}
												class="font-mono text-sm"
											/>
											<p class="text-muted-foreground text-[11px]">{m.oidc_admin_value_help()}</p>
										</div>
									</div>
								</div>
							</div>
						</div>

						{#if oidcStatus.envForced && !oidcStatus.envConfigured}
							<div class="border-destructive/50 bg-destructive/5 rounded-lg border p-4">
								<div class="mb-2 flex items-center gap-2">
									<TriangleAlertIcon class="text-destructive size-4" />
									<h3 class="text-destructive text-base font-semibold">{m.oidc_server_override_warning_title()}</h3>
								</div>
								<p class="text-destructive/80 text-sm">{m.security_server_forces_oidc_missing_env()}</p>
							</div>
						{/if}
					{/if}

					<!-- Redirect URI -->
					<div class="bg-card/50 rounded-lg border p-4">
						<div class="mb-3 flex items-center gap-2">
							<InfoIcon class="size-4 text-blue-600" />
							<h3 class="text-base font-semibold">{m.oidc_redirect_uri_title()}</h3>
						</div>
						<p class="text-muted-foreground mb-3 text-sm">{m.oidc_redirect_uri_description()}</p>
						<div class="flex items-center gap-2">
							<code class="bg-muted flex-1 break-all rounded p-2 font-mono text-xs">{redirectUri}</code>
							<Button
								size="sm"
								variant="outline"
								onclick={() => copyToClipboard(redirectUri)}
								class="flex-shrink-0"
								title={m.common_copy()}
							>
								<CopyIcon class="size-3" />
							</Button>
						</div>
					</div>
				</div>

				<!-- Side panel (lg+) -->
				<aside class="hidden lg:block">
					<div class="sticky top-4 space-y-4">
						<div class="bg-card/50 rounded-lg border p-4">
							<h3 class="text-sm font-semibold">{m.oidc_status_title()}</h3>
							<ul class="mt-3 space-y-2">
								{#each statusItems as s (s.label)}
									<li class="flex items-center justify-between">
										<span class="text-muted-foreground text-sm">{s.label}</span>
										<span class={'rounded px-2 py-0.5 text-xs ring-1 ' + chipClass(s.value)}
											>{s.value ? m.common_yes() : m.common_no()}</span
										>
									</li>
								{/each}
							</ul>
						</div>

						<div class="bg-card/50 rounded-lg border p-4">
							<h3 class="text-sm font-semibold">{m.oidc_tips_title()}</h3>
							<ul class="text-muted-foreground mt-3 list-disc space-y-1 pl-5 text-xs">
								<li>{m.oidc_tip_1()}</li>
								<li>{m.oidc_tip_2()}</li>
								<li>{m.oidc_tip_3()}</li>
							</ul>
						</div>
					</div>
				</aside>
			</div>
		</div>

		<Dialog.Footer class="flex flex-shrink-0 gap-3 border-t pt-4">
			<Button variant="outline" onclick={() => (open = false)} class="flex-1 sm:flex-none">{m.common_close()}</Button>
			{#if !isOidcViewMode}
				<Button onclick={onSave} class="flex-1 sm:flex-none">
					<KeyIcon class="mr-2 size-4" />
					{m.common_save()}
				</Button>
			{/if}
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
