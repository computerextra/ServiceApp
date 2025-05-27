import BackButton from "@/components/BackButton";
import LoadingSpinner from "@/components/LoadingSpinner";
import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { zodResolver } from "@hookform/resolvers/zod";
import axios, { AxiosRequestConfig, RawAxiosRequestHeaders } from "axios";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { SyncAussteller } from "../../wailsjs/go/main/App";

const createAusstellerImageProps = z.object({
  Artikelnummer: z.string(),
  Url: z.string().url(),
});
type createAusstellerImageProps = z.infer<typeof createAusstellerImageProps>;

type WarenlieferungResponse = {
  ok: string;
  error: string;
};
const config: AxiosRequestConfig = {
  headers: {
    Accept: "application/json",
  } as RawAxiosRequestHeaders,
};

const createAusstellerImage = async (
  props: createAusstellerImageProps
): Promise<WarenlieferungResponse> => {
  const res = await axios.post<{ ok: boolean; error: string | null }>(
    "https://aussteller.computer-extra.de/php/update.php",
    {
      Artikelnummer: props.Artikelnummer,
      Link: props.Url,
    },
    config
  );
  if (!res.data.ok && res.data.error != null) {
    return { error: res.data.error, ok: "false" };
  } else {
    return { error: "false", ok: "true" };
  }
};

export default function Aussteller() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | undefined>(undefined);

  const form = useForm<z.infer<typeof createAusstellerImageProps>>({
    resolver: zodResolver(createAusstellerImageProps),
  });

  const sync = async () => {
    setLoading(true);
    const res = await SyncAussteller();
    if (res != "OK") {
      setError(res);
      return;
    } else {
      setLoading(false);
      setError("Daten erfolgreich synchronisiert");
    }
  };

  const onSubmit = async (
    values: z.infer<typeof createAusstellerImageProps>
  ) => {
    setLoading(true);
    const res = await createAusstellerImage(values);
    if (res.error != "false") {
      setError(res.error);
      setLoading(false);
    } else {
      setError("Bilder Link erfolgreich geschrieben");
      form.reset({
        Artikelnummer: undefined,
        Url: undefined,
      });
      setLoading(false);
    }
  };

  return (
    <>
      <BackButton href="/" />
      <h1 className="mb-5 text-4xl">Aussteller</h1>
      {loading && <LoadingSpinner />}
      {!loading && (
        <Button onClick={sync} className="mt-5">
          Sync Aussteller
        </Button>
      )}
      {error != null && (
        <h2 className="my-8 text-5xl font-bold text-red-600">{error}</h2>
      )}
      {!loading && (
        <>
          <h2 className="pb-0 mt-12">Neues Bild für Aussteller anlegen</h2>
          <Form {...form}>
            <form
              onSubmit={form.handleSubmit(onSubmit)}
              className="mt-2 space-y-8 "
            >
              <FormField
                control={form.control}
                name="Artikelnummer"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Artikelnummer</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>

                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="Url"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Url</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>

                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit">Submit</Button>
            </form>
          </Form>
        </>
      )}
    </>
  );
}
