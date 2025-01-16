# Project Metrics Tooling

This repo contains tooling to enable us to programmatically keep track of various project metrics.

At the moment, it only contains a PoC for a simple pipeline to get metrics out of GitHub and into MongoDB Atlas. In the
future, we can move this into the `mongodb` org and add tooling for our various projects.

## GitHub Repo Metrics

### Get Metrics from GitHub

This is a simple PoC that uses [octokit](https://github.com/octokit/octokit.js) to get the following data out of GitHub
for a given repository over a trailing 14 day period:

- Views
- Unique Views
- Stars
- Watchers
- Forks

The intent is to also get the following maintenance-related stats for a given repository over a trailing 14 day period:

- Code frequency
- Commit count

However, at present, GitHub does not have any data cached for the test repo, so I'll iterate on this in a future version.

This code is in the `get-github-metrics.js` file.

> **Note**: The GitHub API does not provide the option to specify a date range for these metrics. The API _only_ provides
> this data for the trailing 14 day period, fixed. In the future, we'll want to set up a server to run this job regularly
> since we cannot specify a date range.

### Write metrics to Atlas

This PoC uses the [MongoDB Node.js Driver](https://www.mongodb.com/docs/drivers/node/current/) to write the data to the
**Developer Docs** -> **Project Metrics** project in Atlas.

This code is in the `write-to-db.js` file.

In the future, we can set up Charts to visualize this data and share it with stakeholders.

### Run the tool

#### Prerequisites

To run the tool, you need:

**Atlas**:

- An Atlas Database User with write permissions for the **Developer Docs** -> **Project Metrics** project.
- A valid connection string for the cluster above.

Contact a member of the Developer Docs team to be added to this project and get the connection string.

**GitHub**:

- A [GitHub Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
  with `repo` permissions

For this project, as a MongoDB org member, you must also auth your PAT with SSO.

**System**:

- Node.js/npm installed

#### Steps

##### Create a `.env` file

Create a `.env` file that contains the following details:

```
ATLAS_CONNECTION_STRING="yourConnectionString"
GITHUB_TOKEN="yourToken"
```

Replace the placeholder values with your connection string and GitHub token.

The `.env` file is in the `.gitignore`, so no worries about accidentally committing credentials.

##### Install the dependencies

From the root of the directory, run the following command to install project dependencies:

```
npm install
```

##### Run the utility

From the root of the directory, run the following command to run the utility:

```
node --env-file=.env index.js
```

You should see output similar to:

```
A document was inserted with the _id: 678197a0ffe1539ff213bd86
```