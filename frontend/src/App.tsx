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
import AbteilungOverview from "./Pages/CMS/Abteilungen/Overview";
import AbteilungNew from "./Pages/CMS/Abteilungen/New";
import AbteilungEdit from "./Pages/CMS/Abteilungen/Edit";
import MitarbeiterNew from "./Pages/CMS/Mitarbeiter/New";
import MitarbeiterEdit from "./Pages/CMS/Mitarbeiter/Edit";
import PartnerNew from "./Pages/CMS/Partner/New";
import PartnerEdit from "./Pages/CMS/Partner/Edit";
import AngebotNew from "./Pages/CMS/Angebote/New";
import JobsNew from "./Pages/CMS/Jobs/New";
import JobsEdit from "./Pages/CMS/Jobs/Edit";
import MitarbeiterOverview from "./Pages/CMS/Mitarbeiter/Overview";
import PartnerOverview from "./Pages/CMS/Partner/Overview";
import AngebotOverview from "./Pages/CMS/Angebote/Overview";
import AngebotEdit from "./Pages/CMS/Angebote/Edit";
import JobsOverview from "./Pages/CMS/Jobs/Overview";

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
            <Route path="CMS">
              <Route index element={<Cms />} />

              <Route path="Abteilungen">
                <Route index element={<AbteilungOverview />} />
                <Route path="Neu" element={<AbteilungNew />} />
                <Route path=":aid" element={<AbteilungEdit />} />
              </Route>

              <Route path="Mitarbeiter">
                <Route index element={<MitarbeiterOverview />} />
                <Route path="Neu" element={<MitarbeiterNew />} />
                <Route path=":mid" element={<MitarbeiterEdit />} />
              </Route>

              <Route path="Partner">
                <Route index element={<PartnerOverview />} />
                <Route path="Neu" element={<PartnerNew />} />
                <Route path=":pid" element={<PartnerEdit />} />
              </Route>

              <Route path="Angebote">
                <Route index element={<AngebotOverview />} />
                <Route path="Neu" element={<AngebotNew />} />
                <Route path=":aid" element={<AngebotEdit />} />
              </Route>

              <Route path="Jobs">
                <Route index element={<JobsOverview />} />
                <Route path="Neu" element={<JobsNew />} />
                <Route path=":jid" element={<JobsEdit />} />
              </Route>
            </Route>
          </Routes>
        </div>
      </HashRouter>
    </ThemeProvider>
  );
}

export default App;
