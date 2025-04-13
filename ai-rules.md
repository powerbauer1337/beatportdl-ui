# AI Rules for Beatport Track Download Extension

This document outlines the AI rules for the Beatport Track Download Extension project, combining insights from existing project files and general coding best practices.

## General Principles

- **Clean Code:** Write code that is easy to read, understand, and maintain. Follow the principles outlined in the "Clean Code Guidelines" (based on `rules/clean_code.mdc`).
- **Clear and Concise Language:** Use clear and concise language in code, comments, and documentation to ensure accessibility for all users and contributors (based on `rules/readme.mdc`).
- **Meaningful Names:** Choose descriptive and meaningful names for variables, functions, and classes that clearly convey their purpose.
- **Smart Comments:** Focus on writing self-documenting code. Use comments sparingly to explain "why" rather than "what". Document APIs, complex algorithms, and non-obvious side effects.
- **Single Responsibility:** Ensure each function has a single, well-defined purpose. Keep functions small and focused.
- **DRY (Don't Repeat Yourself):** Avoid code duplication by extracting reusable logic into functions or modules.
- **Code Organization:** Maintain a consistent and logical structure for files and folders. Group related code together.
- **Encapsulation:** Hide implementation details and expose clear interfaces to interact with different parts of the code.
- **Error Handling:** Implement robust error handling to gracefully manage potential issues, including network errors, server unavailability, and invalid data.
- **Testing:** Write comprehensive tests to ensure code correctness and prevent regressions. Test edge cases and error conditions.
- **Version Control:** Follow the "Gitflow Workflow Rules" (based on the provided description) for branching, committing, and merging changes. Use clear and descriptive commit messages.

## Technology-Specific Rules

### JavaScript (Browser Extension)

- **ES6+ Syntax:** Use modern JavaScript syntax (ES6+) for improved readability and functionality.
- **Consistent Coding Style:** Follow a consistent coding style (e.g., Airbnb JavaScript Style Guide) for formatting and conventions.
- **DOM Manipulation:** Use standard DOM APIs (`document.createElement`, `element.appendChild`, `document.querySelector`, etc.) for interacting with the Beatport page.
- **Asynchronous Operations:** Use `fetch` API for making HTTP requests to the server. Handle promises correctly with `.then` and `.catch` or `async/await`.
- **JSON Communication:** Use JSON for all requests and responses between the extension and the `beatportdl` server.
- **User Interface:** Create a user-friendly and visually consistent interface that integrates seamlessly with the Beatport website.
- **Error Handling:** Provide clear and informative error messages to the user in case of failures.
- **Status Updates:** Implement a mechanism (e.g., polling) to provide feedback to the user on the download status.
- **Security:** Sanitize user inputs and be mindful of potential security vulnerabilities.

### Go (Server)

- **Effective Go:** Adhere to the guidelines in "Effective Go" for writing idiomatic Go code.
- **Clear Naming:** Use clear and concise naming conventions for variables, functions, types, and packages.
- **Comments:** Comment complex logic and exported functions/types to improve understanding and maintainability.
- **Error Handling:** Handle errors appropriately using Go's error handling mechanisms. Always check for errors and return them to the caller when necessary.
- **Concurrency:** Use goroutines and channels for managing concurrent operations, ensuring thread safety when accessing shared resources.
- **API Endpoints:** Implement API endpoints that adhere to RESTful principles, using appropriate HTTP methods and status codes.
- **JSON Communication:** Use Go's `encoding/json` package for encoding and decoding JSON data for API communication.
- **Data Validation:** Validate incoming data to ensure it meets the required format and constraints.
- **Security:** Implement appropriate security measures to protect against common vulnerabilities.

## API Communication

- **JSON Format:** All communication between the extension and the server must use JSON format.
- **Request Structure:** The extension should send download requests to the `/download` endpoint with a JSON body containing track information (URL, ID, title, artist).
- **Response Structure:** The server should respond with appropriate status codes and JSON bodies, including success/failure messages and any relevant data (e.g., download status, error messages, and optionally, a download ID).
- **Status Endpoint:** The extension should poll the `/status` endpoint to get updates on download progress. The server should return a JSON response with status information for ongoing downloads.

## Git Workflow

- **Gitflow:** Follow the Gitflow workflow for managing branches and releases.
- **Feature Branches:** Create feature branches for developing new features or making significant changes.
- **Release Branches:** Create release branches for preparing releases.
- **Hotfix Branches:** Create hotfix branches for addressing urgent issues in production.
- **Commit Messages:** Write clear and descriptive commit messages that follow a consistent format (e.g., `type(scope): description`).

## Code Quality Maintenance

- **Refactor Continuously:** Regularly refactor code to improve its structure, readability, and maintainability.
- **Fix Technical Debt Early:** Address any technical debt (e.g., temporary workarounds, suboptimal solutions) as soon as possible.
- **Leave Code Cleaner Than You Found It:** When making changes, strive to improve the existing code, even if it's not directly related to the changes.

## Example Commit Message
