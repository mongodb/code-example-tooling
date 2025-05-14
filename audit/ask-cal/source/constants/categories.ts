export type CodeExampleCategory =  
  | "Syntax example"
  | "Usage example"
  | "Non-MongoDB command"
  | "Example return object"
  | "Example configuration object";
  
export const CodeExampleDisplayValues: Record<CodeExampleCategory, string> = {
  "Syntax example": "Syntax example",
  "Usage example": "Usage example",
  "Non-MongoDB command": "Non-MongoDB command",
  "Example return object": "Return object",
  "Example configuration object": "Configuration object",
};
