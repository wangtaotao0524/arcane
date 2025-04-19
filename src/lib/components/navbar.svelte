<script lang="ts">
  import {
    Home,
    Box,
    Image,
    Network,
    HardDrive,
    Settings,
    Menu,
    X,
    type Icon as IconType,
  } from "@lucide/svelte";
  import { page } from "$app/stores";
  import { fly } from "svelte/transition";
  import { Button } from "$lib/components/ui/button/index.js";

  // Define MenuItem type with proper IconType
  type MenuItem = {
    href: string;
    label: string;
    icon: typeof IconType;
  };

  // State for mobile sidebar
  let isOpen = $state(false);

  // Menu items with proper typing
  let {
    items = [
      { href: "/", label: "Dashboard", icon: Home },
      { href: "/containers", label: "Containers", icon: Box },
      { href: "/images", label: "Images", icon: Image },
      { href: "/networks", label: "Networks", icon: Network },
      { href: "/volumes", label: "Volumes", icon: HardDrive },
      { href: "/settings", label: "Settings", icon: Settings },
    ] as MenuItem[],
  } = $props();
</script>

<!-- Mobile menu button -->
<Button
  variant="default"
  size="icon"
  class="md:hidden fixed top-4 right-4 z-50"
  onclick={() => (isOpen = !isOpen)}
>
  {#if isOpen}
    <X size={20} />
  {:else}
    <Menu size={20} />
  {/if}
</Button>

<!-- Sidebar navigation -->
<div
  class={`
  ${isOpen ? "translate-x-0" : "-translate-x-full md:translate-x-0"} 
  fixed md:sticky top-0 left-0 h-screen w-64 
  bg-card border-r border-border shadow-md 
  z-40 transition-transform duration-200 ease-in-out
`}
>
  <!-- Logo/App name -->
  <div class="flex items-center gap-2 p-4 border-b border-border">
    <span class="text-2xl">🐳</span>
    <span class="text-xl font-bold">Arcane</span>
  </div>

  <!-- Navigation links -->
  <nav class="p-2">
    {#each items as item}
      {@const isActive = $page.url.pathname === item.href}
      {@const Icon = item.icon}

      <a
        href={item.href}
        class={`flex items-center gap-3 px-3 py-2 my-1 rounded-md text-sm font-medium
        ${
          isActive
            ? "bg-primary/10 text-primary"
            : "text-foreground hover:bg-accent hover:text-accent-foreground"
        } 
        transition-colors`}
      >
        <Icon size={18} />
        <span>{item.label}</span>
      </a>
    {/each}
  </nav>
</div>

<!-- Overlay for mobile -->
{#if isOpen}
  <button
    type="button"
    class="md:hidden fixed inset-0 w-full h-full bg-background/80 backdrop-blur-sm z-30 border-none"
    aria-label="Close menu"
    onclick={() => (isOpen = false)}
    transition:fly={{ duration: 150, y: -5 }}
  ></button>
{/if}
