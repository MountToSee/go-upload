package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	port      string
	uploadDir string
)

func main() {
	// Parse command line arguments
	flag.StringVar(&port, "h", "8000", "Server port")
	flag.StringVar(&uploadDir, "d", "/tmp/upload", "Upload directory")
	flag.Parse()

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	// Setup HTTP handlers
	http.HandleFunc("/", handleRequest)

	// Start server
	addr := ":" + port
	log.Printf("Starting file server on port %s, serving directory: %s", port, uploadDir)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPut:
		handlePut(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handle GET requests - list files in directory
func handleGet(w http.ResponseWriter, r *http.Request) {
	// Clean the path to prevent directory traversal attacks
	requestPath := filepath.Clean(r.URL.Path)
	if requestPath == "." {
		requestPath = "/"
	}
	
	// Build the full path
	fullPath := filepath.Join(uploadDir, requestPath)

	// Check if path exists
	info, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		http.Error(w, "Path not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Error accessing path: %v", err), http.StatusInternalServerError)
		return
	}

	// If it's a file, serve the file
	if !info.IsDir() {
		serveFile(w, r, fullPath)
		return
	}

	// If it's a directory, list its contents
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading directory: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<html><head><title>Directory listing for %s</title></head><body>\n", r.URL.Path)
	fmt.Fprintf(w, "<h1>Directory listing for %s</h1>\n", r.URL.Path)
	fmt.Fprintf(w, "<hr>\n<ul>\n")

	// Add parent directory link if not at root
	if requestPath != "/" {
		parentPath := filepath.Dir(requestPath)
		if parentPath == "." {
			parentPath = "/"
		}
		fmt.Fprintf(w, "<li><a href=\"%s\">../</a></li>\n", parentPath)
	}

	// List all entries
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		linkPath := filepath.Join(r.URL.Path, entry.Name())
		linkPath = filepath.ToSlash(linkPath) // Convert to forward slashes for URLs
		fmt.Fprintf(w, "<li><a href=\"%s\">%s</a></li>\n", linkPath, name)
	}

	fmt.Fprintf(w, "</ul>\n<hr>\n</body></html>\n")
}

// serveFile serves a file with appropriate headers based on file type
func serveFile(w http.ResponseWriter, r *http.Request, filePath string) {
	// Get the MIME type based on file extension
	ext := filepath.Ext(filePath)
	mimeType := mime.TypeByExtension(ext)
	
	// Determine if the file is a text file
	isTextFile := isTextMimeType(mimeType)
	
	if isTextFile {
		// Text files: display in browser
		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
		log.Printf("Serving text file for viewing: %s (type: %s)", filePath, mimeType)
	} else {
		// Non-text files: force download
		fileName := filepath.Base(filePath)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		log.Printf("Serving file for download: %s (type: %s)", filePath, mimeType)
	}
	
	http.ServeFile(w, r, filePath)
}

// isTextMimeType checks if a MIME type represents a text file
func isTextMimeType(mimeType string) bool {
	if mimeType == "" {
		return false
	}
	
	// Text MIME types that should be viewable in browser
	textPrefixes := []string{
		"text/",           // text/plain, text/html, text/css, etc.
		"application/json",
		"application/xml",
		"application/javascript",
		"application/x-javascript",
	}
	
	for _, prefix := range textPrefixes {
		if strings.HasPrefix(mimeType, prefix) {
			return true
		}
	}
	
	return false
}

// Handle PUT requests - upload files
func handlePut(w http.ResponseWriter, r *http.Request) {
	// Clean the path to prevent directory traversal attacks
	requestPath := filepath.Clean(r.URL.Path)
	if requestPath == "/" || requestPath == "." {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Remove leading slash for filepath.Join
	requestPath = strings.TrimPrefix(requestPath, "/")
	
	// Build the full path
	fullPath := filepath.Join(uploadDir, requestPath)

	// Create parent directories if they don't exist
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy the uploaded data to the file
	written, err := io.Copy(file, r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Uploaded file: %s (%d bytes)", fullPath, written)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File uploaded successfully: %s (%d bytes)\n", requestPath, written)
}
