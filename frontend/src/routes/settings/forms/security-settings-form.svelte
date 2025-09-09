<script lang="ts">
	import { z } from 'zod/v4';
	import { createForm, preventDefault } from '$lib/utils/form.utils';
	import * as FieldSet from '$lib/components/ui/field-set';
	import { Button } from '$lib/components/ui/button';
	import FormInput from '$lib/components/form/form-input.svelte';
	import SwitchWithLabel from '$lib/components/form/labeled-switch.svelte';
	import OidcConfigDialog from '$lib/components/dialogs/oidc-config-dialog.svelte';
	import { toast } from 'svelte-sonner';
	import type { Settings } from '$lib/types/settings.type';
	import type { OidcStatusInfo } from '$lib/types/settings.type';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import { m } from '$lib/paraglide/messages';

	let {
		settings,
		oidcStatus,
		callback
	}: {
		settings: Settings;
		oidcStatus: OidcStatusInfo;
		callback: (appConfig: Partial<Settings>) => Promise<void>;
	} = $props();

	let isLoading = $state({ saving: false });
	let showOidcConfigDialog = $state(false);

	let oidcConfigForm = $state({
		clientId: '',
		clientSecret: '',
		issuerUrl: '',
		scopes: 'openid email profile',
		adminClaim: '',
		adminValue: ''
	});

	const formSchema = z
		.object({
			authLocalEnabled: z.boolean(),
			authOidcEnabled: z.boolean(),
			authSessionTimeout: z
				.number(m.security_session_timeout_required())
				.int(m.security_session_timeout_integer())
				.min(15, m.security_session_timeout_min())
				.max(1440, m.security_session_timeout_max()),
			authPasswordPolicy: z.enum(['basic', 'standard', 'strong'])
		})
		.superRefine((data, ctx) => {
			// If server forces OIDC, the constraint is already satisfied
			if (oidcStatus.envForced) return;
			if (!data.authLocalEnabled && !data.authOidcEnabled) {
				ctx.addIssue({
					code: z.ZodIssueCode.custom,
					message: m.security_enable_one_provider(),
					path: ['authLocalEnabled']
				});
			}
		});

	let { inputs: formInputs, ...form } = $derived(createForm<typeof formSchema>(formSchema, settings));

	// Helper: treat OIDC as active if forced by server or enabled in form
	const isOidcActive = () => $formInputs.authOidcEnabled.value || oidcStatus.envForced;

	async function onSubmit() {
		const data = form.validate();
		if (!data) return;

		isLoading.saving = true;

		let authOidcConfig = settings.authOidcConfig;
		if (data.authOidcEnabled && !oidcStatus.envForced) {
			authOidcConfig = JSON.stringify({
				clientId: oidcConfigForm.clientId,
				clientSecret: oidcConfigForm.clientSecret || '',
				issuerUrl: oidcConfigForm.issuerUrl,
				scopes: oidcConfigForm.scopes,
				adminClaim: oidcConfigForm.adminClaim || '',
				adminValue: oidcConfigForm.adminValue || ''
			});
		}

		await callback({
			authLocalEnabled: data.authLocalEnabled,
			authOidcEnabled: data.authOidcEnabled,
			authSessionTimeout: data.authSessionTimeout,
			authPasswordPolicy: data.authPasswordPolicy
		})
			.then(() => toast.success(m.security_settings_saved()))
			.finally(() => (isLoading.saving = false));
	}

	// Only depend on envForced; open config when enabling and not forced
	function handleOidcSwitchChange(checked: boolean) {
		$formInputs.authOidcEnabled.value = checked;

		if (!checked && !$formInputs.authLocalEnabled.value && !oidcStatus.envForced) {
			$formInputs.authLocalEnabled.value = true;
			toast.info(m.security_local_enabled_info());
		}

		if (checked && !oidcStatus.envForced) {
			showOidcConfigDialog = true;
		}
	}

	function handleLocalSwitchChange(checked: boolean) {
		if (!checked && !isOidcActive()) {
			$formInputs.authLocalEnabled.value = true;
			toast.error(m.security_enable_one_provider_error());
			return;
		}
		$formInputs.authLocalEnabled.value = checked;
	}

	function openOidcDialog() {
		if (settings.authOidcConfig) {
			const cfg = JSON.parse(settings.authOidcConfig);
			oidcConfigForm.clientId = cfg.clientId || '';
			oidcConfigForm.issuerUrl = cfg.issuerUrl || '';
			oidcConfigForm.scopes = cfg.scopes || 'openid email profile';
			oidcConfigForm.adminClaim = cfg.adminClaim || '';
			oidcConfigForm.adminValue = cfg.adminValue || '';
		}
		oidcConfigForm.clientSecret = '';
		showOidcConfigDialog = true;
	}

	async function handleSaveOidcConfig() {
		try {
			isLoading.saving = true;
			$formInputs.authOidcEnabled.value = true;

			const data = form.validate();
			if (!data) {
				isLoading.saving = false;
				return;
			}

			const authOidcConfig = JSON.stringify({
				clientId: oidcConfigForm.clientId,
				clientSecret: oidcConfigForm.clientSecret || '',
				issuerUrl: oidcConfigForm.issuerUrl,
				scopes: oidcConfigForm.scopes,
				adminClaim: oidcConfigForm.adminClaim || '',
				adminValue: oidcConfigForm.adminValue || ''
			});

			await callback({
				authOidcEnabled: true,
				authOidcConfig
			});

			toast.success(m.security_oidc_saved());
			showOidcConfigDialog = false;
		} finally {
			isLoading.saving = false;
		}
	}
