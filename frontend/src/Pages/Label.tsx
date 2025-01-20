import BackButton from "@/components/BackButton";
import LoadingSpinner from "@/components/LoadingSpinner";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { SyncLabel } from "../../wailsjs/go/main/App";

export default function Label() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);

  const sync = async () => {
    setLoading(true);

    const res = await SyncLabel();
    if (res != "OK") {
      setError(res);
      return;
    } else {
      setLoading(false);
      setError("Daten erfolgreich synchronisiert");
    }
  };
  return (
    <>
      <BackButton href="/" />
      <h1 className="text-4xl">Label</h1>
      <h2 className="mb-8">Synchronisieren von Preisschildern</h2>
      {loading && <LoadingSpinner />}
      {error == null && !loading && <Button onClick={sync}>Sync Label</Button>}
      {error != null && (
        <h2 className="text-5xl font-bold text-red-600">{error}</h2>
      )}
    </>
  );
}
