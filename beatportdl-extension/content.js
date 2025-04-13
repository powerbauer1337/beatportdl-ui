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

    // Add click listener (implementation will be in later steps)
    downloadButton.addEventListener('click', () => {
      // Placeholder for download logic
      const trackInfo = extractTrackInfo();
      if (trackInfo) {
        console.log('Track information:', trackInfo);
      } else {
        console.error('Failed to extract track information.');
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
