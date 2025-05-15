import styles from "./ResultsPage.module.css";
import { useState } from "react";

import Card from "@leafygreen-ui/card";
import Code from "@leafygreen-ui/code";
import IconButton from "@leafygreen-ui/icon-button";
import Button from "@leafygreen-ui/button";
import Icon from "@leafygreen-ui/icon";
import { PageLoader } from "@leafygreen-ui/loading-indicator";
import { Body, H2, H3 } from "@leafygreen-ui/typography";
import {
  DisplayMode,
  Drawer,
  DrawerStackProvider,
} from "@leafygreen-ui/drawer";
import { CodeExample } from "../../constants/types";

import { useAcala } from "../../providers/UseAcala";

function Resultspage() {
  const [selectedCodeExample, setSelectedCodeExample] =
    useState<CodeExample | null>(null);
  const [openAiDrawer, setOpenAiDrawer] = useState(false);

  const { results, getAiSummary, aiSummary, loading } = useAcala();

  const handleAiSummary = async (code: string, pageUrl: string) => {
    try {
      await getAiSummary({ code, pageUrl });
    } catch (error) {
      console.error("Error fetching AI summary:", error);
    }
  };

  return (
    <div className={styles.results_page}>
      <div className={styles.horizontal_container}>
        <div className={styles.results_container}>
          {results && (
            <div className={styles.results}>
              <Body>{results.length} results found</Body>

              <div className={styles.results_list}>
                {results.map((result, index) => (
                  <Card
                    as="div"
                    contentStyle="clickable"
                    onClick={() => {
                      setSelectedCodeExample(result);

                      console.log(
                        "Selected code example:",
                        selectedCodeExample
                      );
                    }}
                    key={index}
                  >
                    <H3>{result.pageTitle}</H3>

                    <Code
                      language={result.language}
                      expandable={true}
                      className={styles.code_example}
                    >
                      {result.code}
                    </Code>
                  </Card>
                ))}
              </div>
            </div>
          )}
        </div>

        <div className={styles.example_container}>
          {selectedCodeExample && (
            <>
              <div className={styles.example_header}>
                <H2>{selectedCodeExample.pageTitle}</H2>

                <IconButton
                  aria-label="Some Menu"
                  onClick={() => {
                    window.open(selectedCodeExample.pageUrl, "_blank");
                  }}
                >
                  <Icon
                    glyph="OpenNewTab"
                    size={"xlarge"}
                  />
                </IconButton>
              </div>

              {selectedCodeExample.pageDescription && (
                <Body>{selectedCodeExample.pageDescription}</Body>
              )}

              <div className={styles.example_body}>
                <Code
                  language={
                    selectedCodeExample.language
                      ? selectedCodeExample.language
                      : "java"
                  }
                  className={styles.code_example}
                  showLineNumbers={true}
                  onCopy={() => {
                    console.log("copy code clicked");
                  }}
                >
                  {selectedCodeExample.code}
                </Code>

                <Button
                  leftGlyph={<Icon glyph="Sparkle" />}
                  aria-label="Some Menu"
                  className={styles.summary_button}
                  onClick={() => {
                    setOpenAiDrawer(true);
                    handleAiSummary(
                      selectedCodeExample.code,
                      selectedCodeExample.pageUrl
                    );
                    // setOpenResultsDrawer(false);
                  }}
                >
                  Explain this code
                </Button>
              </div>
            </>
          )}
        </div>

        {selectedCodeExample && (
          <div className={styles.summary_container}>
            <DrawerStackProvider>
              <Drawer
                displayMode={DisplayMode.Overlay}
                onClose={() => {
                  setOpenAiDrawer(false);
                  // setOpenResultsDrawer(true);
                }}
                open={openAiDrawer}
                title="Drawer Title"
              >
                {loading ? (
                  <PageLoader description="Asking the robots..." />
                ) : (
                  <Body
                    baseFontSize={16}
                    className={styles.ai_summary}
                  >
                    {aiSummary}
                  </Body>
                )}
              </Drawer>
            </DrawerStackProvider>
          </div>
        )}
      </div>
    </div>
  );
}

export default Resultspage;
