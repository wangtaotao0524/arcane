<script lang="ts">
  import "../app.css";
  import { ModeWatcher } from "mode-watcher";
  import Nav from "$lib/components/navbar.svelte";
  let { children } = $props();
</script>

<ModeWatcher />

<div class="flex min-h-screen bg-background">
  <Nav />

  <main class="flex-1">
    <header
      class="px-6 py-4 border-b border-border bg-card/80 backdrop-blur flex items-center justify-between shadow-sm"
    >
      <h1 class="text-xl font-bold tracking-tight">Docker Management UI</h1>
      <div class="flex items-center gap-4">
        <!-- Placeholder for actions (theme toggle, profile, etc.) -->
      </div>
    </header>

    <section class="p-6">
      {@render children()}
    </section>
  </main>
</div>
