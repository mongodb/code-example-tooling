# Code Example Tooling

This repository contains tooling that the MongoDB Developer Docs team 
uses to help us audit and track code examples across MongoDB's documentation
corpus.

- `github-metrics`: a Node.js script to get metrics from GitHub and write them to a collection in Atlas.

- `examples-copier`: a Go app that runs as a GitHub App and copies files from the source code repo (generated code examples) to multiple target repos and branches.
