import BackButton from "@/components/BackButton";
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
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router";
import { z } from "zod";
import {
  DeleteAbteilung,
  GetAbteilung,
  UpdateAbteilung,
} from "../../../../wailsjs/go/main/App";

const UpdateAbteilungProps = z.object({
  ID: z.string(),
  Name: z.string(),
});
type UpdateAbteilungProps = z.infer<typeof UpdateAbteilungProps>;

export type Abteilung = {
  ID: string;
  Name: string;
};

export default function AbteilungEdit() {
  const { aid } = useParams();
  const [Abteilung, setAbteilung] = useState<Abteilung | undefined>(undefined);

  useEffect(() => {
    async function c() {
      if (aid == null) return;
      const Abteilung = await GetAbteilung(aid);
      if (typeof Abteilung == "string") {
        alert(Abteilung);
        return;
      }
      setAbteilung(Abteilung);
    }
    void c();
  }, [aid]);

  const form = useForm<z.infer<typeof UpdateAbteilungProps>>({
    resolver: zodResolver(UpdateAbteilungProps),
    defaultValues: {
      ID: Abteilung?.ID,
      Name: Abteilung?.Name,
    },
  });
  const navigate = useNavigate();
  const [msg, setMsg] = useState<undefined | string>(undefined);

  useEffect(() => {
    if (Abteilung == null) return;

    form.reset({
      ID: Abteilung.ID,
      Name: Abteilung.Name,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [Abteilung]);

  const onSubmit = async (values: z.infer<typeof UpdateAbteilungProps>) => {
    const res = await UpdateAbteilung(values.ID, values.Name);
    if (res == "OK") {
      navigate("/CMS/Abteilungen");
    } else {
      setMsg(res);
    }
  };

  const handleDelete = async () => {
    if (aid == null) return;
    await DeleteAbteilung(aid);
    navigate("/CMS/Abteilungen");
  };

  return (
    <>
      <BackButton href="/CMS/Abteilungen" />
      <h1 className="mt-8">{Abteilung?.Name} bearbeiten</h1>

      {msg && <h2 className="text-primary">{msg}</h2>}
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="mt-8 space-y-8">
          <FormField
            control={form.control}
            name="Name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Name</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <div className="flex justify-between">
            <Button type="submit">Speichern</Button>
            <Button
              variant="secondary"
              onClick={(e) => {
                e.preventDefault();
                void handleDelete();
              }}
            >
              Eintrag Löschen
            </Button>
          </div>
        </form>
      </Form>
    </>
  );
}
