// cmd/server/main.go
package main

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"

	"unspok3n/beatportdl/config"
	"unspok3n/beatportdl/internal/beatport"
	"unspok3n/beatportdl/internal/server"
)

type Config struct {
	MaxConcurrentDownloads int `yaml:"max_concurrent_downloads"`
}

type downloadStatus struct {
	TrackURL string                 `json:"track_url"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

var (
	downloads         = make(map[string]*downloadStatus)
	downloadsMutex    = &sync.Mutex{}
	cfg               Config
	downloadSemaphore chan struct{}
)

func main() {
	http.HandleFunc("/download", downloadHandler)
	http.HandleFunc("/config", configureHandler)
	http.HandleFunc("/status", statusHandler)

	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

func init() {
	cfg, err := config.Parse("./config.yml")
	if err != nil {
		log.Printf("Error reading config: %v. Using default values.", err)
		cfg = &config.AppConfig{
			MaxGlobalWorkers:   5,
			MaxDownloadWorkers: 5,
		}
		if err := cfg.Save("./config.yml"); err != nil {
			log.Printf("Failed to write default config: %v", err)
		}
	}
	downloadSemaphore = make(chan struct{}, cfg.MaxDownloadWorkers)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
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
		trackURL, urlOK := track["url"].(string)
		if !urlOK {
			errorMessages = append(errorMessages, "Track: missing or invalid 'url'")
			continue
		}

		parsedURL, err := url.Parse(trackURL)
		if err != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("Track: invalid URL format: %v", err))
			continue
		}
		if parsedURL.Scheme != "https" || parsedURL.Host != "www.beatport.com" || !regexp.MustCompile(`^/(track|release)/`).MatchString(parsedURL.Path) {
			errorMessages = append(errorMessages, "Track: invalid Beatport URL: scheme must be 'https', host must be 'www.beatport.com', and path must start with '/track/' or '/release/'")
			continue
		}

		trackID, idOK := track["id"].(string)
		if !idOK {
			errorMessages = append(errorMessages, "Track: missing or invalid 'id'")
			continue
		}

		title, titleOK := track["title"].(string)
		if !titleOK {
			errorMessages = append(errorMessages, fmt.Sprintf("Track with id '%s': missing or invalid 'title'", trackID))
			continue
		}

		artists, artistsOK := track["artists"].(string)
		if !artistsOK {
			errorMessages = append(errorMessages, fmt.Sprintf("Track with id '%s': missing or invalid 'artists'", trackID))
			continue
		}

		downloadsMutex.Lock()
		downloads[id] = &downloadStatus{
			TrackURL: parsedURL.String(),
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

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Download(s) initiated"})
}

func processDownloadInternal(track map[string]interface{}) (*map[string]interface{}, error) {
	resp := &map[string]interface{}{
		"track":  track,
		"status": "downloading",
	}
	trackURL, ok := track["url"].(string)
	if !ok {
		return resp, server.NewServerError(http.StatusBadRequest, "Invalid or missing URL in track data")
	}

	link, err := beatport.ParseUrl(trackURL)
	if err != nil {
		return resp, server.NewServerError(http.StatusBadRequest, fmt.Sprintf("Error parsing URL '%s': %v", trackURL, err))
	}

	log.Printf("Downloading %s with ID %d", link.Type, link.ID)

	if link.Type != beatport.TrackLink {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusBadRequest, fmt.Sprintf("Unsupported link type: %s", link.Type))
	}

	b := beatport.New(link.Store, "", nil)
	trackInfo, err := b.GetTrack(link.ID)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error getting track info: %v", err))
	}

	downloadInfo, err := b.DownloadTrack(link.ID, "lossless")
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error getting download URL: %v", err))
	}

	if downloadInfo == nil || downloadInfo.Location == "" {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, "Empty download URL")
	}

	downloadURL := downloadInfo.Location
	log.Printf("Downloading from URL: %s", downloadURL)
	httpClient := &http.Client{}
	getResp, err := httpClient.Get(downloadURL)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error during download: %v", err))
	}
	defer getResp.Body.Close()

	if getResp.StatusCode != http.StatusOK {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(getResp.StatusCode, fmt.Sprintf("Download failed with status code: %d", getResp.StatusCode))
	}

	filename := fmt.Sprintf("%d.temp", trackInfo.ID)
	tempDir := "/tmp"
	filePath := filepath.Join(tempDir, filename)
	outFile, err := os.Create(filePath)
	if err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error creating file: %v", err))
	}

	contentLength := getResp.ContentLength
	if contentLength <= 0 {
		log.Printf("Content length is not available for %s, progress updates will not be provided.", downloadURL)
	}

	downloadedBytes := int64(0)
	lastReportedPercent := 0
	buf := make([]byte, 32*1024)
	for {
		n, err := getResp.Body.Read(buf)
		if n > 0 {
			written, err := outFile.Write(buf[:n])
			if err != nil || written < n {
				if err == nil {
					err = fmt.Errorf("short write: wrote %d, expected %d", written, n)
				}
				(*resp)["status"] = "failed"
				return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error writing to file: %v", err))
			}
			downloadedBytes += int64(n)

			if contentLength > 0 {
				percent := int(float64(downloadedBytes) / float64(contentLength) * 100)
				if percent-lastReportedPercent >= 10 {
					lastReportedPercent = percent
					log.Printf("Download progress: %d%%", percent)
					(*resp)["progress"] = percent
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			(*resp)["status"] = "failed"
			return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error during download: %v", err))
		}
	}
	defer outFile.Close()

	log.Printf("Download complete for track ID %d. File temporarily saved to %s", trackInfo.ID, filePath)

	newFilename := fmt.Sprintf("%s - %s.mp3", trackInfo.Artists.Display(100, ""), trackInfo.Name)
	newFilename = regexp.MustCompile(`[^a-zA-Z0-9\s\-_.,()\[\]{}]`).ReplaceAllString(newFilename, "")

	if newFilename == "" || len(newFilename) > 255 {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Invalid filename generated: '%s'", newFilename))
	}

	downloadDir := "./downloads"
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		if err := os.MkdirAll(downloadDir, 0755); err != nil {
			(*resp)["status"] = "failed"
			return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error creating downloads directory: %v", err))
		}
	}
	newFilePath := filepath.Join(downloadDir, newFilename)

	if err := os.Rename(filePath, newFilePath); err != nil {
		(*resp)["status"] = "failed"
		return resp, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error renaming and moving file: %v", err))
	}
	log.Printf("Renamed and moved file to: %s", newFilePath)

	if err := os.Remove(filePath); err != nil {
		log.Printf("Error deleting temporary file %s: %v", filePath, err)
	}

	log.Printf("File tagging would be implemented here for: %s", newFilePath)

	resp["status"] = "completed"
	resp["metadata"] = map[string]interface{}{"filename": newFilename, "path": newFilePath}
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
		if serverErr, ok := err.(*server.ServerError); ok {
			status.Metadata["Code"] = serverErr.Code
			status.Metadata["error"] = serverErr.Message
		} else {
			status.Metadata["error"] = err.Error()
		}
		if resp != nil && resp["metadata"] != nil {
			if _, ok := resp["metadata"].(map[string]interface{})["internal_error"]; !ok {
				resp["metadata"].(map[string]interface{})["internal_error"] = err.Error()
			}
		} else if resp != nil {
			resp["metadata"] = map[string]interface{}{"internal_error": err.Error()}
		} else {
			resp = &map[string]interface{}{"metadata": map[string]interface{}{"internal_error": err.Error()}}
		}
		status.Status = "failed"
		if resp != nil {
			status.Metadata["error"] = resp["metadata"]
		}
		log.Printf("Download failed for %s: %v", status.TrackURL, status.Metadata["error"])
	} else {
		status.Status = "completed"
		if resp != nil {
			status.Metadata = resp["metadata"].(map[string]interface{})
		}
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
	case http.MethodPut, http.MethodPost:
		resp, err := updateConfig(r)
		if err != nil {
			code := http.StatusInternalServerError
			if serverErr, ok := err.(*server.ServerError); ok {
				code = serverErr.Code
			}
			w.WriteHeader(code)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		log.Println("Configuration updated successfully")
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

func updateConfig(r *http.Request) (interface{}, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading config request body: %v", err)
		return nil, server.NewServerError(http.StatusBadRequest, "Error reading request body")
	}
	defer r.Body.Close()

	var newCfg config.AppConfig
	if err := yaml.Unmarshal(body, &newCfg); err != nil {
		return nil, server.NewServerError(http.StatusBadRequest, fmt.Sprintf("Error parsing config: %v", err))
	}

	if newCfg.MaxDownloadWorkers <= 0 {
		return nil, server.NewServerError(http.StatusBadRequest, "Invalid max_concurrent_downloads value")
	}

	cfg.MaxDownloadWorkers = newCfg.MaxDownloadWorkers
	downloadSemaphore = make(chan struct{}, cfg.MaxDownloadWorkers)

	if err := cfg.Save("./config.yml"); err != nil {
		return nil, server.NewServerError(http.StatusInternalServerError, fmt.Sprintf("Error writing config: %v", err))
	}

	return map[string]string{"status": "config updated"}, nil
}