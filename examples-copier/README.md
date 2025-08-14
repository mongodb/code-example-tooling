# GitHub Docs Code Example Copier

A GitHub app that copies generated code snippets and examples from a source repository
(currently https://GitHub.com/mongodb/docs-code-examples)
to multiple target repositories. Driven by a single `config.json` file in the source repo,
it can:

1. Read files from a specific path the source repo, with optional recursion (default: recursive copy)
2. Copy source files to specific paths in a destination repo on a target branch (default: "main")
3. Either:
   - Commit the changes directly
   - Commit the changes via PR and merge automatically
   - Commit the changes via PR without merging (default)

If you _remove_ a code snippet from the source and its path is in the config file, the app will add it 
to the `deprecated_examples.json` file. 
It will *not* delete any files from target repos, so links in our docs will not be 
broken.

## Project Structure

| Directory Name | Use Case                                                                                                                                                                                  |
|----------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `configs`      | The configuration files (.env.*) are here, as is the `environment.go` file, which creates globals for those settings. No other part of the code should read the .env files.               |
| `services`     | The handlers for the services we're interacting with: GitHub and Webhook handlers. For better organization, GitHub handlers have been separated into Auth, Upload, and Download services. |
| `types`        | All the `type StructName struct` should be here. These are the structs needs to map webhook json to objects we can work with.                                                             |

## Logic Flow

At its core, this is a simple Go web server that listens for messages and handles them
as they come in. Handling the messages means:
- Determining if it's a message we care about
- Pulling out the changed file list from the message
- Reading the config file, and if a file is in a config setting,
- Copy/replace the file at the target repo.

Basic flow: 

1. Configure GitHub permissions (`services/GitHub_auth.go`)
2. Listen for PR payloads from the GitHub webhook. (`services/web_server.go`)
3. Is the PR closed and merged? If no, ignore. (`services/webhook_handler.go`)
4. Parse the payload to get the list of changed files. (`services/GitHub_download.go`)
5. Read the config file from the source repo.
6. If the path to a changed file is defined in the config file, and it is not a
   "DELETE" action, copy the file to the specified target repos. (`services/GitHub_upload.go`)
7. If the path to a changed file is defined in the config file, and it *is* a "DELETE"
   action, add the deleted file's name and path to the `deprecated_examples.json` file.
   (`services/GitHub_download.go`)
8. Sit idle until the next payload arrives. Rinse and repeat.

## Install the App on a Target Repo
To install the app on a new target repository:
1. [Give the App repo access](#Give the app repo access)
2. [Install the App in the new source repo](#Install the App on a new Source Repo)

### Give the app repo access
1. Go to the [App's Configuration page](https://github.com/apps/docs-examples-copier/installations/62138132).
You'll need to authorize your GitHub account first. You should then see the following screen:
!["gui request tag"](./readme_files/configure_app.png)
2. In the `Select repositories` dropdown, select the new target repo. Then click 
**Update access**. 
> **NOTE:** if you are not an owner of the target repository, one of the owners will need 
to complete the next steps. You will know this if the repository you selected has a 
`request` tag next to it:      
> !["gui request tag"](./readme_files/request.png)

### Confirm the new target repository
1. In the new target repository's settings, go to the 
[GitHub Apps section](https://GitHub.com/mongodb/stitch-tutorial-todo-backend/settings/installations).
Scroll down and confirm that `Docs Examples Copier` is installed.

## Install the App on a new Source Repo
In the source repo, do the following:
1. [Set up a webhook](#Set Up A Webhook)
2. [Update the .env file](#Update the env file)
3. [Add config.json and deprecated_examples.json files](#Add config.json and deprecated_examples.json files)
4. [Configure Permissions for the Web App](#Configure Permissions for the Web App)

### Set Up A Webhook
Go to the source repo's 
[webhooks settings page](https://GitHub.com/mongodb/docs-code-examples/settings/hooks/).
- Add the new `Payload URL`.
- Set the `Content Type` to `application/json`
- Enable SSL Verification
- Choose `Let me select individual events`
  - Choose *only* `Pull Requests`. Do **not** choose any other "Pull Request"-related
    options! 
- At the bottom of the page, make sure `Active` is checked, and then save your changes.

At this point, with PR-related activity, the payload will be sent to the app. 
The app ignores all PR activity except for when a PR is closed and merged.

### Update the env file
The .env file specifies settings for the source repo.

Update the .env values for your repo:

```dotenv
GITHUB_APP_CLIENT_ID="Client ID of the github app you created."
GITHUB_APP_ID="App ID of the github app."
INSTALLATION_ID="When you install the app, you get an installation ID, something like 73438188"

PROJECT_ID="The Google Cloud Project (GCP) ID"
LOG_NAME="The name of the log in Google Cloud Logging"

REPO_NAME="The name of the *source* repo"
REPO_OWNER="The owner of the *source* repo"
REF="The *source* branch to monitor for changes - e.g. 'main' or 'master'"

COMMITER_EMAIL="The email you want to appear as the committer of the changes, e.g. 'foo@example.com'"
COMMITER_NAME="The name you want to appear as the committer of the changes, e.g. 'GitHub Copier App'"

PORT="leave empty for the default server port, or specify a port, like 8080"
WEBSERVER_PATH="/events"

DEPRECATION_FILE="The path to the deprecation file, e.g. deprecated_examples.json"
CONFIG_FILE="The path to the config file, e.g. config.json"
```


### Add config and deprecation json files

Create the config file to hold config settings and an empty `.json` file to hold deprecated file paths.
See the [config.example.json](configs/config.example.json) for reference.
```json
[
  {
    "source_directory": "generated-examples/go",
    "target_repo": "example-repo",
    "target_branch": "main",
    "target_directory": "go",
    "recursive_copy": true,
    "pr_title": "Update Go Examples",
    "commit_message": "Copy latest Go examples from generated-examples",
    "merge_without_review": false
  }
]
```

Leave the deprecation file an empty array:
```json
[
]
```

### Configure Permissions for the Web App
To configure the app in the source repo, go to the repo's list of web apps. 
You should see the Docs Examples Copier listed:
!["list of web apps"](./readme_files/webapps.png)



## Hosting
This app is hosted in a Google Cloud App Engine, in the organization owned by MongoDB.
The PEM file needed for GitHub Authentication is stored as a secret in the Google Secrets Manager.
For testing locally, you will need to download the auth file from gcloud and store it locally.
See the [Google Cloud documentation](https://cloud.google.com/docs/authentication/application-default-credentials#GAC)
for more information.

### Change Where the App is Hosted
If you deploy this app to a new host/server, you will need to create a new webhook 
in the source repo. See [Set Up A Webhook](#Set Up A Webhook)


## How to Modify and Test
To make changes to this app:
1. Clone this repo.
2. Make the changes. See the next section to understand the project structure.
3. Change the .env.test to match your environment needs, or create a new .env file and reference 
   it in the next step.
4. Test by running `go run app.go -env ./configs/.env.test`
5. Interestingly, you **do not need to change the GitHub app installation**. Why? I think 
   because it is entirely self-contained within this Go app. 

### Testing notes
As of this writing, the source repo (https://GitHub.com/mongodb/docs-code-examples) has 
two webhooks configured: one points to the production version of this application, and 
the other points to a [smee.io proxy](https://smee.io/5Zchxoa1xH7WfYo). 

#### What is smee.io?
Smee.io provides a simple way to point a public endpoint to localhost on your computer. 
It requires the smee cli, which is very lightweight. You run the proxy with a single 
command (`smee -u https://smee.io/5Zchxoa1xH7WfYo`) and any webhooks that go to that 
url will be directed to http://localhost:3000/. This is entirely optional, and there are 
probably other solutions for testing. I just found this dead simple.

**Note** The current production deployment looks for messages on the default 
port and the `/events` route, while your testing hooks (like smee) might only send 
messages on a specific port and/or the default path. This is why you can change the 
port and route in the `.env.*` files.

## Future Work

- ~~BUG/SECURITY: Move .pem to google secret.~~
- ~~Where do we view the log for the app when it hits a snag?~~
     Fixed in 112c8953cbb54d3743b25744fe01f6649f783faa. Added Google 
     Logging and centralized logging service for terminal logging.
- ~~Currently each write is a separate commit. Bad. Fix.~~
     Fixed in f91ccfce74edff56eb305068357a069d12a2020f
- Slack integrations: 
  - notifies a channel when the
    `deprecated_examples.json` file changes, so writers can find those deprecated examples 
    in the docs and update/remove them accordingly. See this 
    [Slack API page](https://api.slack.com/messaging/webhooks).
  - posts log updates (e.g. PR created and ready for review)
- Automate further with hook to Audit DB to get doc files with literal includes & iocode blocks
- ~~Mock tests~~ 
