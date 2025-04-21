import { Octokit } from "octokit";

class ReferrerTraffic {
    constructor(referrer, count, uniques) {
        this.referrer = referrer;
        this.count = count;
        this.uniques = uniques;
    }
}

class TopPaths {
    constructor(path, count, uniques) {
        this.path = path;
        this.count = count;
        this.uniques = uniques;
    }
}

class MetricsDoc {
    constructor({owner, repo, clones, pageViews, metricCounts, referralSources, topPaths}) {
        this.date = new Date().toISOString();
        this.owner = owner;
        this.repo = repo;
        this.clones = clones;
        this.viewCount = pageViews.viewCount;
        this.uniqueViews = pageViews.uniqueViews;
        this.stars = metricCounts.stars;
        this.forks = metricCounts.forks;
        this.watchers = metricCounts.watchers;
        this.referralSources = referralSources;
        this.topPaths = topPaths;
    }
}

async function getGitHubMetrics(owner, repo) {
    const apiToken = process.env.GITHUB_TOKEN
    const octokit = new Octokit({
        auth: apiToken,
    });
    await octokit.rest.users.getAuthenticated();

    const metricsData = {
        owner: owner,
        repo: repo,
        clones: await getRepoClones(octokit, owner, repo),
        pageViews: await getPageViews(octokit, owner, repo),
        metricCounts: await getRepoMetricCounts(octokit, owner, repo),
        referralSources: await getReferralSources(octokit, owner, repo),
        topPaths: await getTopPaths(octokit, owner, repo),
    }
    return new MetricsDoc(metricsData);
}

async function getRepoClones(octokit, owner, repo) {
    const clones = await octokit.rest.repos.getClones({
        owner: owner,
        repo: repo
    });
    return clones.data.count;
}

async function getPageViews(octokit, owner, repo) {
    const pageViews = await octokit.rest.repos.getViews({
        owner: owner,
        repo: repo
    });
    return {
        viewCount: pageViews.data.count,
        uniqueViews: pageViews.data.uniques,
    }
}

async function getRepoMetricCounts(octokit, owner, repo) {
    const repoDetails = await octokit.rest.repos.get({
        owner: owner,
        repo: repo
    });
    const stars = repoDetails.data.stargazers_count;
    const forks = repoDetails.data.forks_count;
    const watchers = repoDetails.data.watchers;
    return {
        stars: stars,
        forks: forks,
        watchers: watchers
    }
}

async function getReferralSources(octokit, owner, repo) {
    const repoDetails = await octokit.rest.repos.getTopReferrers({
        owner: owner,
        repo: repo
    });
    let referralSources = [];
    repoDetails.data.map(item => {
        referralSources.push(new ReferrerTraffic(item.referrer, item.count, item.uniques));
    });
    return referralSources;
}

async function getTopPaths(octokit, owner, repo) {
    const repoDetails = await octokit.rest.repos.getTopPaths({
        owner: owner,
        repo: repo
    });
    let paths = [];
    repoDetails.data.map(item => {
        paths.push(new TopPaths(item.path, item.count, item.uniques));
    });
    return paths;
}

// Currently there is no data to display, so not sure what form the return data takes.
// @todo Add details to work with this data once I can get a return value.
async function getMaintenanceInfo(octokit, owner, repo) {
    const codeFrequency = await octokit.rest.repos.getCodeFrequencyStats({
        owner: owner,
        repo: repo
    });
    const commits = await octokit.rest.repos.getCommitActivityStats({
        owner: owner,
        repo: repo
    });
    if (codeFrequency.status === 202 || commits.status === 202) {
        console.log("GitHub returned a 202, which means the requested data is not currently cached. Try again later.")
    } else {
        console.log(codeFrequency);
        console.log(commits);
    }
}

export {
    getGitHubMetrics
}
