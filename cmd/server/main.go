// cmd/server/main.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"io/ioutil"
	"net/http"
	"sync"
	"os"
	"path/filepath"
	"regexp"
	"net/url"
	"github.com/google/uuid"

	"github.com/your-username/your-repository/internal/server"
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

	errorMessages := make([]string, 0)
    for _, track := range data.Tracks {
        id := uuid.New().String()
		// Extract URL
		url, urlOK := track["url"].(string)
		if !urlOK {
			errorMessages = append(errorMessages, fmt.Sprintf("Track: missing or invalid 'url'"))
			continue
		}

		// Validate and sanitize URL
		parsedURL, err := url.Parse(url)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Track: invalid URL format: %v", err))
			continue
		}
		if parsedURL.Scheme != "https" || parsedURL.Host != "www.beatport.com" || (!regexp.MustCompile(`^/(track|release)/`).MatchString(parsedURL.Path)) {
			errorMessages = append(errorMessages, fmt.Sprintf("Track: invalid Beatport URL: scheme must be 'https', host must be 'www.beatport.com', and path must start with '/track/' or '/release/'"))
			continue
		}

		// Extract and sanitize other metadata with type checking
		trackID, idOK := track["id"].(string) // Assuming ID is a string
		if !idOK {
			errorMessages = append(errorMessages, fmt.Sprintf("Track: missing or invalid 'id'"))
			continue
		}

		title, titleOK := track["title"].(string)
		if !titleOK {
			errorMessages = append(errorMessages, fmt.Sprintf("Track with id '%s': missing or invalid 'title'", trackID)) // Use trackID if available
			continue
		}

		artists, artistsOK := track["artists"].(string)
		if !artistsOK {
			errorMessages = append(errorMessages, fmt.Sprintf("Track with id '%s': missing or invalid 'artists'", trackID)) // Use trackID if available
			continue
		}

		downloadsMutex.Lock()
		downloads[id] = &downloadStatus{
			TrackURL:  parsedURL.String(), // Store the parsed and potentially modified URL
			Status:   "pending",
			Metadata: map[string]interface{}{
				"id":      html.EscapeString(trackID),
				"title":   html.EscapeString(title),
				"artists": html.EscapeString(artists),
			},
		}
		downloadsMutex.Unlock()

        downloadSemaphore <- struct{}{}
        go processDownload(id, track)
	}
	
    w.Header().Set("Content-Type", "application/json")
	if len(errorMessages) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string][]string{"errors": errorMessages})
		return
	}

	w.WriteHeader(http.StatusAccepted) // Use 202 Accepted for async processing
	json.NewEncoder(w).Encode(map[string]string{"message": "Download(s) initiated"})
}

