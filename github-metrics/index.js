import { readFile } from 'fs/promises';
import { getGitHubMetrics } from "./get-github-metrics.js";
import { addMetricsToAtlas } from "./write-to-db.js";
import { RepoDetails } from './RepoDetails.js'; // Import the RepoDetails class
import { shouldRunMetricsCollection, recordSuccessfulRun } from './last-run-tracker.js';

/* To change which repos to track metrics for, update the `repo-details.json` file.
To track metrics for a new repo, add a new entry with the owner and repo name.
You can get the owner and name from the repo URL: `https://github.com/<owner>/<repo>`
For example, to add `https://github.com/mongodb/docs-notebooks`, add:
{
  "owner": "mongodb",
  "repo": "docs-notebooks"
}
NOTE: The GitHub token used to retrieve the info from a repo MUST have repo admin permissions to access all the endpoints in this code. */

// processRepos reads the JSON config file and iterates through the repos specified, converting each to an instance of the RepoDetails class.
async function processRepos() {
    try {
        // Check if we should run based on last run time
        const { shouldRun, lastRun, daysSinceLastRun } = await shouldRunMetricsCollection();
        
        if (!shouldRun) {
            console.log('Skipping metrics collection - not enough time has passed since last run.');
            console.log(`Last run was ${daysSinceLastRun} days ago on ${lastRun?.toISOString()}`);
            process.exit(0); // Exit successfully without running
        }

        console.log('Starting metrics collection...');
        
        // Read the JSON file
        const data = await readFile('repo-details.json', 'utf8');

        // Parse the JSON data into an array
        const reposArray = JSON.parse(data);

        // Convert each repo object into an instance of RepoDetails
        const repos = reposArray.map(
            (repo) => new RepoDetails(repo.owner, repo.repo)
        );

        const metricsDocs = [];

        // Iterate through the repos array
        for (const repo of repos) {
            const metricsDoc = await getGitHubMetrics(repo.owner, repo.repo);
            metricsDocs.push(metricsDoc);
        }

        await addMetricsToAtlas(metricsDocs);
        
        // Record successful run
        await recordSuccessfulRun();
        console.log('✓ Metrics collection completed successfully');
    } catch (error) {
        console.error('Error processing repos:', error);
        throw error; // Re-throw to ensure the job fails
    }
}

// Call the function
processRepos().catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
});

