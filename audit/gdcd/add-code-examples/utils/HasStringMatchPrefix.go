package utils

import (
	"common"
	"strings"
)

func HasStringMatchPrefix(contents string, langCategory string) (string, bool) {
	// These prefixes related to syntax examples
	atlasCli := "atlas "
	mongosh := "mongosh "

	// These prefixes relate to usage examples
	importPrefix := "import "
	fromPrefix := "from "
	namespacePrefix := "namespace "
	packagePrefix := "package "
	usingPrefix := "using "
	mongoConnectionStringPrefix := "mongodb://"
	alternoConnectionStringPrefix := "mongodb+srv://"

	// These prefixes relate to command-line commands that *aren't* MongoDB specific, such as other tools, package managers, etc.
	mkdir := "mkdir "
	cd := "cd "
	docker := "docker "
	dockerCompose := "docker-compose "
	brew := "brew "
	yum := "yum "
	apt := "apt-"
	npm := "npm "
	pip := "pip "
	goRun := "go run "
	node := "node "
	dotnet := "dotnet "
	export := "export "
	sudo := "sudo "
	copyPrefix := "cp "
	tar := "tar "
	jq := "jq "
	vi := "vi "
	cmake := "cmake "
	syft := "syft "
	choco := "choco "

	syntaxExamplePrefixes := []string{atlasCli, mongosh}
	usageExamplePrefixes := []string{importPrefix, fromPrefix, namespacePrefix, packagePrefix, usingPrefix, mongoConnectionStringPrefix, alternoConnectionStringPrefix}
	nonMongoPrefixes := []string{mkdir, cd, docker, dockerCompose, dockerCompose, brew, yum, apt, npm, pip, goRun, node, dotnet, export, sudo, copyPrefix, tar, jq, vi, cmake, syft, choco}

	if langCategory == common.Shell {
		for _, prefix := range syntaxExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return common.SyntaxExample, true
			}
		}
		for _, prefix := range nonMongoPrefixes {
			if strings.HasPrefix(contents, prefix) {
				return common.NonMongoCommand, true
			}
		}
		return "Uncategorized", false
	} else if langCategory == common.Text || langCategory == common.Undefined {
		for _, prefix := range syntaxExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return common.SyntaxExample, true
			}
		}
		for _, prefix := range nonMongoPrefixes {
			if strings.HasPrefix(contents, prefix) {
				return common.NonMongoCommand, true
			}
		}
		for _, prefix := range usageExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return common.UsageExample, true
			}
		}
		return "Uncategorized", false
	} else {
		for _, prefix := range usageExamplePrefixes {
			if strings.HasPrefix(contents, prefix) {
				return common.UsageExample, true
			}
		}
		return "Uncategorized", false
	}
}
