import { HashRouter, Routes, Route, NavLink } from "react-router";
import Home from "./Pages/Home";
import Seriennummer from "./Pages/Seriennummer";
import { ThemeProvider } from "./components/theme-provider";
import NavBar from "./components/NavBar";
import Info from "./Pages/Info";
import Aussteller from "./Pages/Aussteller";
import Label from "./Pages/Label";
import Warenlieferung from "./Pages/Warenlieferung";
import Cms from "./Pages/CMS/Cms";

function App() {
  // Router
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <HashRouter basename={"/"}>
        <NavBar />
        <div className="container mx-auto mt-5 text-center">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/Seriennummer" element={<Seriennummer />} />
            <Route path="/Info" element={<Info />} />
            <Route path="/Aussteller" element={<Aussteller />} />
            <Route path="/Label" element={<Label />} />
            <Route path="/Warenlieferung" element={<Warenlieferung />} />
            <Route path="/CMS" element={<Cms />} />
          </Routes>
        </div>
      </HashRouter>
    </ThemeProvider>
  );
}

export default App;
