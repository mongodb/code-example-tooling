import "./assets/styles/fonts.css";
import "./assets/styles/main.css";
import styles from "./App.module.css";

// React
import { useState } from "react";
import { BrowserRouter, Routes, Route } from "react-router";

// Leafygreen UI components
import LeafyGreenProvider from "@leafygreen-ui/leafygreen-provider";
import { AcalaProvider } from "./providers/AcalaProvider";
import { SearchProvider } from "./providers/SearchProvider";
import Toggle from "@leafygreen-ui/toggle";

// App components
import Homepage from "./pages/home/HomePage";
import Resultspage from "./pages/results/ResultsPage";

function App() {
  const [darkMode, setDarkMode] = useState(false);

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
            <BrowserRouter>
              <Routes>
                <Route
                  path="/"
                  element={<Homepage />}
                />
                <Route
                  path="/results"
                  element={<Resultspage />}
                />
              </Routes>
              {/* <Header
                setIsHomepage={setIsHomepage}
                isHomepage={isHomepage}
              />

              {isHomepage ? (
                <Homepage setIsHomepage={setIsHomepage} />
              ) : (
                <Resultspage />
              )} */}
            </BrowserRouter>
          </SearchProvider>
        </AcalaProvider>
      </div>
    </LeafyGreenProvider>
  );
}

export default App;
