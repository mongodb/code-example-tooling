import styles from "./HomePage.module.css";
import { useState, Dispatch, SetStateAction } from "react";

import { SearchInput } from "@leafygreen-ui/search-input";
import { Combobox, ComboboxOption } from "@leafygreen-ui/combobox";

import LogoBlock from "../../components/logoblock/LogoBlock";
import { useAcala } from "../../providers/UseAcala";

// Types
import {
  CodeExampleDisplayValues,
  CodeExampleCategory,
} from "../../constants/categories";
import { DocsSetDisplayValues, DocsSet } from "../../constants/docsSets";
import {
  ProgrammingLanguageDisplayValues,
  ProgrammingLanguage,
} from "../../constants/programmingLanguages";

import { Facet, FacetGroup } from "../../constants/types";

interface HomepageProps {
  setIsHomepage: Dispatch<SetStateAction<boolean>>;
}

function Homepage({ setIsHomepage }: HomepageProps) {
  // TODO: move this into AcalaProvider. Also consider making search its own
  // provider and hook.
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [facets, setFacets] = useState<FacetGroup>({
    programmingLanguage: "",
    category: "",
    docsSet: "",
  });

  const { search, results, loading } = useAcala();

  const handleSearch = async (event: React.FormEvent<HTMLFormElement>) => {
    // Get the value from the input element. Look for the role "search".
    const inputElement = event.currentTarget.querySelector(
      "input"
    ) as HTMLInputElement;
    const value = inputElement.value;

    console.log("Search value: ", value);

    if (!value) {
      console.error("Search input is empty");
      return;
    }

    setSearchQuery(value);

    // Mock search, as getting CORS errors
    // TODO: make this work for real
    await search({
      bodyContent: {
        QueryString: searchQuery,
        LanguageFacet: facets.programmingLanguage,
        CategoryFacet: facets.category,
        DocsSet: facets.docsSet,
      },
      mock: true,
    });

    setIsHomepage(false);
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
    <div className={styles.homepage}>
      <LogoBlock />

      {/* TODO: abstract to a component */}
      <div className={styles.facet_container}>
        <Combobox
          label="Programming Language"
          placeholder="Select language"
          size="xsmall"
          onChange={(value: string | null) => {
            handleFacetChange({
              facet: "programmingLanguage",
              value: value,
            });
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
            handleFacetChange({
              facet: "category",
              value: value,
            });
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
            handleFacetChange({
              facet: "docsSet",
              value: value,
            });
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

      <SearchInput
        onSubmit={(event) => {
          handleSearch(event);
        }}
        aria-label="search box"
        className={styles.search_input}
        size="large"
      ></SearchInput>
    </div>
  );
}

export default Homepage;
