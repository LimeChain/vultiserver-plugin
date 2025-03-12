/// <reference types="vite-plugin-svgr/client" />

import { BrowserRouter, Route, Routes } from "react-router-dom";
import Plugin from "./screens/plugin/Plugin";
import Marketplace from "./screens/marketplace/Marketplace";
import Layout from "./Layout";
import PluginDetail from "./screens/plugin-detail/PluginDetail";

const App = () => {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Marketplace />} />
          <Route path="/plugin-detail/:id" element={<PluginDetail />} />
          <Route path="/plugin/:id" element={<Plugin />} />
        </Route>
        {/* <Route path="/">
          <Route index element={<Marketplace />} />
        </Route>
        <Route path="/dca">
          <Route index element={<Plugin />} />
        </Route> */}
      </Routes>
    </BrowserRouter>
  );
};

export default App;
