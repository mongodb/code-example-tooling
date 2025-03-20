# Great Docs Code Devourer

The Great Docs Code Devourer (GDCD) pours through the finest hand-curated selection of MongoDB documentation pages to find
every last morsel of our code examples. It then compares these code examples with code that it has previously devoured
to focus on only the new code, updates to existing code, and figuring out which code has been removed from this delectable
docs corpus. The Great Docs Code Devourer relies on a MongoDB Atlas database to savor each and every detail about these
code examples - recording not just the code itself, but delightful metadata to make it easier to recall the code in
novel and useful ways.

To let it feast upon your code, or ask it to recall the finest details of its prior meals, reach out to the Developer
Docs team.

## Why devour code at all?

The Great Docs Code Devourer stores code examples and related metadata to a MongoDB Atlas database. This intersection of
code and metadata enables us get interesting information about the code examples across our documentation corpus, such as:

- Number of code examples for a given programming language, category, product, sub-product, or keyword
- Density of code examples on the page by product, sub-product, or keyword
- Code example complexity (the intersection of the "Usage example" category and code example length)
- Count of one-line code examples
- List of docs projects or pages that have code examples in a given language

The Great Docs Code Devourer does not do any of this querying itself - it only devours code. If you want to get information
from the musings stored by the Great Docs Code Devourer, you want the much more mundane
[pull-audit-data](https://github.com/mongodb/code-example-tooling/tree/main/audit/pull-audit-data) tooling.

## How it works

The Great Docs Code Devourer uses the pipeline described below to get code examples from our documentation. We then store
metadata about the examples, as well as the examples themselves, to a MongoDB Atlas database maintained by the Developer
Docs team.

### Pipeline

- Get the latest information about docs projects from the Snooty Data API
- Get documentation pages for the current active branch in a project from the Snooty Data API for a subset of projects
- Find code examples and related metadata for each docs page in the project
- Sync changes to code examples on a given documentation page with a MongoDB Atlas database

### Metadata

We track various bits of metadata about the code examples, as well as the examples themselves, including:

- The code example text
- File extension
- Programming language
- Category
- Whether the code example category was assigned by an LLM or manually
- Date created, updated, and removed

Every code example is associated with a documentation page that has its own metadata, including:

- Production page URL
- Number of code examples on the page, and their languages
- Product and sub-product name
- Keywords on the documentation page
- Date last updated

## Run the tool

Enlist the aid of the Great Docs Code Devourer at your peril. This beast is an amalgam of tools with some test coverage,
but key bits of business logic remain uncovered by tests. If demand/priority permits, we would love to expand and improve
this tooling.

### Prerequisites

To perform operations with this project, you need:

- Golang installed. Refer to the [Go installation page](https://go.dev/doc/install) for details.
- Ollama installed. Refer to the [Ollama](https://ollama.com/) website for
  installation details.
- The Ollama [qwen2.5-coder](https://ollama.com/library/qwen2.5-coder) model installed.
- - A `.env` file for the appropriate environment with details outlined below.

#### Ollama

This project uses Ollama running locally on the device to categorize new incoming code examples. Refer to the
[Ollama](https://ollama.com/) website for installation details.

##### Model

This project uses the Ollama [qwen2.5-coder](https://ollama.com/library/qwen2.5-coder) model. At the time of writing
this README, this is the latest series of code-specific Qwen models, with a focus on improved code reasoning, code
generation, and code fixing. This model has consistently produced the most accurate results when categorizing code
examples.

To install the model locally, after you have installed Ollama, run the following command:

```shell
ollama pull qwen2.5-coder
```

### Install project dependencies

From the project root, run the following command to install dependencies:

```shell
go get gdcd
```

### Create the relevant env file(s)

This project is set up for three environments:

- production
- development
- testing

Create a `.env.ENVIRONMENT` file for each environment where you want to run the tool. For example, to run the
tool against the production database, create a `.env.production` file.

Your .env file must contain the following keys:

```
MONGODB_URI="YOUR_MONGODB_URI_HERE"
DB_NAME="RELEVANT_DB_NAME_HERE"
```

When you start the app, set an `APP_ENV` key to specify the environment you're running against. The tool loads the
relevant `.env` file and uses the appropriate values.

#### MongoDB URI

You need a connection string for the Code Snippets project in the Developer Docs Atlas organization. Contact the
Developer Docs team for access.

#### Database name

The Developer Docs team maintains various databases for production, testing, and backup. Get the appropriate DB name
from Developer Docs.

### Run the project

With the dependencies installed, and the MONGODB_URI available in your environment, you can run the project from the
command line or from your IDE.

#### Command-line

To run the project from the command line, run the following commands:

```shell
export APP_ENV=production
go build
go run .
```

Substitute `production` with the appropraite environment.

#### IDE

To run the project from an IDE:

- Set the `APP_ENV` environment variable in your IDE to the appropriate environment variable
- Press the play button next to the main() func in main.go - OR -
- Press the Build button in the top right of your IDE
