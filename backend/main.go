package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/yehuoshun/cangye/checkin"
	"github.com/yehuoshun/cangye/db"
	"github.com/yehuoshun/cangye/file"
	"github.com/yehuoshun/cangye/rss"
	"github.com/yehuoshun/cangye/settings"
)

//go:embed web/dist/*
var frontendFS embed.FS

const defaultPort = 27138

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("cangye v0.1.0")
		return
	}

	// Init DB
	if err := db.Init(""); err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer db.Close()

	// Find available port
	port := findPort(defaultPort)

	// Router
	r := mux.NewRouter()
	r.Use(corsMiddleware)

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	file.RegisterRoutes(api)
	file.RegisterOverviewRoutes(api)
	rss.RegisterRoutes(api)
	checkin.RegisterRoutes(api)
	settings.RegisterRoutes(api)

	// Serve embedded frontend
	distFS, err := fs.Sub(frontendFS, "web/dist")
	if err != nil {
		log.Printf("frontend not embedded, serving API only")
	} else {
		spa := spaHandler{staticFS: distFS, indexPath: "index.html"}
		r.PathPrefix("/").Handler(spa)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	url := fmt.Sprintf("http://%s", addr)

	log.Printf("cangye starting on %s", url)

	// Open browser
	go openBrowser(url)

	// Start server
	log.Fatal(http.ListenAndServe(addr, r))
}

func findPort(preferred int) int {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", preferred))
	if err == nil {
		ln.Close()
		return preferred
	}
	// Try +1 up to 10
	for i := 1; i <= 10; i++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", preferred+i))
		if err == nil {
			ln.Close()
			log.Printf("port %d in use, using %d", preferred, preferred+i)
			return preferred + i
		}
	}
	return preferred
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		log.Printf("auto-open browser failed: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(200)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// spaHandler serves SPA from embedded FS
type spaHandler struct {
	staticFS  fs.FS
	indexPath string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = h.indexPath
	}

	// Try requested path
	data, err := fs.ReadFile(h.staticFS, path)
	if err == nil {
		http.ServeContent(w, r, path, time.Now(), bytes.NewReader(data))
		return
	}

	// Fallback to index.html for SPA routing
	data, err = fs.ReadFile(h.staticFS, h.indexPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, h.indexPath, time.Now(), bytes.NewReader(data))
}
