# Beatport Track Download Project

## Project Overview
BeatportDL is a tool for downloading tracks from Beatport (and Beatsource, if applicable) in various formats (FLAC, AAC).  A Beatport Streaming Subscription is required.  The project includes a browser extension, a backend server, and a command-line tool.

## System Architecture

The system has three main components:

*   **Server:** A Go web server that handles download requests initiated by the browser extension. It interacts with the Beatport API to authenticate and retrieve download links. The server manages download concurrency using a semaphore. After a successful download, it performs actions such as applying metadata tags to the file and organizing it within the file system.

*   **Extension:** A browser extension (likely for Chrome) that integrates with the Beatport website. It modifies the website's interface to allow users to initiate downloads directly from track pages or playlists.  The extension communicates with the server using a JSON API, sending download requests and receiving status updates.

*   **Command-line Tool:** A Go command-line application (the original BeatportDL) that offers an alternative way to download tracks.  Users can provide Beatport URLs as arguments or from a text file.  The tool interacts with the Beatport API and handles tagging/file organization in a similar manner to the server.

## Installation

**Server and Command-Line Tool:**

Refer to the [Releases page](https://github.com/unspok3n/beatportdl/releases/) on GitHub for pre-built binaries for Windows, macOS, and Linux.  Download the appropriate binary for your system.

On Unix-based systems (macOS, Linux), you may need to set execute permissions on the downloaded binaries: `chmod +x beatportdl-server` and `chmod +x beatportdl`.

Alternatively, you can build the server and command-line tool from source:

1.  Ensure you have Go installed (version 1.21+ is recommended).
2.  Clone the repository:  `git clone <repository_url>` (replace with the actual repository URL)
3.  Navigate to the project directory: `cd <project_directory>`
4.  Build the server: `go build -o beatportdl-server cmd/server/main.go`
5.  Build the command-line tool: `go build -o beatportdl main.go` (assuming main.go is in the project root)

Building from source requires additional dependencies: [TagLib](https://github.com/taglib/taglib), [zlib](https://github.com/madler/zlib), and a [Zig C/C++ Toolchain](https://github.com/ziglang/zig).  The build process uses a Makefile, which may require customization based on your system's library locations.  See the original project's `README.md` for more detailed build instructions, including how to use environment variables to specify library paths.

**Browser Extension:**

1.  Navigate to the `beatportdl-extension` directory.
2.  Follow the instructions in the extension's `README.md` (if available) or the browser's documentation for loading unpacked extensions (e.g., in Chrome, use "Developer mode" and "Load unpacked").

## Usage

**Server:**

1.  Start the server:  `./beatportdl-server` (or however you named the server executable)

The server typically listens on port 8080. To download tracks, the browser extension (or a custom client) must send HTTP POST requests to the server's `/download` endpoint with a JSON payload containing track information (URL, ID, title, artist).  The server's `/status` endpoint can be queried (via HTTP GET) to retrieve the status of ongoing downloads.

**Browser Extension:**

1.  Navigate to a Beatport track page.
2.  Use the extension's modified UI (e.g., a download button added to the page) to initiate a download for the track or tracks you want.

The extension handles sending the download request to the server and may provide visual feedback on the download progress.

**Command-Line Tool:**

1.  Run the command-line tool, providing Beatport track URLs as arguments:  `./beatportdl <track_url_1> <track_url_2> ...` (or however you named the command-line executable)

Alternatively, you can provide a text file containing a list of Beatport URLs (one URL per line):

`./beatportdl file.txt`

## Contributing

Contributions are welcome! Please follow the Gitflow workflow and submit pull requests for any changes.  See `CONTRIBUTING.md` (or create one) for more detailed guidelines.

*(Placeholder - We'll add more specific contribution guidelines later.)*

## API Documentation

*(Placeholder - To be filled with details about the server's API endpoints, request/response formats, etc.)*

*(This documentation may be moved to a separate file, e.g., `API.md`, later.)*

## AI Rules and Coding Standards

- **General Principles**

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

- **Technology-Specific Rules**

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

- **API Communication**

  - **JSON Format:** All communication between the extension and the server must use JSON format.
  - **Request Structure:** The extension should send download requests to the `/download` endpoint with a JSON body containing track information (URL, ID, title, artist).
  - **Response Structure:** The server should respond with appropriate status codes and JSON bodies, including success/failure messages and any relevant data (e.g., download status, error messages, and optionally, a download ID).
  - **Status Endpoint:** The extension should poll the `/status` endpoint to get updates on download progress. The server should return a JSON response with status information for ongoing downloads.

- **Git Workflow**

  - **Gitflow:** Follow the Gitflow workflow for managing branches and releases.
  - **Feature Branches:** Create feature branches for developing new features or making significant changes.
  - **Release Branches:** Create release branches for preparing releases.
  - **Hotfix Branches:** Create hotfix branches for addressing urgent issues in production.
  - **Commit Messages:** Write clear and descriptive commit messages that follow a consistent format (e.g., `type(scope): description`).

- **Code Quality Maintenance**

  - **Refactor Continuously:** Regularly refactor code to improve its structure, readability, and maintainability.
  - **Fix Technical Debt Early:** Address any technical debt (e.g., temporary workarounds, suboptimal solutions) as soon as possible.
  - **Leave Code Cleaner Than You Found It:** When making changes, strive to improve the existing code, even if it's not directly related to the changes.