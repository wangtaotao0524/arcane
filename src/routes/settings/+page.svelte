<script lang="ts">
  import type { PageData, ActionData } from "./$types";
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import { Settings, Save, RefreshCw } from "@lucide/svelte";
  import * as Switch from "$lib/components/ui/switch/index.js";
  import { enhance } from "$app/forms"; // Optional: for progressive enhancement

  export let data: PageData;
  export let form: ActionData; // To get data back from the form action

  // Initialize state from loaded data or form action result
  let settings = form?.settings ?? data.settings;

  let dockerHost = settings.dockerHost;
  let autoRefresh = settings.autoRefresh;
  let refreshInterval = settings.refreshInterval;
  let darkMode = settings.darkMode;

  let isSubmitting = false; // Manual submission state

  // Test connection function
  function testConnection() {
    console.log("Testing connection to", dockerHost);
    // TODO: Implement actual connection test logic
  }
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Settings</h1>
  <!-- Button moved inside the form below -->
</div>

{#if form?.error}
  <div
    class="mb-4 p-4 bg-destructive/10 text-destructive border border-destructive rounded-md"
  >
    <p><strong>Error:</strong> {form.error}</p>
  </div>
{/if}
{#if form?.success}
  <div
    class="mb-4 p-4 bg-primary/10 text-primary border border-primary rounded-md"
  >
    <p>Settings saved successfully!</p>
  </div>
{/if}

<form
  method="POST"
  use:enhance={() => {
    isSubmitting = true;
    return async ({ update }) => {
      await update(); // Update form data if action returns data
      isSubmitting = false;
    };
  }}
>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
    <!-- Docker Connection Settings -->
    <Card.Root>
      <Card.Header>
        <Card.Title>Docker Connection</Card.Title>
        <Card.Description>Configure Docker host connection</Card.Description>
      </Card.Header>
      <Card.Content class="space-y-4">
        <div>
          <Label for="dockerHost">Docker Host</Label>
          <Input
            id="dockerHost"
            name="dockerHost"
            bind:value={dockerHost}
            placeholder="unix:///var/run/docker.sock"
            class="mt-1"
          />
          <p class="text-sm text-muted-foreground mt-1">
            Enter Docker host URL (e.g., unix:///var/run/docker.sock or
            tcp://localhost:2375)
          </p>
        </div>

        <Button
          variant="secondary"
          class="mt-2"
          onclick={testConnection}
          type="button"
        >
          <RefreshCw class="w-4 h-4 mr-2" />
          Test Connection
        </Button>
      </Card.Content>
    </Card.Root>

    <!-- Application Settings -->
    <Card.Root>
      <Card.Header>
        <Card.Title>Application Settings</Card.Title>
        <Card.Description>Configure Arcane behavior</Card.Description>
      </Card.Header>
      <Card.Content>
        <div class="space-y-6">
          <div class="flex items-center justify-between rounded-lg border p-4">
            <div class="space-y-0.5">
              <Label for="autoRefreshSwitch" class="text-base"
                >Auto Refresh</Label
              >
              <p class="text-sm text-muted-foreground">
                Automatically refresh data periodically.
              </p>
            </div>
            <Switch.Root
              id="autoRefreshSwitch"
              name="autoRefresh"
              bind:checked={autoRefresh}
            />
          </div>

          {#if autoRefresh}
            <div class="space-y-1">
              <Label for="refreshInterval">Refresh Interval (seconds)</Label>
              <Input
                id="refreshInterval"
                name="refreshInterval"
                type="number"
                bind:value={refreshInterval}
                min="5"
                max="60"
              />
              {#if form?.error && form.values?.refreshInterval && typeof form.values.refreshInterval === "string" && (parseInt(form.values.refreshInterval, 10) < 5 || parseInt(form.values.refreshInterval, 10) > 60)}
                <p class="text-sm text-destructive">
                  Must be between 5 and 60.
                </p>
              {/if}
            </div>
          {/if}

          <div class="flex items-center justify-between rounded-lg border p-4">
            <div class="space-y-0.5">
              <Label for="darkModeSwitch" class="text-base">Dark Mode</Label>
              <p class="text-sm text-muted-foreground">
                Enable the dark color theme.
              </p>
            </div>
            <Switch.Root
              id="darkModeSwitch"
              name="darkMode"
              bind:checked={darkMode}
            />
          </div>
        </div>
      </Card.Content>
    </Card.Root>
  </div>

  <!-- Submit Button inside the form -->
  <div class="mt-6 flex justify-end">
    <Button type="submit" disabled={isSubmitting}>
      <Save class="w-4 h-4 mr-2" />
      {isSubmitting ? "Saving..." : "Save Settings"}
    </Button>
  </div>
</form>
