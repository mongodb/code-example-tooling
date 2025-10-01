# Debug Failing Comparison Utility

## Context
The current context is a <PROGRAMMING LANGUAGE> project designed to test the
MongoDB <DRIVER> Driver code examples shown in the MongoDB documentation.

This project contains a utility that compares the actual output from running a
code example against a defined, expected output. The entrypoint for that
utility is <ENTRYPOINT_FILEPATH>.

## Problem Statement 

A specific test is executing a function from an example file and comparing its
output to the expected output defined in <EXPECTED_OUTPUT_PATH_OR_VAR>. The test currently
fails because the outputs differ, but they should match. This suggests an issue in the comparison utility's implementation.

## Constraints

- Do not modify the example file or expected output.

## Task

- Investigate and debug why the comparison utility returns false for this case.
- Fix any issues found in the comparison utility.
- (Optional but recommended) If helpful, update or add tests for identified edge cases to ensure robustness.

If you need more information (such as sample actual/expected outputs or error messages), let me know.

## Success Criteria

The comparison utility should return `true` when the actual and expected outputs match for this test.
