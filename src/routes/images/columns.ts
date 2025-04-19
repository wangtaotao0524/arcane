import type { ServiceImage } from "$lib/services/docker-service";
import type { ColumnDef } from "@tanstack/table-core";
import { renderComponent } from "$lib/components/ui/data-table/index.js";
import ImageActions from "./ImageActions.svelte";

// Helper to format bytes (keep as is)
function formatBytes(bytes: number, decimals = 2): string {
  if (!+bytes) return "0 Bytes";
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ["Bytes", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`;
}

export const columns: ColumnDef<ServiceImage>[] = [
  {
    accessorKey: "repo",
    header: "Repository",
    cell: ({ row }) => {
      // Return the plain repository name string
      return row.original.repo;
      // Note: Styling for '<none>' will be lost without a component
    },
  },
  {
    accessorKey: "tag",
    header: "Tag",
    cell: ({ row }) => {
      // Return the plain tag string
      return row.original.tag;
      // Note: Styling for '<none>' will be lost without a component
    },
  },
  {
    accessorKey: "id",
    header: "Image ID",
    cell: ({ row }) => {
      // Calculate and return the short ID string
      const shortId =
        row.original.id.split(":")[1]?.substring(0, 12) ||
        row.original.id.substring(0, 12);
      return shortId;
      // Note: Monospace font styling will be lost without a component
    },
    enableSorting: false,
  },
  {
    accessorKey: "size",
    header: "Size",
    cell: ({ row }) => {
      // formatBytes already returns a string
      return formatBytes(row.original.size);
    },
  },
  {
    id: "actions",
    header: "Actions",
    cell: ({ row }) => {
      // Keep using renderComponent for the actions dropdown
      return renderComponent(ImageActions, {
        id: row.original.id,
        repoTag: row.original.repoTags?.[0],
      });
    },
    enableSorting: false,
  },
];
