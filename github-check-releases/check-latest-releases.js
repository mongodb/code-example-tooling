async function checkLatestReleases(octokit, details) {
    const latestRelease = await octokit.rest.repos.getLatestRelease({
        owner: details.owner,
        repo: details.repo
    });
    const tagName = latestRelease.data.tag_name;

    if (tagName !== details.testSuiteVersion) {
        console.log(`Warning: for ${details.productName}, test suite version ${details.testSuiteVersion} is behind latest release version ${tagName}`);
        // TODO: Add code here to call a Jira endpoint to create a ticket to update the test suite. Store the Jira endpoint as a .env var to avoid exposing it in the public repo.
    } else {
        console.log(`Test suite version for ${details.productName} is up-to-date.`)
    }
}

export { checkLatestReleases };