import BackButton from "@/components/BackButton";
import LoadingSpinner from "@/components/LoadingSpinner";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import {
  GenerateWarenlieferung,
  SendWarenlieferung,
} from "../../wailsjs/go/main/App";

export default function Warenlieferung() {
  const [loading, setLoading] = useState(false);
  const [generated, setGenerated] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);

  const generate = async () => {
    setLoading(true);

    const res = await GenerateWarenlieferung();
    if (res != "OK") {
      setError(res);
      return;
    } else {
      setGenerated(true);
      setLoading(false);
    }
  };

  const send = async () => {
    setLoading(true);

    const res = await SendWarenlieferung();
    if (res != "OK") {
      setError(res);
      return;
    } else {
      setGenerated(false);
      setLoading(false);
      setError("Warenlieferung erfolgreich gesendet!");
    }
  };

  return (
    <>
      <BackButton href="/" />
      <h1 className="text-4xl">Warenlieferung</h1>
      <h2 className="mb-8">
        Erstellen und versenden der täglichen Warenlieferung
      </h2>
      {loading && <LoadingSpinner />}
      {error == null && !loading && !generated && (
        <Button onClick={generate}>Warenlieferung generieren</Button>
      )}
      {error == null && generated && !loading && (
        <Button onClick={send}>Warenlieferung senden</Button>
      )}
      {error != null && (
        <h2 className="text-5xl font-bold text-red-600">{error}</h2>
      )}
    </>
  );
}
