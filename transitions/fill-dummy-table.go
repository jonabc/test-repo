package transitions

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/github/go-dbmigrator"
)

// Get20200604095405 returns a transition that inserts the current time into a dummy table
func Get20200604095405(db *sql.DB) dbmigrator.MigrationFunc {
	return func(ctx context.Context) error {
		// Note: you should throttle this write! Don't use this query as a good practice.
		// You'll likely have your own logging, statting, etc. functionality to import here as well
		created := time.Now()
		result, err := db.ExecContext(
			ctx,
			"INSERT INTO dummy_table(created_at) VALUES(?)",
			created,
		)
		if err != nil {
			return fmt.Errorf("executing insert: %v", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("getting last insert ID: %v", err)
		}
		log.Printf("ID: %v, created: %v\n", id, created)

		return nil
	}
}
