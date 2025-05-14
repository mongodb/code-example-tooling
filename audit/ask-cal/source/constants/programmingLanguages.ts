export type ProgrammingLanguage =  
  | "bash"  
  | "c"  
  | "cpp"  
  | "csharp"  
  | "go"  
  | "java"  
  | "javascript"  
  | "json"  
  | "kotlin"  
  | "php"  
  | "python"  
  | "ruby"  
  | "rust"  
  | "scala"  
  | "shell"  
  | "swift"  
  | "text"  
  | "typescript"  
  | "undefined"
  | "xml"  
  | "yaml";  

export const ProgrammingLanguageDisplayValues: Record<ProgrammingLanguage, string> = {  
  bash: "Bash",  
  c: "C",  
  cpp: "C++",  
  csharp: "C#",  
  go: "Go",  
  java: "Java",  
  javascript: "JavaScript",  
  json: "JSON",  
  kotlin: "Kotlin",  
  php: "PHP",  
  python: "Python",  
  ruby: "Ruby",  
  rust: "Rust",  
  scala: "Scala",  
  shell: "Shell",  
  swift: "Swift",  
  text: "Text",  
  typescript: "TypeScript",  
  undefined: "Undefined",  
  xml: "XML",  
  yaml: "YAML",  
};  
