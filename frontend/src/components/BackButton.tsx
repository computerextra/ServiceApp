import { Link } from "react-router";
import { Button } from "./ui/button";

export default function BackButton({ href }: { href: string }) {
  return (
    <div className="fixed">
      <Button className="mb-2" variant="secondary" asChild>
        <Link to={href}>Zurück</Link>
      </Button>
    </div>
  );
}
