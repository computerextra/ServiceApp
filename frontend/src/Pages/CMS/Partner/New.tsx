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
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router";
import { CreatePartner } from "../../../../wailsjs/go/main/App";
import { z } from "zod";

const CreatePartnerProps = z.object({
  Name: z.string(),
  Image: z.string(),
  Link: z.string(),
});
type CreatePartnerProps = z.infer<typeof CreatePartnerProps>;

export default function PartnerNew() {
  const navigate = useNavigate();
  const form = useForm<z.infer<typeof CreatePartnerProps>>({
    resolver: zodResolver(CreatePartnerProps),
  });

  const onSubmit = async (values: z.infer<typeof CreatePartnerProps>) => {
    const res = await CreatePartner(values.Name, values.Link, values.Image);
    if (res) navigate("/CMS/Partner");
  };

  return (
    <>
      <BackButton href="/CMS/Partner" />
      <h1 className="my-8">Neuer Partner</h1>
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

          <Button type="submit">Speichern</Button>
        </form>
      </Form>
    </>
  );
}
