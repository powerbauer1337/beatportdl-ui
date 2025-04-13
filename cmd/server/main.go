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

func processDownloadInternal(track map[string]interface{}) (map[string]interface{}, error) {
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

	if linkType != beatport.LinkTypeTrack {
    errMsg := fmt.Sprintf("Unsupported link type: %s", linkType)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}

	trackID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		errMsg := fmt.Sprintf("Invalid track ID: %s", id)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}

	b := beatport.New(nil, "", nil) // Assuming a basic Beatport client is sufficient here
	downloadInfo, err := b.DownloadTrack(trackID, "lossless") // Assuming a basic Beatport client
	if err != nil {
		errMsg := fmt.Sprintf("Error getting download URL: %v", err)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}

	if downloadInfo == nil || downloadInfo.Location == "" {
		errMsg := "Empty download URL"
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}

	// Download the file
	downloadURL := downloadInfo.Location
	log.Printf("Downloading from URL: %s", downloadURL)

	httpClient := &http.Client{}
	getResp, err := httpClient.Get(downloadURL)
	if err != nil {
		errMsg := fmt.Sprintf("Error during download: %v", err)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Download failed with status code: %d", getResp.StatusCode)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}

	// Store the file temporarily
	filename := fmt.Sprintf("%d.temp", trackID)
	tempDir := "/tmp" //In a real production app, make sure this directory exists and is writable.
	filePath := filepath.Join(tempDir, filename)

	outFile, err := os.Create(filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Error creating file: %v", err)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, getResp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("Error saving download: %v", err)
		resp["error"] = errMsg
		resp["status"] = "failed"
		return resp, fmt.Errorf(errMsg)
	}

	log.Printf("Download complete for track ID %d. File saved to %s", trackID, filePath)
	resp["status"] = "completed"
	resp["metadata"] = map[string]interface{}{ "filename": filename, "path": filePath }
	return resp, nil
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

    resp, err := processDownloadInternal(track)
	if err != nil {
		log.Printf("processDownloadInternal error: %v", err)
		// resp already contains error details, no need to modify it here.
	}

    downloadsMutex.Lock()
    defer downloadsMutex.Unlock()

    if resp["error"] != nil {
		if resp["metadata"] == nil {
			resp["metadata"] = make(map[string]interface{})
		}
		if err != nil {
			resp["metadata"].(map[string]interface{})["internal_error"] = err.Error() //Add stack trace or other debugging info here
		}
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