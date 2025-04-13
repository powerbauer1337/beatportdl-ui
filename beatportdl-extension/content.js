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
    downloadButton.id = 'beatportdl-download-button';

    // Basic styling to match Beatport's design (adjust as needed)
    downloadButton.style.backgroundColor = '#e20074'; /* Beatport pink */
    downloadButton.style.color = 'white';
    downloadButton.style.border = 'none';
    downloadButton.style.padding = '8px 15px';
    downloadButton.style.borderRadius = '4px';
    downloadButton.style.cursor = 'pointer';

    // Add click listener
    downloadButton.addEventListener('click', async () => {
      const trackInfo = extractTrackInfo();
      if (trackInfo) {
        // Change button text and disable
        downloadButton.textContent = 'Preparing Download...';
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
            // Display an error message to the user
            downloadButton.textContent = 'Download Failed: ' + error.message;
            downloadButton.disabled = false;
          }
        };

        sendDownloadRequest();

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
                const error = downloadStatus.Metadata?.error;

                if (status === 'downloading') {
                  downloadButton.textContent = 'Downloading...';
                } else if (status === 'completed') {
                  downloadButton.textContent = 'Download Complete!';
                  clearInterval(intervalId); // Stop polling
                  downloadButton.disabled = false;
                } else if (status === 'failed') {
                  downloadButton.textContent = `Download Failed: ${error || 'Unknown error'}`;
                  clearInterval(intervalId); // Stop polling
                  downloadButton.disabled = false;
                }
              }
            } catch (error) {
              console.error('Error polling for status:', error);
              downloadButton.textContent = 'Download Failed: ' + error.message;
              // Consider offering a retry instead of just disabling
              clearInterval(intervalId); // Stop polling
              downloadButton.disabled = false;
            }
          }, 2000); // Poll every 2 seconds

          // Initial status (in case of immediate failure)
          if (downloadButton.textContent !== 'Download Complete!') {
            downloadButton.textContent = 'Preparing Download...';
          }
        };
      }
    });

    // Find a suitable location to insert the button (e.g., near the track title)
    const titleElement = document.querySelector('h1[itemprop="name"]'); // Adjust selector as needed
    if (titleElement && titleElement.parentNode) {
      titleElement.parentNode.appendChild(downloadButton);
    } else {
      console.error('Could not find track title element.');
      // Fallback: Add to the end of the body (for debugging/testing)
      document.body.appendChild(downloadButton);
    }
  }
}

// Inject the button on page load
injectDownloadButton();
