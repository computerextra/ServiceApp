import { Button } from "@/components/ui/button";
import { NavLink } from "react-router";

export default function Home() {
  return (
    <>
      <h1 className="text-4xl">ServiceApp</h1>
      <h2 className="mt-2 text-2xl">Schnellwahl</h2>
      <div className="grid grid-cols-2 gap-8 mt-4">
        <Button asChild>
          <NavLink to="/Seriennummer">Seriennummer</NavLink>
        </Button>
        <Button asChild>
          <NavLink to="/Info">Info an Kunde</NavLink>
        </Button>
        <Button asChild>
          <NavLink to="/Aussteller">Aussteller</NavLink>
        </Button>
        <Button asChild>
          <NavLink to="/Label">Label Sync</NavLink>
        </Button>
        <Button asChild>
          <NavLink to="/Warenlieferung">Warenlieferung</NavLink>
        </Button>
        <Button asChild>
          <NavLink to="/CMS">CMS</NavLink>
        </Button>
      </div>
    </>
  );
}
