<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import { Trash2 } from "@lucide/svelte";

  interface Props {
    id: string | undefined;
    name: string;
  }

  let { id, name }: Props = $props();

  function deleteNetwork() {
    // TODO: Implement network deletion logic
    alert(`Implement deletion for network: ${name} (ID: ${id ?? "N/A"})`);
  }
</script>

<Button
  variant="destructive"
  size="sm"
  onclick={deleteNetwork}
  title="Delete Network"
>
  <Trash2 class="w-4 h-4" />
</Button>
<!-- Add other actions like inspect if needed -->
