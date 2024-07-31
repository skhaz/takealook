package functions

import (
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
)

type CollectorFunctions struct {
	db *sql.DB
}

func Collector(db *sql.DB) {
	collector := CollectorFunctions{db: db}

	for {
		if err := collector.run(); err != nil {
			log.Error("collector run error", zap.Error(err))
			time.Sleep(time.Second * 10)
		}
	}
}

func (cf *CollectorFunctions) run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("collector panic", zap.Any("error", r))
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	var query = "DELETE FROM data WHERE error IS NOT NULL AND created_at <= datetime('now', '-10 minutes')"

	for {
		if _, err := cf.db.Exec(query); err != nil {
			log.Error("failed to exec query", zap.Error(err))
			return err
		}

		time.Sleep(time.Minute * 1)
	}
}
