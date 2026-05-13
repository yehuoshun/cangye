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

	var rootCol, subCol, pathCount, vfCount, scCount, tagCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE parent_id IS NULL").Scan(&rootCol)
	db.DB.QueryRow("SELECT COUNT(*) FROM collections WHERE parent_id IS NOT NULL").Scan(&subCol)
	db.DB.QueryRow("SELECT COUNT(*) FROM collection_paths").Scan(&pathCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM virtual_files").Scan(&vfCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM scan_cache").Scan(&scCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM tags").Scan(&tagCount)
	stats["root_collections"] = rootCol
	stats["sub_collections"] = subCol
	stats["paths"] = pathCount
	stats["virtual_files"] = vfCount
	stats["scanned_files"] = scCount
	stats["tags"] = tagCount

	common.JSON(w, 200, map[string]interface{}{"stats": stats})
}
