# Overview

This application finds the most similar Git tag for a given tag in a target repository.

# Coding Style

*   Follow standard Go conventions and formatting (`make lint` to verify).
*   Wrap returned errors with `errors.Join` using a distinct `Err...` variable.
*   Use `defer func() { _ = closer.Close() }()` when closing resources.
*   Always specify parameter types in function signatures.
*   Use "range over integers" for `for` loops where appropriate.

# Commands

*   `make mockgen`: Generate mocks.
*   `make fmt`: Check code quality.
*   `make test`: Run all tests.