import "./styles/fonts.css";
import "./styles/main.css";

import { useState } from "react";
import LeafyGreenProvider from "@leafygreen-ui/leafygreen-provider";
import Button from "@leafygreen-ui/button";
import Toggle from "@leafygreen-ui/toggle";
import { H1, Body } from "@leafygreen-ui/typography";

function App() {
  const [darkMode, setDarkMode] = useState(true);

  return (
    <LeafyGreenProvider darkMode={darkMode}>
      <header>
        <div>
          <Toggle
            aria-label="Dark mode toggle"
            checked={darkMode}
            onChange={setDarkMode}
            className="toggle-style"
            size="small"
          />
        </div>
      </header>

      <main className={`App ${darkMode && "darkmode-bg"}`}>
        <H1>Vite + React + LeafyGreen</H1>

        <section>
          <Body baseFontSize={16}>
            Click the button below to see a LeafyGreen component in action!
          </Body>
          <Button>Tada!</Button>
        </section>
      </main>
    </LeafyGreenProvider>
  );
}

export default App;
