// content.js

// Function to inject checkboxes and a button into the Beatport page
function injectControls() {
  if (document.getElementById('beatportdl-copy-button')) return; // Verhindere doppelte Einbindung

  const trackElements = document.querySelectorAll('li.bucket-item.track');
  if (trackElements.length === 0) {
    console.log('No tracks found on this page.');
    return;
  }

  // Use a Set to store selected track URLs and avoid duplicates
  const selectedTracks = new Set();
  const trackInfoMap = new Map(); // Neues Map-Objekt für Track-Informationen

  // Add checkboxes to each track
  trackElements.forEach((trackElement, index) => {
    if (trackElement.querySelector('.beatportdl-checkbox')) return; // Checkbox schon vorhanden?

    const titleElement = trackElement.querySelector('[data-testid="track-title"]') ||
                       trackElement.querySelector('a[data-qa-id="track-title"]') ||
                       trackElement.querySelector('a span');
    const artistElement = trackElement.querySelector('[data-testid="track-artist"]') ||
                          trackElement.querySelector('a[data-qa-id="artist-name"]') ||
                          trackElement.querySelector('div a');
    const linkElement = trackElement.querySelector('a[href*="/track/"]');

    if (titleElement && artistElement && linkElement) {
      const trackTitle = titleElement.textContent.trim();
      const artist = artistElement.textContent.trim();
      const url = linkElement.href;

      const checkbox = document.createElement('input');
      checkbox.type = 'checkbox';
      checkbox.id = 'track-checkbox-' + index;
      checkbox.dataset.url = url;
      checkbox.style.marginRight = '5px';
      checkbox.classList.add('beatportdl-checkbox'); // Checkbox markieren

      checkbox.addEventListener('change', (event) => {
        const url = event.target.dataset.url;
        if (event.target.checked) {
          selectedTracks.add(url);
          const trackInfo = extractTrackInfo(trackElement); // Track-Info extrahieren
          if (trackInfo) {
            trackInfoMap.set(url, trackInfo); // In Map speichern
          }
        } else {
          selectedTracks.delete(url);
          trackInfoMap.delete(url); // Aus Map entfernen
        }
      });

      // Insert checkbox before the track title
      titleElement.parentNode.insertBefore(checkbox, titleElement);
    }
  });

  // Create and add the "Copy URLs" button
  const copyButton = document.createElement('button');
  copyButton.textContent = 'Copy Selected URLs';
  copyButton.style.marginTop = '10px';
  copyButton.id = 'beatportdl-copy-button';
  copyButton.addEventListener('click', () => {
    if (selectedTracks.size > 0) {
      //  const urls = Array.from(selectedTracks).join('\n'); // Bisher: Nur URLs
      const tracksWithInfo = Array.from(selectedTracks).map(url => ({ url, ...trackInfoMap.get(url) })); // Neue Datenstruktur
      const jsonData = JSON.stringify(tracksWithInfo, null, 2); // JSON-String mit Formatierung
      navigator.clipboard.writeText(jsonData) // JSON-Daten kopieren
        .then(() => {
          console.log('Track data copied to clipboard!');
          alert('Track data copied to clipboard!');
        })
        .catch(err => {
          console.error('Error copying track data:', err);
          alert('Error copying track data: ' + err);
        });
    } else {
      console.log('No tracks selected.');
      alert('No tracks selected.');
    }
  });
  // Create and add the "Download" button
  const downloadButton = document.createElement('button');
  downloadButton.textContent = 'Download';
  downloadButton.style.marginTop = '10px';
  downloadButton.id = 'beatportdl-download-button';
  downloadButton.addEventListener('click', () => {
    if (selectedTracks.size > 0) {
      const tracksWithInfo = Array.from(selectedTracks).map(url => ({ url, ...trackInfoMap.get(url) }));
      sendToLocalServer(tracksWithInfo); // Daten an den Server senden
    } else {
      console.log('No tracks selected.');
      alert('No tracks selected.');
    }
  });

  // Insert the "Download" button next to the "Copy URLs" button
  const copyButton = document.getElementById('beatportdl-copy-button');
  if (copyButton && copyButton.parentNode) {
    copyButton.parentNode.insertBefore(downloadButton, copyButton.nextSibling);
  } else {
    console.error('Could not find copy button.');
    // Fallback: Add to the end of the body
    document.body.appendChild(downloadButton);
  }
}

  // Function to send data to the local server
  function sendToLocalServer(tracks) {
    fetch('http://localhost:3000/download', { // Angepasste URL
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(tracks)
    })
    .then(response => response.json()) // Antwort als JSON parsen
    .then(data => {
      if (data.success) {
        alert(data.message); // Erfolgsmeldung aus der Serverantwort
      } else {
        alert(`Fehler: ${data.message}`); // Fehlermeldung aus der Serverantwort
      }
    })
    .catch(error => {
      alert('Server nicht erreichbar: ' + error);
    });
  }


// Use MutationObserver with debouncing and observe only the track list container
let injectTimeout;
const observer = new MutationObserver(() => {
  clearTimeout(injectTimeout);
  injectTimeout = setTimeout(() => {
    injectControls();
  }, 500);
});

const container = document.querySelector('ul.bucket-items');
if (container) {
  observer.observe(container, { childList: true, subtree: true });
}

// Initial injection
injectControls();

console.log('Beatport Download Helper content script loaded.');

function extractTrackInfo(trackElement) { // Übergabe des Track-Elements
  try {
    const track = {};

    // JSON-Extraktion (angepasst auf das spezifische Track-Element)
    let data;
    const scripts = trackElement.querySelectorAll('script'); // Nur Skripte innerhalb des Track-Elements
    for (const script of scripts) {
      if (script.textContent.includes('window.Playables')) {
        const jsonStr = script.textContent.match(/window.Playabless*=s*({.*?});/)?.[1];
        if (jsonStr) {
          data = JSON.parse(jsonStr);
          track.id = data?.tracks?.[0]?.id;
          track.title = data?.tracks?.[0]?.name;
          track.artists = data?.tracks?.[0]?.artists?.map(a => a.name).join(', ');
        }
        break;
      }
    }

    // DOM-Extraktion (Fallback, ebenfalls angepasst)
    if (!data || !track.title || !track.artists || !track.id) {
      track.title = track.title || trackElement.querySelector('.ec-item__title')?.textContent.trim();
      track.artists = track.artists || trackElement.querySelector('.ec-item__artists')?.textContent.trim();
      track.id = track.id || trackElement.querySelector('[data-ec-id]')?.dataset.ecId;
    }

    return track.id && track.title && track.artists ? track : null;
  } catch (error) {
    console.error('Fehler bei der Track-Extraktion:', error);
    return null;
  }
}