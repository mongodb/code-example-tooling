import "./styles/fonts.css";
import "./styles/main.css";

import { useState } from "react";

// Leafygreen UI components
import LeafyGreenProvider from "@leafygreen-ui/leafygreen-provider";

import Homepage from "./pages/home/HomePage";
import Resultspage from "./pages/results/home/ResultsPage";

function App() {
  const [darkMode, setDarkMode] = useState(false);
  const [isHomepage, setIsHomepage] = useState(true);

  return (
    <LeafyGreenProvider darkMode={darkMode}>
      <div className={`App ${darkMode && "darkmode-bg"}`}>
        {/* 
        <Sidebar />
        <Viewer />
        <Footer /> */}

        {isHomepage ? (
          <Homepage
            setIsHomepage={setIsHomepage}
            setDarkMode={setDarkMode}
            darkMode={darkMode}
          />
        ) : (
          <Resultspage />
        )}
      </div>
    </LeafyGreenProvider>
  );
}

export default App;
