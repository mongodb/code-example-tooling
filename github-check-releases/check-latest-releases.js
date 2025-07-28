import { createJiraTicket } from "./create-jira-ticket.js";

async function checkLatestReleases(octokit, details) {
    const latestRelease = await octokit.rest.repos.getLatestRelease({
        owner: details.owner,
        repo: details.repo
    });
    const tagName = latestRelease.data.tag_name;

    if (tagName !== details.testSuiteVersion) {
        console.log(`Warning: for ${details.productName}, test suite version ${details.testSuiteVersion} is behind latest release version ${tagName}. Creating Jira ticket.`);

        const jiraTicketId = await createJiraTicket(details, tagName);
        if (jiraTicketId) {
            console.log(`Created Jira ticket to track updating this test suite: ${jiraTicketId}`);
        }
    } else {
        console.log(`Test suite version for ${details.productName} is up-to-date.`)
    }
}

export { checkLatestReleases };