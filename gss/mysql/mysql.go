package mysql

import (
	"database/sql"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jonabc/test-repo/gss/config"
	"github.com/luna-duclos/instrumentedsql"
	driverTrace "github.com/luna-duclos/instrumentedsql/opentracing"
)

var (
	registerDriverOnce sync.Once
)

// OpenDB returns an opened connection to the database
func OpenDB(cfg *config.Config) (*sql.DB, error) {
	driverName := "instrumented-mysql"
	registerDriverOnce.Do(func() {
		// IMPORTANT: see these docs about working with instrumentedsql and the risks associated:
		// https://github.com/github/go/blob/f02dbd39f9f2587ba2498aa2ec7c7b5b25cab540/docs/database_access.md#database-instrumentation
		driver := instrumentedsql.WrapDriver(mysql.MySQLDriver{},
			instrumentedsql.WithTracer(driverTrace.NewTracer(true)),
			instrumentedsql.WithOmitArgs(),
		)
		sql.Register(driverName, driver)
	})

	mysqlConfig, err := cfg.NewDatabaseConfig()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, mysqlConfig.FormatDSN())
	if err != nil {
		return nil, err
	}

	// align each connection's lifetime with the HAProxy drain timeout for the corresponding GLB cluster
	// see https://github.com/github/turboscan/commit/c114d90035db547c868fabea0a6c2955d189821b
	db.SetConnMaxLifetime(5 * time.Minute)
	// try to prevent consuming all available connections on the db
	db.SetMaxOpenConns(20)

	return db, nil
}
