// content.js

// Function to extract track information
function extractTrackInfo() {
  const trackURL = window.location.href;
  const titleElement = document.querySelector('h1[itemprop="name"]');
  const artistElements = document.querySelectorAll('a[itemprop="byArtist"]');

  if (!titleElement || !artistElements.length) {
    console.error('Could not extract track information. Elements not found.');
    return null;
  }

  const trackTitle = titleElement.textContent.trim();
  const trackArtists = Array.from(artistElements).map(artist => artist.textContent.trim()).join(', ');

  // Extract track ID from the URL (assuming it's the last part of the path)
  const urlParts = trackURL.split('/');
  const trackID = urlParts[urlParts.length - 1];

  return {
    url: trackURL,
    id: trackID,
    title: trackTitle,
    artists: trackArtists
  };
}


// Function to create a button with specific styling
function createStyledButton(text, additionalStyles = {}) {
  const button = document.createElement('button');
  button.textContent = text;
  Object.assign(button.style, {
    marginTop: '10px',
    backgroundColor: '#e20074', // Beatport pink
    color: 'white',
    border: 'none',
    padding: '8px 15px',
    borderRadius: '4px',
    cursor: 'pointer',
    ...additionalStyles,
  });
  return button;
}


// Function to handle download errors and update the button state
function handleDownloadError(downloadButton, retryButton, spinner, message, showRetry = false) {
  downloadButton.textContent = message;
  spinner.style.display = 'none'; // Hide spinner on error
  if (showRetry) {
    downloadButton.style.display = 'none'; // Hide main button
    retryButton.style.display = 'inline-block'; // Show retry button
  } else {
    // Optionally re-enable the main button if no retry
    downloadButton.disabled = false;
  }
}

// Function to send a download request to the server, using configured server URL
async function sendDownloadRequest(trackInfo, config) {
  try {
    const response = await fetch('https://localhost:8080/download', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ tracks: [trackInfo] }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();
    return data;

  } catch (error) {
      // Improved error logging including server message if available
      let errorMessage = `Error during download request: ${error.message}`;
      if (error instanceof TypeError && error.message === "Failed to fetch") {
          errorMessage += ". Ensure the server is running and accessible.";
      }
      console.error(errorMessage);
      // Re-throw the error to be caught by the caller
      throw error;
  }
}

// Function to detect track pages and inject a download button, using configured settings
async function injectDownloadButton() {
    // Check if the URL matches the track page pattern

  if (window.location.href.includes('/track/')) {
    const downloadButton = createStyledButton('Download Track', { id: 'beatportdl-download-button' });

    // Add spinner element
    const spinner = document.createElement('span'); // Create spinner
    spinner.className = 'beatportdl-spinner';
    spinner.style.display = 'none'; // Initially hidden
    spinner.style.marginLeft = '5px';
    downloadButton.appendChild(spinner);

    // Add retry button (initially hidden)
    const retryButton = createStyledButton('Retry Download', { display: 'none', marginLeft: '10px' });

    // Container for button and retry
    const buttonContainer = document.createElement('div');
    buttonContainer.style.marginLeft = '10px';
    buttonContainer.insertAdjacentElement('beforeend', downloadButton);
    buttonContainer.insertAdjacentElement('beforeend', retryButton);


    // Add CSS for the spinner (injected into the page)
    const style = document.createElement('style');
    style.textContent = `
      .beatportdl-spinner {
        border: 3px solid rgba(255, 255, 255, 0.3);
        border-radius: 50%;
        border-top: 3px solid white;
        width: 16px;
        height: 16px;
        animation: spin 1s linear infinite;
      }
      @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
      }
    `;

    // Apply Beatport-like button style -  This is a guess and may need adjustment
    downloadButton.classList.add(' വാങ്ങ'); // Assuming 'button' is a common class. Inspect Beatport for the correct class.
    Object.assign(downloadButton.style, {
        fontFamily: 'Roboto, sans-serif', // Common font
        fontSize: '14px',
        fontWeight: '500',
        // Add more styles as needed to match Beatport's buttons
    });



    document.head.appendChild(style); // Inject CSS

    const config = await getExtensionConfig();

    await checkServerHealth(downloadButton, retryButton, spinner);

    downloadButton.addEventListener('click', async () => {
      const trackInfo = extractTrackInfo();
      if (trackInfo) {
        // Change button text and disable
        downloadButton.textContent = 'Preparing Download...';
        spinner.style.display = 'inline';
        downloadButton.disabled = true;

        initiateDownloadWithRetries(trackInfo, downloadButton, retryButton, spinner, config);
      }
    });

    // Retry button event listener
    retryButton.addEventListener('click', () => {
      retryButton.style.display = 'none';
      downloadButton.style.display = 'inline-block';      
      initiateDownloadWithRetries(extractTrackInfo(), downloadButton, retryButton, spinner, config);
    });

    // Find a suitable location to insert the button (e.g., near the track title)
    const titleElement = document.querySelector('h1[itemprop="name"]');

    if (titleElement && titleElement.parentNode) {
      // Create a container for the title and button
      const titleButtonContainer = document.createElement('div');
      titleButtonContainer.style.display = 'flex';
      titleButtonContainer.style.alignItems = 'center'; // Vertically align items
      titleButtonContainer.appendChild(titleElement);
      titleButtonContainer.appendChild(buttonContainer);

      titleElement.parentNode.appendChild(titleButtonContainer);
    } else {
      console.error('Could not find track title element. Appending button to body.');
      document.body.insertAdjacentElement('beforeend', buttonContainer);
    }
  }

}

