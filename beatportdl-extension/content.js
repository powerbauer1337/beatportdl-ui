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





// Function to detect track pages and inject a download button
function injectDownloadButton() {
  // Check if the URL matches the track page pattern
  if (window.location.href.includes('/track/')) {
    // Create the download button
    const downloadButton = document.createElement('button');
    downloadButton.textContent = 'Download Track';
    downloadButton.style.marginTop = '10px';
    downloadButton.id = 'beatportdl-download-button'; // Keep the ID

    // Basic styling to match Beatport's design (adjust as needed)
    downloadButton.style.backgroundColor = '#e20074'; // Beatport pink
    downloadButton.style.color = 'white';
    downloadButton.style.border = 'none';
    downloadButton.style.padding = '8px 15px';
    downloadButton.style.borderRadius = '4px';
    downloadButton.style.cursor = 'pointer';

    // Add spinner element
    const spinner = document.createElement('span');
    spinner.className = 'beatportdl-spinner';
    spinner.style.display = 'none'; // Initially hidden
    spinner.style.marginLeft = '5px';
    downloadButton.appendChild(spinner);

    // Add retry button (initially hidden)
    const retryButton = document.createElement('button');
    retryButton.textContent = 'Retry Download';
    retryButton.style.display = 'none'; // Initially hidden
    retryButton.style.marginLeft = '10px';
    retryButton.style.backgroundColor = '#e20074';
    retryButton.style.color = 'white';
    retryButton.style.border = 'none';
    retryButton.style.padding = '8px 15px';
    retryButton.style.borderRadius = '4px';
    retryButton.style.cursor = 'pointer';

    // Container for button and retry
    const buttonContainer = document.createElement('div');
    buttonContainer.style.marginTop = '10px';
    buttonContainer.appendChild(downloadButton);
    buttonContainer.appendChild(retryButton);


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
    document.head.appendChild(style);

    // Add click listener
    downloadButton.addEventListener('click', async () => {
      const trackInfo = extractTrackInfo();
      if (trackInfo) {
        // Change button text and disable
        downloadButton.textContent = 'Preparing Download...';
        spinner.style.display = 'inline-block';
        downloadButton.disabled = true;

        // Send download request
        const sendDownloadRequest = async () => {
          try {
            const response = await fetch('http://localhost:8080/download', {
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
            console.log('Download request sent:', data);

            // Start polling for status updates
            pollForStatus(trackInfo.url);

          } catch (error) {
            console.error('Error sending download request:', error);
            downloadButton.textContent = 'Download Failed';
            spinner.style.display = 'none';
            downloadButton.style.display = 'none';
            retryButton.style.display = 'inline-block';
          }
        };

        initiateDownloadWithRetries(sendDownloadRequest);

        // Function to poll for status updates
        const pollForStatus = (trackURL) => {
          const intervalId = setInterval(async () => {
            try {
              const response = await fetch('http://localhost:8080/status');
              if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
              }

              const data = await response.json();
              console.log('Status update:', data);

              // Find the relevant download
              const downloadStatus = data.find(status => status.TrackURL === trackURL);

              if (downloadStatus) {
                const status = downloadStatus.Status;
                const error = downloadStatus.Metadata?.error || 'Unknown error';

                if (status === 'downloading') {
                  downloadButton.textContent = 'Downloading';
                } else if (status === 'completed') {
                  downloadButton.textContent = 'Download Complete';
                  spinner.style.display = 'none';
                  clearInterval(intervalId);
                  downloadButton.disabled = false;
                } else if (status === 'failed') {
                  downloadButton.textContent = `Download Failed: ${error || 'Unknown error'}`;
                  spinner.style.display = 'none';
                  clearInterval(intervalId);
                  // Show retry
                  downloadButton.style.display = 'none';
                  retryButton.style.display = 'inline-block'; // Show retry
                  downloadButton.disabled = false; // Enable retry
                }
              }
            } catch (error) {
              console.error('Error polling for status:', error);
              handleDownloadError('Download Failed: Server not responding.', true);
              clearInterval(intervalId);
            }
          }, 1000);

          // Timeout for polling
          setTimeout(() => {
            clearInterval(intervalId);
            if (downloadButton.textContent !== 'Download Complete') {
              console.error('Polling timed out.');
              handleDownloadError('Download Failed: Server timed out.', true);
            }
          }, 5000);

          const handleDownloadError = (message, showRetry = false) => {
            downloadButton.textContent = message;
            spinner.style.display = 'none';
            if (showRetry) {
              downloadButton.style.display = 'none';
              retryButton.style.display = 'inline-block';
              retryButton.disabled = false;
            } else {
              downloadButton.disabled = false;
            }
          };

          // Retry Logic
          const initiateDownloadWithRetries = async (downloadFunc) => {
            let retries = 2;
            const attemptDownload = async () => {
              try {
                await downloadFunc();
              } catch (err) {
                console.error(`Download attempt failed: ${err}. Retries left: ${retries}`);
                if (retries > 0) {
                  retries--;
                  await new Promise(resolve => setTimeout(resolve, 1000)); // 1-second delay
                  await attemptDownload();
                } else {
                  handleDownloadError("Download Failed: Max retries exceeded", true);
                }
              }
            };
            await attemptDownload();
          };

          retryButton.addEventListener('click', () => {
            retryButton.style.display = 'none';
            downloadButton.style.display = 'inline-block';
            initiateDownloadWithRetries(sendDownloadRequest); // Retry with full logic
          });
          if (downloadButton.textContent !== 'Download Complete!') {
            downloadButton.textContent = 'Preparing Download...';
          }
        };

        retryButton.addEventListener('click', () => {
          retryButton.style.display = 'none'; // Hide retry
          downloadButton.style.display = 'inline-block'; // Show main button
          sendDownloadRequest(); // Retry download
        });

      }
    });

    const handleDownloadError = (message, showRetry = false) => {
      downloadButton.textContent = message;
      spinner.style.display = 'none';
      if (showRetry) {
        downloadButton.style.display = 'none';
        retryButton.style.display = 'inline-block';
        retryButton.disabled = false;
      } else {
        downloadButton.disabled = false;
      }
    });

    // Find a suitable location to insert the button (e.g., near the track title)
    const titleElement = document.querySelector('h1[itemprop="name"]'); // Adjust selector as needed
    if (titleElement && titleElement.parentNode) {

      //Health check
      const healthCheck = async () => {
        try{
            await fetch('http://localhost:8080/');
        }catch (e){
            handleDownloadError("Server Unavailable");
        }};
      titleElement.parentNode.appendChild(buttonContainer);
    } else {
      console.error('Could not find track title element.');
      // Fallback: Add to the end of the body (for debugging/testing)
      document.body.appendChild(buttonContainer);
    }


  }
}

// Inject the button on page load
injectDownloadButton();
