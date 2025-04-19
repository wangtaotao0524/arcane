<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { Network, Plus, Trash2 } from "@lucide/svelte";

  export let data;
  const { networks } = data;
</script>

<div class="flex justify-between items-center mb-6">
  <h1 class="text-2xl font-bold">Networks</h1>
  <Button variant="default">
    <Plus class="w-4 h-4 mr-2" />
    Create Network
  </Button>
</div>

<Card.Root>
  <Card.Header>
    <Card.Title>Docker Networks</Card.Title>
    <Card.Description>Manage container communication</Card.Description>
  </Card.Header>
  <Card.Content>
    <div class="overflow-x-auto">
      <table class="w-full">
        <thead>
          <tr class="border-b">
            <th class="text-left py-2">Name</th>
            <th class="text-left py-2">Driver</th>
            <th class="text-left py-2">Scope</th>
            <th class="text-left py-2">Subnet</th>
            <th class="text-left py-2">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each networks as network}
            <tr class="border-b">
              <td class="py-2 font-medium">{network.name}</td>
              <td>{network.driver}</td>
              <td>{network.scope}</td>
              <td class="font-mono text-xs">{network.subnet || "N/A"}</td>
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
