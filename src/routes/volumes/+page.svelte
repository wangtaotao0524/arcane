<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { HardDrive, Plus, Trash2 } from "@lucide/svelte";

  export let data;
  const { volumes } = data;
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Volumes</h1>
  <Button variant="default">
    <Plus class="w-4 h-4 mr-2" />
    Create Volume
  </Button>
</div>

<Card.Root>
  <Card.Header>
    <Card.Title>Docker Volumes</Card.Title>
    <Card.Description>Manage persistent data storage</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b">
            <th class="text-left py-2">Name</th>
            <th class="text-left py-2">Driver</th>
            <th class="text-left py-2">Mount Point</th>
            <th class="text-left py-2">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each volumes as volume}
            <tr class="border-b">
              <td class="py-2 font-medium">{volume.name}</td>
              <td>{volume.driver}</td>
              <td class="font-mono text-xs">{volume.mountpoint}</td>
              <td class="py-2">
                <Button variant="destructive" size="sm">
                  <Trash2 class="w-4 h-4" />
                </Button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </Card.Content>
</Card.Root>
