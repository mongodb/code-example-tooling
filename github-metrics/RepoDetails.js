// This class is used to map the deserialized owner and repo name for a given repository to the corresponding `repo-details.json` config data.

export class RepoDetails {
    constructor(owner, repo) {
        this.owner = owner; // the GitHub organization or member who owns the repo
        this.repo = repo; // the name of the repo within the organization or member
    }
}
