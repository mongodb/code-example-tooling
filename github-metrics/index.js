import { getGitHubMetrics } from "./get-github-metrics.js";
import { addMetricsToAtlas } from "./write-to-db.js";

/* To change the repo we're tracking metrics for, change the owner and/or repo name here.
The URL pattern is: `https://github.com/<owner>/<repo>`
For example, the URL for the `mongodb`-owned `docs-notebooks` repo is: `https://github.com/mongodb/docs-notebooks`
NOTE: The GitHub token used to retrieve the info from a repo MUST have repo admin permissions to access all the endpoints in this code. */

const owner = "mongodb"; // the GitHub organization or member who owns the repo
const repo = "docs-notebooks"; // the name of the repo within the organization or member
const metricsDoc = await getGitHubMetrics(owner, repo);
await addMetricsToAtlas(metricsDoc);
