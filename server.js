const http = require('http');
const fs = require('fs');
const { exec } = require('child_process');
const path = require('path');

// Konfigurierbarer Pfad zu beatportdl (mit Standardwert)
const beatportdlPath = process.env.BEATPORTDL_PATH || '/usr/local/bin/beatportdl';

// Log-Datei
const logFile = 'server.log';
const logStream = fs.createWriteStream(logFile, { flags: 'a' });

function log(message) {
  const timestamp = new Date().toISOString();
  const logMessage = `[${timestamp}] ${message}\n`;
  console.log(logMessage); // Ausgabe auf der Konsole
  logStream.write(logMessage); // Ausgabe in die Log-Datei
}

const server = http.createServer((req, res) => {
  if (req.method === 'POST' && req.url === '/download') { // Pfad für Download-Anfragen
    let body = '';
    req.on('data', chunk => {
      body += chunk.toString();
    });
    req.on('end', () => {
      try {
        const tracks = JSON.parse(body);
        const urls = tracks.map(track => track.url);
        const tempFile = path.join(__dirname, 'temp_urls.txt'); // Temporäre Datei im Server-Verzeichnis
        fs.writeFileSync(tempFile, urls.join('\n'));

        log(`Starte Download für ${urls.length} Tracks...`);
        exec(`${beatportdlPath} -f ${tempFile}`, (error, stdout, stderr) => {
          if (error) {
            log(`Fehler beim Ausführen von beatportdl: ${error}`);
            log(`beatportdl stderr: ${stderr}`);
            const errorMessage = stderr || 'Unbekannter Fehler beim Download.';
            res.statusCode = 500;
            res.setHeader('Content-Type', 'application/json');
            res.end(JSON.stringify({ success: false, message: errorMessage }));
            return;
          }

          log(`beatportdl stdout: ${stdout}`);
          log(`Download abgeschlossen.`);
          res.statusCode = 200;
          res.setHeader('Content-Type', 'application/json');
          res.end(JSON.stringify({ success: true, message: 'Download gestartet.' }));
        });
      } catch (e) {
        log(`Fehler bei der Verarbeitung der Anfrage: ${e}`);
        res.statusCode = 400;
        res.setHeader('Content-Type', 'application/json');
        res.end(JSON.stringify({ success: false, message: 'Ungültige Anfrage.' }));
      }
    });
  } else {
    log(`Ungültige Anfrage: ${req.method} ${req.url}`);
    res.statusCode = 405;
    res.setHeader('Content-Type', 'application/json');
    res.end(JSON.stringify({ success: false, message: 'Methode nicht erlaubt.' }));
  }
});

const port = 3000;
server.listen(port, () => {
  log(`Server läuft auf Port ${port}`);
});

// Sicherstellen, dass der Log-Stream beim Beenden des Prozesses geschlossen wird
process.on('exit', () => {
  logStream.end();
});