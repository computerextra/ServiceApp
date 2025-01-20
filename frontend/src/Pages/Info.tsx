import { Button } from "@/components/ui/button";
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { SendInfo } from "../../wailsjs/go/main/App";
import BackButton from "@/components/BackButton";

const formSchema = z.object({
  Auftrag: z.string(),
  Mail: z.string().email(),
});

export default function Info() {
  const [message, setMessage] = useState<string | undefined>(undefined);
  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });

  useEffect(() => {
    const controller = new AbortController();

    setTimeout(() => setMessage(undefined), 2000, {
      signal: controller.signal,
    });

    return () => controller.abort();
  }, [message]);

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    // BUG: App stürzt ab, wenn eine Mail gesendet wird!
    const res = await SendInfo(values.Auftrag, values.Mail);
    if (res != "OK") {
      setMessage(res);
    } else {
      setMessage("Mail Erfolgreich verschickt");
    }

    form.reset(
      { Auftrag: undefined, Mail: undefined },
      {
        keepDefaultValues: false,
        keepDirty: false,
        keepErrors: false,
        keepIsValid: false,
        keepIsSubmitted: false,
      }
    );
  };

  return (
    <>
      <BackButton href="/" />
      <h1 className="text-4xl">Info</h1>
      <h2>Info Mail an Kunde</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
          <FormField
            control={form.control}
            name="Mail"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Mail</FormLabel>
                <FormControl>
                  <Input type="email" required placeholder="Mail" {...field} />
                </FormControl>

                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="Auftrag"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Auftrag</FormLabel>
                <FormControl>
                  <Input
                    type="text"
                    required
                    placeholder="Auftrag"
                    {...field}
                  />
                </FormControl>
                <FormDescription>Nummer AU/LI/RE</FormDescription>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button type="submit">Senden</Button>
        </form>
      </Form>

      <p className="text-2xl font-semibold">{message}</p>
    </>
  );
}
