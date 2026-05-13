package file

import (
	"database/sql"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yehuoshun/cangye/common"
	"github.com/yehuoshun/cangye/db"
)

// --- Models ---

type Collection struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Icon      string  `json:"icon"`
	ParentID  *string `json:"parent_id"`
	SortOrder int     `json:"sort_order"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
	// denormalized
	PathCount  int    `json:"path_count,omitempty"`
	FileCount  int    `json:"file_count,omitempty"`
	PrefixType string `json:"prefix_type,omitempty"`
}

type CollectionPath struct {
	ID           string `json:"id"`
	CollectionID string `json:"collection_id"`
	Path         string `json:"path"`
	AutoScan     bool   `json:"auto_scan"`
	SortOrder    int    `json:"sort_order"`
	CreatedAt    string `json:"created_at"`
}

type VirtualFile struct {
	ID           string  `json:"id"`
	CollectionID string  `json:"collection_id"`
	Path         string  `json:"path"`
	DisplayName  *string `json:"display_name"`
	Size         int64   `json:"size"`
	SortOrder    int     `json:"sort_order"`
	CreatedAt    string  `json:"created_at"`
	// denormalized from scan
	FileName string  `json:"file_name,omitempty"`
	MimeType *string `json:"mime_type,omitempty"`
	ModTime  *string `json:"mod_time,omitempty"`
}

type FileEntry struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	Size       int64   `json:"size"`
	ModTime    string  `json:"mod_time"`
	MimeType   string  `json:"mime_type"`
	Source     string  `json:"source"` // "scan" or "virtual"
	IsDir      bool    `json:"is_dir"`
	Prefix     string  `json:"prefix,omitempty"`
	PrefixType string  `json:"prefix_type,omitempty"`
	Icon       string  `json:"icon,omitempty"`
}

type ReorderRequest struct {
	Items []ReorderItem `json:"items"`
}
type ReorderItem struct {
	ID        string `json:"id"`
	SortOrder int    `json:"sort_order"`
}

type CreatePathRequest struct {
	Path     string `json:"path"`
	AutoScan *bool  `json:"auto_scan"`
}

// --- MIME helpers ---

var mimeMap = map[string]string{
	".txt": "text/plain", ".md": "text/markdown", ".go": "text/x-go",
	".js": "text/javascript", ".ts": "text/typescript", ".vue": "text/x-vue",
	".html": "text/html", ".css": "text/css", ".json": "application/json",
	".xml": "text/xml", ".yaml": "text/yaml", ".yml": "text/yaml",
	".py": "text/x-python", ".java": "text/x-java", ".rs": "text/x-rust",
	".c": "text/x-c", ".h": "text/x-c", ".cpp": "text/x-c++",
	".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg",
	".gif": "image/gif", ".webp": "image/webp", ".svg": "image/svg+xml",
	".mp4": "video/mp4", ".webm": "video/webm", ".mkv": "video/x-matroska",
	".mp3": "audio/mpeg", ".wav": "audio/wav", ".flac": "audio/flac",
	".pdf": "application/pdf", ".zip": "application/zip", ".rar": "application/x-rar",
	".doc": "application/msword", ".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls": "application/vnd.ms-excel", ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt": "application/vnd.ms-powerpoint", ".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
}

func getMime(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	if m, ok := mimeMap[ext]; ok {
		return m
	}
	return "application/octet-stream"
}

func getCategory(mime string) string {
	if strings.HasPrefix(mime, "image/") {
		return "image"
	}
	if strings.HasPrefix(mime, "video/") {
		return "video"
	}
	if strings.HasPrefix(mime, "audio/") {
		return "audio"
	}
	if strings.HasPrefix(mime, "text/") || mime == "application/json" || mime == "application/xml" {
		return "text"
	}
	return "other"
}

func getIcon(cat, name string) string {
	icons := map[string]string{
		"image": "🖼️",
		"video": "🎬",
		"audio": "🎵",
		"text":  "📄",
		"other": "📎",
	}
	if icon, ok := icons[cat]; ok {
		return icon
	}
	return "📎"
}

// --- Route Registration ---

func RegisterRoutes(r *mux.Router) {
	// Collections
	r.HandleFunc("/api/collections", handleCollectionsList).Methods("GET")
	r.HandleFunc("/api/collections", handleCollectionsCreate).Methods("POST")
	r.HandleFunc("/api/collections/reorder", handleCollectionsReorder).Methods("PUT")
	r.HandleFunc("/api/collections/{id}", handleCollectionsGet).Methods("GET")
	r.HandleFunc("/api/collections/{id}", handleCollectionsUpdate).Methods("PUT")
	r.HandleFunc("/api/collections/{id}", handleCollectionsDelete).Methods("DELETE")
	r.HandleFunc("/api/collections/{id}/children", handleCollectionsChildren).Methods("GET")

	// Paths
	r.HandleFunc("/api/collections/{id}/paths", handlePathsList).Methods("GET")
	r.HandleFunc("/api/collections/{id}/paths", handlePathsCreate).Methods("POST")
	r.HandleFunc("/api/paths/{id}", handlePathsUpdate).Methods("PUT")
	r.HandleFunc("/api/paths/{id}", handlePathsDelete).Methods("DELETE")
	r.HandleFunc("/api/paths/{id}/scan", handlePathsScan).Methods("POST")

	// Virtual files
	r.HandleFunc("/api/collections/{id}/vfiles", handleVFilesList).Methods("GET")
	r.HandleFunc("/api/collections/{id}/vfiles", handleVFilesCreate).Methods("POST")
	r.HandleFunc("/api/vfiles/{id}", handleVFilesUpdate).Methods("PUT")
	r.HandleFunc("/api/vfiles/{id}", handleVFilesDelete).Methods("DELETE")

	// Browse & Preview
	r.HandleFunc("/api/collections/{id}/browse", handleBrowse).Methods("GET")
	r.HandleFunc("/api/preview/content", handlePreviewContent).Methods("GET")
	r.HandleFunc("/api/preview/thumbnail", handlePreviewThumbnail).Methods("GET")
	r.HandleFunc("/api/office", handleOffice).Methods("POST")
	r.HandleFunc("/api/open-external", handleOpenExternal).Methods("POST")
}

// --- Collections handlers ---

func handleCollectionsList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.name, c.icon, c.parent_id, c.sort_order, c.created_at, c.updated_at
		FROM collections c WHERE c.parent_id IS NULL ORDER BY c.sort_order, c.name
	`)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	collections := []Collection{}
	for rows.Next() {
		var c Collection
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &parentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		if parentID.Valid {
			c.ParentID = &parentID.String
		}
		collections = append(collections, c)
	}

	// enrich with counts
	for i := range collections {
		db.DB.QueryRow("SELECT COUNT(*) FROM collection_paths WHERE collection_id=?", collections[i].ID).Scan(&collections[i].PathCount)
		db.DB.QueryRow("SELECT COUNT(*) FROM virtual_files WHERE collection_id=?", collections[i].ID).Scan(&collections[i].FileCount)
	}

	common.JSON(w, 200, collections)
}

