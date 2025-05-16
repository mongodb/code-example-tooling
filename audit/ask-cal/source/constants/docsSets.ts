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
  | "docs-django"
  | "docs"
  | "docs-k8s-operator"
  | "relational-migrator"
  | "docs-entity-framework"
  | "docs-golang"
  | "docs-java"
  | "docs-java-rs"
  | "kafka-connector"
  | "kotlin"
  | "kotlin-sync"
  | "laravel"
  | "mck"
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
  "cloud-docs": "Atlas",
  "atlas-architecture": "Atlas Architecture Center",
  "charts": "Atlas Charts",
  "atlas-cli": "Atlas CLI",
  "atlas-operator": "Atlas Kubernetes Operator",
  "bi-connector": "BI Connector",
  "c": "C Driver",
  "cpp-driver": "C++ Driver",
  "csharp": "C#/.NET Driver",
  "cloud-manager": "Cloud Manager",
  "compass": "Compass",
  "docs": "Database Manual",
  "database-tools": "Database Tools",
  "docs-django": "Django MongoDB Backend",
  "docs-k8s-operator": "Enterprise Kubernetes Operator",
  "docs-entity-framework": "EF Core Provider",
  "docs-golang": "Go Driver",
  "docs-java": "Java Sync Driver",
  "docs-java-rs": "Java Reactive Streams Driver",
  "kafka-connector": "Kafka Connector",
  "kotlin": "Kotlin Coroutine",
  "kotlin-sync": "Kotlin Sync Driver",
  "laravel": "Laravel MongoDB",
  "mongocli": "MongoCLI",
  "cloudgov": "MongoDB Atlas for Government",
  "cluster-sync": "MongoDB Cluster-to-Cluster Sync",
  "mck": "MongoDB Controllers for Kubernetes Operators",
  "mongodb-shell": "mongosh",
  "mongoid": "Mongoid",
  "node": "Node.js Driver",
  "ops-manager": "Ops Manager",
  "php-library": "PHP Library Manual",
  "pymongo": "PyMongo Driver",
  "pymongo-arrow": "PyMongoArrow",
  "relational-migrator": "Relational Migrator",
  "ruby-driver": "Ruby Driver",
  "rust": "Rust Driver",
  "scala": "Scala Driver",
  "spark-connector": "Spark Connector",
};
