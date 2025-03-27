# Query Docs Feedback

This project contains an aggregation pipeline and Go type definitions to work with MongoDB documentation feedback.
Specifically, the included aggregation pipeline finds feedback where readers have left comments that contain a substring
we think may be related to code examples. We omit substrings that may be related to broken links.

This project sorts the feedback by collection, prints the collection counts to the console, and
creates a `report.csv` file on the file system in this directory that contains four columns:

- `EntryNumber`: integer; an arbitrarily-assigned incrementing integer to make it easier to differentiate entries and
  track your position when working with the csv
- `DocsProperty`: string; the name of the docs project that the feedback pertains to - i.e. `cloud-docs` or `kafka-connector`
- `URL`: string; the URL of the page that the feedback pertains to
- `Comment`: string; the text of the feedback that the user left about the given documentation page

## Prerequisites

To perform operations with this project, you need:

- `Golang` installed. Refer to the [Go installation page](https://go.dev/doc/install) for details.
- A `MONGODB_URI` key in your environment with a connection string
- The `DB_NAME` and `COLLECTION_NAME` in your environment with relevant names.

### Install project dependencies

From this directory, run the following command to install dependencies:

```shell
go get query-docs-feedback
```

### Create relevant keys in your environment

You need a connection string for the relevant cluster in Atlas. Refer to the article in the internal documentation
about how to access docs feedback data to get the relevant information. The project maintains multiple databases
and collections for different environments; get the DB and collection names that are relevant to the data you want to
query.

#### In a `.env` file

In this directory, create a `.env` file.

Add the following keys with your relevant details:

```
MONGODB_URI="YOUR-CONNECTION-STRING-HERE"
DB_NAME="YOUR_DB_NAME_HERE"
COLLECTION_NAME="YOUR-COLLECTION-NAME-HERE"
```

#### In your IDE

If you prefer to use your IDE to build and run the project, add the `MONGODB_URI` key, the `DB_NAME` key, and the
`COLLECTION_NAME` key using your IDE's paradigm for handling environment variables.

## Run the project

With the dependencies installed, and the required keys available in your environment, you can run the
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
