import { createContext } from "react";
import {
  CodeExample,
  RequestProperties,
  AiSummaryPayload,
} from "../constants/types";

interface AcalaContextType {
  search: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  searchQueryId: string | null;
  reportFeedback: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  requestExample: ({ bodyContent, mock }: RequestProperties) => Promise<void>;
  getAiSummary: (payload: AiSummaryPayload) => Promise<void>;
  aiSummary: string | null;
  results: CodeExample[];
  loading: boolean;
  apiError: string | null;
}

export const AcalaContext = createContext<AcalaContextType | undefined>(
  undefined
);
