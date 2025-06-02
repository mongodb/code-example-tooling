# Great Docs Code Devourer (Code Ingest Tool) 

The Great Docs Code Devourer (GDCD) processes MongoDB documentation pages to extract code examples. It compares these 
examples with previously stored code to identify new, updated, or removed examples. GDCD stores all code examples and 
metadata in a MongoDB Atlas database maintained by the Developer Docs team.


Contact the Developer Docs team to use this tool or access its data.

## Why We Devour Code

The database of devoured code examples enables powerful analysis of the documentation corpus, including:

- Code example counts by programming language, category, product, sub-product, or keyword
- Density of code examples on the page by product, sub-product, or keyword
- Code example complexity (the intersection of the "Usage example" category and code example length)
- Distribution of short vs. comprehensive examples
- Language coverage across documentation

For querying this data, use the companion project,
[Database of Devoured Example Code (DODEC)](https://github.com/mongodb/code-example-tooling/tree/main/audit/dodec).

## How it works

GDCD follows this pipeline:

1. Retrieves the latest docs project information from the Snooty Data API
2. For each specified project, gets docs pages for current active branch (latest published version)
3. Extracts code examples and metadata from each docs page
4. Syncs changes to MongoDB Atlas

### LLM-Based Code Categorization

We use the Ollama [qwen2.5-coder](https://ollama.com/library/qwen2.5-coder) model to categorize new incoming 
code examples. At the time of this writing, it is the latest series of code-specific Qwen models focused on improved code 
reasoning, code generation, and code fixing. This model has consistently produced the most accurate results when 
categorizing code examples. Refer to the [Ollama](https://ollama.com/) website for more details.

### Metadata Tracked

We track various metadata about the code examples and their associated documentation pages:

For each code example:
- Code example text 
- File extension and programming language
- Category
- Categorization method (LLM or manual)
- Date created, updated, and removed

For each docs page:
- Production URL
- Example counts by language
- Product and sub-product
- Keywords
- Last updated date

## Installing the Tool

### Prerequisites

Before you begin, contact the Developer Docs team for the required connection details and access. 

- [Go](https://go.dev/doc/install)
- [Ollama](https://ollama.com/) installed locally
- The [qwen2.5-coder](https://ollama.com/library/qwen2.5-coder) model
- Environment configuration details from the Developer Docs team

### Setup
1. Install Ollama from [ollama.com](https://ollama.com/), then install the required model:
    ```shell
    ollama pull qwen2.5-coder
    ```
2. Install dependencies. From the project root, run the following:
    ```shell
    go get gdcd
    ```
3. Create the relevant env configuration files in the project root. This project is set up for three environments. You will most likely be running against prod.  
   1. Create a `.env.ENVIRONMENT` file for the `ENVIRONMENT` where you want to run the tool:
      - `production` 
      - `development`
      - `testing`
      
      (for example, create `.env.production` to run against the prod database)
   2. Add the following:
         ```dotenv
         MONGODB_URI="YOUR_MONGODB_URI_HERE"
         DB_NAME="RELEVANT_DB_NAME_HERE"
         ```
      - `MONGODB_URI`: Connection string for the Code Snippets project in the Developer Docs Atlas organization. 
        Contact the Developer Docs team for access.
      - `DB_NAME`: The database to run the tool on. We maintain several databases for production, testing, and backup purposes. 
        Contact the Developer Docs team for the appropriate DB name.

## Running the Tool

Set the `APP_ENV` variable to the environment where you want to run the tool, then run from `main.go`. 
Env values:
- `production`
- `development`
- `testing`

You can do this from the command line or your IDE: 

- **Command Line**

    To run from the terminal, set the variable, then run from the project root. 
    For example, to run against the `production` environment:
    ```shell
    export APP_ENV=production
    go build
    go run .
    ```
- **IDE**:
    
    To run from an IDE configuration: 
    1. Set the `APP_ENV` environment variable (e.g. `APP_ENV=production`) 
    2. Run `main.go`

The progress bar should immediately output to console and continue to display progress until all 
projects are parsed. Depending on your machine and the amount of projects specified, this can be a 
long-running program (~1-2hrs ). 

## Troubleshooting
### Ollama Issues
```text
Error: "failed to generate a response from the given prompt (is Ollama running locally?)"
```
1. Check if Ollama is running locally
2. Verify model availability:

  ```shell
  ollama list
  ```
  If `qwen2.5-coder` isn't listed, install it:

  ```shell
  ollama pull qwen2.5-coder
  ```

### Connection Issues
```text
Error: "failed to connect to MongoDB"
```
1. Verify you've set the correct `APP_ENV` variable and corresponding `.env.ENVIRONMENT` exists in project root
2. Check your connection string in the corresponding `.env.ENVIRONMENT` file
3. Check connectivity to Atlas and that your IP is whitelisted 

### Other Issues

Contact the Developer Docs team for assistance with environment setup or access.

## Disclaimer

Enlist the aid of the Great Docs Code Devourer at your peril! 

This beast is an amalgam of tools with some test coverage, but key bits of business logic still remain uncovered by tests. 
If demand/priority permits, we would love to expand and improve this tooling.