// Default configuration values
const defaultConfig = {
  serverUrl: 'https://localhost:8080',
  timeout: 10000,
  maxRetries: 2,
};

// Simulate extension configuration retrieval (replace with actual extension API call)
async function getExtensionConfig() {
  // In a real extension, this would use chrome.storage.sync.get or similar
  return defaultConfig;
}

const initiateDownloadWithRetries = async (trackInfo, downloadButton, retryButton, spinner, config) => {
  let retries = config.maxRetries || defaultConfig.maxRetries;
  const attemptDownload = async () => {
    try {
      const downloadRequest = await sendDownloadRequest(trackInfo, config);
      console.log('Download request sent:', downloadRequest);
      const polling = pollForStatus(trackInfo.url, downloadButton, retryButton, spinner, config);

      setTimeout(() => {
        polling.clearInterval();\n        if (downloadButton.textContent !== \'Download Complete\') {\n
          handleDownloadError(downloadButton, retryButton, spinner, 'Download Failed: Server timed out.', true);\n
        }
      }, timeout);\n
    } catch (error) {
      console.error('Error sending download request:', error);
      if (retries > 0) {
        retries--;
        console.log(`Retrying download. Attempts left: ${retries}`);
        await new Promise(resolve => setTimeout(resolve, 1000)); // 1-second delay
        await attemptDownload();
      } else {
        handleDownloadError(downloadButton, retryButton, spinner, "Download Failed: Max retries exceeded", true);
      }
    }
  };
  await attemptDownload();
};

const pollForStatus = (trackURL, downloadButton, retryButton, spinner, config) => {
  const intervalId = setInterval(async () => {
     try {
      const response = await fetch('http://localhost:8080/status');
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      console.log('Status update:', data);
      const downloadStatus = data.find(status => status.TrackURL === trackURL);
      if (downloadStatus) {
        checkDownloadStatus(downloadStatus, downloadButton, retryButton, spinner);
      }
    } catch (error) {
      console.error('Error polling for status:', error);
      handleDownloadError(downloadButton, retryButton, spinner, 'Download Failed: Server not responding.', true);
      clearInterval(intervalId);
    }
  }, (config.timeout || defaultConfig.timeout) / 5);
  setTimeout(() => {
    clearInterval(intervalId);    if (downloadButton.textContent !== 'Download Complete') {      console.error('Polling timed out.');      handleDownloadError(downloadButton, retryButton, spinner, 'Download Failed: Server timed out.', true);    }  }, config.timeout); // Use timeout here as well
  return { intervalId, clearInterval: () => clearInterval(intervalId) };
};
const checkDownloadStatus = (downloadStatus, downloadButton, retryButton, spinner, polling) => {
    const status = downloadStatus.Status;

    if (status === 'downloading') {
        const progress = downloadStatus.Metadata?.progress;
        if (progress !== undefined) {
            downloadButton.textContent = `Downloading (${Math.round(progress)}%)`;
        } else {
            downloadButton.textContent = 'Downloading';
        }
    } else if (status === 'completed') {
        downloadButton.textContent = 'Download Complete!';
        spinner.style.display = 'none'; // Ensure spinner is hidden on completion
        downloadButton.disabled = false;
        polling.clearInterval();
    } else if (status === 'failed') {
        const error = downloadStatus.Metadata?.error || 'Download Failed';
        const errorMessage = `Download Failed: ${error}`;
        handleDownloadError(downloadButton, retryButton, spinner, errorMessage, true);
    }
};


// Function to check server health, using configured server URL
async function checkServerHealth(downloadButton, retryButton, spinner, config) {
  try {
        // Use config.serverUrl for health check
        const response = await fetch('https://localhost:8080/health');
        if (!response.ok) {
            throw new Error(`Server health check failed with status: ${response.status}`);
        }
    } catch (error) {
        console.error('Server health check failed:', error);
        handleDownloadError(downloadButton, retryButton, spinner, 'Server Unavailable');
    }
}
// Inject the button on page load
injectDownloadButton();
