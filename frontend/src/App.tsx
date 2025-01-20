import { HashRouter, Routes, Route, NavLink } from "react-router";
import Home from "./Pages/Home";
import Seriennummer from "./Pages/Seriennummer";
import { ThemeProvider } from "./components/theme-provider";
import NavBar from "./components/NavBar";

function App() {
  // Router
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <HashRouter basename={"/"}>
        <NavBar />
        <div className="mt-5 container text-center mx-auto">
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/Seriennummer" element={<Seriennummer />} />
            <Route path="/Info" element={<>Info</>} />
            <Route path="/Aussteller" element={<>Aussteller</>} />
            <Route path="/Label" element={<>Label</>} />
            <Route path="/Warenlieferung" element={<>Warenlieferung</>} />
            <Route path="/CMS" element={<>CMS</>} />
          </Routes>
        </div>
      </HashRouter>
    </ThemeProvider>
  );
}

export default App;
