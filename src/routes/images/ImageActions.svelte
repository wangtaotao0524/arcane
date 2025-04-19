<script lang="ts">
  import Ellipsis from "@lucide/svelte/icons/ellipsis";
  import { Button } from "$lib/components/ui/button/index.js";
  import * as DropdownMenu from "$lib/components/ui/dropdown-menu/index.js";
  import { Trash2 } from "@lucide/svelte"; // Import Trash icon

  // Accept image ID and potentially other relevant info
  let { id, repoTag }: { id: string; repoTag?: string } = $props();

  function removeImage() {
    // TODO: Implement API call to remove image
    console.log("Attempting to remove image:", id, repoTag);
    alert(`Implement remove action for image ID: ${id.substring(0, 12)}`);
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        variant="ghost"
        size="icon"
        class="relative size-8 p-0"
      >
        <span class="sr-only">Open menu</span>
        <Ellipsis />
      </Button>
    {/snippet}
  </DropdownMenu.Trigger>
  <DropdownMenu.Content>
    <DropdownMenu.Group>
      <DropdownMenu.Label>Actions</DropdownMenu.Label>
      <DropdownMenu.Separator />
      <DropdownMenu.Item
        onclick={removeImage}
        class="text-destructive focus:text-destructive"
      >
        <Trash2 class="mr-2 h-4 w-4" />
        Remove Image
      </DropdownMenu.Item>
      <!-- Add other actions if needed -->
    </DropdownMenu.Group>
  </DropdownMenu.Content>
</DropdownMenu.Root>
