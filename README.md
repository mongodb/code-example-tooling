# Code Example Tooling

This repository contains tooling that the MongoDB Developer Docs team
uses to help us audit and track code examples across MongoDB's documentation
corpus.

- `audit`: Two Go projects, plus Go type definitions and constants that are
  common to both of them:
  - `gdcd`: The Great Docs Code Devourer gets code examples from MongoDB
    documentation, along with a selection of metadata, adds a category, and
    writes the info to a database in Atlas.
  - `dodec`: The Database of Devoured Example Code: query and perform a few
    manual updates on the database of code examples and related metadata.
- `examples-copier`: a Go app that runs as a GitHub App and copies files from the
   source code repo (generated code examples) to multiple target repos and branches.
- `github-check-releasese`: a Node.js script to get the latest release versions
  of products we use in our test suites, and let us know what's outdated.
- `github-metrics`: a Node.js script to get metrics from GitHub and write them
  to a database in Atlas.
- `query-docs-feedback`: A Go project and type definitions to query the MongoDB
  docs feedback for feedback related to code examples, and output the result
  to a report as .csv.
