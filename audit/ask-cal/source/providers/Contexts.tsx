import { createContext } from "react";
import {
  CodeExample,
  RequestProperties,
  AiSummaryPayload,
  HandleRequestProperties,
  SearchResponse,
} from "../constants/types";

interface AcalaContextType {
  handleRequest: ({
    options,
    requestType,
  }: HandleRequestProperties) => Promise<SearchResponse>;
  reportFeedback: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  requestExample: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  getAiSummary: (payload: AiSummaryPayload) => Promise<void>;
  aiSummary: string | null;
  loadingRequest: boolean;
  apiError: string | null;
}

interface SearchContextType {
  search: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  searchQueryId: string | null;
  loadingSearch: boolean;
  results: CodeExample[];
}

export const AcalaContext = createContext<AcalaContextType | undefined>(
  undefined
);

export const SearchContext = createContext<SearchContextType | undefined>(
  undefined
);
