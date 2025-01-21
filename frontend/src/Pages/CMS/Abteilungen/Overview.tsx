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
import { MoreHorizontal } from "lucide-react";
import { useEffect, useState } from "react";
import { Link } from "react-router";
import { z } from "zod";
import { GetAbteilungen } from "../../../../wailsjs/go/main/App";

const Abteilung = z.object({
  ID: z.string(),
  Name: z.string(),
});
type Abteilung = z.infer<typeof Abteilung>;

const columns: ColumnDef<Abteilung>[] = [
  {
    accessorKey: "Name",
    header: "Name",
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
              <Link to={`/CMS/Abteilungen/${payment.ID}`}>Bearbeiten</Link>
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      );
    },
  },
];

export default function AbteilungOverview() {
  const [Abteilungen, setAbteilungen] = useState<Abteilung[] | undefined>(
    undefined
  );

  useEffect(() => {
    async function x() {
      const Abteilungen = await GetAbteilungen();
      if (typeof Abteilungen == "string") {
        alert(Abteilungen);
        return;
      } else {
        setAbteilungen(Abteilungen);
      }
    }

    void x();
  }, []);

  return (
    <>
      <BackButton href="/CMS/" />
      <h1 className="mb-8">CMS - Abteilungen</h1>
      <Button asChild className="mb-2">
        <Link to="/CMS/Abteilungen/Neu">Neue Abteilung</Link>
      </Button>
      {Abteilungen && (
        <DataTable
          columns={columns}
          data={Abteilungen}
          placeholder="Suche nach Name"
          search="Name"
        />
      )}
    </>
  );
}
