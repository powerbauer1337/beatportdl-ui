package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/unspok3n/beatportdl-ui/config"
)

type Track struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
	URL    string `json:"url"`
}

func NewServerError(code int, message string) *ServerError {
	return &ServerError{
		Code:    code,
		Message: message,
	}
}

func Start(cfg *config.AppConfig) error {
	http.HandleFunc("/download", downloadHandler(cfg))
	return nil
}

func downloadHandler(cfg *config.AppConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var tracks []Track
		if err := json.NewDecoder(r.Body).Decode(&tracks); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		for _, track := range tracks {
			fmt.Printf("Received track: %s - %s (%s)\n", track.Title, track.Artist, track.URL)

			// TODO: Adapt download logic from main.go here
			//  This is a placeholder.  The actual download process needs to be
			//  integrated, handling errors, progress, and using the configuration.

			// Example (incomplete):
			// bp := beatport.New(cfg.Auth.Username, cfg.Auth.Password)
			// if err := bp.Login(); err != nil { ... }
			// trackInfo, err := bp.GetTrack(track.URL) // Adapt to use track.URL
			// if err != nil { ... }
			// downloader := beatport.Downloader{Config: cfg, Beatport: bp}
			// if err := downloader.DownloadTrack(trackInfo); err != nil { ... }
			// tagger := taglib.Tagger{Config: cfg}
			// if err := tagger.TagFile( ... ); err != nil { ... }  // Path to downloaded file needed

			// Dummy response for now:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Download started for: " + track.Title))
		}
	}
}
