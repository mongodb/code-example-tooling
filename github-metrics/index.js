import { getGitHubMetrics } from "./get-github-metrics.js";
import { addMetricsToAtlas } from "./write-to-db.js";

/* To change the repos we're tracking metrics for, change the owner and/or repo name here.
The URL pattern is: `https://github.com/<owner>/<repo>`
For example, the URL for the `mongodb`-owned `docs-notebooks` repo is: `https://github.com/mongodb/docs-notebooks`
NOTE: The GitHub token used to retrieve the info from a repo MUST have repo admin permissions to access all the endpoints in this code. */

class RepoDetails {
    constructor(owner, repo) {
        this.owner = owner; // the GitHub organization or member who owns the repo
        this.repo = repo; // the name of the repo within the organization or member
    }
}

const docsNotebooksRepo = new RepoDetails("mongodb", "docs-notebooks");
const atlasArchitectureGoSdkRepo = new RepoDetails("mongodb", "atlas-architecture-go-sdk");

const repos = [docsNotebooksRepo, atlasArchitectureGoSdkRepo];

const metricsDocs = [];

for (const repo of repos) {
    const metricsDoc = await getGitHubMetrics(repo.owner, repo.repo);
    metricsDocs.push(metricsDoc);
}

await addMetricsToAtlas(metricsDocs);
