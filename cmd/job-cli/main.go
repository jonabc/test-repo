package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/github/go-dbmigrator"
	migratorCfg "github.com/github/go-dbmigrator/config"
	"github.com/github/go-dbmigrator/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jonabc/test-repo/gss/config"
	"github.com/jonabc/test-repo/transitions"
	"github.com/urfave/cli/v2"
)

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

func realMain() error {

	app := &cli.App{
		Name:  "transition",
		Usage: "Run transitions, or GHE migrations and transitions with this small CLI",
		Commands: []*cli.Command{
			{
				Name:    "enterprise",
				Aliases: []string{"e"},
				Usage:   "Run database migrations and transitions in GHE",
				Action:  EnterpriseAction,
			},
			{
				Name:    "cloud",
				Aliases: []string{"c"},
				Usage:   "Run transitions only for cloud environments",
				Action:  CloudAction,
			},
		},
	}

	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)
	if err != nil {
		return fmt.Errorf("running CLI: %v", err)
	}

	return nil
}

// EnterpriseAction runs migrations and transitions together, for use in GHE environments.
func EnterpriseAction(c *cli.Context) error {
	// Your application will likely have its own configuration, app context, and database helper code to plug in here.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading configuration: %v", err)
	}

	dbCfg, err := cfg.NewDatabaseConfig()
	if err != nil {
		return fmt.Errorf("loading DB config: %v", err)
	}

	db, err := sql.Open("mysql", dbCfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("opening sql DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging DB: %v", err)
	}

	transitions, err := GetEnterpriseTransitions(db)
	if err != nil {
		return fmt.Errorf("getting enterprise transitions: %v", err)
	}

	root, err := migratorCfg.RootDir()
	if err != nil {
		return fmt.Errorf("getting root directory: %v", err)
	}

	fullSourceURL := fmt.Sprintf("file://%s", filepath.Join(root, "migrations"))

	// TODO: construct migrator options with logger and pass it in
	migrator, err := mysql.NewWithDatabaseInstance(nil, db, fullSourceURL, "test_migrations")
	if err != nil {
		return fmt.Errorf("creating migrator: %v", err)
	}
	defer migrator.Close()

	return migrator.Migrate(c.Context, transitions)
}

// CloudAction runs the data transitions only, for use in cloud environments (locally and with Moda jobs).
func CloudAction(c *cli.Context) error {
	// Your application will likely have its own configuration, app context, and database helper code to plug in here.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading configuration: %v", err)
	}

	dbCfg, err := cfg.NewDatabaseConfig()
	if err != nil {
		return fmt.Errorf("loading database config: %v", err)
	}

	db, err := sql.Open("mysql", dbCfg.FormatDSN())
	if err != nil {
		return fmt.Errorf("opening sql DB: %v", err)
	}
	if err := db.Ping(); err != nil {
		return fmt.Errorf("pinging DB: %v", err)
	}

	transitions, err := GetCloudTransitions(db)
	if err != nil {
		return fmt.Errorf("getting cloud transitions: %v", err)
	}
	trans := transitions.LatestTransition()
	return trans.Run(c.Context)
}

// GetCloudTransitions returns all transitions to be applied for the database
func GetCloudTransitions(db *sql.DB) (*dbmigrator.Transitioner, error) {
	trans := dbmigrator.NewTransitioner()
	err := trans.Add(20200604095405, transitions.Get20200604095405(db))
	if err != nil {
		return nil, err
	}

	return trans, nil
}

// GetEnterpriseTransitions returns all transitions to be applied in GHE
func GetEnterpriseTransitions(db *sql.DB) (*dbmigrator.Transitioner, error) {
	trans := dbmigrator.NewTransitioner()
	err := trans.Add(20200604095405, transitions.Get20200604095405(db))
	if err != nil {
		return nil, err
	}
	return trans, nil
}
