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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router";
import { z } from "zod";
import {
  DeleteMitarbeiter,
  GetAbteilungen,
  GetMitarbeiter,
  UpdateMitarbeiter,
} from "../../../../wailsjs/go/main/App";
import { Abteilung } from "../Abteilungen/Edit";

const Mitarbeiter = z.object({
  ID: z.string(),
  Name: z.string(),
  Short: z.string(),
  Sex: z.string(),
  Tags: z.string(),
  Focus: z.string(),
  Abteilungid: z.string(),
  Image: z.boolean(),
});
export type Mitarbeiter = z.infer<typeof Mitarbeiter>;
const UpdateMitarbeiterProps = Mitarbeiter;
type UpdateMitarbeiterProps = z.infer<typeof UpdateMitarbeiterProps>;

export default function MitarbeiterEdit() {
  const { mid } = useParams();
  const [Mitarbeiter, setMitarbeiter] = useState<Mitarbeiter | undefined>(
    undefined
  );
  const [Abteilungen, setAbteilungen] = useState<Abteilung[] | undefined>(
    undefined
  );

  useEffect(() => {
    async function c() {
      if (mid == null) return;
      const ma = await GetMitarbeiter(mid);
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
  }, [mid]);

  const navigate = useNavigate();
  const form = useForm<z.infer<typeof UpdateMitarbeiterProps>>({
    resolver: zodResolver(UpdateMitarbeiterProps),
    defaultValues: {
      Abteilungid: Mitarbeiter?.Abteilungid ?? "",
      Focus: Mitarbeiter?.Focus ?? "",
      ID: Mitarbeiter?.ID ?? "",
      Image: Mitarbeiter?.Image ?? false,
      Name: Mitarbeiter?.Name ?? "",
      Sex: Mitarbeiter?.Sex ?? "",
      Short: Mitarbeiter?.Short ?? "",
      Tags: Mitarbeiter?.Tags ?? "",
    },
  });

  useEffect(() => {
    if (Mitarbeiter == null) return;

    form.reset({
      Abteilungid: Mitarbeiter.Abteilungid,
      Focus: Mitarbeiter.Focus,
      ID: Mitarbeiter.ID,
      Image: Mitarbeiter.Image,
      Name: Mitarbeiter.Name,
      Sex: Mitarbeiter.Sex,
      Short: Mitarbeiter.Short,
      Tags: Mitarbeiter.Tags,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [Mitarbeiter]);

  const onSubmit = async (values: z.infer<typeof UpdateMitarbeiterProps>) => {
    const res = await UpdateMitarbeiter(
      values.ID,
      values.Name,
      values.Short,
      values.Image ? "true" : "false",
      values.Sex,
      values.Tags,
      values.Focus,
      values.Abteilungid
    );
    if (res) navigate("/CMS/Mitarbeiter");
  };

  const handleDelete = async () => {
    if (mid == null) return;
    await DeleteMitarbeiter(mid);
  };

  return (
    <>
      <BackButton href="/CMS/Mitarbeiter" />
      <h1 className="my-8">{Mitarbeiter?.Name} bearbeiten</h1>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
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
          <FormField
            control={form.control}
            name="Short"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Short</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="Focus"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Focus</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="Tags"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Tags</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="Sex"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Geschlecht</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Bitte Wählen" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    <SelectItem value="m">Männlich</SelectItem>
                    <SelectItem value="w">Weiblich</SelectItem>
                  </SelectContent>
                </Select>
                <FormMessage />
              </FormItem>
            )}
          />
          <div>
            <h3 className="mb-4 text-lg font-medium">Bild auf Webseite</h3>
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="Image"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between p-4 border rounded-lg">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">Bild</FormLabel>
                    </div>
                    <FormControl>
                      <Switch
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                  </FormItem>
                )}
              />
            </div>
          </div>
          <FormField
            control={form.control}
            name="Abteilungid"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Abteilung</FormLabel>
                <Select
                  onValueChange={field.onChange}
                  defaultValue={field.value}
                >
                  <FormControl>
                    <SelectTrigger>
                      <SelectValue placeholder="Bitte Wählen" />
                    </SelectTrigger>
                  </FormControl>
                  <SelectContent>
                    {Abteilungen?.map((x) => (
                      <SelectItem key={x.ID} value={x.ID}>
                        {x.Name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
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
              Löschen
            </Button>
          </div>
        </form>
      </Form>
    </>
  );
}