func handleCollectionsCreate(w http.ResponseWriter, r *http.Request) {
	var c Collection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	c.ID = uuid.New().String()
	if c.Icon == "" {
		c.Icon = "📁"
	}

	_, err := db.DB.Exec(
		"INSERT INTO collections (id, name, icon, parent_id, sort_order) VALUES (?,?,?,?,?)",
		c.ID, c.Name, c.Icon, c.ParentID, c.SortOrder,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}

	common.JSON(w, 201, c)
}

func handleCollectionsGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c Collection
	var parentID sql.NullString
	err := db.DB.QueryRow(
		"SELECT id, name, icon, parent_id, sort_order, created_at, updated_at FROM collections WHERE id=?",
		id,
	).Scan(&c.ID, &c.Name, &c.Icon, &parentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		common.Error(w, 404, "not found")
		return
	}
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	if parentID.Valid {
		c.ParentID = &parentID.String
	}
	common.JSON(w, 200, c)
}

func handleCollectionsUpdate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c Collection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	_, err := db.DB.Exec(
		"UPDATE collections SET name=?, icon=?, sort_order=?, updated_at=datetime('now') WHERE id=?",
		c.Name, c.Icon, c.SortOrder, id,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	c.ID = id
	common.JSON(w, 200, c)
}

func handleCollectionsDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	// delete scan_cache for paths in this collection
	db.DB.Exec("DELETE FROM scan_cache WHERE collection_path_id IN (SELECT id FROM collection_paths WHERE collection_id=?)", id)
	db.DB.Exec("DELETE FROM collection_paths WHERE collection_id=?", id)
	db.DB.Exec("DELETE FROM virtual_files WHERE collection_id=?", id)
	db.DB.Exec("DELETE FROM collection_tags WHERE collection_id=?", id)
	db.DB.Exec("DELETE FROM collections WHERE id=?", id)
	common.JSON(w, 200, map[string]string{"status": "deleted"})
}

