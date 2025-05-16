import styles from "./ResultsPage.module.css";
import { useState } from "react";

import Card from "@leafygreen-ui/card";
import Code from "@leafygreen-ui/code";
import Button from "@leafygreen-ui/button";
import Icon from "@leafygreen-ui/icon";
import Badge from "@leafygreen-ui/badge";
import { PageLoader } from "@leafygreen-ui/loading-indicator";
import { Body, H2, H3, Link } from "@leafygreen-ui/typography";
import {
  DisplayMode,
  Drawer,
  DrawerStackProvider,
} from "@leafygreen-ui/drawer";
import { CodeExample } from "../../constants/types";

import { useAcala } from "../../providers/UseAcala";
import CodeExamplePlaceholder from "../../components/code-example-placeholder/CodeExamplePlaceholder";

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

  const parseLanguage = (language: string) => {
    switch (language) {
      case "undefined":
        return "javascript";
      case "text":
        return "javascript";
      default:
        return language;
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
                    }}
                    key={index}
                  >
                    <H3>{result.pageTitle}</H3>

                    <Code
                      language={parseLanguage(result.language)}
                      expandable={true}
                      className={styles.code_example}
                    >
                      {result.code}
                    </Code>

                    <div className={styles.badge_container}>
                      <Badge variant="blue">{result.language}</Badge>
                      <Badge variant="green">{result.category}</Badge>
                    </div>
                  </Card>
                ))}
              </div>
            </div>
          )}

          {!results && (
            <div>
              <Body>No results found. Try a different search query.</Body>
            </div>
          )}
        </div>

        <div className={styles.example_container}>
          {!selectedCodeExample && <CodeExamplePlaceholder />}

          {selectedCodeExample && (
            <>
              <H2>{selectedCodeExample.pageTitle}</H2>
              <Link href={selectedCodeExample.pageUrl}>
                {" "}
                {selectedCodeExample.pageUrl}{" "}
              </Link>

              {selectedCodeExample.pageDescription && (
                <Body
                  baseFontSize={16}
                  className={styles.page_description}
                >
                  {selectedCodeExample.pageDescription}
                </Body>
              )}

              <div className={styles.example_body}>
                <Code
                  language={parseLanguage(selectedCodeExample.language)}
                  className={styles.code_example}
                  showLineNumbers={true}
                  onCopy={() => {
                    navigator.clipboard.writeText(selectedCodeExample.code);
                  }}
                >
                  {selectedCodeExample.code}
                </Code>

                {!openAiDrawer && (
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
                    }}
                  >
                    Explain this code
                  </Button>
                )}
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
                title="AI Summary"
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
