import fs from 'fs/promises'; // Use fs/promises for asynchronous file reading
import { Octokit } from "octokit";
import {RepoDetails} from './RepoDetails.js'; // Import the RepoDetails class
import {checkLatestReleases} from './check-latest-releases.js';
import {checkForJiraEnvDetails} from "./create-jira-ticket.js";

// processRepos reads the JSON config file and iterates through the repos specified, converting each to an instance of the RepoDetails class.
async function processRepos() {
    const apiToken = process.env.GITHUB_TOKEN
    const octokit = new Octokit({
        auth: apiToken,
    });

    if (apiToken === "") {
        console.error('No API token provided - make sure you have created a .env file with your API token.')
    }

    try {
        await octokit.rest.users.getAuthenticated();
    } catch (error) {
        console.error('Error authenticating with GitHub:', error)
    }

    const hasJiraEnvDetails = checkForJiraEnvDetails();
    if (!hasJiraEnvDetails) {
        console.error('Cannot create Jira tickets due to missing environment variable(s). Please update your .env file.');
    }

    try {
        // Read the JSON file
        const data = await fs.readFile('repo-details.json', 'utf8');

        // Parse the JSON data into an array
        const reposArray = JSON.parse(data);

        // Convert each repo object into an instance of RepoDetails
        const repos = reposArray.map(
            (repo) =>
                new RepoDetails(repo.owner, repo.repo, repo.productName, repo.testSuiteVersion)
        );

        // Iterate through the repos array
        for (const repo of repos) {
            await checkLatestReleases(octokit, repo, hasJiraEnvDetails);
        }
    } catch (error) {
        console.error('Error processing repos:', error);
    }
}

// Call the function
processRepos()
