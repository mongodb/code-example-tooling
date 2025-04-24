import { getGitHubMetrics } from "./get-github-metrics.js";
import { addMetricsToAtlas } from "./write-to-db.js";

/* To change which repos to track metrics for, update the `repos` array before running the utility. 
To track metrics for a new repo, set the owner and name first. 
You can get the owner and name from the repo URL: `https://github.com/<owner>/<repo>`
For example, to add `https://github.com/mongodb/docs-notebooks`, set `mongodb` as the 
owner and `docs-notebooks` as the repo name.
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
