<!-- filepath: /Users/kylemendell/dev/ofkm/arcane/src/routes/volumes/VolumeActions.svelte -->
<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import { Trash2 } from "@lucide/svelte";

  export let name: string;

  function deleteVolume() {
    // TODO: Implement volume deletion logic
    alert(`Implement deletion for volume: ${name}`);
  }
</script>

<Button
  variant="destructive"
  size="sm"
  onclick={deleteVolume}
  title="Delete Volume"
>
  <Trash2 class="w-4 h-4" />
</Button>
<!-- Add other actions like inspect if needed -->
