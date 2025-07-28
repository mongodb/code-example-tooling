import axios from 'axios';

function checkForJiraEnvDetails() {
    const missingVariables = [];

    // List of required environment variables
    const requiredEnvVars = {
        JIRA_BASE_URL: process.env.JIRA_BASE_URL,
        JIRA_API_TOKEN: process.env.JIRA_API_TOKEN,
        JIRA_PROJECT_KEY: process.env.JIRA_PROJECT_KEY,
        JIRA_ISSUE_TYPE: process.env.JIRA_ISSUE_TYPE,
        JIRA_COMPONENT: process.env.JIRA_COMPONENT,
    };

    // Check for missing values and collect variable names
    for (const [key, value] of Object.entries(requiredEnvVars)) {
        if (!value) {
            missingVariables.push(key);
        }
    }

    // Handle case where at least one variable is missing
    if (missingVariables.length > 0) {
        console.error(
            `Missing required environment variable(s): ${missingVariables.join(
                ', '
            )}. Please ensure the following variables are set in your .env file:\n` +
            Object.keys(requiredEnvVars)
                .map((key) => `- ${key}`)
                .join('\n')
        );
        return false; // Validation failed
    }

    return true; // Validation succeeded
}


async function createJiraTicket(details, latestVersion) {
    // Ensure environment variables are configured
    if (!checkForJiraEnvDetails()) {
        // Exit early with meaningful feedback
        console.error('Failed to create Jira ticket due to missing environment variable(s). Please update your .env file.');
        return;
    }

    const jiraBaseUrl = process.env.JIRA_BASE_URL;
    const apiToken = process.env.JIRA_API_TOKEN;
    const projectKey = process.env.JIRA_PROJECT_KEY;
    const issueType = process.env.JIRA_ISSUE_TYPE;
    const componentName = process.env.JIRA_COMPONENT;

    const endpoint = `${jiraBaseUrl}/rest/api/2/issue`;

    const issueData = {
        fields: {
            project: {
                key: projectKey,
            },
            summary: `${details.productName}: Update test suite to latest release version ${latestVersion}`,
            description: `The test suite version for ${details.productName} (${details.testSuiteVersion}) is behind the latest release version ${latestVersion}. Please update the test suite, and bump the version in code-example-tooling/github-check-releases.`,
            issuetype: {
                name: issueType,
            },
            components: [
                {
                    "name": componentName,
                }
            ],
            labels: ["feature", "dd-maintenance"],
        },
    };

    try {
        const response = await axios.post(endpoint, issueData, {
            headers: {
                Authorization: `Bearer ${apiToken}`,
                'Content-Type': 'application/json',
            },
        });

        return response.data.key;
    } catch (error) {
        console.error(
            `Failed to create Jira ticket: ${error.response?.data || error.message}`
        );
    }
}

export { createJiraTicket };
