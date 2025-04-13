// cmd/server/main.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/your-username/your-repository/internal/beatport" // Replace with the actual path to the beatport package
	"github.com/your-username/your-repository/config"        // Replace with the actual path to the config package
	"gopkg.in/yaml.v2"
)

var downloads = make(map[string]*downloadStatus)
var downloadsMutex = &sync.Mutex{}

func main() {
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/config", configHandler)

	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var data struct {
		URLs []string `json:"urls"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	if len(data.URLs) == 0 {
		http.Error(w, "No URLs provided", http.StatusBadRequest)
		return
	}

	results := make([]*downloadStatus, len(data.URLs))
	for i, url := range data.URLs {
		status := &downloadStatus{URL: url, Status: "pending"}
		addDownload(status)
		results[i] = status
		go processDownload(status)
	}

	w.Header().Set("Content-Type", "application/json")
	if hasErrors(results) {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Println("Error encoding JSON response:", err)
	}

}

func addDownload(status *downloadStatus) {
	downloadsMutex.Lock()
	defer downloadsMutex.Unlock()
	downloads[status.URL] = status
}

func getDownload(url string) *downloadStatus {
	downloadsMutex.Lock()
	defer downloadsMutex.Unlock()
	return downloads[url]
}

func updateDownloadStatus(url string, status string, errorMsg string) {
	downloadsMutex.Lock()
	defer downloadsMutex.Unlock()
	if s, ok := downloads[url]; ok {
		s.Status = status
		s.Error = errorMsg
	}
}

type downloadStatus struct {
	URL    string `json:"url"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func processDownload(status *downloadStatus) {
	linkType, id, err := beatport.ParseUrl(url)
	if err != nil {
		updateDownloadStatus(status.URL, "failed", fmt.Sprintf("Error parsing URL: %v", err))
		return
	}

	// Implement the actual download logic here, using existing BeatportDL functions
	// Example (replace with actual logic):
	switch linkType {
	case beatport.LinkTypeTrack:
		updateDownloadStatus(status.URL, "downloading", "")
		err := beatport.DownloadTrack(id) // Replace with actual BeatportDL function
		if err != nil {
			updateDownloadStatus(status.URL, "failed", fmt.Sprintf("Error downloading track: %v", err))
			return
		}
		updateDownloadStatus(status.URL, "completed", "")
	default:
		updateDownloadStatus(status.URL, "failed", fmt.Sprintf("Unsupported link type: %s", linkType))
	}
}

func hasErrors(results []*downloadStatus) bool {
	for _, result := range results {
		if result.Status == "failed" {
			return true
		}
	}
	return false
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	downloadsMutex.Lock()
	defer downloadsMutex.Unlock()

	var statusResponse struct {
		Total     int                      `json:"total"`
		Downloads map[string]*downloadStatus `json:"downloads"`
	}

	statusResponse.Total = len(downloads)
	statusResponse.Downloads = downloads

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(statusResponse); err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getConfig(w, r)
	case http.MethodPut, http.MethodPost:
		updateConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.ReadConfig()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}

func updateConfig(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	var cfg config.Config
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing config: %v", err), http.StatusBadRequest)
		return
	}

	if err := config.WriteConfig(cfg); err != nil {
		http.Error(w, fmt.Sprintf("Error writing config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "config updated"}); err != nil {
		log.Println("Error encoding JSON response:", err)
	}
}