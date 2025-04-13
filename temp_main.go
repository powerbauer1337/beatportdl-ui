package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v2"

	"github.com/aashutoshrathi/beatportdl/internal/beatport"
	"github.com/aashutoshrathi/beatportdl/config"
)

type DownloadStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

var downloads []DownloadStatus

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	for _, urlStr := range req.URLs {
		go func(urlStr string) {
			status := DownloadStatus{Status: "in progress"}
			downloads = append(downloads, status)
			url, err := beatport.ParseUrl(urlStr)
			if err != nil {
				status.Status = "error"
				status.Error = fmt.Sprintf("Invalid URL: %v", err)
				log.Printf("Error parsing URL %s: %v", urlStr, err)
				return
			}

			if err := beatport.Download(url); err != nil { // Assuming a Download function exists
				status.Status = "error"
				status.Error = fmt.Sprintf("Download failed: %v", err)
				log.Printf("Error downloading from %s: %v", urlStr, err)
				return
			}
			status.Status = "completed"
			log.Printf("Successfully downloaded from %s", urlStr)
		}(urlStr)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "downloads started"})
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(downloads)
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.LoadConfig() // Assuming LoadConfig exists
	if err != nil {
		http.Error(w, fmt.Sprintf("Error loading config: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cfg)
	} else if r.Method == http.MethodPut || r.Method == http.MethodPost {
		var newCfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
			http.Error(w, fmt.Sprintf("Invalid config data: %v", err), http.StatusBadRequest)
			return
		}
		// Merge new config with existing one, or replace entirely (depending on desired behavior)
		// Example merge:  (Adapt based on your config structure)
		// if newCfg.DownloadQuality != "" { cfg.DownloadQuality = newCfg.DownloadQuality }
		cfg = newCfg // Example replace
		if err := config.SaveConfig(cfg); err != nil { // Assuming SaveConfig exists
			http.Error(w, fmt.Sprintf("Error saving config: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "config updated"})
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	serverMode := flag.Bool("server", false, "Start in server mode")
	flag.Parse()

	if *serverMode {
		http.HandleFunc("/download", downloadHandler)
		http.HandleFunc("/status", statusHandler)
		http.HandleFunc("/config", configHandler)
		fmt.Println("BeatportDL server started. Listening on port 8080...")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	} else {
		// Run in command-line mode (existing BeatportDL logic)
		//  Replace this with the actual command-line execution logic of BeatportDL
		fmt.Println("Running BeatportDL in command-line mode...")
		// Example:
		// cfg, err := config.LoadConfig()
		// if err != nil {
		//  log.Fatalf("Error loading config: %v", err)
		// }

		cfg, err := config.LoadConfig()
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}

		if err := beatport.Run(cfg, flag.Args()); err != nil {
			log.Fatalf("Error during command-line execution: %v", err)
		}

		fmt.Println("Command-line execution finished.")
	}
}