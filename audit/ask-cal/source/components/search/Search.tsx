import styles from "./Search.module.css";
import { useState } from "react";

import { SearchInput } from "@leafygreen-ui/search-input";
import { Combobox, ComboboxOption } from "@leafygreen-ui/combobox";

import { useAcala } from "../../providers/UseAcala";

import { FacetGroup, Facet } from "../../constants/types";
import {
  CodeExampleDisplayValues,
  CodeExampleCategory,
} from "../../constants/categories";
import { DocsSetDisplayValues, DocsSet } from "../../constants/docsSets";
import {
  ProgrammingLanguageDisplayValues,
  ProgrammingLanguage,
} from "../../constants/programmingLanguages";

type SearchProps = {
  isHomepage: boolean;
  setIsHomepage: React.Dispatch<React.SetStateAction<boolean>>;
};

function Search({ isHomepage, setIsHomepage }: SearchProps) {
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [facets, setFacets] = useState<FacetGroup>({
    programmingLanguage: "",
    category: "",
    docsSet: "",
  });

  const { search } = useAcala();

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

    setIsHomepage(!isHomepage);
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
    <div
      className={
        !isHomepage ? styles.search_block : styles.search_block_homepage
      }
    >
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
    </div>
  );
}

export default Search;
