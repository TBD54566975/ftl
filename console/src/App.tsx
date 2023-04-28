import { Route, Routes } from "react-router-dom";
import Modules from "./components/Modules";
import Logs from "./components/Logs";
import Layout from "./components/Layout";

function App() {
  return (
    <>
      <Routes>
        <Route path="/" element={<Layout />}>
          <Route index element={<Modules />} />
          <Route path="logs" element={<Logs />} />
        </Route>
      </Routes>
    </>
  );
}

export default App;
