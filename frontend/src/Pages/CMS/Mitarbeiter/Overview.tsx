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
import { Check, Cross, MoreHorizontal } from "lucide-react";
import { Link } from "react-router";
import { Mitarbeiter } from "./Edit";
import { Abteilung } from "../Abteilungen/Edit";
import { useEffect, useState } from "react";
import {
  GetAbteilungen,
  GetAllMitarbeiter,
} from "../../../../wailsjs/go/main/App";
import { DataTable } from "@/components/DataTable";

export default function MitarbeiterOverview() {
  const [Mitarbeiter, setMitarbeiter] = useState<Mitarbeiter[] | undefined>(
    undefined
  );
  const [Abteilungen, setAbteilungen] = useState<Abteilung[] | undefined>(
    undefined
  );

  useEffect(() => {
    async function c() {
      const ma = await GetAllMitarbeiter();
      if (typeof ma == "string") {
        alert(ma);
        return;
      } else {
        setMitarbeiter(ma);
      }
      const A = await GetAbteilungen();
      if (typeof A == "string") {
        alert(A);
        return;
      } else {
        setAbteilungen(A);
      }
    }
    void c();
  }, []);

  const columns: ColumnDef<Mitarbeiter>[] = [
    {
      accessorKey: "Name",
      header: "Name",
    },
    {
      accessorKey: "Short",
      header: "Short",
    },
    {
      accessorKey: "Sex",
      header: "Sex",
      cell: ({ row }) => {
        const x = row.original;
        return <p>{x.Sex == "m" ? "Männlich" : "Weiblich"}</p>;
      },
    },
    {
      accessorKey: "Image",
      header: "Image",
      cell: ({ row }) => {
        const x = row.original;
        if (x.Image) return <Check className="text-green-500" />;
        return <Cross className="rotate-45 text-primary" />;
      },
    },
    {
      accessorKey: "Focus",
      header: "Focus",
    },
    {
      accessorKey: "Tags",
      header: "Tags",
    },
    {
      id: "Abteilung",
      header: "Abteilung",
      cell: ({ row }) => {
        const x = row.original;
        const Abteilung = Abteilungen?.find((y) => y.ID == x.Abteilungid);
        if (Abteilung) {
          return <p>{Abteilung.Name}</p>;
        }
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
                <Link to={`/CMS/Mitarbeiter/${payment.ID}`}>Bearbeiten</Link>
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];

  return (
    <>
      <BackButton href="/CMS/" />
      <h1 className="mb-8">CMS - Mitarbeiter</h1>
      <Button asChild className="mb-2">
        <Link to="/CMS/Mitarbeiter/Neu">Neuen Mitarbeiter</Link>
      </Button>
      {Mitarbeiter && (
        <DataTable
          columns={columns}
          data={Mitarbeiter}
          placeholder="Suche nach Name"
          search="Name"
        />
      )}
    </>
  );
}
