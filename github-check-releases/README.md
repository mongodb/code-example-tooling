# GitHub Check Releases

This directory contains tooling to check GitHub releases programmatically, and
validate them against the versions of the MongoDB products we use in our code
example test suites.

This is a simple PoC with three main components:

- A [JSON file](repo-details.json) that contains the list of products, GitHub
  repo information, and test suite versions for the specified repos.
- It gets release data for the MongoDB products from GitHub using
  [octokit](https://github.com/octokit/octokit.js), and compares it against the
  version in our JSON config file.
- For any test suites that need a version bump, it automatically creates the
  appropriate Jira tickets using [axios](https://github.com/axios/axios) with
  the Jira API.

## Add or change products and test suite versions

This project pulls the configuration data from
[repo-details.json](repo-details.json) to check the release versions for the
MongoDB products we use in our code example test suites.

### Add a new product

To add a new product, create a new entry in the `repo-details.json` file in
the following format:

```json
{
  "owner": "<repo-owner>",
  "repo": "<repo-name>",
  "productName": "<MongoDB Product Name>",
  "testSuiteVersion": "<x.y.z>"
}
```

You can get the owner and name from the repo URL: `https://github.com/<owner>/<repo>`
For example, if you want to add the Node.js Driver, you'd add the following to
the `repo-details.json`:

```
"owner": "mongodb", // repo owner from URL
"repo": "node-mongodb-native", // repo name from URL
"productName": "Node.js Driver", // unique, human-readable product name to use in report outputs
"testSuiteVersion": "v6.17.0" // taken from the `package.json` dependencies
```

- repo URL: `https://github.com/mongodb/node-mongodb-native`
- product name: the "Node.js Driver"
- the test suite's dependency file's `code-example-tests/javascript/driver/package.json`
  `dependencies` stanza:

  ```json
  "dependencies": {
    "bluehawk": "^1.6.0",
    "mongodb": "6.17.0" // the test suite version for the Driver
  }
  ```

### Change the test suite version

> IMPORTANT: You must manually update the version in this project and in the test suite.

If the version is out of date, bump the version in the test suite, and then
update it in [repo-details.json](repo-details.json) in this project.

## GitHub release information

### Get release information from GitHub

This is a simple PoC that uses [octokit](https://github.com/octokit/octokit.js)
to get release data for MongoDB products from GitHub.

This code is in the [`check-latest-releases.js`](check-latest-releases.js) file.

## Run the tool

### Prerequisites

To run the tool, you need:

**GitHub**:

- A [GitHub Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
  (PAT) with `repo` permissions

You must also authorize your PAT with SSO for this project as a MongoDB org
member.

**Jira**:

- A Jira PAT - refer to the internal wiki for instructions on how to create it
- Information about the Jira base URL, project, and component(s) to create the
  ticket with the appropriate data

**System**:

- Node.js/npm installed

### Steps

1. **Create a `.env` file**

   Create a `.env` file that contains the following details:

   ```
   GITHUB_TOKEN="yourToken"
   JIRA_BASE_URL="https://your-jira-instance-url"
   JIRA_API_TOKEN="your-api-token"
   JIRA_PROJECT_KEY="your-project-key"
   JIRA_ISSUE_TYPE="Task"
   JIRA_COMPONENT="your-team-or-project-component"
   ```

   Replace the placeholder values with your GitHub token, Jira token, and
   Jira project/component details.

   > Note: The `.env` file is in the `.gitignore`, but still use caution to avoid accidentally committing credentials.

2. **Install the dependencies**

   From the root of the directory, run the following command to install project
   dependencies:

   ```
   npm install
   ```

3. **Run the utility**

   From the root of the directory, run the following command to run the utility:

   ```
   node --env-file=.env index.js
   ```

   You should see output similar to:

   ```console
   Warning: for C# Driver, test suite version v3.4.0 is behind latest release version v3.4.2.
   Creating Jira ticket.
   Created Jira ticket to track updating this project: DOCSP-52411
   Warning: for Node.js Driver, test suite version v6.17.0 is behind latest release version v6.18.0.
   Creating Jira ticket.
   Created Jira ticket to track updating this project: DOCSP-52412
   Test suite version for PyMongo Driver is up-to-date.
   ```
