import BackButton from "@/components/BackButton";
import { DataTable } from "@/components/DataTable";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { ColumnDef } from "@tanstack/react-table";
import { useState } from "react";
import { SearchKunde } from "../../wailsjs/go/main/App";
import type { main } from "../../wailsjs/go/models";

const columns: ColumnDef<main.Sg_Adressen>[] = [
  {
    accessorKey: "KundNr.String",
    header: "Kundennummer",
    cell: ({ row }) => {
      const c = row.original;

      if (c.KundNr.Valid)
        return <span className="text-primary">{c.KundNr.String}</span>;
      if (c.LiefNr.Valid)
        return <span className="text-destructive">{c.LiefNr.String}</span>;
    },
  },

  {
    accessorKey: "Suchbegriff.String",
    header: "Suchbegriff",
    cell: ({ row }) => {
      const c = row.original;

      return <p className="text-start">{c.Suchbegriff.String}</p>;
    },
  },
  {
    accessorKey: "Telefon1.String",
    header: "Telefon",
    cell: ({ row }) => {
      const c = row.original;

      return (
        <div className="text-start">
          {c.Telefon1.Valid && (
            <a
              className="underline text-primary"
              href={`tel:${c.Telefon1.String}`}
            >
              {c.Telefon1.String}
            </a>
          )}
          {c.Telefon2.Valid && (
            <a
              className="underline text-primary"
              href={`tel:${c.Telefon2.String}`}
            >
              {c.Telefon2.String}
            </a>
          )}
        </div>
      );
    },
  },
  {
    accessorKey: "Mobiltelefon1.String",
    header: "Mobiltelefon",
    cell: ({ row }) => {
      const c = row.original;

      return (
        <div className="text-start">
          {c.Mobiltelefon1.Valid && (
            <a
              className="underline text-primary"
              href={`tel:${c.Mobiltelefon1.String}`}
            >
              {c.Mobiltelefon1.String}
            </a>
          )}
          {c.Mobiltelefon2.Valid && (
            <a
              className="underline text-primary"
              href={`tel:${c.Mobiltelefon2.String}`}
            >
              {c.Mobiltelefon2.String}
            </a>
          )}
        </div>
      );
    },
  },
  {
    accessorKey: "EMail1.String",
    header: "EMail",
    cell: ({ row }) => {
      const c = row.original;

      return (
        <div className="text-start">
          {c.EMail1.Valid && (
            <a
              className="underline text-primary"
              href={`mailto:${c.EMail1.String}`}
            >
              {c.EMail1.String}
            </a>
          )}
          {c.EMail2.Valid && (
            <a
              className="underline text-primary"
              href={`mailto:${c.EMail2.String}`}
            >
              {c.EMail2.String}
            </a>
          )}
        </div>
      );
    },
  },
  {
    accessorKey: "KundUmsatz.Float64",
    header: "KundUmsatz",
    cell: ({ row }) => {
      const c = row.original;
      let umsatz = 0;

      if (c.KundUmsatz.Valid) {
        umsatz =
          Math.round((c.KundUmsatz.Float64 + Number.EPSILON) * 100) / 100;
      }
      if (c.LiefUmsatz.Valid) {
        umsatz =
          Math.round((c.LiefUmsatz.Float64 + Number.EPSILON) * 100) / 100;
      }

      return <p className="text-start">{umsatz} €</p>;
    },
  },
];

export default function Sage() {
  const [search, setSearch] = useState("");
  const [results, setResults] = useState<main.Sg_Adressen[] | undefined>();
  const [loading, setLoading] = useState(false);

  const onSubmit = async () => {
    setLoading(true);
    const res = await SearchKunde(search);
    if (res) {
      setResults(res);
      setLoading(false);
    }
  };

  return (
    <>
      <BackButton href="/" />
      <h1 className="text-4xl">Kundensuche</h1>
      <form
        className="space-y-4"
        onSubmit={(e) => {
          e.preventDefault();
          void onSubmit();
        }}
      >
        <div className="grid w-full max-w-sm items-center gap-1.5">
          <Label htmlFor="email">Suchbegriff</Label>
          <Input
            type="text"
            id="search"
            name="search"
            disabled={loading}
            defaultValue={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Suchbegriff"
          />
          <Button type="submit" variant="default" disabled={loading}>
            {loading ? "Bitte warten..." : "Suchen"}
          </Button>
        </div>
      </form>
      <Separator />
      {results && results.length > 0 && (
        <div className="mt-5">
          <DataTable columns={columns} data={results} placeholder="PLATZ DA" />
        </div>
      )}
    </>
  );
}
