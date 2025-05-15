import styles from "./ResultsPage.module.css";
import { useState } from "react";

import Card from "@leafygreen-ui/card";
import Code from "@leafygreen-ui/code";
import IconButton from "@leafygreen-ui/icon-button";
import Button from "@leafygreen-ui/button";
import Icon from "@leafygreen-ui/icon";
import { Body, H2, H3 } from "@leafygreen-ui/typography";
import { CodeExample } from "../../../constants/types";

import { useAcala } from "../../../providers/UseAcala";

function Resultspage() {
  const [selectedCodeExample, setSelectedCodeExample] =
    useState<CodeExample | null>(null);

  const { results } = useAcala();

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
                      className={styles.results_code_example}
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
              <div className={styles.example_header_container}>
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

                <Button
                  leftGlyph={<Icon glyph="Sparkle" />}
                  aria-label="Some Menu"
                  className={styles.copy_button}
                  onClick={() => {
                    console.log("copy code clicked");
                  }}
                >
                  Explain this code
                </Button>
              </div>
              {selectedCodeExample.pageDescription && (
                <Body>{selectedCodeExample.pageDescription}</Body>
              )}
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
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default Resultspage;
