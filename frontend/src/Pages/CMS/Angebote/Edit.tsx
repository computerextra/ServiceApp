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
import { cn } from "@/lib/utils";
import { zodResolver } from "@hookform/resolvers/zod";
import { format } from "date-fns";
import { de } from "date-fns/locale";
import { CalendarIcon } from "lucide-react";
import { useEffect, useState } from "react";
import { DayPicker } from "react-day-picker";
import "react-day-picker/dist/style.css";
import { useForm } from "react-hook-form";
import { useNavigate, useParams } from "react-router";
import { z } from "zod";
import {
  DeleteAngebot,
  GetAngebot,
  UpdateAngebot,
} from "../../../../wailsjs/go/main/App";

const Angebot = z.object({
  ID: z.string(),
  Title: z.string(),
  Subtitle: z.string(),
  DateStart: z.string(),
  DateStop: z.string(),
  Link: z.string(),
  Image: z.string(),
  Anzeigen: z.boolean(),
});
export type Angebot = z.infer<typeof Angebot>;
const UpdateAngebotProps = Angebot;
type UpdateAngebotProps = z.infer<typeof UpdateAngebotProps>;

export default function AngebotEdit() {
  const { aid } = useParams();
  const [data, setData] = useState<Angebot | undefined>(undefined);

  useEffect(() => {
    async function x() {
      if (aid == null) return;
      const Angebot = await GetAngebot(aid);
      if (typeof Angebot == "string") {
        alert(Angebot);
        return;
      } else {
        setData({
          Anzeigen: Angebot.Anzeigen.Bool,
          DateStart: Angebot.DateStart,
          DateStop: Angebot.DateStop,
          ID: Angebot.ID,
          Image: Angebot.Image,
          Link: Angebot.Link,
          Subtitle: Angebot.Subtitle.String,
          Title: Angebot.Title,
        });
      }
    }
    void x();
  }, [aid]);

  const form = useForm<z.infer<typeof UpdateAngebotProps>>({
    resolver: zodResolver(UpdateAngebotProps),
    defaultValues: {
      Anzeigen: data?.Anzeigen,
      DateStart: data?.DateStart ?? "",
      DateStop: data?.DateStop ?? "",
      ID: data?.ID ?? "",
      Image: data?.Image ?? "",
      Link: data?.Link ?? "",
      Subtitle: data?.Subtitle ?? "",
      Title: data?.Title ?? "",
    },
  });
  const navigate = useNavigate();

  useEffect(() => {
    if (data == null) return;

    form.reset({
      Anzeigen: data.Anzeigen,
      DateStart: data.DateStart,
      DateStop: data.DateStop,
      ID: data.ID,
      Image: data.Image,
      Link: data.Link,
      Subtitle: data.Subtitle,
      Title: data.Title,
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data]);

  const onSubmit = async (values: z.infer<typeof UpdateAngebotProps>) => {
    const start = new Date(values.DateStart);
    const end = new Date(values.DateStop);
    const data: UpdateAngebotProps = {
      ID: values.ID,
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
    const res = await UpdateAngebot(
      data.ID,
      data.Title,
      data.Subtitle,
      data.DateStart,
      data.DateStop,
      data.Link,
      data.Image,
      data.Anzeigen ? "true" : "false"
    );
    if (res) navigate("/CMS/Angebote");
  };

  const handleDelete = async () => {
    if (aid == null) return;
    await DeleteAngebot(aid);
    navigate("/CMS/Angebote");
  };

  return (
    <>
      <BackButton href="/CMS/Angebote" />
      <h1 className="mt-8">{data?.Title} bearbeiten</h1>{" "}
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
