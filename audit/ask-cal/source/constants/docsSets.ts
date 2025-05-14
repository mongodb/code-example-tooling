export type DocsSet =  
  | "atlas-architecture"
  | "atlas-cli"
  | "atlas-operator"
  | "bi-connector"
  | "c"
  | "charts"
  | "cloud-docs"
  | "cloud-manager"
  | "cloudgov"
  | "cluster-sync"
  | "compass"
  | "cpp-driver"
  | "csharp"
  | "database-tools"
  | "django"
  | "docs"
  | "docs-k8s-operator"
  | "docs-relational-migrator"
  | "entity-framework"
  | "golang"
  | "java"
  | "java-rs"
  | "kafka-connector"
  | "kotlin"
  | "kotlin-sync"
  | "laravel"
  | "mongocli"
  | "mongodb-shell"
  | "mongoid"
  | "node"
  | "ops-manager"
  | "php-library"
  | "pymongo"
  | "pymongo-arrow"
  | "ruby-driver"
  | "rust"
  | "scala"
  | "spark-connector";
  
export const DocsSetDisplayValues: Record<DocsSet, string> = {
  "atlas-architecture": "Atlas Architecture Center",
  "atlas-cli": "Atlas CLI",
  "atlas-operator": "Atlas Kubernetes Operator",
  "bi-connector": "BI Connector",
  "c": "C Driver",
  "charts": "Atlas Charts",
  "cloud-docs": "Atlas",
  "cloud-manager": "Cloud Manager",
  "cloudgov": "MongoDB Atlas for Government",
  "cluster-sync": "MongoDB Cluster-to-Cluster Sync",
  "compass": "Compass",
  "cpp-driver": "C++ Driver",
  "csharp": "C#/.NET Driver",
  "database-tools": "Database Tools",
  "django": "Django MongoDB Backend",
  "docs": "Database Manual",
  "docs-k8s-operator": "Enterprise Kubernetes Operator",
  "docs-relational-migrator": "Relational Migrator",
  "entity-framework": "EF Core Provider",
  "golang": "Go Driver",
  "java": "Java Sync Driver",
  "java-rs": "Java Reactive Streams Driver",
  "kafka-connector": "Kafka Connector",
  "kotlin": "Kotlin Coroutine",
  "kotlin-sync": "Kotlin Sync Driver",
  "laravel": "Laravel MongoDB",
  "mongocli": "MongoCLI",
  "mongodb-shell": "mongosh",
  "mongoid": "Mongoid",
  "node": "Node.js Driver",
  "ops-manager": "Ops Manager",
  "php-library": "PHP Library Manual",
  "pymongo": "PyMongo Driver",
  "pymongo-arrow": "PyMongoArrow",
  "ruby-driver": "Ruby Driver",
  "rust": "Rust Driver",
  "scala": "Scala Driver",
  "spark-connector": "Spark Connector",
};
