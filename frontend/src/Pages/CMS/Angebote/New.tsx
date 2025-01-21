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
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Switch } from "@/components/ui/switch";
import { CreateAngebot } from "../../../../wailsjs/go/main/App";
import { cn } from "@/lib/utils";
import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { de } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";
import { DayPicker } from "react-day-picker";
import { useForm } from "react-hook-form";
import { useNavigate } from "react-router";
import { z } from "zod";
import "react-day-picker/dist/style.css";

const CreateAngebotProps = z.object({
  Title: z.string(),
  Subtitle: z.string(),
  DateStart: z.string(),
  DateStop: z.string(),
  Link: z.string(),
  Image: z.string(),
  Anzeigen: z.boolean(),
});
type CreateAngebotProps = z.infer<typeof CreateAngebotProps>;

export default function AngebotNew() {
  const form = useForm<z.infer<typeof CreateAngebotProps>>({
    resolver: zodResolver(CreateAngebotProps),
  });
  const navigate = useNavigate();

  const onSubmit = async (values: z.infer<typeof CreateAngebotProps>) => {
    const start = new Date(values.DateStart);
    const end = new Date(values.DateStop);
    const Angebot: CreateAngebotProps = {
      Anzeigen: values.Anzeigen,
      Image: values.Image,
      Link: values.Link,
      Title: values.Title,
      Subtitle: values.Subtitle,
      DateStart: `${start.getDate()}.${
        start.getMonth() + 1
      }.${start.getFullYear()}`,
      DateStop: `${end.getDate()}.${end.getMonth() + 1}.${end.getFullYear()}`,
    };
    await CreateAngebot(
      Angebot.Title,
      Angebot.Subtitle,
      Angebot.DateStart,
      Angebot.DateStop,
      Angebot.Link,
      Angebot.Image,
      Angebot.Anzeigen ? "true" : "false"
    );
    navigate("/CMS/Angebote");
  };
  return (
    <>
      <BackButton href="/CMS/Angebote" />
      <h1 className="mt-8">Neues Angebot anglegen</h1>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="mt-8 space-y-8">
          <FormField
            control={form.control}
            name="Title"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Title</FormLabel>
                <FormControl>
                  <Input {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <FormField
            control={form.control}
            name="Subtitle"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Subtitle</FormLabel>
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
          <div className="grid grid-cols-2 gap-8">
            <FormField
              control={form.control}
              name="DateStart"
              render={({ field }) => (
                <FormItem className="flex flex-col">
                  <FormLabel>Laufzeit Von</FormLabel>
                  <Popover>
                    <PopoverTrigger asChild>
                      <FormControl>
                        <Button
                          variant={"outline"}
                          className={cn(
                            "w-[240px] pl-3 text-left font-normal",
                            !field.value && "text-muted-foreground"
                          )}
                        >
                          {field.value ? (
                            format(field.value, "PPPP", { locale: de })
                          ) : (
                            <span>Bitte Auswählen</span>
                          )}
                          <CalendarIcon className="w-4 h-4 ml-auto opacity-50" />
                        </Button>
                      </FormControl>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <DayPicker
                        locale={de}
                        mode="single"
                        captionLayout="dropdown-buttons"
                        selected={new Date(field.value)}
                        onSelect={(e) => field.onChange(e?.toDateString())}
                        initialFocus
                      />
                    </PopoverContent>
                  </Popover>

                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="DateStop"
              render={({ field }) => (
                <FormItem className="flex flex-col">
                  <FormLabel>Laufzeit Bis</FormLabel>
                  <Popover>
                    <PopoverTrigger asChild>
                      <FormControl>
                        <Button
                          variant={"outline"}
                          className={cn(
                            "w-[240px] pl-3 text-left font-normal",
                            !field.value && "text-muted-foreground"
                          )}
                        >
                          {field.value ? (
                            format(field.value, "PPPP", { locale: de })
                          ) : (
                            <span>Bitte Auswählen</span>
                          )}
                          <CalendarIcon className="w-4 h-4 ml-auto opacity-50" />
                        </Button>
                      </FormControl>
                    </PopoverTrigger>
                    <PopoverContent className="w-auto p-0" align="start">
                      <DayPicker
                        locale={de}
                        mode="single"
                        captionLayout="dropdown-buttons"
                        selected={new Date(field.value)}
                        onSelect={(e) => field.onChange(e?.toDateString())}
                        initialFocus
                      />
                    </PopoverContent>
                  </Popover>

                  <FormMessage />
                </FormItem>
              )}
            />
          </div>
          <div>
            <h3 className="mb-4 text-lg font-medium">Anzeige auf Webseite</h3>
            <div className="space-y-4">
              <FormField
                control={form.control}
                name="Anzeigen"
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
          <Button type="submit">Speichern</Button>
        </form>
      </Form>
    </>
  );
}
