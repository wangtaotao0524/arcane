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
    ServerCog,
    Sliders,
    Palette,
  } from "@lucide/svelte";
  import * as Switch from "$lib/components/ui/switch/index.js";
  import { enhance } from "$app/forms";
  import * as Alert from "$lib/components/ui/alert/index.js";

  let { data, form }: { data: PageData; form: ActionData } = $props();

  // Initialize state variables *only* from the initial data load
  let dockerHost = $state(data.settings.dockerHost);
  let autoRefresh = $state(data.settings.autoRefresh);
  let refreshInterval = $state(data.settings.refreshInterval);
  let darkMode = $state(data.settings.darkMode);

  let isSubmitting = $state(false);
  let testStatus: "idle" | "testing" | "success" | "error" = $state("idle");
  let testMessage: string | null = $state(null);

  // Keep the effect for resetting test status when dockerHost input changes
  $effect(() => {
    const currentHost = dockerHost; // Create dependency on dockerHost
    if (currentHost !== undefined) {
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

<div class="space-y-6">
  <!-- Header -->
  <div class="flex flex-col sm:flex-row justify-between sm:items-center gap-4">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">Settings</h1>
      <p class="text-sm text-muted-foreground mt-1">
        Configure Arcane's connection and behavior
      </p>
    </div>
    <Button
      type="submit"
      form="settings-form"
      disabled={isSubmitting || testStatus === "testing"}
      class="h-10"
    >
      <Save class="w-4 h-4 mr-2" />
      {isSubmitting ? "Saving..." : "Save Settings"}
    </Button>
  </div>

  <!-- Alerts -->
  {#if form?.error}
    <Alert.Root variant="destructive">
      <XCircle class="h-4 w-4 mr-2" />
      <Alert.Title>Error Saving Settings</Alert.Title>
      <Alert.Description>{form.error}</Alert.Description>
    </Alert.Root>
  {/if}

  {#if form?.success}
    <Alert.Root
      variant="default"
      class="bg-primary/10 text-primary border border-primary"
    >
      <CheckCircle class="h-4 w-4 mr-2" />
      <Alert.Title>Settings Saved</Alert.Title>
      <Alert.Description
        >Your settings have been updated successfully.</Alert.Description
      >
    </Alert.Root>
  {/if}

  <!-- Settings Form -->
  <form
    id="settings-form"
    method="POST"
    use:enhance={() => {
      isSubmitting = true;
      testStatus = "idle";
      testMessage = null;
      return async ({ result, update }) => {
        isSubmitting = false;

        // IMPORTANT: Update local state *before* calling update()
        if (result.type === "success" && result.data?.settings) {
          const newSettings = result.data.settings as typeof data.settings;
          dockerHost = newSettings.dockerHost;
          autoRefresh = newSettings.autoRefresh;
          refreshInterval = newSettings.refreshInterval;
          darkMode = newSettings.darkMode;
        }

        await update({ reset: false });
      };
    }}
  >
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Docker Connection Card -->
      <Card.Root class="border shadow-sm">
        <Card.Header class="pb-3">
          <div class="flex items-center gap-2">
            <div class="bg-blue-500/10 p-2 rounded-full">
              <ServerCog class="h-5 w-5 text-blue-500" />
            </div>
            <div>
              <Card.Title>Docker Connection</Card.Title>
              <Card.Description
                >Configure Docker host connection</Card.Description
              >
            </div>
          </div>
        </Card.Header>
        <Card.Content class="space-y-6">
          <div class="space-y-2">
            <Label for="dockerHost" class="text-sm font-medium"
              >Docker Host</Label
            >
            <Input
              id="dockerHost"
              name="dockerHost"
              bind:value={dockerHost}
              placeholder="unix:///var/run/docker.sock"
              class="mt-1"
              disabled={testStatus === "testing"}
            />
            <p class="text-xs text-muted-foreground">
              Enter Docker host URL (e.g., unix:///var/run/docker.sock or
              tcp://localhost:2375)
            </p>
          </div>

          <div>
            <Button
              variant="secondary"
              class="w-full sm:w-auto"
              onclick={testConnection}
              type="button"
              disabled={testStatus === "testing" || !dockerHost}
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
              <div
                class="flex items-center gap-2 text-sm text-green-600 mt-3 bg-green-50 p-3 rounded-md border border-green-200"
              >
                <CheckCircle class="w-4 h-4 flex-shrink-0" />
                <span>{testMessage || "Connection successful!"}</span>
              </div>
            {:else if testStatus === "error"}
              <div
                class="flex items-center gap-2 text-sm text-destructive mt-3 bg-destructive/10 p-3 rounded-md border border-destructive/20"
              >
                <XCircle class="w-4 h-4 flex-shrink-0" />
                <span>{testMessage || "Connection failed."}</span>
              </div>
            {/if}
          </div>
        </Card.Content>
      </Card.Root>

      <!-- Application Settings Card -->
      <div class="space-y-6">
        <!-- Auto Refresh Settings Card -->
        <Card.Root class="border shadow-sm">
          <Card.Header class="pb-3">
            <div class="flex items-center gap-2">
              <div class="bg-amber-500/10 p-2 rounded-full">
                <Sliders class="h-5 w-5 text-amber-500" />
              </div>
              <div>
                <Card.Title>Auto Refresh</Card.Title>
                <Card.Description
                  >Control data refresh behavior</Card.Description
                >
              </div>
            </div>
          </Card.Header>
          <Card.Content class="space-y-6">
            <div
              class="flex items-center justify-between rounded-lg border p-4 bg-muted/30"
            >
              <div class="space-y-0.5">
                <Label for="autoRefreshSwitch" class="text-base font-medium"
                  >Auto Refresh</Label
                >
                <p class="text-sm text-muted-foreground">
                  Automatically refresh data periodically
                </p>
              </div>
              <Switch.Root
                id="autoRefreshSwitch"
                name="autoRefresh"
                bind:checked={autoRefresh}
              />
            </div>

            {#if autoRefresh}
              <div class="space-y-2 px-1">
                <Label for="refreshInterval" class="text-sm font-medium"
                  >Refresh Interval (seconds)</Label
                >
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
                {:else}
                  <p class="text-xs text-muted-foreground">
                    Set between 5-60 seconds. Lower values increase server load.
                  </p>
                {/if}
              </div>
            {/if}
          </Card.Content>
        </Card.Root>

        <!-- UI Settings Card -->
        <Card.Root class="border shadow-sm">
          <Card.Header class="pb-3">
            <div class="flex items-center gap-2">
              <div class="bg-purple-500/10 p-2 rounded-full">
                <Palette class="h-5 w-5 text-purple-500" />
              </div>
              <div>
                <Card.Title>UI Settings</Card.Title>
                <Card.Description>Customize the appearance</Card.Description>
              </div>
            </div>
          </Card.Header>
          <Card.Content>
            <div
              class="flex items-center justify-between rounded-lg border p-4 bg-muted/30"
            >
              <div class="space-y-0.5">
                <Label for="darkModeSwitch" class="text-base font-medium"
                  >Dark Mode</Label
                >
                <p class="text-sm text-muted-foreground">
                  Enable the dark color theme
                </p>
              </div>
              <Switch.Root
                id="darkModeSwitch"
                name="darkMode"
                bind:checked={darkMode}
              />
            </div>
          </Card.Content>
        </Card.Root>
      </div>
    </div>
  </form>
</div>
