package main

import (
	"log/slog"
	"yourapp/config"
	"yourapp/internal/server"
)

func main() {
	cfg := config.Load()
	db := config.NewDatabase(cfg.Database.ConnString())
	server := server.NewHTTPServer(db, cfg.Log, cfg)

	if err := server.Start(); err != nil {
		cfg.Log.Error("server failed", slog.String("error", err.Error()))
		return
	}
}
