import { NavLink } from "react-router";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
  navigationMenuTriggerStyle,
} from "./ui/navigation-menu";

export default function NavBar() {
  return (
    <div className="container mx-auto mt-1">
      <NavigationMenu className="mx-auto">
        <NavigationMenuList>
          <NavigationMenuItem>
            <NavLink to="/">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Home
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/Seriennummer">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Seriennummer
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/Info">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Info
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/Aussteller">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Aussteller
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/Label">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Label
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/Warenlieferung">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Warenlieferung
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/CMS">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                CMS
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
          <NavigationMenuItem>
            <NavLink to="/Sage">
              <NavigationMenuLink className={navigationMenuTriggerStyle()}>
                Sage
              </NavigationMenuLink>
            </NavLink>
          </NavigationMenuItem>
        </NavigationMenuList>
      </NavigationMenu>
    </div>
  );
}
