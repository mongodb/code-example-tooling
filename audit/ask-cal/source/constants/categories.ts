// TODO: refactor constants and types to use objects instead of relying
// on mapping keys to values and vice versa.
// This includes changing the shape that the backend expects and returns.
// For example,
// const Example = {
//   syntaxExample: {
//     value: "Example return object",
//     displayValue: "Return object",
//   }
// }

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
