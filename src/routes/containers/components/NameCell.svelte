<script lang="ts">
  interface Props {
    id: string;
    name: string;
  }

  let { id, name }: Props = $props();
</script>

<a href="/containers/{id}" class="font-medium hover:underline">
  {name}
</a>
