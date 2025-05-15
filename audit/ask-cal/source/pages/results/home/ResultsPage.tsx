import styles from "./ResultsPage.module.css";
import { useState } from "react";

import { SearchInput } from "@leafygreen-ui/search-input";
import { Combobox, ComboboxOption } from "@leafygreen-ui/combobox";
import Card from "@leafygreen-ui/card";
import Code from "@leafygreen-ui/code";
import IconButton from "@leafygreen-ui/icon-button";
import Button from "@leafygreen-ui/button";
import Icon from "@leafygreen-ui/icon";
import { FacetGroup, Facet, CodeExample } from "../../../constants/types";

import { useAcala } from "../../../providers/UseAcala";

import {
  CodeExampleDisplayValues,
  CodeExampleCategory,
} from "../../../constants/categories";
import { DocsSetDisplayValues, DocsSet } from "../../../constants/docsSets";
import {
  ProgrammingLanguageDisplayValues,
  ProgrammingLanguage,
} from "../../../constants/programmingLanguages";
import { Body, H2, H3 } from "@leafygreen-ui/typography";

function Resultspage() {
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [facets, setFacets] = useState<FacetGroup>({
    programmingLanguage: "",
    category: "",
    docsSet: "",
  });
  const [selectedCodeExample, setSelectedCodeExample] =
    useState<CodeExample | null>(null);

  const { search, results } = useAcala();

  // TODO: this also exists on the homepage. Move to a common place.
  const handleSearch = async (event: React.FormEvent<HTMLFormElement>) => {
    setSearchQuery("");
    // Get the value from the input element. Look for the role "search".
    const inputElement = event.currentTarget.querySelector(
      "input"
    ) as HTMLInputElement;
    const value = inputElement.value;

    if (!value) {
      console.error("Search input is empty");
      return;
    }

    setSearchQuery(value);

    // Mock search, as getting CORS errors
    // TODO: make this work for real
    await search({
      bodyContent: {
        queryString: searchQuery,
        LanguageFacet: facets.programmingLanguage,
        CategoryFacet: facets.category,
        docsSet: facets.docsSet,
      },
      mock: true,
    });
  };

  const handleFacetChange = ({ facet, value }: Facet) => {
    setFacets((previous) => {
      const updatedFacetGroup: FacetGroup = {
        programmingLanguage:
          facet === "programmingLanguage"
            ? (value as ProgrammingLanguage)
            : previous?.programmingLanguage,
        category:
          facet === "category"
            ? (value as CodeExampleCategory)
            : previous?.category,
        docsSet: facet === "docsSet" ? (value as DocsSet) : previous?.docsSet,
      };
      return updatedFacetGroup;
    });
  };

  return (
    <div className={styles.results_page}>
      <header>
        <SearchInput
          onSubmit={(event) => {
            handleSearch(event);
          }}
          aria-label="search box"
          className={styles.search_input}
        ></SearchInput>

        <div className={styles.facet_container}>
          <Combobox
            label="Programming Language"
            placeholder="Select language"
            size="xsmall"
            onChange={(value: string | null) => {
              if (value) {
                handleFacetChange({
                  facet: "programmingLanguage",
                  value: value,
                });
              }
            }}
          >
            {Object.values(ProgrammingLanguageDisplayValues).map((language) => (
              <ComboboxOption
                key={language}
                value={language}
              />
            ))}
          </Combobox>

          <Combobox
            label="Category"
            placeholder="Select category"
            size="xsmall"
            onChange={(value: string | null) => {
              if (value) {
                handleFacetChange({
                  facet: "category",
                  value: value,
                });
              }
            }}
          >
            {Object.values(CodeExampleDisplayValues).map((category) => (
              <ComboboxOption
                key={category}
                value={category}
              />
            ))}
          </Combobox>

          <Combobox
            label="Documentation set"
            placeholder="Select docs set"
            size="xsmall"
            onChange={(value: string | null) => {
              if (value) {
                handleFacetChange({
                  facet: "docsSet",
                  value: value,
                });
              }
            }}
          >
            {Object.values(DocsSetDisplayValues).map((docsSet) => (
              <ComboboxOption
                key={docsSet}
                value={docsSet}
              />
            ))}
          </Combobox>
        </div>
      </header>

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
