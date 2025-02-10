# LLM-Assisted Categorization in Go with Ollama

## Description

This is a small, unstructured proof-of-concept project using Golang with
Ollama local LLM to perform categorization. The project currently:

- Builds a list of file paths recursively from the specified start directory
- Reads the contents of each file into memory and asks the LLM to categorize it
- Write two reports to file as JSON in an `output` directory:
  - A report of category counts broken down by language
  - A report with details about each snippet

The prompt is structured to categorize code examples based on definitions that
the docs organization is currently codifying.

## Install the dependencies

### Golang

This test suite requires you to have `Golang` installed. If you do not yet
have Go installed, refer to [the Go installation page](https://go.dev/doc/install)
for details.

### Go Dependencies

From the project root, run the following command to install
dependencies:

```shell
go get code-example-categorization
```

### Ollama

This project uses Ollama running locally on the device to perform the
categorization task. Refer to the [Ollama](https://ollama.com/) website for
installation details.

#### Model

This project uses the Ollama [qwen2.5-coder](https://ollama.com/library/qwen2.5-coder)
model. At the time of writing this README, this is the latest series of
code-specific Qwen models, with a focus on improved code reasoning, code
generation, and code fixing. This model has consistently produced the most
accurate results when categorizing code examples. For different types of
categorization tasks, you may need a different model.

To install the model locally, after you have installed Ollama, run the following
command:

```shell
ollama pull qwen2.5-coder
```

If you want to use a different model, pull a different model from Ollama, and
change the model name in `constants.go`. The model name is a constant so it's
available to both the project and the tests.

## Run the project

With the model and dependencies installed, and Ollama running on your machine,
you can run the project from an IDE or from the command line.

### Change the start directory path (optional)

As written, this project categorizes files in the `examples/` directory of
this repository. If you'd like to categorize files in a different part of your
file system, change the path in `constants.go`.

This project currently categorizes _all_ files in the given directory. If you
want to differentiate between code examples and other types of files, add
logic to handle files with different extensions or naming structures, or add
other logic as needed to differentiate between files to process and files to
ignore.

### IDE

To run the project from an IDE, press the `play` button next to the `main()`
func in `main.go`. Alternately, press the `Build` button in the top right of
your IDE.

### Command-line

To run the project from the command line, run the following commands:

```
go build
go run .
```

## Run the tests

This project includes basic tests to verify the functionality. You might want
to run the tests again if you change the prompt or the model, or if you need
to add new tests because you're modifying the logic for traversing files or
generating an artifact.

### IDE

#### Run a single test

To run a test from an IDE, press the `play` button next to the test
function you want to run.

#### Run all the tests

In any test file, press the `play` button next to the package declaration and
select `Run` -> `go test test-code-example-categorization`

### Command-line

#### Run a single test

To run a test from the command line, run the following command:

```
go test code-example-categorization -run TestFuncName
```

Example:

```
go test code-example-categorization -run TestGetSnippetHash
```

#### Run all the tests

To run all tests from the command line, run the following command:

```
go test
```