func processDownloadInternal(track map[string]interface{}) (*map[string]interface{}, error) {
	resp := &map[string]interface{}{
		"track":  track, // Include the full track info
		"status": "downloading",
	}
	url, ok := track["url"].(string)
	if !ok {
		return resp, server.NewServerError(400, "Invalid or missing URL in track data")
	}

	linkType, id, err := beatport.ParseUrl(url)
	if err != nil {
		return resp, server.NewServerError(400, fmt.Sprintf("Error parsing URL '%s': %v", url, err))
	}

	log.Printf("Downloading %s with ID %s", linkType, id)

	if linkType != beatport.LinkTypeTrack {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(400, fmt.Sprintf("Unsupported link type: %s", linkType))
	}

	trackID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(400, fmt.Sprintf("Invalid track ID: %s", id))
	}

	b := beatport.New(nil, "", nil) // Assuming a basic Beatport client is sufficient here
	downloadInfo, err := b.DownloadTrack(trackID, "lossless") // Assuming a basic Beatport client
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, fmt.Sprintf("Error getting download URL: %v", err))
	}

	if downloadInfo == nil || downloadInfo.Location == "" {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, "Empty download URL")
	}

	// Download the file
	downloadURL := downloadInfo.Location
	log.Printf("Downloading from URL: %s", downloadURL)
	httpClient := &http.Client{}
	getResp, err := httpClient.Get(downloadURL)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, fmt.Sprintf("Error during download: %v", err))
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(getResp.StatusCode, fmt.Sprintf("Download failed with status code: %d", getResp.StatusCode))
	}

	// Store the file temporarily
	filename := fmt.Sprintf("%d.temp", trackID)
	tempDir := "/tmp"
	filePath := filepath.Join(tempDir, filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, fmt.Sprintf("Error creating file: %v", err))
	}

	// Get total file size for progress reporting
	contentLength := getResp.ContentLength
	if contentLength <= 0 {
		log.Printf("Content length is not available for %s, progress updates will not be provided.", downloadURL)
	}

	// Track downloaded bytes and progress
	downloadedBytes := int64(0)
	lastReportedPercent := 0
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := getResp.Body.Read(buf)
		if n > 0 {
			written, err := outFile.Write(buf[:n])
			if err != nil || written < n {
				if err == nil {
					err = fmt.Errorf("short write: wrote %d, expected %d", written, n)
				}
				(*resp)["status"] = "failed"
				return resp, server.NewServerError(500, fmt.Sprintf("Error writing to file: %v", err))
			}
			downloadedBytes += int64(n)

			// Calculate and report progress
			if contentLength > 0 {
				percent := int(float64(downloadedBytes) / float64(contentLength) * 100)
				if percent-lastReportedPercent >= 10 {
					lastReportedPercent = percent
					log.Printf("Download progress: %d%%", percent)
					(*resp)["progress"] = percent // Update progress in response
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			(*resp)["status"] = "failed"
			return resp, server.NewServerError(500, fmt.Sprintf("Error during download: %v", err))
		}
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, getResp.Body)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, fmt.Sprintf("Error saving download: %v", err))
	}
	log.Printf("Download complete for track ID %d. File temporarily saved to %s", trackID, filePath)

	// --- Start Download Completion Actions ---

	// 1. Get metadata
	title, titleOK := track["title"].(string)
	artists, artistsOK := track["artists"].(string)
	if !titleOK || !artistsOK {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, "Missing or invalid track title or artists in metadata")
	}

	// 2. Construct new filename
	newFilename := fmt.Sprintf("%s - %s.mp3", artists, title)
	// Basic filename sanitization (replace invalid characters) - improve as needed
	newFilename = regexp.MustCompile(`[^a-zA-Z0-9\s\-_.,()\[\]{}]`).ReplaceAllString(newFilename, "")

	if newFilename == "" || len(newFilename) > 255 { // Basic check for filename validity
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, fmt.Sprintf("Invalid filename generated: '%s'", newFilename))
	}

	// 3. Determine destination path (using a default for now)
	downloadDir := "./downloads" // In a real app, get this from config
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		os.MkdirAll(downloadDir, 0755) //Create the directory if it doesn't exist. Handle potential error in production.
	}
	newFilePath := filepath.Join(downloadDir, newFilename)

	// 4. Rename and move the file
	err = os.Rename(filePath, newFilePath)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(500, fmt.Sprintf("Error renaming and moving file: %v", err))
	}
	log.Printf("Renamed and moved file to: %s", newFilePath)

	// 5. Delete temporary file
	err = os.Remove(filePath)
	if err != nil {
		log.Printf("Error deleting temporary file %s: %v", filePath, err) // Log but don't fail the download
	}

	// 6. Tag the file (Placeholder - implement actual tagging logic)
	log.Printf("File tagging would be implemented here for: %s", newFilePath)

	// --- End Download Completion Actions ---

	resp["status"] = "completed" // Update status after all actions are complete
	resp["metadata"] = map[string]interface{}{"filename": newFilename, "path": newFilePath} // Update metadata
	return resp, nil // Return successful response
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
		if serverErr, ok := err.(*server.ServerError); ok {
			status.Metadata["Code"] = serverErr.Code
			status.Metadata["error"] = serverErr.Message
		} else {
			status.Metadata["error"] = err.Error()
		}
		if resp != nil && resp["metadata"] != nil {
			if _, ok := resp["metadata"].(map[string]interface{})["internal_error"]; !ok {
				resp["metadata"].(map[string]interface{})["internal_error"] = err.Error() // Add stack trace or other debugging info here if not already present
			}
		} else if resp != nil {
			resp["metadata"] = map[string]interface{}{"internal_error": err.Error()}
		} else {
			resp = &map[string]interface{}{"metadata": map[string]interface{}{"internal_error": err.Error()}}
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
		serverErr := server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error encoding status: %v", err))
		w.WriteHeader(serverErr.Code)
		json.NewEncoder(w).Encode(map[string]string{"error": serverErr.Message})
		return

	}
	log.Println("Returned download status")
}

func configureHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet: 
		getConfig(w, r)
	case http.MethodPut, http.MethodPost:{
			code := http.StatusInternalServerError
			if serverErr, ok := err.(*server.ServerError); ok {
				code = serverErr.Code
			}
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "config updated"})
		log.Println("Configuration updated successfully")}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}


func updateConfig(r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading config request body: %v", err)
		return nil, server.NewServerError(http.StatusBadRequest, "Error reading request body")
	}
	defer r.Body.Close()

	var newCfg Config
	if err := yaml.Unmarshal(body, &newCfg); err != nil {
		http.Error(w, fmt.Sprintf("Error parsing config: %v", err), http.StatusBadRequest)
		return nil, server.NewServerError(http.StatusBadRequest, fmt.Sprintf("Error parsing config: %v", err))
	}

	if newCfg.MaxConcurrentDownloads <= 0 {
		return nil, server.NewServerError(http.StatusBadRequest, "Invalid max_concurrent_downloads value")
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