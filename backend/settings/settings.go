package settings

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/yehuoshun/cangye/common"
	"github.com/yehuoshun/cangye/db"
)

// --- Models ---

type Prefix struct {
	Prefix      string `json:"prefix"`
	Type        string `json:"type"`
	MapPath     string `json:"map_path"`
	URLTemplate string `json:"url_template"`
	CreatedAt   string `json:"created_at"`
}

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// --- Route Registration ---

func RegisterRoutes(r *mux.Router) {
	// Settings
	r.HandleFunc("/api/settings/{key}", handleSettingsGet).Methods("GET")
	r.HandleFunc("/api/settings/{key}", handleSettingsPut).Methods("PUT")

	// Prefixes
	r.HandleFunc("/api/prefixes", handlePrefixesList).Methods("GET")
	r.HandleFunc("/api/prefixes/{prefix}", handlePrefixesPut).Methods("PUT")

	// Tags
	r.HandleFunc("/api/tags/search", handleTagsSearch).Methods("GET")
	r.HandleFunc("/api/tags", handleTagsCreate).Methods("POST")
	r.HandleFunc("/api/tags/{id}", handleTagsDelete).Methods("DELETE")

	// File tags
	r.HandleFunc("/api/files/{id}/tags", handleFileTagsGet).Methods("GET")
	r.HandleFunc("/api/files/{id}/tags", handleFileTagsPut).Methods("PUT")

	// Collection tags
	r.HandleFunc("/api/collections/{id}/tags", handleCollectionTagsGet).Methods("GET")
	r.HandleFunc("/api/collections/{id}/tags", handleCollectionTagsPut).Methods("PUT")
}

// --- Settings ---

func handleSettingsGet(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	var value string
	err := db.DB.QueryRow("SELECT value FROM settings WHERE key=?", key).Scan(&value)
	if err == sql.ErrNoRows {
		common.JSON(w, 200, map[string]interface{}{"key": key, "value": nil})
		return
	}
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	common.JSON(w, 200, map[string]string{"key": key, "value": value})
}

func handleSettingsPut(w http.ResponseWriter, r *http.Request) {
	key := mux.Vars(r)["key"]
	var body struct{ Value string `json:"value"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	_, err := db.DB.Exec(
		"INSERT INTO settings (key, value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		key, body.Value,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	common.JSON(w, 200, map[string]string{"key": key, "value": body.Value})
}

// --- Prefixes ---

func handlePrefixesList(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT prefix, type, COALESCE(map_path,''), COALESCE(url_template,''), created_at FROM prefix_config ORDER BY prefix")
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	prefixes := []Prefix{}
	for rows.Next() {
		var p Prefix
		if err := rows.Scan(&p.Prefix, &p.Type, &p.MapPath, &p.URLTemplate, &p.CreatedAt); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		prefixes = append(prefixes, p)
	}
	common.JSON(w, 200, prefixes)
}

func handlePrefixesPut(w http.ResponseWriter, r *http.Request) {
	prefix := mux.Vars(r)["prefix"]
	// Add ":" suffix for lookup if not already
	lookupPrefix := prefix
	if len(prefix) > 0 && prefix[len(prefix)-1] != ':' {
		lookupPrefix = prefix + ":"
	}

	var body struct {
		Type        string `json:"type"`
		MapPath     string `json:"map_path"`
		URLTemplate string `json:"url_template"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}

	_, err := db.DB.Exec(
		"INSERT INTO prefix_config (prefix, type, map_path, url_template) VALUES (?,?,?,?) ON CONFLICT(prefix) DO UPDATE SET type=excluded.type, map_path=excluded.map_path, url_template=excluded.url_template",
		lookupPrefix, body.Type, body.MapPath, body.URLTemplate,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	common.JSON(w, 200, map[string]string{"status": "ok"})
}

// --- Tags ---

func handleTagsSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var rows *sql.Rows
	var err error
	if q == "" {
		rows, err = db.DB.Query("SELECT id, name, color FROM tags ORDER BY name LIMIT 20")
	} else {
		rows, err = db.DB.Query("SELECT id, name, color FROM tags WHERE name LIKE ? ORDER BY name LIMIT 20", "%"+q+"%")
	}
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		tags = append(tags, t)
	}
	common.JSON(w, 200, tags)
}

func handleTagsCreate(w http.ResponseWriter, r *http.Request) {
	var t Tag
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}
	t.ID = uuid.New().String()
	if t.Color == "" {
		t.Color = "gray"
	}

	_, err := db.DB.Exec("INSERT INTO tags (id, name, color) VALUES (?,?,?)", t.ID, t.Name, t.Color)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	common.JSON(w, 201, t)
}

func handleTagsDelete(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	db.DB.Exec("DELETE FROM file_tags WHERE tag_id=?", id)
	db.DB.Exec("DELETE FROM collection_tags WHERE tag_id=?", id)
	db.DB.Exec("DELETE FROM tags WHERE id=?", id)
	common.JSON(w, 200, map[string]string{"status": "deleted"})
}

// --- File Tags ---

func handleFileTagsGet(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	rows, err := db.DB.Query(
		`SELECT t.id, t.name, t.color FROM tags t
		 JOIN file_tags ft ON ft.tag_id = t.id
		 WHERE ft.file_id=? ORDER BY t.name`,
		fileID,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		tags = append(tags, t)
	}
	common.JSON(w, 200, tags)
}

func handleFileTagsPut(w http.ResponseWriter, r *http.Request) {
	fileID := mux.Vars(r)["id"]
	var body struct {
		TagIDs []string `json:"tag_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM file_tags WHERE file_id=?", fileID)
	for _, tagID := range body.TagIDs {
		tx.Exec("INSERT OR IGNORE INTO file_tags (file_id, tag_id) VALUES (?,?)", fileID, tagID)
	}
	tx.Commit()

	common.JSON(w, 200, map[string]string{"status": "ok"})
}

// --- Collection Tags ---

func handleCollectionTagsGet(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	rows, err := db.DB.Query(
		`SELECT t.id, t.name, t.color FROM tags t
		 JOIN collection_tags ct ON ct.tag_id = t.id
		 WHERE ct.collection_id=? ORDER BY t.name`,
		id,
	)
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			common.Error(w, 500, err.Error())
			return
		}
		tags = append(tags, t)
	}
	common.JSON(w, 200, tags)
}

func handleCollectionTagsPut(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var body struct {
		TagIDs []string `json:"tag_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		common.Error(w, 400, "invalid body")
		return
	}

	tx, err := db.DB.Begin()
	if err != nil {
		common.Error(w, 500, err.Error())
		return
	}
	defer tx.Rollback()

	tx.Exec("DELETE FROM collection_tags WHERE collection_id=?", id)
	for _, tagID := range body.TagIDs {
		tx.Exec("INSERT OR IGNORE INTO collection_tags (collection_id, tag_id) VALUES (?,?)", id, tagID)
	}
	tx.Commit()

	common.JSON(w, 200, map[string]string{"status": "ok"})
}
