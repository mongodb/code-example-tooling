import "./styles/fonts.css";
import "./styles/main.css";

import { useState } from "react";

// Leafygreen UI components
import LeafyGreenProvider from "@leafygreen-ui/leafygreen-provider";

import Header from "./components/header/Header";
import Sidebar from "./components/sidebar/Sidebar";
import Viewer from "./components/viewer/Viewer";
import Footer from "./components/footer/Footer";

function App() {
  const [darkMode, setDarkMode] = useState(true);

  return (
    <LeafyGreenProvider darkMode={darkMode}>
      <div className={`App ${darkMode && "darkmode-bg"}`}>
        <Header
          darkMode={darkMode}
          setDarkMode={setDarkMode}
        />
        <Sidebar />
        <Viewer />
        <Footer />
      </div>
    </LeafyGreenProvider>
  );
}

export default App;
