package snooty

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"snooty-api-parser/types"
)

func ReadDocsForGitHubUser(reader bufio.Reader, username string) []types.PageWrapper {
	var docsPages []types.PageWrapper
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
			maybePage = GetPageFromResponse(trimmedLine, username)
			if maybePage != nil {
				docsPages = append(docsPages, *maybePage)
			}
		}
	}
	return docsPages
}
