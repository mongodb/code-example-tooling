import { useState, ReactNode } from "react";
import { AcalaContext } from "./AcalaContext";

// List of possible requests against the API
enum RequestType {
  Search = "search",
  ReportFeedback = "reportFeedback",
  RequestExample = "requestExample",
}

interface RequestProperties {
  bodyContent: unknown;
  mock: boolean;
}

interface HandleRequestProperties {
  url: string;
  options: RequestInit;
  requestType: RequestType;
}

// TODO: remove the mock features when the API is ready

// Acala stands for "Ask CAL API".
export const AcalaProvider = ({ children }: { children: ReactNode }) => {
  const [loading, setLoading] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [results, setResults] = useState<unknown[]>([]);
  const [searchQueryId, setSearchQueryId] = useState<string | null>(null);

  const baseUrl =
    window.location.hostname === "localhost"
      ? "http://localhost:8888"
      : "https://ask-cal.netlify.app";

  const handleRequest = async ({
    url,
    options,
    requestType,
  }: HandleRequestProperties) => {
    setLoading(true);
    setApiError(null);

    console.log("API request:", { url, options, requestType });

    try {
      const response = await fetch(url, options);
      if (!response.ok) throw new Error(response.statusText);

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
      setLoading(false);
    }
  };

  const handleMockRequest = async ({
    url,
    options,
    requestType,
  }: HandleRequestProperties) => {
    setLoading(true);
    setApiError(null);

    setResults([]);

    console.log("Mock request:", { url, options });

    try {
      const response = {
        ok: true,
        json: async () => {
          await new Promise((resolve) => setTimeout(resolve, 500));

          switch (requestType) {
            case RequestType.Search:
              return {
                queryId: "fake-string-id",
                codeExamples: [
                  { someShape: "really", page: "something" },
                  { someShape: "another", page: "yaaaay!" },
                ],
              };
            case RequestType.ReportFeedback:
              return { success: true };
            case RequestType.RequestExample:
              return { success: true };
            default:
              throw new Error("Invalid request type");
          }
        },
      };

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
      setLoading(false);
    }
  };

  const search = async ({ bodyContent, mock }: RequestProperties) => {
    const url = `${baseUrl}/.netlify/functions/search`;
    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyContent),
    };

    if (mock) {
      const data = await handleMockRequest({
        url,
        options,
        requestType: RequestType.Search,
      });

      setSearchQueryId(data.queryId as string);

      if (data.codeExamples) {
        setResults(data.codeExamples);
        return results;
      }

      return;
    }

    const data = await handleRequest({
      url,
      options,
      requestType: RequestType.Search,
    });

    setResults(data);

    return results;
  };

  const reportFeedback = async ({ bodyContent, mock }: RequestProperties) => {
    const url = `${baseUrl}/.netlify/functions/feedback`;
    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyContent),
    };

    if (mock) {
      return await handleMockRequest({
        url,
        options,
        requestType: RequestType.ReportFeedback,
      });
    }

    return await handleRequest({
      url,
      options,
      requestType: RequestType.ReportFeedback,
    });
  };

  const requestExample = async ({ bodyContent, mock }: RequestProperties) => {
    const url = `${baseUrl}/.netlify/functions/request-example`;
    const options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(bodyContent),
    };

    if (mock) {
      return await handleMockRequest({
        url,
        options,
        requestType: RequestType.RequestExample,
      });
    }

    return await handleRequest({
      url,
      options,
      requestType: RequestType.RequestExample,
    });
  };

  return (
    <AcalaContext.Provider
      value={{
        search,
        searchQueryId,
        reportFeedback,
        requestExample,
        results,
        loading,
        apiError,
      }}
    >
      {children}
    </AcalaContext.Provider>
  );
};
