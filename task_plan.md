## Beatport Track Download Extension - AI Agent Optimized Plan (Structured Task Prompting)

**Coding Rules:**

*   **JavaScript:** Use ES6+ syntax. Follow a consistent coding style (e.g., Airbnb JavaScript Style Guide). Comment complex logic.
*   **Go:** FollowEffective Go guidelines. Use clear and concise naming conventions. Comment complex logic and exported functions/types. Handle errors appropriately.
*   **API Communication:** Use JSON for all requests and responses between the extension and the server.

**Git Strategy:**

*   Create a new branch for each iteration (e.g., `feat/extension-iteration-1`).
*   Commit changes frequently with descriptive commit messages.
*   Push branches to the remote repository.

---

**Phase: 1 - Basic Download Initiation**

**Goal:** Enable the extension to detect individual track pages, extract track information, and send a download request to the `beatportdl` server.

**Step: 1.1 - Track Page Detection and Download Button**

*   **Component:** Browser Extension (JS)
*   **Goal:** Detect Beatport track pages and inject a "Download Track" button.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`, implement logic to detect track pages by checking if the URL contains the pattern `/track/`.
    2.  Create a "Download Track" button element using DOM manipulation (`document.createElement`).
    3.  Insert the button into a user-friendly location on the page (e.g., near the track title) using `element.appendChild`.
    4.  Style the button to visually integrate with Beatport's design using CSS.
*   **Verification:**
    1.  Manually navigate to a Beatport track page in the browser.
    2.  Verify that the "Download Track" button appears on the page in the expected location.
*   **Git:**
    1.  Create branch: `feat/extension-iteration-1`.
    2.  Commit: "feat(extension): add download button and track page detection".

**Step: 1.2 - Track Information Extraction**

*   **Component:** Browser Extension (JS)
*   **Goal:** Extract track URL, ID, title, and artist(s) from the Beatport track page.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`, implement a function to extract track information.
    2.  Use `window.location.href` to get the track URL.
    3.  Use `document.querySelector` and DOM manipulation to extract:
        *   Track ID (inspect Beatport's HTML source to identify the relevant element/attribute).
        *   Track title (identify the element containing the title).
        *   Track artist(s) (identify the element(s) containing the artist names).
*   **Verification:**
    1.  Manually navigate to a Beatport track page.
    2.  Use the browser's developer console to call the extraction function.
    3.  Verify that the function returns an object with the correct track URL, ID, title, and artist(s).
*   **Git:**
    1.  Commit: "feat(extension): extract track information".

**Step: 1.3 - Download Request**

*   **Component:** Browser Extension (JS)
*   **Goal:** Send a download request to the `beatportdl` server with the extracted track information.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`, attach a click event listener to the "Download Track" button.
    2.  When the button is clicked:
        *   Call the track information extraction function.
        *   Construct a JSON object with the extracted data in the format:
```
json
{
                "tracks": [
                    {
                        "url": "track_url_here",
                        "id": "track_id_here",
                        "title": "track_title_here",
                        "artists": "track_artists_here"
                    }
                ]
            }
```
*   Use the `fetch` API to send a `POST` request to the `beatportdl` server's `/download` endpoint.
        *   Include the JSON object in the request body.
        *   Set the `Content-Type` header to `application/json`.
*   **Verification:**
    1.  Manually navigate to a Beatport track page.
    2.  Click the "Download Track" button.
    3.  Use the browser's network console to:
        *   Verify that a `POST` request is sent to the correct URL (`/download`).
        *   Inspect the request headers to ensure `Content-Type` is `application/json`.
        *   Inspect the request body to verify it contains the track information in the correct JSON format.
*   **Git:**
    1.  Commit: "feat(extension): send download request to server".

**Step: 1.4 - Minimal User Feedback**

*   **Component:** Browser Extension (JS)
*   **Goal:** Provide immediate feedback to the user after sending the download request.
*   **Instructions:**
    1.  In the click event listener in `beatportdl-extension/content.js`:
        *   Change the text of the "Download Track" button to "Download Started...".
        *   Optionally, display a small, temporary message (e.g., "Downloading...") near the button.
*   **Verification:**
    1.  Manually navigate to a Beatport track page.
    2.  Click the "Download Track" button.
    3.  Verify that the button text changes to "Download Started..." (or the chosen message).
*   **Git:**
    1.  Commit: "feat(extension): provide initial download feedback".

---

**Phase: 2 - Enhanced Feedback & Error Handling**

**Goal:** Improve the user experience by providing download status updates and handling common error scenarios.

**Step: 2.1 - Download Status Feedback**

*   **Component:** Browser Extension (JS)
*   **Goal:** Poll the server for download status and update the UI accordingly.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`:
        *   When the "Download Track" button is clicked, change the button text to "Preparing Download..." and disable the button.
        *   Implement a function to poll the `beatportdl` server's `/status` endpoint using `setInterval` (every 2-3 seconds).
        *   In the polling function:
            *   Send a `GET` request to `/status`.
            *   Parse the JSON response.
            *   Iterate through the download statuses in the response.
            *   Identify the relevant download by matching the `TrackURL` with the URL of the track being downloaded.
            *   Extract the `Status` field (and any error information from `Metadata["error"]` if the status is "failed").
            *   Update the button text based on the status:
                *   "Preparing Download..." (initial or no match)
                *   "Downloading..." (status is "downloading")
                *   "Download Complete!" (status is "completed")
                *   "Download Failed: {error}" (status is "failed")
            *   Re-enable the button only when the download is complete or has failed.
*   **Verification:**
    1.  Manually test by initiating a download.
    2.  Verify that the button text updates correctly to reflect the different download statuses.
    3.  Use the browser's network console to observe the polling requests to `/status`.
*   **Git:**
    1.  Commit: "feat(extension): poll for download status and update UI".

**Step: 2.2 - Basic Error Handling**

*   **Component:** Browser Extension (JS)
*   **Goal:** Handle network errors and server-side download failures.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`:
        *   In the `fetch` promise's `.catch` blocks (for both the initial download request and the status polling), display a user-friendly error message near the download button (e.g., "Download Failed: Network error") and re-enable the button (or offer a retry mechanism).
        *   In the status polling function, if a matching track is found and its status is "failed", display the server-provided error message (extracted from `Metadata["error"]`) and re-enable the button (or offer a retry).
*   **Verification:**
    1.  Manually test by:
        *   Simulating network errors (e.g., by temporarily disabling the network).
        *   If possible, triggering server-side errors (e.g., by modifying the extension to send invalid track data).
    2.  Verify that appropriate error messages are displayed in the UI and the button is re-enabled.
*   **Git:**
    1.  Commit: "feat(extension): handle network and server errors".

---

**Phase: 3 - UI Refinements & Robustness**

**Goal:** Refine the user interface, add robust error handling, and address potential edge cases.

**Step: 3.1 - User Interface Refinements**

*   **Component:** Browser Extension (JS)
*   **Goal:** Enhance the UI with a progress indicator and a retry button.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`:
        *   Consider replacing the text-based status updates with a visual progress indicator (e.g., a spinner animation using CSS).
        *   Add a "Retry" button that appears if a download fails, allowing the user to easily attempt the download again.
        *   Further polish the styling of all UI elements (button, messages, progress indicator) to ensure visual consistency with Beatport's design and a clean user experience.
*   **Verification:**
    1.  Manually test the extension.
    2.  Verify that the progress indicator (if implemented) is displayed correctly.
    3.  Verify that the "Retry" button appears on download failure and functions as expected.
    4.  Assess the overall visual presentation and ensure it aligns with Beatport's design.
*   **Git:**
    1.  Commit: "feat(extension): refine UI with progress indicator and retry button".

**Step: 3.2 - Robust Error Handling**

*   **Component:** Browser Extension (JS)
*   **Goal:** Implement timeout, retry logic, and handle server unavailability.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`:
        *   Implement a timeout for the status polling requests (e.g., 5-10 seconds). If no response is received within the timeout, display an error message (e.g., "Download Failed: Server not responding") and offer a retry or re-enable the button.
        *   Implement a retry strategy: automatically retry failed downloads 2-3 times with a short delay (e.g., 1 second) between attempts. If all retries fail, display a final error message.
        *   Handle server unavailability: before sending the initial download request, attempt a "health check" by fetching a simple resource from the server (if such an endpoint exists). If the server is unreachable, display a clear message to the user.
*   **Verification:**
    1.  Manually test by:
        *   Simulating slow network conditions or server timeouts.
        *   Temporarily making the server unavailable.
    2.  Verify that the extension handles these scenarios gracefully and displays appropriate error messages.
*   **Git:**
    1.  Commit: "feat(extension): implement robust error handling".

**Step: 3.3 - Address Status Tracking Limitations**

*   **Component:** Browser Extension (JS)
*   **Goal:** Mitigate potential race conditions in status tracking.
*   **Instructions:**
    1.  In `beatportdl-extension/content.js`:
        *   To address the potential race condition of matching downloads by `TrackURL`, implement more frequent status checks (reduce the polling interval).
        *   If possible, explore more robust matching strategies. For example, if timestamps or other unique metadata become available in the `/status` response in the future, incorporate them into the matching logic.
*   **Verification:**
    1.  Carefully monitor the extension's behavior during testing, especially when initiating multiple downloads.
    2.  If the server is modified to return a unique download ID (see "Future Improvements" below), update the extension to use this ID for status tracking instead of `TrackURL`.
*   **Git:**
    1.  Commit: "feat(extension): mitigate status tracking race condition".

**Future Improvements (Optional, Requires Server Changes):**

*   **Server-Generated Download ID:** Modify the `beatportdl` server's `/download` endpoint (in `cmd/server/main.go`) to return a unique download ID (UUID from `github.com/google/uuid`) in the JSON response when a download is initiated. This will significantly improve the robustness and accuracy of status tracking in the extension.  The response should include a `download_ids` field with an array of IDs (even if only one track is downloaded). Example:
```
json
{
        "message": "Download(s) initiated",
        "download_ids": ["unique_download_id"]
    }
```
*   **Extension Update for Download ID:** If the server is modified as described above, update the extension's status tracking logic in `beatportdl-extension/content.js` to use the server-generated download ID instead of matching by `TrackURL`. This will involve:
    1.  Capturing the `download_ids` from the `/download` response.
    2.  Modifying the status polling function to filter the `/status` response based on these IDs.

This structured task plan provides a detailed, step-by-step guide for developing the Beatport track download extension, optimized for clarity, AI-readability, and maintainability. Remember to adapt and refine the plan as needed based on your specific implementation and testing results.