func handleCollectionsReorder(w http.ResponseWriter, r *http.Request) {
	var req ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	for _, item := range req.Items {
		db.DB.Exec("UPDATE collections SET sort_order=? WHERE id=?", item.SortOrder, item.ID)
	}
	common.JSON(w, 200, map[string]string{"status": "ok"})
}

func handleCollectionsChildren(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	rows, err := db.DB.Query(
		"SELECT id, name, icon, parent_id, sort_order, created_at, updated_at FROM collections WHERE parent_id=? ORDER BY sort_order, name",
		id,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	children := []Collection{}
	for rows.Next() {
		var c Collection
		var parentID sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &parentID, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		if parentID.Valid {
			c.ParentID = &parentID.String
		}
		children = append(children, c)
	}
	common.JSON(w, 200, children)
}

// --- Paths handlers ---

func handlePathsList(w http.ResponseWriter, r *http.Request) {
	collectionID := mux.Vars(r)["id"]
	rows, err := db.DB.Query(
		"SELECT id, collection_id, path, auto_scan, sort_order, created_at FROM collection_paths WHERE collection_id=? ORDER BY sort_order",
		collectionID,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	paths := []CollectionPath{}
	for rows.Next() {
		var p CollectionPath
		var autoScan int
		if err := rows.Scan(&p.ID, &p.CollectionID, &p.Path, &autoScan, &p.SortOrder, &p.CreatedAt); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		p.AutoScan = autoScan == 1
		paths = append(paths, p)
	}
	common.JSON(w, 200, paths)
}

func handlePathsCreate(w http.ResponseWriter, r *http.Request) {
	collectionID := mux.Vars(r)["id"]
	var req CreatePathRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	autoScan := 1
	if req.AutoScan != nil && !*req.AutoScan {
		autoScan = 0
	}
	id := uuid.New().String()
	_, err := db.DB.Exec(
		"INSERT INTO collection_paths (id, collection_id, path, auto_scan) VALUES (?,?,?,?)",
		id, collectionID, req.Path, autoScan,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	common.JSON(w, 201, CollectionPath{ID: id, CollectionID: collectionID, Path: req.Path, AutoScan: autoScan == 1})
}

func handlePathsUpdate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var p CollectionPath
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	autoScan := 1
	if !p.AutoScan {
		autoScan = 0
	}
	db.DB.Exec("UPDATE collection_paths SET path=?, auto_scan=? WHERE id=?", p.Path, autoScan, id)
	common.JSON(w, 200, map[string]string{"status": "ok"})
}

func handlePathsDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	db.DB.Exec("DELETE FROM scan_cache WHERE collection_path_id=?", id)
	db.DB.Exec("DELETE FROM collection_paths WHERE id=?", id)
	common.JSON(w, 200, map[string]string{"status": "deleted"})
}

func handlePathsScan(w http.ResponseWriter, r *http.Request) {
	pathID := mux.Vars(r)["id"]

	var dirPath string
	err := db.DB.QueryRow("SELECT path FROM collection_paths WHERE id=?", pathID).Scan(&dirPath)
	if err != nil {
		common.Error(w, 404, "path not found")
		return
	}

	// clear old cache
	db.DB.Exec("DELETE FROM scan_cache WHERE collection_path_id=?", pathID)

	files := []FileEntry{}
	filepath.WalkDir(dirPath, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, _ := d.Info()
		mime := getMime(p)
		size := int64(0)
		modTime := ""
		if info != nil {
			size = info.Size()
			modTime = info.ModTime().Format(time.RFC3339)
		}

		db.DB.Exec(
			"INSERT OR REPLACE INTO scan_cache (collection_path_id, file_path, file_name, file_size, mod_time, mime_type) VALUES (?,?,?,?,?,?)",
			pathID, p, filepath.Base(p), size, modTime, mime,
		)

		files = append(files, FileEntry{
			Name:     filepath.Base(p),
			Path:     p,
			Size:     size,
			ModTime:  modTime,
			MimeType: mime,
			Source:   "scan",
		})
		return nil
	})

	common.JSON(w, 200, map[string]interface{}{
		"path_id": pathID,
		"count":   len(files),
		"files":   files,
	})
}

