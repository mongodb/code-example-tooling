import styles from "./Search.module.css";

// React
import { useState } from "react";
import { useNavigate } from "react-router";

// Leafygreen UI components
import { SearchInput } from "@leafygreen-ui/search-input";
import { Combobox, ComboboxOption } from "@leafygreen-ui/combobox";

// App components
import { useSearch } from "../../providers/Hooks";

// Types
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
};

function Search({ isHomepage }: SearchProps) {
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [facets, setFacets] = useState<FacetGroup>({
    programmingLanguage: "",
    category: "",
    docsSet: "",
  });

  const { search } = useSearch();
  const navigate = useNavigate();

  // TODO: this also exists on the homepage. Move to a common place.
  const handleSearch = async (event: React.FormEvent<HTMLFormElement>) => {
    // setSearchQuery("");
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
      mock: false,
    });

    if (isHomepage) {
      navigate("/results");
    }
  };

  const handleFacetChange = ({ facet, value }: Facet) => {
    switch (facet) {
      case "programmingLanguage": {
        const languageKey = mapLanguageValueToKey(value as string);

        if (!languageKey) {
          console.error("Invalid programming language key");
          return;
        }

        setFacets((previous) => ({
          ...previous,
          programmingLanguage: languageKey,
        }));

        break;
      }
      case "category": {
        const categoryKey = mapCategoryValueToKey(value as string);

        if (!categoryKey) {
          console.error("Invalid category key");
          return;
        }

        setFacets((previous) => ({
          ...previous,
          category: categoryKey,
        }));

        break;
      }
      case "docsSet": {
        const docsSetKey = mapDocsSetValueToKey(value as string);

        if (!docsSetKey) {
          console.error("Invalid docs set key");
          return;
        }

        setFacets((previous) => ({
          ...previous,
          docsSet: docsSetKey,
        }));

        break;
      }
      default:
        break;
    }
  };

  const mapLanguageValueToKey = (value: string) => {
    const languageKey = Object.keys(ProgrammingLanguageDisplayValues).find(
      (key) =>
        ProgrammingLanguageDisplayValues[key as ProgrammingLanguage] === value
    );

    return (languageKey as ProgrammingLanguage) || "";
  };

  const mapCategoryValueToKey = (value: string) => {
    const categoryKey = Object.keys(CodeExampleDisplayValues).find((key) =>
      CodeExampleDisplayValues[key as CodeExampleCategory].includes(value)
    );

    return (categoryKey as CodeExampleCategory) || "";
  };

  const mapDocsSetValueToKey = (value: string) => {
    const docsSetKey = Object.keys(DocsSetDisplayValues).find(
      (key) => DocsSetDisplayValues[key as DocsSet] === value
    );

    return (docsSetKey as DocsSet) || "";
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
        onChange={(event) => {
          setSearchQuery(event.target.value);
        }}
        aria-label="search box"
      />

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
