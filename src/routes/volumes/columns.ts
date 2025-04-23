import type { ColumnDef } from "@tanstack/table-core";
import { renderComponent } from "$lib/components/ui/data-table/index.js";
import VolumeActions from "./VolumeActions.svelte";
import MountpointCell from "./MountpointCell.svelte";
import CustomBadge from "$lib/components/badges/custom-badge.svelte";
import type { ServiceVolume } from "$lib/services/docker-service";

// Update the type to include the inUse property
export type VolumeInfo = ServiceVolume & { inUse?: boolean };

export const columns: ColumnDef<VolumeInfo>[] = [
  {
    accessorKey: "name",
    header: "Name",
  },
  {
    accessorKey: "driver",
    header: "Driver",
  },
  {
    accessorKey: "scope",
    header: "Scope",
  },
  {
    accessorKey: "usage",
    header: "Status",
    cell: ({ row }) => {
      const inUse = row.original.inUse;
      if (inUse === undefined) {
        return null;
      }

      if (inUse) {
        return renderComponent(CustomBadge, {
          variant: "status",
          text: "In use",
        });
      } else {
        return renderComponent(CustomBadge, {
          variant: "status",
          text: "Unused",
        });
      }
    },
    enableSorting: true,
  },
  {
    accessorKey: "mountpoint",
    header: "Mountpoint",
    cell: ({ row }) => {
      return renderComponent(MountpointCell, {
        mountpoint: row.original.mountpoint,
      });
    },
  },
  {
    id: "actions",
    header: "",
    cell: ({ row }) => {
      return renderComponent(VolumeActions, {
        name: row.original.name,
        inUse: !!row.original.inUse,
      });
    },
    enableSorting: false,
  },
];
