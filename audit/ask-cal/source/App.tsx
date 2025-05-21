import "./assets/styles/fonts.css";
import "./assets/styles/main.css";
import styles from "./App.module.css";

import { useState } from "react";

// Leafygreen UI components
import LeafyGreenProvider from "@leafygreen-ui/leafygreen-provider";
import { AcalaProvider } from "./providers/AcalaProvider";
import { SearchProvider } from "./providers/SearchProvider";
import Toggle from "@leafygreen-ui/toggle";

import Homepage from "./pages/home/HomePage";
import Resultspage from "./pages/results/ResultsPage";
import Header from "./components/header/Header";

function App() {
  const [darkMode, setDarkMode] = useState(false);
  const [isHomepage, setIsHomepage] = useState(true);

  return (
    <LeafyGreenProvider darkMode={darkMode}>
      <div className={`App ${darkMode && "darkmode-bg"}`}>
        <Toggle
          aria-label="Dark mode toggle"
          checked={darkMode}
          onChange={setDarkMode}
          size="small"
          className={styles.theme_toggle}
        />
        <AcalaProvider>
          <SearchProvider>
            <Header
              setIsHomepage={setIsHomepage}
              isHomepage={isHomepage}
            />

            {isHomepage ? (
              <Homepage setIsHomepage={setIsHomepage} />
            ) : (
              <Resultspage />
            )}
          </SearchProvider>
        </AcalaProvider>
      </div>
    </LeafyGreenProvider>
  );
}

export default App;
