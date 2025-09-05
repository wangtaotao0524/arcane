<script lang="ts">
	import { z } from 'zod/v4';
	import { createForm, preventDefault } from '$lib/utils/form.utils';
	import { Button } from '$lib/components/ui/button';
	import * as FieldSet from '$lib/components/ui/field-set';
	import FormInput from '$lib/components/form/form-input.svelte';
	import type { Settings } from '$lib/types/settings.type';
	import { toast } from 'svelte-sonner';
	import SwitchWithLabel from '$lib/components/form/labeled-switch.svelte';

	let {
		settings,
		callback
	}: {
		settings: Settings;
		callback: (appConfig: Partial<Settings>) => Promise<void>;
	} = $props();

	let isLoading = $state(false);

	const formSchema = z.object({
		stacksDirectory: z.string().min(1, 'Projects directory is required'),
		baseServerUrl: z.string().min(1, 'Base server URL is required'),
		enableGravatar: z.boolean()
	});

	let { inputs: formInputs, ...form } = $derived(createForm<typeof formSchema>(formSchema, settings));

	async function onSubmit() {
		const data = form.validate();
		if (!data) return;
		isLoading = true;

		await callback(data)
			.then(() => toast.success('Settings Updated Successfully'))
			.finally(() => (isLoading = false));
	}
</script>

<form onsubmit={preventDefault(onSubmit)} class="space-y-6">
	<div class="w-full p-6">
		<FieldSet.Root>
			<FieldSet.Content class="grid grid-cols-1 gap-6 md:grid-cols-2">
				<div class="bg-background/40 min-w-0 space-y-4 rounded-lg border p-5 shadow-sm">
					<div class="space-y-1">
						<h3 class="text-base font-medium">Project Paths</h3>
						<p class="text-muted-foreground text-sm">Configure where Arcane looks for your Compose files.</p>
					</div>

					<FormInput
						label="Projects Directory"
						placeholder="data/projects"
						bind:input={$formInputs.stacksDirectory}
						helpText="Directory where Docker Compose files are stored (this is inside the container)"
					/>

					<FormInput
						label="Base Server URL"
						placeholder="localhost"
						bind:input={$formInputs.baseServerUrl}
						helpText="The Base URL for your docker host. Used when opening containers from the service links"
					/>
				</div>

				<div class="bg-background/40 min-w-0 space-y-4 rounded-lg border p-5 shadow-sm">
					<div class="space-y-1">
						<h3 class="text-base font-medium">User Avatars</h3>
						<p class="text-muted-foreground text-sm">Control Gravatar usage for profile images.</p>
					</div>

					<SwitchWithLabel
						id="enableGravatar"
						label="Enable Gravatar"
						description="Whether to use Gravatar-based avatars for user accounts"
						bind:checked={$formInputs.enableGravatar.value}
					/>
				</div>
			</FieldSet.Content>

			<FieldSet.Footer>
				<div class="flex w-full place-items-center justify-between">
					<span class="text-muted-foreground text-sm">Save your updated settings.</span>
					<Button type="submit" disabled={isLoading} size="sm">{isLoading ? 'Saving…' : 'Save'}</Button>
				</div>
			</FieldSet.Footer>
		</FieldSet.Root>
	</div>
</form>
