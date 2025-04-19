<script lang="ts">
  import * as Card from "$lib/components/ui/card/index.js";

  // Placeholder data; replace with API calls later
  let containers = [
    { id: "1a2b3c", name: "web", status: "running", image: "nginx:latest" },
    { id: "4d5e6f", name: "db", status: "exited", image: "postgres:15" },
  ];
  let images = [
    { id: "abc123", repo: "nginx", tag: "latest", size: "133MB" },
    { id: "def456", repo: "postgres", tag: "15", size: "350MB" },
  ];
  let dockerInfo = {
    version: "25.0.0",
    os: "Linux",
    containers: 2,
    images: 2,
  };
</script>

<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
  <Card.Root>
    <Card.Header>
      <Card.Title>Docker Info</Card.Title>
      <Card.Description>System overview</Card.Description>
    </Card.Header>
    <Card.Content>
      <ul class="text-sm">
        <li><strong>Version:</strong> {dockerInfo.version}</li>
        <li><strong>OS:</strong> {dockerInfo.os}</li>
        <li><strong>Containers:</strong> {dockerInfo.containers}</li>
        <li><strong>Images:</strong> {dockerInfo.images}</li>
      </ul>
    </Card.Content>
    <Card.Footer />
  </Card.Root>

  <Card.Root class="md:col-span-2">
    <Card.Header>
      <Card.Title>Quick Actions</Card.Title>
      <Card.Description>Manage all containers</Card.Description>
    </Card.Header>
    <Card.Content>
      <div class="flex gap-4">
        <button
          class="bg-primary text-primary-foreground px-4 py-2 rounded hover:bg-primary/80"
          >Start All</button
        >
        <button
          class="bg-secondary text-secondary-foreground px-4 py-2 rounded hover:bg-secondary/80"
          >Stop All</button
        >
        <button
          class="bg-accent text-accent-foreground px-4 py-2 rounded hover:bg-accent/80"
          >Prune</button
        >
      </div>
    </Card.Content>
    <Card.Footer />
  </Card.Root>
</div>

<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
  <Card.Root>
    <Card.Header>
      <Card.Title>Containers</Card.Title>
      <Card.Description>Active and stopped containers</Card.Description>
    </Card.Header>
    <Card.Content>
      <table class="w-full text-sm">
        <thead>
          <tr>
            <th class="text-left">Name</th>
            <th class="text-left">Image</th>
            <th class="text-left">Status</th>
          </tr>
        </thead>
        <tbody>
          {#each containers as c}
            <tr>
              <td>{c.name}</td>
              <td>{c.image}</td>
              <td>
                <span
                  class={c.status === "running"
                    ? "text-green-600"
                    : "text-gray-500"}
                >
                  {c.status}
                </span>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </Card.Content>
    <Card.Footer />
  </Card.Root>

  <Card.Root>
    <Card.Header>
      <Card.Title>Images</Card.Title>
      <Card.Description>Available Docker images</Card.Description>
    </Card.Header>
    <Card.Content>
      <table class="w-full text-sm">
        <thead>
          <tr>
            <th class="text-left">Repository</th>
            <th class="text-left">Tag</th>
            <th class="text-left">Size</th>
          </tr>
        </thead>
        <tbody>
          {#each images as img}
            <tr>
              <td>{img.repo}</td>
              <td>{img.tag}</td>
              <td>{img.size}</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </Card.Content>
    <Card.Footer />
  </Card.Root>
</div>
