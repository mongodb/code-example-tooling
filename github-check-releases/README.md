# GitHub Check Releases

This directory contains tooling to enable us to check GitHub releases programmatically, and validate them against the
version(s) of the product(s) we use in our code example test suite(s).

This is a simple PoC with two main components:

- A [JSON file](repo-details.json) that contains the list of products, GitHub repo information, and test suite versions.
- It uses [octokit](https://github.com/octokit/octokit.js) to get release data for those MongoDB products from GitHub, and compare it against the version 
  in our JSON file.

Future work may include automatically creating a Jira ticket to update the test suite when the version is out of date.

## Add or change products and test suite versions

This project pulls data from [repo-details.json](repo-details.json) to check the release versions for the MongoDB
products we use in our code example test suites. It deserializes the data using `RepoDetails` declared in the
[RepoDetails file](RepoDetails.js) to map the owner and repo name for a given product whose version we want to check.

### Add a new product

To add a new product, create a new entry for it in the `repo-details.json` file. 

You can get the owner and name from the repo URL: `https://github.com/<owner>/<repo>`
For example, to add `https://github.com/mongodb/node-mongodb-native`, set `mongodb` as the
owner and `node-mongodb-native` as the repo name.

You can get the test suite version from the test suite's dependency file. For example, the Node.js Driver version is in
`code-example-tests/javascript/driver/package.json`, in the `dependencies` stanza:

```
"dependencies": {
  "bluehawk": "^1.6.0",
  "mongodb": "6.17.0" // This is the test suite version for the Driver
}
```

The product name is just a human-readable name we can use in the output to report on the changes - i.e. `"Node.js Driver"`.

### Change the test suite version

If the version is out of date, bump the version in the test suite, and then update it in [repo-details.json](repo-details.json) here
in this project.

## GitHub release information

### Get release information from GitHub

This is a simple PoC that uses [octokit](https://github.com/octokit/octokit.js) to get release data for MongoDB products from GitHub.

This code is in the `check-latest-releases.js` file.

## Run the tool

### Prerequisites

To run the tool, you need:

**GitHub**:

- A [GitHub Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens) (PAT) with `repo` permissions

For this project, as a MongoDB org member, you must also auth your PAT with SSO.

**System**:

- Node.js/npm installed

### Steps

1. **Create a `.env` file**

   Create a `.env` file that contains the following details:

   ```
   GITHUB_TOKEN="yourToken"
   ```

   Replace the placeholder value with your GitHub token.

   > Note: The `.env` file is in the `.gitignore`, so no worries about accidentally committing credentials.

2. **Install the dependencies**

   From the root of the directory, run the following command to install project dependencies:

   ```
   npm install
   ```

3. **Run the utility**

   From the root of the directory, run the following command to run the utility:

   ```
   node --env-file=.env index.js
   ```

   You should see output similar to:

   ```
   Warning: for C# Driver, test suite version v3.4.0 is behind latest release version v3.4.2
   Warning: for Node.js Driver, test suite version v6.17.0 is behind latest release version v6.18.0
   Test suite version for PyMongo Driver is up-to-date.
   ```
