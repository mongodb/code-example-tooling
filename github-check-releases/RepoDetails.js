export class RepoDetails {
    constructor(owner, repo, productName, testSuiteVersion) {
        this.owner = owner; // the GitHub organization or member who owns the repo
        this.repo = repo; // the name of the repo within the organization or member
        this.productName = productName;
        this.testSuiteVersion = testSuiteVersion; // the version of the project currently used in the test suite
    }
}
