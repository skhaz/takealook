package functions

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
)

var counterMap sync.Map

func Increment(key uint64) {
	for {
		actual, loaded := counterMap.Load(key)
		if !loaded {
			pointer := new(uint64)
			_, loaded = counterMap.LoadOrStore(key, pointer)
			if loaded {
				continue
			}
			atomic.AddUint64(pointer, 1)
			return
		}
		atomic.AddUint64(actual.(*uint64), 1)
		return
	}
}

func Synchronize(db *sql.DB) {
	sf := &SynchronizeFunctions{db: db}
	for {
		sf.run()
		time.Sleep(time.Second * 10)
	}
}

type SynchronizeFunctions struct {
	db *sql.DB
}

func (sf *SynchronizeFunctions) run() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := sf.sync(); err != nil {
			log.Error("error during synchronization", zap.Error(err))
		}
	}
}

func (sf *SynchronizeFunctions) sync() error {
	tx, err := sf.db.Begin()
	if err != nil {
		log.Error("error starting transaction", zap.Error(err))
		return err
	}

	stmt, err := tx.Prepare("UPDATE data SET count = count + ? WHERE id = ?")
	if err != nil {
		log.Error("error preparing statement", zap.Error(err))
		//nolint:golint,errcheck
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	var failed bool
	counterMap.Range(func(key, value interface{}) bool {
		id := key.(uint64)
		count := value.(*uint64)

		if _, err := stmt.Exec(*count, id); err != nil {
			log.Error("error executing update", zap.Error(err))
			failed = true
			return false
		}
		return true
	})

	if failed {
		//nolint:golint,errcheck
		tx.Rollback()
		return fmt.Errorf("failed to update some counters")
	}

	if err := tx.Commit(); err != nil {
		log.Error("error committing transaction", zap.Error(err))
		//nolint:golint,errcheck
		tx.Rollback()
		return err
	}

	counterMap = sync.Map{}
	return nil
}
