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
	http.HandleFunc("/status", statusHandler)

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
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received download request: %s", string(body))

	defer r.Body.Close()



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

	responses := make([]map[string]interface{}, 0, len(data.URLs))
	hasError := false

	for i, url := range data.URLs {
		go func(url string) {
			resp := processDownload(url)
			responses = append(responses, resp)
			if resp["error"] != nil {
				hasError = true
			}
		}(url)
	}

	w.Header().Set("Content-Type", "application/json")
	if hasError {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("Download request resulted in errors")
	} else {
		log.Println("Download request processed successfully")
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		log.Println("Error encoding JSON response:", err)
	}

}

func processDownload(url string) map[string]interface{} {
	resp := make(map[string]interface{})
	resp["url"] = url

	linkType, id, err := beatport.ParseUrl(url)
	if err != nil {
		errMsg := fmt.Sprintf("Error parsing URL: %v", err)
		resp["error"] = errMsg
		log.Printf("Error processing download for URL %s: %s", url, errMsg)
		return resp
	}

	resp["status"] = "downloading"
	log.Printf("Downloading %s with ID %s", linkType, id)

	// Replace this with your actual download logic based on linkType
	var downloadErr error
	switch linkType {
	case beatport.LinkTypeTrack:
		downloadErr = beatport.DownloadTrack(id)
	case beatport.LinkTypeRelease:
		// Assuming there's a DownloadRelease function
		// downloadErr = beatport.DownloadRelease(id) 
		downloadErr = fmt.Errorf("release downloads not yet supported")
	default:
		downloadErr = fmt.Errorf("unsupported link type: %s", linkType)
	}

	if downloadErr != nil {
		errMsg := fmt.Sprintf("Error during download: %v", downloadErr)
		resp["error"] = errMsg
		resp["status"] = "failed"
		log.Printf("Error downloading %s: %s", url, errMsg)
	} else {
		resp["status"] = "completed"
		log.Printf("Successfully downloaded %s", url)
	}

	return resp
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return

	}

	// Placeholder for actual status retrieval.
	// In a real implementation, you'd track download statuses
	// and return them here.  For now, we return a dummy response.
	status := map[string]string{"status": "Endpoint not yet fully implemented"}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
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
		log.Println("Handled GET request to /config")
	case http.MethodPut, http.MethodPost:
		updateConfig(w, r)
		log.Println("Handled PUT/POST request to /config")
	default:
		log.Printf("Method %s not allowed for /config", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.ReadConfig()

	if err != nil {
		errMsg := fmt.Sprintf("Error reading config: %v", err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(cfg); err != nil {
		log.Printf("Error encoding config response: %v", err)
		http.Error(w, "Error encoding config", http.StatusInternalServerError)
		return
	}

	log.Println("Returned current configuration")
}

func updateConfig(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading config request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	log.Printf("Received config update request: %s", string(body))



	var cfg config.Config
	if err := yaml.Unmarshal(body, &cfg); err != nil {

		errMsg := fmt.Sprintf("Error parsing config: %v", err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}


	if err := config.WriteConfig(cfg); err != nil {
		errMsg := fmt.Sprintf("Error writing config: %v", err)
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "config updated"})
	log.Println("Configuration updated successfully")
}