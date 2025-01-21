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
import {
  DeletePartner,
  GetPartner,
  UpdatePartner,
} from "../../../../wailsjs/go/main/App";
import { z } from "zod";

const Partner = z.object({
  ID: z.string(),
  Name: z.string(),
  Image: z.string(),
  Link: z.string(),
});
export type Partner = z.infer<typeof Partner>;
const UpdatePartnerProps = Partner;
type UpdatePartnerProps = z.infer<typeof UpdatePartnerProps>;

export default function PartnerEdit() {
  const { pid } = useParams();
  const [data, setData] = useState<Partner | undefined>(undefined);

  useEffect(() => {
    async function c() {
      if (pid == null) return;
      const x = await GetPartner(pid);
      if (typeof x == "string") {
        alert(x);
        return;
      } else {
        setData(x);
      }
    }
    void c();
  }, [pid]);

  const navigate = useNavigate();
  const form = useForm<z.infer<typeof UpdatePartnerProps>>({
    resolver: zodResolver(UpdatePartnerProps),
    defaultValues: {
      ID: data?.ID ?? "",
      Image: data?.Image ?? "",
      Link: data?.Link ?? "",
      Name: data?.Name ?? "",
    },
  });

  useEffect(() => {
    if (data == null) return;

    form.reset({
      ID: data.ID,
      Image: data.Image,
      Link: data.Link,
      Name: data.Name,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  const onSubmit = async (values: z.infer<typeof UpdatePartnerProps>) => {
    const res = await UpdatePartner(
      values.ID,
      values.Name,
      values.Link,
      values.Image
    );
    if (res) navigate("/CMS/Partner");
  };

  const handleDelete = async () => {
    if (pid == null) return;
    await DeletePartner(pid);
    navigate("/CMS/Partner");
  };

  return (
    <>
      <BackButton href="/CMS/Partner" />
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
          <FormField
            control={form.control}
            name="Link"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Link</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="Image"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Image</FormLabel>
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
              Löschen
            </Button>
          </div>
        </form>
      </Form>
    </>
  );
}
