# Database of Devoured Example Code (DoDEC)

This project contains scaffold and several aggregation pipelines to work with the Database of Devoured Example Code.
The Database of Devoured Example Code contains code examples and related metadata that has been ingested by the [Great
Docs Code Devourer](https://github.com/mongodb/code-example-tooling/tree/main/audit/gdcd).

This DoDEC tooling can currently perform the following tasks:

**Retrieve data**

Get counts broken down in the following ways:

- Total counts:
  - [Total count by docs property](src/aggregations/GetCollectionCount.go)
- Programming language counts:
  - [By docs property](src/aggregations/GetLanguageCounts.go)
  - [By product](src/aggregations/GetProductLanguageCounts.go)
  - [By Atlas sub-product](src/aggregations/GetSubProductLanguageCounts.go)
  - [For a specific language by docs property](src/aggregations/GetSpecificLanguageCounts.go)
  - [Manually get counts from the languages array on device versus using an aggregation](/src/aggregations/GetLangCountFromLangArrayManually.go)
  - [From code nodes instead of the languages array (slower but used to diagnose data issues)](/src/aggregations/GetLangCountFromNodes.go)
- Category counts:
  - [By docs property](src/aggregations/GetCategoryCounts.go)
  - [By product](src/aggregations/GetProductCategoryCounts.go)
  - [By Atlas sub-product](src/aggregations/GetSubProductCategoryCounts.go)
  - [For one specific category by product](src/aggregations/GetSpecificCategoryByProduct.go)
- Complexity counts:
  - [One line usage examples by docs property](src/aggregations/GetOneLineUsageExampleCounts.go)
  - [Minimum, median, maximum character counts for code examples in the collection, and one-liner counts](src/aggregations/GetMinMedianMaxCodeLength.go)

Get IDs for pages based on various criteria:
- [Find docs pages in Atlas that have had code examples added, updated, or removed within the last week](/src/aggregations/GetDocsIdsWithRecentActivity.go)
- [Find docs pages in Atlas where the Product name is missing](/src/aggregations/FindDocsMissingProduct.go)
- [Find docs pages in Atlas that have a count mismatch between the languages array and the languages in code nodes](/src/aggregations/GetPagesWithNodeLangCountMismatch.go)

**Update Documents**
- [Add `product` and `sub_product` fields](src/updates/AddProductNames.go) to their relevant documents across the 37
  docs properties
- [Rename a field](src/updates/RenameField.go) in the document across the 37 docs properties
- [Rename a value](src/updates/RenameValue.go) in the document across the 37 docs properties
- [Copy the current production DB for testing](/src/updates/CopyDBForTesting.go)

**Print to console**

The aggregations listed in the *Retrieve data* section above store data for each collection in maps with varying
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
- [Print Page IDs with changes in the last week](/src/utils/PrintPageIdChangesCountMap.go) where each row represents a
  page that has had changes, and each column lists the number of code examples added, updated, or removed on the page
- [Print page IDs with node language count mismatch](/src/utils/PrintPageIdChangesCountMap.go) with each table representing
  a different docs project, and each row is a Page ID for a document in that project that has a node/languages array
  count mismatch

## Prerequisites

To perform operations with this project, you need:

- `Golang` installed. Refer to the [Go installation page](https://go.dev/doc/install) for details.
- A `MONGODB_URI` key in your environment with a connection string

### Install project dependencies

From the `src` directory, run the following command to install
dependencies:

```shell
go get dodec
```

### Create relevant keys in your environment

You need a connection string for the `Code Snippets` project in the `Developer Docs` Atlas organization. Contact the
Developer Docs team for access.

#### In a `.env` file

In the `src` directory, create a `.env` file.

Add a `MONGODB_URI` key with your valid connection string:

```
MONGODB_URI="YOUR-CONNECTION-STRING-HERE"
```

Add a `DB_NAME` key with the name of the database you want to work with:

```
DB_NAME="code_metrics"
```

#### In your IDE

If you prefer to use your IDE to build and run the project, add the `MONGODB_URI` key and `DB_NAME` key using your
IDE's paradigm for handling environment variables.

## Run the project

With the dependencies installed, and the `MONGODB_URI` and `DB_NAME` available in your environment, you can run the
project from the command line or from your IDE.

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

The `code_metrics` database currently consists of 38 collections that represent 38 docs projects. Our list of curated
docs projects to parse for the audit contains 39 projects, but 1 of them (Rails MongoDB) has no live, published
documentation pages. Rails MongoDB is currently on hold.

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

The summary document has a schema that conforms to the
[CollectionReport](https://github.com/mongodb/code-example-tooling/blob/main/audit/common/CollectionReport.go) type.

#### Docs page document

The remaining documents in the collection each map to an individual docs page. The docs page documents have a schema that
conforms to the [DocsPage](https://github.com/mongodb/code-example-tooling/blob/main/audit/common/DocsPage.go) type.

Each docs page has a `nodes` array, which may be `null`, or may contain `CodeNode` elements. The `CodeNode` elements
contain metadata about the code examples, as well as the examples themselves. To work with only the `CodeNode` elements
that are currently live on the documentation page, omit `CodeNode` instances where `is_removed` is present and `true`.
`CodeNode` elements that have not been removed from the page lack the `is_removed` flag.
