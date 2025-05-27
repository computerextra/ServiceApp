import { Button } from "@/components/ui/button";

import { Badge } from "@/components/ui/badge";
import { useEffect, useState } from "react";
import { Link } from "react-router";
import { GetCmsCounts } from "../../../wailsjs/go/main/App";

type Counts = {
  Abteilung: number;
  Angebote: number;
  Jobs: number;
  Mitarbeiter: number;
  Partner: number;
};

export default function Cms() {
  const [Counts, setCounts] = useState<Counts | undefined>(undefined);

  useEffect(() => {
    async function x() {
      const Counts = (await GetCmsCounts()) as Counts;
      console.log(Counts);
      setCounts(Counts);
    }
    void x();
  }, []);

  return (
    <>
      <h1>CMS Übersicht</h1>
      <div className="flex flex-col items-start mt-8">
        <Button asChild variant="link" className="my-2">
          <Link to="/CMS/Abteilungen">
            Abteilungen <Badge>{Counts?.Abteilung}</Badge>
          </Link>
        </Button>
        <Button asChild variant="link" className="my-2">
          <Link to="/CMS/Angebote">
            Angebote <Badge>{Counts?.Angebote}</Badge>
          </Link>
        </Button>
        <Button asChild variant="link" className="my-2">
          <Link to="/CMS/Jobs">
            Jobs <Badge>{Counts?.Jobs}</Badge>
          </Link>
        </Button>
        <Button asChild variant="link" className="my-2">
          <Link to="/CMS/Mitarbeiter">
            Mitarbeiter <Badge>{Counts?.Mitarbeiter}</Badge>
          </Link>
        </Button>
        <Button asChild variant="link" className="my-2">
          <Link to="/CMS/Partner">
            Partner <Badge>{Counts?.Partner}</Badge>
          </Link>
        </Button>
      </div>
    </>
  );
}
