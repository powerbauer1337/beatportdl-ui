package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/your-username/your-repo/cmd/server" // Replace with your actual import path
	// Add other necessary BeatportDL imports here, e.g.:
	// "github.com/your-username/your-repo/internal/beatport" 
	// "github.com/your-username/your-repo/config"
)

func main() {
	serverMode := flag.Bool("server", false, "Start in server mode")
	flag.Parse()

	if *serverMode {
		// Start HTTP server in a separate goroutine
		go func() {
			if err := server.StartServer(); err != nil { // Replace with your actual server start function
				log.Printf("Error starting server: %v", err)
			}
		}()
		fmt.Println("BeatportDL server started. Listening on port 8080...") // Adjust port if needed

		// Keep the main goroutine alive while the server is running.  
		// You might want to add a more sophisticated mechanism to handle server shutdown.
		select {} 
	} else {
		// Run in command-line mode (existing BeatportDL logic)
		//  Replace this with the actual command-line execution logic of BeatportDL
		fmt.Println("Running BeatportDL in command-line mode...")
		// Example:
		// cfg, err := config.LoadConfig()
		// if err != nil {
		//  log.Fatalf("Error loading config: %v", err)
		// }
		// 
		// if err := beatport.RunCommandLine(cfg, flag.Args()); err != nil {
		//	log.Fatalf("Error during command-line execution: %v", err)
		// }

		fmt.Println("Command-line execution finished.")
	}
}

// Note: You need to create a StartServer function in your server package (e.g., cmd/server/server.go)
// This function should contain the server logic (as implemented previously in cmd/server/main.go),
// but it should not contain its own main function.  Instead it should return an error if the server fails to start.

// Example server.go:

// package server

// import (
// 	"fmt"
// 	"net/http"
// 	"encoding/json"
// 	// ... other imports ...
// )

// // ... downloadStatus, config loading/saving functions ...

// func StartServer() error {
// 	// ... HTTP handlers for /download, /status, /config ...
// 	http.HandleFunc("/download", downloadHandler)
// 	http.HandleFunc("/status", statusHandler)
// 	http.HandleFunc("/config", configHandler)

// 	fmt.Println("Server listening on port 8080")
// 	if err := http.ListenAndServe(":8080", nil); err != nil {
// 		return fmt.Errorf("error starting server: %w", err)
// 	}
// 	return nil
// }

//  ... Handler functions (downloadHandler, statusHandler, configHandler) ...