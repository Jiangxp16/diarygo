package server

import (
	"fmt"
	"net/http"

	"diarygo/internal/backup"
	"diarygo/internal/config"
	"diarygo/internal/db"
	"diarygo/internal/i18n"
)

func InitServer() {
	cfg := config.GetRepository()
	db.Init(cfg.Get("global", "db_name"))
	defer db.Close()

	i18n.Init(cfg.Get("global", "language"))

	m := backup.StartBackup(cfg)
	if m != nil {
		defer m.Stop()
	}

	RegisterRoutes()

	port := cfg.Get("global", "port")

	fmt.Printf("Server started at http://localhost:%s", port)
	http.ListenAndServe(":"+port, nil)
}
