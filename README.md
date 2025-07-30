# Code Example Tooling

This repository contains tooling that the MongoDB Developer Docs team
uses audit and track code examples across MongoDB's documentation
corpus.

- `audit`: Two Go projects, plus Go type definitions and constants that are
  common to both of them:
- `gdcd`, or the Great Docs Code Devourer:  an ingestion tool that gets and categorizes code examples from the current 
  MongoDB documentation corpus, with a selection of metadata, and writes the info to a 
  database in Atlas.
- `dodec`, or the Database of Devoured Example Code: a query tool that lets us find code examples and related 
  metadata in the database for reporting or to perform manual updates.
- `examples-copier`: a Go app that runs as a GitHub App and copies files from the
   source code repo (generated code examples) to multiple target repos and branches.
- `github-check-releases`: a Node.js script that gets the latest release versions
  of products we use in our test suites, and creates Jira tickets for outdated versions.
- `github-metrics`: a Node.js script that gets engagement metrics from GitHub for specified repos and writes them
  to a database in Atlas.
- `query-docs-feedback`: a Go project with type definitions that queries the MongoDB
Docs Feedback received for any feedback related to code examples, and outputs the result
to a report as `.csv`.
