import { useContext } from "react";
import { AcalaContext, SearchContext } from "./Contexts";

export const useAcala = () => {
  const context = useContext(AcalaContext);

  if (!context) throw new Error("useAcala must be used within AcalaProvider");

  return context;
};

export const useSearch = () => {
  const context = useContext(SearchContext);

  if (!context) throw new Error("useSearch must be used within SearchProvider");

  return context;
};
