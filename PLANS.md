# Plan for Git Tag Similarity Application

This document outlines the plan to build an application that compares two Git tags and calculates their similarity based on commit history.

## 1. Project Restructuring and Refactoring

The initial step is to refactor the existing code in `main.go` to improve modularity, making it easier to extend and maintain.

*   **Create a `Repository` struct:** This struct will encapsulate all interactions with the `go-git` repository object. It will hold the repository path and the `go-git` repository instance.
*   **Isolate Tag and Commit Logic:**
    *   Create a function to fetch all tag references from the repository.
    *   Create a function that, for a given tag, traverses its history and returns a set of all its parent commit hashes.
*   **Abstract Similarity Calculation:**
    *   Create a dedicated function to calculate the Jaccard similarity score between two commit sets.
*   **Centralize the Main Logic:** The `main` function will be simplified to orchestrate the application flow.

### Testing Plan for Step 1

*   **Interface-Based Design:** To ensure testability, a `Repository` interface will be defined to abstract Git operations. The concrete implementation using `go-git` will implement this interface. This allows for dependency injection of a mock object in later tests.
*   **Unit Test Similarity Calculation:** The Jaccard similarity function will be unit-tested in isolation with a variety of inputs (e.g., empty sets, identical sets, disjoint sets) to verify its mathematical correctness.

## 2. Command-Line Interface (CLI)

The application will be controlled via command-line arguments.

*   **Use the `flag` package:** Implement CLI argument parsing using Go's standard `flag` package.
*   **Define Required Arguments:**
    *   `-repo`: A string flag for the absolute path to the target Git repository.
    *   `-tag1`: A string flag for the first tag name to compare.
    *   `-tag2`: A string flag for the second tag name to compare.
*   **Input Validation:** Implement checks to ensure that required arguments are provided and that the repository and both tags exist. If validation fails, print a helpful usage message and exit gracefully.

### Testing Plan for Step 2

*   **Unit Test CLI Logic:** The command-line argument handling logic will be unit-tested to verify that it correctly parses flags, handles missing or invalid arguments, and displays usage information as expected.

## 3. Core Tag Comparison Logic

This is the central part of the application that compares two tags.

*   **Fetch Commit Sets:** Calculate the commit set for both tag1 and tag2.
*   **Calculate Similarity:** Compute the Jaccard similarity score between the two commit sets.
*   **Calculate Differences:** For each tag, identify commits that are unique to that tag (present in one but not the other).

### Testing Plan for Step 3

*   **Mocking:** A mock implementation of the `Repository` interface will be used for testing.
*   **Unit Test Core Logic:** The tag comparison logic will be tested by injecting the mock `Repository`. This will allow for the simulation of various scenarios (tags with identical, overlapping, or distinct histories; error conditions) to verify correct behavior.

## 4. Output Formatting

The output should be clear, concise, and informative for the user.

*   **Similarity Score:** Display the similarity percentage between the two tags (formatted to two decimal places).
*   **Commit Differences:** For each tag, display a list of commits that are unique to that tag:
    *   Commit SHA (7 characters)
    *   Commit message (first line)
*   **Summary Information:** Include total commit counts for each tag and the number of shared commits.

### Testing Plan for Step 4

*   **Capture and Assert Output:** The tests for the main application logic will capture the standard output and assert that it is formatted correctly.

## 5. Integration Testing

After the individual components are unit-tested, an end-to-end integration test will be performed.

*   **Test Repository Setup:** The test will programmatically create a temporary, local Git repository with a known and controlled structure of commits and tags.
*   **End-to-End Validation:** The compiled application will be executed as a subprocess with arguments pointing to the temporary repository. The test will then assert that the final output is correct, ensuring that the concrete `go-git` implementation and all components work together as expected.
