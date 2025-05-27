import BackButton from "@/components/BackButton";
import { DataTable } from "@/components/DataTable";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ColumnDef } from "@tanstack/react-table";
import { Check, Cross, MoreHorizontal } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router";
import { GetAngebote } from "../../../../wailsjs/go/main/App";
import type { Angebot } from "./Edit";

const columns: ColumnDef<Angebot>[] = [
  {
    accessorKey: "Title",
    header: "Title",
  },
  {
    accessorKey: "Subtitle",
    header: "Subtitle",
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
    accessorKey: "DateStart",
    header: "Gültig Von",
    cell: ({ row }) => {
      const x = row.original;

      return (
        <p>
          {new Date(x.DateStart).toLocaleDateString("de-DE", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
          })}
        </p>
      );
    },
  },
  {
    accessorKey: "DateStop",
    header: "Gültig Bis",
    cell: ({ row }) => {
      const x = row.original;

      return (
        <p>
          {new Date(x.DateStop).toLocaleDateString("de-DE", {
            day: "2-digit",
            month: "2-digit",
            year: "numeric",
          })}
        </p>
      );
    },
  },
  {
    accessorKey: "Anzeigen",
    header: "Online",
    cell: ({ row }) => {
      const x = row.original;
      if (x.Anzeigen) return <Check className="text-green-500" />;
      else return <Cross className="rotate-45 text-primary" />;
    },
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
              <Link to={`/CMS/Angebote/${payment.ID}`}>Bearbeiten</Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  },
];

export default function AngebotOverview() {
  const [Angebote, setAngebote] = useState<Angebot[] | undefined>(undefined);

  useEffect(() => {
    async function x() {
      const A = await GetAngebote();
      if (typeof A == "string") {
        alert(A);
        return;
      } else {
        const B: Angebot[] = A.map((x) => {
          return {
            ...x,
            Subtitle: x.Subtitle.String,
            Anzeigen: x.Anzeigen.Bool,
          };
        });
        setAngebote(B);
      }
    }
    void x();
  }, []);

  return (
    <>
      <BackButton href="/CMS/" />
      <h1 className="mb-8">CMS - Angebote</h1>
      <Button asChild className="mb-2">
        <Link to="/CMS/Angebote/Neu">Neues Angebot</Link>
      </Button>
      {Angebote && (
        <DataTable
          columns={columns}
          data={Angebote}
          placeholder="Suche nach Name"
          search="Title"
        />
      )}
    </>
  );
}
