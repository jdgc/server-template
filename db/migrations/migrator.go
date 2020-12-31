// https;//techinscribed.com/create-db-migrations-in-go-from-scratch

package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/template"
	"time"
)

type Migration struct {
	Version string
	Up      func(*sql.Tx) error
	Down    func(*sql.Tx) error

	processed bool
}

type Migrator struct {
	db         *sql.DB
	Versions   []string
	Migrations map[string]*Migration
}

var migrator = &Migrator{
	Versions:   []string{},
	Migrations: map[string]*Migration{},
}

func Init(db *sql.DB) (*Migrator, error) {
	err := db.Ping()
	if err != nil {
		fmt.Printf("Failed to establish DB connection: %s\n", err.Error())
		return migrator, err
	}

	migrator.db = db

	// create schema table on first run
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
			version varchar(255)
	);`); err != nil {
		fmt.Println("Unable to create `schema_migrations` table", err)
		return migrator, err
	}

	rows, err := db.Query("SELECT version FROM schema_migrations;")
	if err != nil {
		fmt.Printf("Unable to retrieve version from schema_migrations table: %s\n", err.Error())
		return migrator, err
	}

	defer rows.Close()

	// search and mark as processed
	for rows.Next() {
		var version string
		err := rows.Scan(&version)
		if err != nil {
			return migrator, err
		}

		if migrator.Migrations[version] != nil {
			migrator.Migrations[version].processed = true
		}
	}

	return migrator, err
}

func (m *Migrator) MigrationStatus() error {
	for _, v := range m.Versions {
		mg := m.Migrations[v]

		if mg.processed {
			fmt.Println(fmt.Sprintf("Migration %s completed", v))
		} else {
			fmt.Println(fmt.Sprintf("Migration %s pending", v))
		}
	}

	return nil
}

func (m *Migrator) AddMigration(mg *Migration) {
	// add migration to collection with version as key
	m.Migrations[mg.Version] = mg

	// insert into versions array
	index := 0
	for index < len(m.Versions) {
		if m.Versions[index] > mg.Version {
			break
		}
		index++
	}

	m.Versions = append(m.Versions, mg.Version)
	copy(m.Versions[index+1:], m.Versions[index:])
	m.Versions[index] = mg.Version
}

func Create(name string) error {
	version := time.Now().Format("20060102150405")

	in := struct {
		Version string
		Name    string
	}{
		Version: version,
		Name:    name,
	}

	var out bytes.Buffer

	t := template.Must(template.ParseFiles("./db/migrations/template.txt"))
	err := t.Execute(&out, in)
	if err != nil {
		fmt.Printf("Unable to execute migration template: %s", err.Error())
	}

	f, err := os.Create(fmt.Sprintf("./db/migrations/%s_%s.go", version, name))
	if err != nil {
		fmt.Printf("Unable to create migration file: %s", err.Error())
	}
	defer f.Close()

	if _, err := f.WriteString(out.String()); err != nil {
		fmt.Printf("Unable to write migration file: %s", err.Error())
	}

	fmt.Println("Generated new migration file: ", f.Name())
	return nil
}

// NOTE: runs all migrations in a single transaction
func (m *Migrator) Up(step int) error {
	tx, err := m.db.BeginTx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		fmt.Printf("failed to initialize db transaction: %s\n", err.Error())
		return err
	}

	count := 0
	for _, v := range m.Versions {
		if step > 0 && count == step {
			break
		}

		mg := m.Migrations[v]

		if mg.processed {
			continue
		}

		fmt.Println("Running Migration: ", mg.Version)
		if err := mg.Up(tx); err != nil {
			tx.Rollback()
			fmt.Printf("Error running migration: %s\n", err.Error())
			return err
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations VALUES($1)",
			mg.Version); err != nil {
			fmt.Printf("Error running migration: %s\n", err.Error())
			tx.Rollback()
			return err
		}
		fmt.Println("Migration Succeeded")

		count++
	}

	tx.Commit()

	return nil
}

func (m *Migrator) Down(step int) error {
	tx, err := m.db.BeginTx(context.TODO(), &sql.TxOptions{})
	if err != nil {
		return err
	}

	count := 0
	for _, v := range reverse(m.Versions) {
		if step > 0 && count == step {
			break
		}

		mg := m.Migrations[v]

		if !mg.processed {
			continue
		}

		fmt.Println("Reverting Migration: ", mg.Version)
		if err := mg.Down(tx); err != nil {
			fmt.Printf("Error running migration: %s\n", err.Error())
			tx.Rollback()
			return err
		}

		if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = $1",
			mg.Version); err != nil {
			fmt.Printf("Error running migration: %s\n", err.Error())

			tx.Rollback()
			return err
		}
		fmt.Println("Revert Succeeded")

		count++
	}

	tx.Commit()

	return nil
}

func reverse(arr []string) []string {
	for i := 0; i < len(arr)/2; i++ {
		j := len(arr) - i - 1
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}
