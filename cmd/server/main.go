// cmd/server/main.go
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/your-username/your-repository/internal/beatport" // Replace with the actual path to the beatport package
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, World!")
	})
	http.HandleFunc("/download", downloadHandler)

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

	results := make([]downloadResult, len(data.URLs))
	for i, url := range data.URLs {
		results[i] = processDownload(url)
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

type downloadResult struct {
	URL    string `json:"url"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func processDownload(url string) downloadResult {
	result := downloadResult{URL: url}
	linkType, id, err := beatport.ParseUrl(url)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("Error parsing URL: %v", err)
		return result
	}

	// Implement the actual download logic here, using existing BeatportDL functions
	// Example (replace with actual logic):
	switch linkType {
	case beatport.LinkTypeTrack:
		result.Status = "started"
		// Call BeatportDL's DownloadTrack function here, e.g.:
		// if err := beatport.DownloadTrack(id); err != nil {
		// 	result.Status = "failed"
		// 	result.Error = fmt.Sprintf("Error downloading track: %v", err)
		// }
	default:
		result.Status = "failed"
		result.Error = fmt.Sprintf("Unsupported link type: %s", linkType)
	}

	return result
}

func hasErrors(results []downloadResult) bool {
	for _, result := range results {
		if result.Status == "failed" {
			return true
		}
	}
	return false
}