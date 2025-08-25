package database

import "github.com/MH-PAVEL/uni-backend-go/internal/config"

// DbName returns the database name from config
func DbName() string {
    cfg := config.AppConfig
    if cfg == nil {
        return ""
    }
    return cfg.Database.Name
}
