import { useContext } from "react";
import { AcalaContext } from "./AcalaContext";

export const useAcala = () => {
  const context = useContext(AcalaContext);

  if (!context) throw new Error("useAcala must be used within AcalaProvider");

  return context;
};
