<script lang="ts">
  import "../app.css";
  import { ModeWatcher } from "mode-watcher";
  import Nav from "$lib/components/navbar.svelte";
  let { children } = $props();
</script>

<svelte:head><title>Arcane</title></svelte:head>

<ModeWatcher />

<div class="flex min-h-screen bg-background">
  <Nav />

  <main class="flex-1">
    <section class="p-6">
      {@render children()}
    </section>
  </main>
</div>