</script>

<form onsubmit={preventDefault(onSubmit)} class="space-y-6">
	<div class="w-full p-6">
		<FieldSet.Root>
			<FieldSet.Content class="flex flex-col gap-8">
				<div class="min-w-0 space-y-4">
					<h2 class="text-muted-foreground text-sm font-semibold">{m.security_authentication_heading()}</h2>

					<SwitchWithLabel
						id="localAuthSwitch"
						label={m.security_local_auth_label()}
						description={m.security_local_auth_description()}
						bind:checked={$formInputs.authLocalEnabled.value}
						onCheckedChange={handleLocalSwitchChange}
					/>

					<div class="space-y-2">
						<SwitchWithLabel
							id="oidcAuthSwitch"
							label={m.security_oidc_auth_label()}
							description={oidcStatus.envForced ? m.security_oidc_auth_description_forced() : m.security_oidc_auth_description()}
							disabled={oidcStatus.envForced}
							checked={$formInputs.authOidcEnabled.value}
							onCheckedChange={handleOidcSwitchChange}
						/>

						{#if isOidcActive()}
							<div class="pl-11">
								{#if oidcStatus.envForced}
									{#if !oidcStatus.envConfigured}
										<Button variant="link" class="text-destructive h-auto p-0 text-xs hover:underline" onclick={openOidcDialog}>
											{m.security_server_forces_oidc_missing_env()}
										</Button>
									{:else}
										<Button variant="link" class="h-auto p-0 text-xs text-sky-600 hover:underline" onclick={openOidcDialog}>
											{m.security_oidc_configured_forced_view()}
										</Button>
									{/if}
								{:else}
									<Button variant="link" class="h-auto p-0 text-xs text-sky-600 hover:underline" onclick={openOidcDialog}>
										{m.security_manage_oidc_config()}
									</Button>
								{/if}
							</div>
						{/if}
					</div>
				</div>

				<div class="min-w-0 space-y-4">
					<h2 class="text-muted-foreground text-sm font-semibold">{m.security_session_heading()}</h2>

					<FormInput
						type="number"
						id="sessionTimeout"
						label={m.security_session_timeout_label()}
						placeholder={m.security_session_timeout_placeholder()}
						bind:input={$formInputs.authSessionTimeout}
						description={m.security_session_timeout_description()}
					/>

					<div class="space-y-2">
						<span class="text-sm font-medium" id="passwordPolicyLabel">{m.security_password_policy_label()}</span>
						<Tooltip.Provider>
							<div class="mt-2 grid grid-cols-3 gap-2" role="group" aria-labelledby="passwordPolicyLabel">
								<Tooltip.Root>
									<Tooltip.Trigger>
										<Button
											variant={$formInputs.authPasswordPolicy.value === 'basic' ? 'default' : 'outline'}
											class={$formInputs.authPasswordPolicy.value === 'basic'
												? 'arcane-button-create w-full'
												: 'arcane-button-restart w-full'}
											onclick={() => ($formInputs.authPasswordPolicy.value = 'basic')}
											type="button"
											>{m.security_password_policy_basic()}
										</Button>
									</Tooltip.Trigger>
									<Tooltip.Content side="top" align="center">{m.security_password_policy_basic_tooltip()}</Tooltip.Content>
								</Tooltip.Root>

								<Tooltip.Root>
									<Tooltip.Trigger>
										<Button
											variant={$formInputs.authPasswordPolicy.value === 'standard' ? 'default' : 'outline'}
											class={$formInputs.authPasswordPolicy.value === 'standard'
												? 'arcane-button-create w-full'
												: 'arcane-button-restart w-full'}
											onclick={() => ($formInputs.authPasswordPolicy.value = 'standard')}
											type="button"
											>{m.security_password_policy_standard()}
										</Button>
									</Tooltip.Trigger>
									<Tooltip.Content side="top" align="center">{m.security_password_policy_standard_tooltip()}</Tooltip.Content>
								</Tooltip.Root>

								<Tooltip.Root>
									<Tooltip.Trigger>
										<Button
											variant={$formInputs.authPasswordPolicy.value === 'strong' ? 'default' : 'outline'}
											class={$formInputs.authPasswordPolicy.value === 'strong'
												? 'arcane-button-create w-full'
												: 'arcane-button-restart w-full'}
											onclick={() => ($formInputs.authPasswordPolicy.value = 'strong')}
											type="button"
											>{m.security_password_policy_strong()}
										</Button>
									</Tooltip.Trigger>
									<Tooltip.Content side="top" align="center">{m.security_password_policy_strong_tooltip()}</Tooltip.Content>
								</Tooltip.Root>
							</div>
						</Tooltip.Provider>
					</div>
				</div>
			</FieldSet.Content>

			<FieldSet.Footer>
				<div class="flex w-full place-items-center justify-between">
					<span class="text-muted-foreground text-sm">{m.security_save_instructions()}</span>
					<Button type="submit" disabled={isLoading.saving} size="sm"
						>{isLoading.saving ? m.common_saving() : m.common_save()}</Button
					>
				</div>
			</FieldSet.Footer>
		</FieldSet.Root>
	</div>

	<OidcConfigDialog
		bind:open={showOidcConfigDialog}
		currentSettings={settings}
		{oidcStatus}
		bind:oidcForm={oidcConfigForm}
		onSave={handleSaveOidcConfig}
	/>
</form>
