import { useState, ReactNode } from "react";
import { SearchContext } from "./Contexts";
import { useAcala } from "./Hooks";

import {
  SearchResponse,
  CodeExample,
  RequestType,
  RequestProperties,
} from "../constants/types";

export const SearchProvider = ({ children }: { children: ReactNode }) => {
  const [loadingSearch, setLoadingSearch] = useState(false);
  const [results, setResults] = useState<CodeExample[]>([]);
  const [searchQueryId, setSearchQueryId] = useState<string | null>(null);
  const [requestObject, setRequestObject] = useState<RequestProperties | null>(
    null
  );

  const { handleRequest } = useAcala();

  const search = async ({ bodyContent }: RequestProperties) => {
    // Handle empty input before making the request
    if (bodyContent.queryString === "") {
      console.error("Search input element has no value");

      return;
    }

    setLoadingSearch(true);
    // Store the request info for referencing elsewhere
    setRequestObject({ bodyContent });

    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyContent),
    };

    const data = (await handleRequest({
      options,
      requestType: RequestType.Search,
    })) as SearchResponse;
    const rawResults = data.codeExamples;

    // for every result in rawResults, look at the pageTitle and remove
    // the substring " - MongoDB Docs" from the end of the string.
    rawResults.forEach((result) => {
      if (result.pageTitle.endsWith(" - MongoDB Docs")) {
        result.pageTitle = result.pageTitle.slice(
          0,
          result.pageTitle.length - " - MongoDB Docs".length
        );
      }
    });

    setSearchQueryId(data.queryId as string);
    setResults(rawResults);
    setLoadingSearch(false);

    return;
  };

  return (
    <SearchContext.Provider
      value={{
        search,
        searchQueryId,
        requestObject,
        loadingSearch,
        results,
      }}
    >
      {children}
    </SearchContext.Provider>
  );
};
