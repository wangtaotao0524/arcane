<script lang="ts">
  import { Badge } from "$lib/components/ui/badge/index.js"; // Assuming you use shadcn-svelte Badge

  type ContainerState =
    | "created"
    | "running"
    | "paused"
    | "restarting"
    | "removing"
    | "exited"
    | "dead"
    | string; // Allow other potential states

  let { state }: { state: ContainerState } = $props();

  // Determine badge variant and text based on state
  let variant: "default" | "secondary" | "destructive" | "outline" =
    $derived.by(() => {
      switch (state) {
        case "running":
          return "default"; // Typically green/primary in shadcn
        case "exited":
        case "dead":
          return "destructive"; // Red
        case "paused":
        case "restarting":
        case "created":
          return "secondary"; // Gray/Yellowish
        default:
          return "outline"; // Default outline for unknown states
      }
    });

  let text = $derived(state.charAt(0).toUpperCase() + state.slice(1)); // Capitalize state
</script>

<Badge {variant} class="capitalize">
  {text}
</Badge>
