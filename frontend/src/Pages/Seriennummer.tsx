import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Terminal } from "lucide-react";
import { GetSeriennummer } from "../../wailsjs/go/main/App";
import BackButton from "@/components/BackButton";

const formSchema = z.object({
  Artikelnummer: z.string(),
});

export default function Seriennummer() {
  const [string, setString] = useState<string | undefined>(undefined);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
  });

  const onSubmit = async (values: z.infer<typeof formSchema>) => {
    const res = await GetSeriennummer(values.Artikelnummer);
    if (res) {
      const x = `${values.Artikelnummer}: ${res}`;
      setString(x);
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(x);
      } else {
        // Use the 'out of viewport hidden text area' trick
        const textArea = document.createElement("textarea");
        textArea.value = x;

        // Move textarea out of the viewport
        textArea.style.position = "absolute";
        textArea.style.left = "-99999999px";

        document.body.prepend(textArea);
        textArea.select();

        try {
          document.execCommand("copy");
        } catch (error) {
          alert(error);
        } finally {
          textArea.remove();
        }
      }

      setTimeout(() => {
        setString(undefined);
        form.reset({
          Artikelnummer: "",
        });
      }, 2000);
    }
  };

  return (
    <>
      <BackButton href="/" />
      <h1 className="text-4xl">Seriennummer</h1>
      <h2>Eingabe von Artikelnummern</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
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
          <Button type="submit">Senden</Button>
        </form>
      </Form>

      {string && (
        <Alert className="mt-8">
          <Terminal className="w-4 h-4" />
          <AlertTitle>Text Kopiert.</AlertTitle>
          <AlertDescription>
            {string} wurde in die Zwischenablage kopiert.
          </AlertDescription>
        </Alert>
      )}
    </>
  );
}