// --- Virtual files handlers ---

func handleVFilesList(w http.ResponseWriter, r *http.Request) {
	collectionID := mux.Vars(r)["id"]
	rows, err := db.DB.Query(
		"SELECT id, collection_id, path, display_name, size, sort_order, created_at FROM virtual_files WHERE collection_id=? ORDER BY sort_order, display_name",
		collectionID,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	files := []VirtualFile{}
	for rows.Next() {
		var f VirtualFile
		var displayName sql.NullString
		if err := rows.Scan(&f.ID, &f.CollectionID, &f.Path, &displayName, &f.Size, &f.SortOrder, &f.CreatedAt); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		if displayName.Valid {
			f.DisplayName = &displayName.String
		}
		files = append(files, f)
	}
	common.JSON(w, 200, files)
}

func handleVFilesCreate(w http.ResponseWriter, r *http.Request) {
	collectionID := mux.Vars(r)["id"]
	var f VirtualFile
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	f.ID = uuid.New().String()
	f.CollectionID = collectionID

	_, err := db.DB.Exec(
		"INSERT INTO virtual_files (id, collection_id, path, display_name, size, sort_order) VALUES (?,?,?,?,?,?)",
		f.ID, collectionID, f.Path, f.DisplayName, f.Size, f.SortOrder,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	common.JSON(w, 201, f)
}

func handleVFilesUpdate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var f VirtualFile
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	db.DB.Exec(
		"UPDATE virtual_files SET path=?, display_name=?, size=?, sort_order=? WHERE id=?",
		f.Path, f.DisplayName, f.Size, f.SortOrder, id,
	)
	common.JSON(w, 200, map[string]string{"status": "ok"})
}

func handleVFilesDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	db.DB.Exec("DELETE FROM file_tags WHERE file_id=?", id)
	db.DB.Exec("DELETE FROM virtual_files WHERE id=?", id)
	common.JSON(w, 200, map[string]string{"status": "deleted"})
}

// --- Browse / Preview handlers ---

