package snooty

import (
	"bufio"
	"bytes"
	"gdcd/types"
	"io"
	"log"
)

// ReadPagesForGitHubUser creates a slice of []types.PageWrapper with logic to avoid double-counting pages as a workaround
// for an outstanding DOP bug.
func ReadPagesForGitHubUser(reader bufio.Reader) []types.PageWrapper {
	var allIncomingDocsPages []types.PageWrapper
	for {
		line, err := reader.ReadBytes('\n') // Read until newline
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("Error reading response: %v", err)
		}

		trimmedLine := bytes.TrimSpace(line)
		var maybePage *types.PageWrapper
		if len(trimmedLine) > 0 { // Process non-empty lines
			maybePage = GetPageFromResponse(trimmedLine)
			if maybePage != nil {
				allIncomingDocsPages = append(allIncomingDocsPages, *maybePage)
			}
		}
	}

	// A DOP bug has introduced duplicate pages for some projects - one page with a GitHub username "netlify", and one with
	// a GitHub username "docs-builder-bot". We don't want to double-count pages/code examples, so this logic should only
	// count pages once depending on the GitHub username. Not all projects have duplicate usernames. Per DOP, prefer
	// "netlify" if it exists. If there is no netlify user, we can just return the docs-builder-bot documents, or if there
	// is no GitHub user at all, we return the noUsernamePages.
	// When this DOP ticket is resolved, we can remove this logic: https://jira.mongodb.org/browse/DOP-5440
	var docsBuilderBotPages []types.PageWrapper
	var netlifyPages []types.PageWrapper
	var noUsernamePages []types.PageWrapper

	for _, page := range allIncomingDocsPages {
		switch page.Data.GitHubUsername {
		case GitHubUsernameNetlify:
			netlifyPages = append(netlifyPages, page)
		case GitHubUsernameDocsBuilderBot:
			docsBuilderBotPages = append(docsBuilderBotPages, page)
		default:
			noUsernamePages = append(noUsernamePages, page)
		}
	}

	if len(netlifyPages) > 0 {
		return netlifyPages
	} else if len(docsBuilderBotPages) > 0 {
		return docsBuilderBotPages
	} else {
		return noUsernamePages
	}
}
