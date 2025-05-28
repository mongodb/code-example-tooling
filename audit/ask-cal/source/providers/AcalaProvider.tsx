import { useState, ReactNode } from "react";
import { AcalaContext } from "./Contexts";

import {
  RequestType,
  RequestProperties,
  Requests,
  HandleRequestProperties,
  AiSummaryPayload,
} from "../constants/types";

// Acala stands for "Ask CAL API".
export const AcalaProvider = ({ children }: { children: ReactNode }) => {
  const [loadingRequest, setLoadingRequest] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [aiSummary, setAiSummary] = useState<string | null>(null);

  const baseUrl =
    window.location.hostname === "localhost"
      ? "http://localhost:8888"
      : "https://ask-cal.netlify.app";

  const handleRequest = async ({
    options,
    requestType,
  }: HandleRequestProperties) => {
    setLoadingRequest(true);
    setApiError(null);

    // Map request types to Netlify Function endpoints
    let requestUrl: string;

    switch (requestType) {
      case RequestType.Search:
        requestUrl = `${baseUrl + Requests.Search.url}`;

        break;
      case RequestType.ReportFeedback:
        requestUrl = new URL(baseUrl, Requests.ReportFeedback.url).toString();

        break;
      case RequestType.RequestExample:
        requestUrl = new URL(baseUrl, Requests.RequestExample.url).toString();

        break;
      case RequestType.GetAiSummary:
        requestUrl = new URL(baseUrl, Requests.GetAiSummary.url).toString();

        break;
      default:
        throw new Error("Invalid request type");
    }

    try {
      const response = await fetch(requestUrl, options);
      if (!response.ok) throw new Error(response.statusText);

      if (requestType === RequestType.GetAiSummary && response.body) {
        // Stream handling
        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let result = "";
        let done = false;

        setLoadingRequest(false);

        // TODO: investigate why it isn't rendering chunks and we only get
        // the final result.
        while (!done) {
          const { value, done: streamDone } = await reader.read();
          done = streamDone;
          if (value) {
            const chunk = decoder.decode(value, { stream: !done });
            result += chunk;
            setAiSummary(result);
          }
        }
        return { summary: result };
      }

      return await response.json();
    } catch (error: unknown) {
      if (error instanceof Error) {
        setApiError(error.message);
      } else if (error instanceof TypeError) {
        setApiError("Network error. Please try again later.");
      } else if (error instanceof SyntaxError) {
        setApiError("Invalid response format.");
      }

      throw error;
    } finally {
      setLoadingRequest(false);
    }
  };

  const getAiSummary = async (payload: AiSummaryPayload) => {
    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    };

    const data = await handleRequest({
      options,
      requestType: RequestType.GetAiSummary,
    });

    if (data) {
      setAiSummary(data.summary);
    } else {
      setApiError("Failed to fetch AI summary.");
    }

    return;
  };

  const reportFeedback = async ({ bodyContent }: RequestProperties) => {
    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyContent),
    };

    return await handleRequest({
      options,
      requestType: RequestType.ReportFeedback,
    });
  };

  const requestExample = async ({ bodyContent }: RequestProperties) => {
    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyContent),
    };

    return await handleRequest({
      options,
      requestType: RequestType.RequestExample,
    });
  };

  return (
    <AcalaContext.Provider
      value={{
        handleRequest,
        reportFeedback,
        requestExample,
        getAiSummary,
        aiSummary,
        loadingRequest,
        apiError,
      }}
    >
      {children}
    </AcalaContext.Provider>
  );
};