func handleBrowse(w http.ResponseWriter, r *http.Request) {
	collectionID := mux.Vars(r)["id"]

	entries := []FileEntry{}

	// get prefix info for this collection's paths
	prefixInfo := map[string]string{} // prefix -> type
	prefixRows, _ := db.DB.Query("SELECT prefix, type FROM prefix_config")
	if prefixRows != nil {
		defer prefixRows.Close()
		for prefixRows.Next() {
			var pfx, ptype string
			prefixRows.Scan(&pfx, &ptype)
			prefixInfo[pfx] = ptype
		}
	}

	// scanned files from collection paths
	pathRows, err := db.DB.Query(
		"SELECT id, path FROM collection_paths WHERE collection_id=?",
		collectionID,
	)
	if err == nil {
		defer pathRows.Close()
		for pathRows.Next() {
			var pathID, dirPath string
			pathRows.Scan(&pathID, &dirPath)

			cacheRows, err := db.DB.Query(
				"SELECT file_path, file_name, file_size, mod_time, mime_type FROM scan_cache WHERE collection_path_id=?",
				pathID,
			)
			if err != nil {
				continue
			}
			for cacheRows.Next() {
				var fp, fn, mt string
				var fs int64
				var modTime sql.NullString
				cacheRows.Scan(&fp, &fn, &fs, &modTime, &mt)
				cat := getCategory(mt)
				entries = append(entries, FileEntry{
					ID:       "scan:" + fp,
					Name:     fn,
					Path:     fp,
					Size:     fs,
					ModTime:  modTime.String,
					MimeType: mt,
					Source:   "scan",
					IsDir:    false,
					Icon:     getIcon(cat, fn),
				})
			}
			cacheRows.Close()
		}
	}

	// virtual files
	vfRows, err := db.DB.Query(
		"SELECT id, path, display_name, size, created_at FROM virtual_files WHERE collection_id=?",
		collectionID,
	)
	if err == nil {
		defer vfRows.Close()
		for vfRows.Next() {
			var f VirtualFile
			var displayName sql.NullString
			vfRows.Scan(&f.ID, &f.Path, &displayName, &f.Size, &f.CreatedAt)
			name := f.Path
			if displayName.Valid && displayName.String != "" {
				name = displayName.String
			}
			mime := getMime(f.Path)
			cat := getCategory(mime)

			// detect prefix
			pfx := ""
			ptype := ""
			for p := range prefixInfo {
				if strings.HasPrefix(f.Path, p) {
					pfx = p
					ptype = prefixInfo[p]
					break
				}
			}

			entries = append(entries, FileEntry{
				ID:         f.ID,
				Name:       name,
				Path:       f.Path,
				Size:       f.Size,
				MimeType:   mime,
				Source:     "virtual",
				IsDir:      false,
				Prefix:     pfx,
				PrefixType: ptype,
				Icon:       getIcon(cat, name),
			})
		}
	}

	// sort
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name < entries[j].Name
	})

	common.JSON(w, 200, entries)
}

func handlePreviewContent(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		common.Error(w, 400, "path required")
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		common.Error(w, 500, "cannot read: "+err.Error())
		return
	}
	mime := getMime(path)
	w.Header().Set("Content-Type", mime+"; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func handlePreviewThumbnail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		common.Error(w, 400, "path required")
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		common.Error(w, 500, "cannot read: "+err.Error())
		return
	}
	mime := getMime(path)
	if !strings.HasPrefix(mime, "image/") {
		common.Error(w, 400, "not an image")
		return
	}
	w.Header().Set("Content-Type", mime)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Write(data)
}

func handleOffice(w http.ResponseWriter, r *http.Request) {
	// TODO: Office file parsing
	common.Error(w, 501, "office preview not implemented")
}

func handleOpenExternal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	if body.Path == "" {
		common.Error(w, 400, "path required")
		return
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", body.Path)
	case "darwin":
		cmd = exec.Command("open", body.Path)
	default:
		cmd = exec.Command("xdg-open", body.Path)
	}
	if err := cmd.Start(); err != nil {
		log.Println("[open-external] error:", err)
		common.JSON(w, 200, map[string]string{"status": "attempted", "error": err.Error()})
		return
	}
	common.JSON(w, 200, map[string]string{"status": "ok"})
}
