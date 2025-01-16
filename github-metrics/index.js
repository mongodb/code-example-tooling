import { getGitHubMetrics } from "./get-github-metrics.js";
import { addMetricsToAtlas } from "./write-to-db.js";

// To change the repo we're tracking, change the owner and/or repo name here.
// Owner corresponds to the GitHub organization or member who owns the repo
// Repo corresponds to the name of the repo within the organization or member
// The URL for this repo is: https://github.com/mongodb/docs-notebooks
// The URL pattern to figure out owner/repo is: https://github.com/owner/repo
// The GitHub token used to retrieve the info must have repo admin permissions to access all the endpoints in this code
const owner = "mongodb";
const repo = "docs-notebooks";
const metricsDoc = await getGitHubMetrics(owner, repo);
await addMetricsToAtlas(metricsDoc);
