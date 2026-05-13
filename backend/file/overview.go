package file

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/yehuoshun/cangye/common"
	"github.com/yehuoshun/cangye/db"
)

func RegisterOverviewRoutes(r *mux.Router) {
	r.HandleFunc("/api/overview/stats", handleOverviewStats).Methods("GET")
}

func handleOverviewStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]int{}

	db.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE parent_id IS NULL").Scan(&stats["root_collections"])
	db.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE parent_id IS NOT NULL").Scan(&stats["sub_collections"])
	db.DB.QueryRow("SELECT COUNT(*) FROM collection_paths").Scan(&stats["paths"])
	db.DB.QueryRow("SELECT COUNT(*) FROM virtual_files").Scan(&stats["virtual_files"])
	db.DB.QueryRow("SELECT COUNT(*) FROM scan_cache").Scan(&stats["scanned_files"])
	db.DB.QueryRow("SELECT COUNT(*) FROM tags").Scan(&stats["tags"])

	common.JSON(w, 200, map[string]interface{}{"stats": stats})
}
