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
import { useNavigate } from "react-router";
import { z } from "zod";
import {
  CreateMitarbeiter,
  GetAbteilungen,
} from "../../../../wailsjs/go/main/App";
import { Abteilung } from "../Abteilungen/Edit";

const CreateMitarbeiterProps = z.object({
  Name: z.string(),
  Short: z.string(),
  Sex: z.string(),
  Tags: z.string(),
  Focus: z.string(),
  Abteilungid: z.string(),
  Image: z.boolean(),
});
type CreateMitarbeiterProps = z.infer<typeof CreateMitarbeiterProps>;

export default function MitarbeiterNew() {
  const [Abteilungen, setAbteilungen] = useState<Abteilung[] | undefined>(
    undefined
  );
  const navigate = useNavigate();
  const form = useForm<z.infer<typeof CreateMitarbeiterProps>>({
    resolver: zodResolver(CreateMitarbeiterProps),
  });

  useEffect(() => {
    async function c() {
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

  const onSubmit = async (values: z.infer<typeof CreateMitarbeiterProps>) => {
    const res = await CreateMitarbeiter(
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

  return (
    <>
      <BackButton href="/CMS/Mitarbeiter" />
      <h1 className="my-8">Neuen Mitarbeiter anlegen</h1>
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

          <Button type="submit">Speichern</Button>
        </form>
      </Form>
    </>
  );
}
