import BackButton from "@/components/BackButton";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ColumnDef } from "@tanstack/react-table";
import { MoreHorizontal } from "lucide-react";
import { Link } from "react-router";
import { Partner } from "./Edit";
import { DataTable } from "@/components/DataTable";
import { useEffect, useState } from "react";
import { GetAllPartner } from "../../../../wailsjs/go/main/App";

const columns: ColumnDef<Partner>[] = [
  {
    accessorKey: "Name",
    header: "Name",
  },
  {
    accessorKey: "Link",
    header: "Link",
  },
  {
    accessorKey: "Image",
    header: "Image",
  },
  {
    id: "actions",
    cell: ({ row }) => {
      const payment = row.original;

      return (
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" className="w-8 h-8 p-0">
              <span className="sr-only">Open menu</span>
              <MoreHorizontal className="w-4 h-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuLabel>Actions</DropdownMenuLabel>
            <DropdownMenuItem asChild>
              <Link to={`/CMS/Partner/${payment.ID}`}>Bearbeiten</Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  },
];

export default function PartnerOverview() {
  const [data, setData] = useState<Partner[] | undefined>(undefined);

  useEffect(() => {
    async function c() {
      const x = await GetAllPartner();
      if (typeof x == "string") {
        alert(x);
        return;
      } else {
        setData(x);
      }
    }
    void c();
  }, []);

  return (
    <>
      <BackButton href="/CMS/" />
      <h1 className="mb-8">CMS - Partner</h1>
      <Button asChild className="mb-2">
        <Link to="/CMS/Partner/Neu">Neuen Partner</Link>
      </Button>
      {data && (
        <DataTable
          columns={columns}
          data={data}
          placeholder="Suche nach Name"
          search="Name"
        />
      )}
    </>
  );
}
