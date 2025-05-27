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
import { Switch } from "@/components/ui/switch";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router";
import { z } from "zod";
import { DeleteJob, GetJob, UpdateJob } from "../../../../wailsjs/go/main/App";

const Job = z.object({
  ID: z.string(),
  Name: z.string(),
  Online: z.boolean(),
});
export type Job = z.infer<typeof Job>;
const UpdateJobProps = Job;
type UpdateJobProps = z.infer<typeof UpdateJobProps>;

export default function JobsEdit() {
  const { jid } = useParams();
  const [data, setData] = useState<Job | undefined>(undefined);

  useEffect(() => {
    async function c() {
      if (jid == null) return;
      const x = await GetJob(jid);
      if (typeof x == "string") {
        alert(x);
        return;
      } else {
        setData(x);
      }
    }
    void c();
  }, [jid]);

  const form = useForm<z.infer<typeof UpdateJobProps>>({
    resolver: zodResolver(UpdateJobProps),
    defaultValues: {
      ID: data?.ID ?? "",
      Name: data?.Name ?? "",
      Online: data?.Online ?? false,
    },
  });
  const navigate = useNavigate();

  useEffect(() => {
    if (data == null) return;

    form.reset({
      ID: data.ID,
      Name: data.Name,
      Online: data.Online,
    });
  }, [data]);

  const onSubmit = async (values: z.infer<typeof UpdateJobProps>) => {
    const res = await UpdateJob(
      values.ID,
      values.Name,
      values.Online ? "true" : "false"
    );
    if (res) navigate("/CMS/Jobs");
  };

  const handleDelete = async () => {
    if (jid == null) return;
    await DeleteJob(jid);
    navigate("/CMS/Jobs");
  };

  return (
    <>
      <BackButton href="/CMS/Jobs" />
      <h1 className="my-8">{data?.Name} bearbeiten</h1>

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
          <div>
            <h3 className="mb-4 text-lg font-medium">Anzeige auf Webseite</h3>
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="Online"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-center justify-between p-4 border rounded-lg">
                    <div className="space-y-0.5">
                      <FormLabel className="text-base">Online</FormLabel>
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
