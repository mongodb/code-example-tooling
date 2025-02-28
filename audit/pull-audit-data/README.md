# Work with Code Example Audit Database

## Overview

This project contains scaffold and several aggregation pipelines to work with the code example audit database. This
project currently supports:

**Retrieving data**

Getting counts broken down in the following ways:

- Total counts:
  - [Total count by docs property](src/aggregations/GetCollectionCount.go)
- Programming language counts:
  - [By docs property](src/aggregations/GetLanguageCounts.go)
  - [By product](src/aggregations/GetProductLanguageCounts.go)
  - [By Atlas sub-product](src/aggregations/GetSubProductLanguageCounts.go)
- Category counts:
  - [By docs property](src/aggregations/GetCategoryCounts.go)
  - [By product](src/aggregations/GetProductCategoryCounts.go)
  - [By Atlas sub-product](src/aggregations/GetSubProductCategoryCounts.go)
  - [For one specific category by product](src/aggregations/GetSpecificCategoryByProduct.go)
- Complexity counts:
  - [One line usage examples by docs property](src/aggregations/GetOneLineUsageExampleCounts.go)
  - [Minimum, median, maximum character counts for code examples in the collection, and one-liner counts](src/aggregations/GetMinMedianMaxCodeLength.go)

**Updating Documents**
- [Adding `product` and `sub_product` fields](src/updates/AddProductNames.go) to their relevant documents across the 37
  docs properties
- [Renaming a field](src/updates/RenameField.go) in the document across the 37 docs properties

**Printing to console**

The aggregations listed in the *Retrieving data* section above store data for each collection in maps with varying
structures. The `utils` directory contains convenience functions to print the map output to console as one or more
tables.

- [Print one table](src/utils/PrintSimpleCountDataToConsole.go) with rows representing each collection, product, category,
  or programming language. Use where the aggregation returns a `simpleMap` as defined in [PerformAggregation](src/PerformAggregation.go)
- [Print multiple tables](src/utils/PrintNestedOneLevelCountDataToConsole.go) with each row representing a category or 
  programming language, and each table representing a higher-level division such as product or docs property. Use where
  the aggregation returns a `nestedOneLevelMap` as defined in [PerformAggregation](src/PerformAggregation.go)
- [Print multiple tables from two-level nested maps](src/utils/PrintNestedTwoLevelCountDataToConsole.go) with each row
  representing a category or programming language, and each table representing a higher-level division such as product
  or docs property. Use where the aggregation returns a `nestedTwoLevelMap` as defined in [PerformAggregation](src/PerformAggregation.go)
- [Print a table with code example length details](src/utils/PrintLengthCountMapToConsole.go) where each row represents
  a docs property, and columns provide details about the minimum, median, and maximum character counts for code examples
  in the collection, as well as a count of "short" examples (under 80 characters) in the collection. Use with the
  [GetMinMedianMaxCodeLength](src/aggregations/GetMinMedianMaxCodeLength.go) aggregation.

## Prerequisites

To perform operations with this project, you need:

- The required dependencies
- A `MONGODB_URI` key in your environment with a connection string

### Install the dependencies

### Golang

This project requires you to have `Golang` installed. If you do not yet
have Go installed, refer to [the Go installation page](https://go.dev/doc/install)
for details.

### Go Dependencies

From the `src` directory, run the following command to install
dependencies:

```shell
go get pull-audit-data
```

### Create a `MONGODB_URI` key in your environment

You need a connection string for the `Code Snippets` project in the `Developer Docs` Atlas organization. Contact the
Developer Docs team for access.

#### In a `.env` file

In the `src` directory, create a `.env` file.

Add a `MONGODB_URI` key with your valid connection string:

```
MONGODB_URI="YOUR-CONNECTION-STRING-HERE"
```

#### In your IDE

If you prefer to use your IDE to build and run the project, add the `MONGODB_URI` key with your valid connection string
using your IDE's paradigm for handling environment variables.

## Run the project

With the dependencies installed, and the `MONGODB_URI` available in your environment, you can run the project from the
command line or from your IDE.

### Command-line

To run the project from the command line, run the following commands:

```
go build
go run .
```

### IDE

To run the project from an IDE, press the `play` button next to the `main()`
func in `main.go`. Alternately, press the `Build` button in the top right of
your IDE.

## Database and Collection data structure

The `code_metrics` database currently consists of 37 collections that represent 37 docs projects. Our list of curated
docs projects to parse for the audit contains 39 projects, but 2 of them (Atlas Architecture Center and Rails MongoDB)
had no live, published documentation pages at the time of the audit.

Atlas Architecture Center has gone live since the audit was completed, so a new collection will be added for this project
the next time we run the audit tooling. Rails MongoDB is currently on hold.

**Work with all collections in the database**

This project assumes you want to iterate through all the collections in the database, and get various counts for each
collection. However, the aggregations are structured to be callable for a single collection, and returning values as a
map, which you can update with values where each collection becomes a key in the map.

**Work with a single collection in the database**

You can work directly with a single collection by calling the aggregation function you want with that specific
collection name, instead of getting the list of collections and iterating through it to perform aggregations across all
collections.

### Documents

Every collection contains documents that map to one of two schemas:

- A Summary document with an array containing details about the docs versions and audit dates. To omit this document from
  aggregation pipelines (or view it directly), look for `"_id": "summaries"`
- Documents that map to pages in the documentation. Each document represents a specific documentation page, with
  page-level metadata as well as a collection of `nodes` that represent the code examples on that page.

#### Summary document

The summary document has a schema that conforms to the [Summary](src/types/Summary.go) type.

#### Docs page document

The remaining documents in the collection each map to an individual docs page. The docs page documents have a schema that
conforms to the [DocsPage](src/types/DocsPage.go) type.

Each docs page has a `nodes` array, which may be `null`, or may contain `CodeNode` elements. The `CodeNode` elements
contain metadata about the code examples, as well as the examples themselves.
