package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	runner "github.com/TheBizzle/AStar-Studio/internal/runner"
	testset "github.com/TheBizzle/PathFindingCore-Golang/testset"
)

type SendableTest struct {
	Name    string `json:"name"`
	Delim   string `json:"delim"`
	Content string `json:"content"`
}

func rootHandler(res http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.NotFound(res, req)
		return
	}
	http.ServeFile(res, req, "frontend/index.html")
}

func exampleMapsHandler(res http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Printf("Error closing request body: %v", err)
		}
	}()

	if req.Method == http.MethodGet {
		var sendables []SendableTest
		for i, test := range testset.Tests {
			name := fmt.Sprintf("Example #%d", i+1)
			sendable := SendableTest{Name: name, Delim: test.MapStr.Delim, Content: test.MapStr.Contents}
			sendables = append(sendables, sendable)
		}
		if err := json.NewEncoder(res).Encode(sendables); err != nil {
			http.Error(res, "Error encoding response", http.StatusInternalServerError)
		}
	} else {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func pathfindingHandler(res http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := req.Body.Close(); err != nil {
			log.Printf("Error closing request body: %v", err)
		}
	}()

	if req.Method == http.MethodPost {
		const oneMB = 1 << 20
		body, err := io.ReadAll(http.MaxBytesReader(res, req.Body, oneMB))
		if err != nil {
			http.Error(res, "Error reading body", http.StatusInternalServerError)
			return
		}

		wasSuccessful, pmapStr, timingStrs := runner.RunAStars(string(body))
		var writeErr error
		if wasSuccessful {
			_, writeErr = fmt.Fprintf(res, "%d,%v,%v", 0, strings.Join(timingStrs, "&"), pmapStr)
		} else {
			_, writeErr = fmt.Fprintf(res, "%d,%v", 1, pmapStr)
		}

		if writeErr != nil {
			log.Printf("Error when writing response: %v", writeErr)
		}
	} else {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/example-maps", exampleMapsHandler)
	http.HandleFunc("/find-me-a-path", pathfindingHandler)

	portNum := 8080
	fmt.Printf("Pathfinding server running on port %d\n", portNum)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", portNum), nil); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}
