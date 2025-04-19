<script lang="ts">
  import type { PageData, ActionData } from "./$types";
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Input } from "$lib/components/ui/input/index.js";
  import { Label } from "$lib/components/ui/label/index.js";
  import {
    Settings,
    Save,
    RefreshCw,
    CheckCircle,
    XCircle,
  } from "@lucide/svelte";
  import * as Switch from "$lib/components/ui/switch/index.js";
  import { enhance } from "$app/forms";

  // Use $props() for component inputs
  let { data, form }: { data: PageData; form: ActionData } = $props();

  // Derive initial settings from props, reacting if form action updates props
  let initialSettings = $derived(form?.settings ?? data.settings);

  // Use $state for mutable component state bound to inputs
  let dockerHost = $state("");
  let autoRefresh = $state(false);
  let refreshInterval = $state(30);
  let darkMode = $state(false);

  let isSubmitting = $state(false);
  let testStatus: "idle" | "testing" | "success" | "error" = $state("idle");
  let testMessage: string | null = $state(null);

  // Effect to update local state when initialSettings (derived from props) changes
  $effect(() => {
    // Avoid resetting if the values are already the same (prevents potential loops)
    if (dockerHost !== initialSettings.dockerHost)
      dockerHost = initialSettings.dockerHost;
    if (autoRefresh !== initialSettings.autoRefresh)
      autoRefresh = initialSettings.autoRefresh;
    if (refreshInterval !== initialSettings.refreshInterval)
      refreshInterval = initialSettings.refreshInterval;
    if (darkMode !== initialSettings.darkMode)
      darkMode = initialSettings.darkMode;
  });

  // Effect to reset test status when dockerHost input changes
  $effect(() => {
    // This effect now correctly tracks the $state variable 'dockerHost'
    // Need to access dockerHost to create the dependency
    const currentHost = dockerHost;
    if (currentHost !== undefined) {
      // Check if initialized
      testStatus = "idle";
      testMessage = null;
    }
  });

  async function testConnection() {
    testStatus = "testing";
    testMessage = null;
    try {
      const hostParam = encodeURIComponent(dockerHost);
      const response = await fetch(
        `/api/docker/test-connection?host=${hostParam}`
      );
      const result = await response.json();

      if (response.ok && result.success) {
        testStatus = "success";
        testMessage = result.message || "Connection successful!";
      } else {
        testStatus = "error";
        testMessage =
          result.error ||
          "Connection failed. Check Docker status and host setting.";
      }
    } catch (error: any) {
      testStatus = "error";
      testMessage =
        "Failed to run connection test: " + (error.message || "Unknown error");
      console.error("Error during connection test fetch:", error);
    }
  }
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Settings</h1>
</div>

{#if form?.error}
  <div
    class="mb-4 p-4 bg-destructive/10 text-destructive border border-destructive rounded-md"
  >
    <p><strong>Error saving settings:</strong> {form.error}</p>
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
    testStatus = "idle"; // Reset test status on save attempt
    testMessage = null;
    return async ({ update }) => {
      await update({ reset: false }); // Use reset: false to prevent SvelteKit from resetting form state automatically
      isSubmitting = false;
      // The $effect will handle updating local state from props automatically
    };
  }}
>
  <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
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
            disabled={testStatus === "testing"}
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
          disabled={testStatus === "testing"}
        >
          {#if testStatus === "testing"}
            <RefreshCw class="w-4 h-4 mr-2 animate-spin" />
            Testing...
          {:else}
            <RefreshCw class="w-4 h-4 mr-2" />
            Test Connection
          {/if}
        </Button>

        {#if testStatus === "success"}
          <div class="flex items-center gap-2 text-sm text-green-600 mt-2">
            <CheckCircle class="w-4 h-4" />
            <span>{testMessage || "Connection successful!"}</span>
          </div>
        {:else if testStatus === "error"}
          <div class="flex items-center gap-2 text-sm text-destructive mt-2">
            <XCircle class="w-4 h-4" />
            <span>{testMessage || "Connection failed."}</span>
          </div>
        {/if}
      </Card.Content>
    </Card.Root>

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
              {#if form?.error && form.values?.refreshInterval && (parseInt(String(form.values.refreshInterval), 10) < 5 || parseInt(String(form.values.refreshInterval), 10) > 60)}
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

  <div class="mt-6 flex justify-end">
    <Button type="submit" disabled={isSubmitting || testStatus === "testing"}>
      <Save class="w-4 h-4 mr-2" />
      {isSubmitting ? "Saving..." : "Save Settings"}
    </Button>
  </div>
</form>
