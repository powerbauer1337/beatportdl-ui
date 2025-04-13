// cmd/server/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"sync"
	"regexp"
	"github.com/google/uuid"

	"github.com/your-username/your-repository/internal/beatport" // Replace with the actual path to the beatport package
	"github.com/your-username/your-repository/config"        // Replace with the actual path to the config package
	"gopkg.in/yaml.v2"
)

type Config struct {
	MaxConcurrentDownloads int `yaml:"max_concurrent_downloads"`
}

type downloadStatus struct {
	TrackURL string                 `json:"track_url"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata,omitempty"` // Include other track metadata
}

var (
	downloads        = make(map[string]*downloadStatus)
	downloadsMutex   = &sync.Mutex{}
	cfg              Config
	downloadSemaphore chan struct{}
)

func main() {
	http.HandleFunc("/download", downloadHandler2)
	http.HandleFunc("/config", configureHandler)
	http.HandleFunc("/status", statusHandler)

	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func init() {
	// Load configuration, or use default if not found.
	cfg, err := config.ReadConfig()
	if err != nil {
		log.Printf("Error reading config: %v. Using default values.", err)
		cfg.MaxConcurrentDownloads = 5 // Default value
		// If config file doesn't exist, attempt to create one with default values.
		if err := config.WriteConfig(cfg); err != nil {
			log.Printf("Failed to write default config: %v", err)
		}
	}
	downloadSemaphore = make(chan struct{}, cfg.MaxConcurrentDownloads)
}

func downloadHandler2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, fmt.Sprintf("Error reading request body: %v", err), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    var data struct {
        Tracks []map[string]interface{} `json:"tracks"`
    }
    if err := json.Unmarshal(body, &data); err != nil {
        http.Error(w, fmt.Sprintf("Error parsing JSON: %v", err), http.StatusBadRequest)
        return
    }

    if len(data.Tracks) == 0 {
        http.Error(w, "No tracks provided", http.StatusBadRequest)
        return
    }

	// Basic Beatport URL pattern check (improve as needed)
	beatportURLPattern := regexp.MustCompile(`^https://www\.beatport\.com/(track|release)/[^/]+/[^/]+$`)

	responses := make([]map[string]interface{}, 0, len(data.Tracks)) // Initialize with capacity
    for _, track := range data.Tracks {
        id := uuid.New().String()
		// Extract URL
		url, urlOK := track["url"].(string)
		if !urlOK {
			responses = append(responses, map[string]interface{}{"error": "missing or invalid 'url' in track data"})
			continue
		}

		// Validate URL format
		if !beatportURLPattern.MatchString(url) {
			responses = append(responses, map[string]interface{}{"error": fmt.Sprintf("invalid Beatport URL format: %s", url)})
			continue
		}

		// Extract other metadata with type checking
		trackID, idOK := track["id"].(string)
		if !idOK {
			responses = append(responses, map[string]interface{}{"error": "missing or invalid 'id' in track data"})
			continue
		}

		title, titleOK := track["title"].(string)
		if !titleOK {
			responses = append(responses, map[string]interface{}{"error": "missing or invalid 'title' in track data"})
			continue
		}

		artists, artistsOK := track["artists"].(string)
		if !artistsOK {
			responses = append(responses, map[string]interface{}{"error": "missing or invalid 'artists' in track data"})
			continue
		}

		downloadsMutex.Lock()
		downloads[id] = &downloadStatus{
			TrackURL: url,
			Status:   "pending",
			Metadata: map[string]interface{}{
				"id":      trackID,
				"title":   title,
				"artists": artists,
			},
		}
		downloadsMutex.Unlock()

        downloadSemaphore <- struct{}{}
        go processDownload(id, track)
	}
	
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted) // Use 202 Accepted for async processing
    json.NewEncoder(w).Encode(map[string]string{"message": "Download(s) initiated"})
}

func processDownloadInternal(track map[string]interface{}) map[string]interface{} {
	resp := make(map[string]interface{})
	resp["track"] = track // Include the full track info in the response
	url, ok := track["url"].(string)
	if !ok {
		return map[string]interface{}{"error": "invalid or missing url in track data"}
	}

	linkType, id, err := beatport.ParseUrl(url)
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing URL '%s': %v", url, err)
		resp["error"] = errMsg
		return resp
	}

	resp["status"] = "downloading" // Use string constants for statuses
	log.Printf("Downloading %s with ID %s", linkType, id)

	// Replace this with your actual download logic based on linkType
	var downloadErr error
	switch linkType {
	case beatport.LinkTypeTrack:
		downloadErr = beatport.DownloadTrack(id) // Assuming beatport.DownloadTrack expects just the ID
	default:
		downloadErr = fmt.Errorf("unsupported link type '%s'", linkType)
	}
	if downloadErr != nil {
		errMsg := fmt.Sprintf("Error during download: %v", downloadErr)
		resp["error"] = errMsg
		resp["status"] = "failed" // Use string constants for statuses
	} else {
		resp["status"] = "completed" // Use string constants for statuses
	}
	return resp
}

func processDownload(downloadID string, track map[string]interface{}) {
    defer func() { <-downloadSemaphore }()

    downloadsMutex.Lock()
    status := downloads[downloadID]
    if status == nil {
        log.Printf("Warning: Download status not found for ID: %s", downloadID)
        downloadsMutex.Unlock()
        return
    }
    status.Status = "downloading"
    downloadsMutex.Unlock()

    resp := processDownloadInternal(track)

    downloadsMutex.Lock()
    defer downloadsMutex.Unlock()

    if resp["error"] != nil {
        status.Status = "failed"
		status.Metadata["error"] = resp["error"]
        log.Printf("Download failed for %s: %v", status.TrackURL, resp["error"])
    } else {
        status.Status = "completed"
        log.Printf("Download completed for %s", status.TrackURL)
    }
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
    }
    downloadsMutex.Lock()
    defer downloadsMutex.Unlock()

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(downloads); err != nil {
		log.Printf("Error encoding status response: %v", err)
		http.Error(w, "Error encoding status", http.StatusInternalServerError)
		return
	}
	log.Println("Returned download status")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		log.Printf("Error encoding config response: %v", err)
		http.Error(w, "Error encoding config", http.StatusInternalServerError)
		return
	}
	log.Println("Returned current configuration")
}

func configureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getConfig(w, r)
	}
	// Add PUT/POST handling if needed for updating config.
}

func updateConfig(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading config request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var newCfg Config
	if err := yaml.Unmarshal(body, &newCfg); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing config: %v", err), http.StatusBadRequest)
		return
	}

	if newCfg.MaxConcurrentDownloads <= 0 {
		http.Error(w, "Invalid max_concurrent_downloads value", http.StatusBadRequest)
		return
	}

	cfg.MaxConcurrentDownloads = newCfg.MaxConcurrentDownloads
	downloadSemaphore = make(chan struct{}, cfg.MaxConcurrentDownloads) //Reinitialize semaphore

	// Save the updated configuration (assuming config.WriteConfig can handle the new struct)
	if err := config.WriteConfig(cfg); err != nil {
		http.Error(w, fmt.Sprintf("Error writing config: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "config updated"})
	log.Println("Configuration updated successfully")
}