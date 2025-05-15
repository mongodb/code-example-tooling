import { createContext } from "react";
import { CodeExample, RequestProperties } from "../constants/types";

interface AcalaContextType {
  search: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  searchQueryId: string | null;
  reportFeedback: ({ bodyContent, mock }: RequestProperties) => Promise<any>;
  requestExample: ({ bodyContent, mock }: RequestProperties) => Promise<any>;
  results: CodeExample[];
  loading: boolean;
  apiError: string | null;
}

export const AcalaContext = createContext<AcalaContextType | undefined>(
  undefined
);
