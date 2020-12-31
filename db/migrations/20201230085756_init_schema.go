package migrations

import (
	"database/sql"
	"fmt"
)

func init() {
	migrator.AddMigration(&Migration{
		Version: "20201230085756",
		Up:      mig_20201230085756_init_schema_up,
		Down:    mig_20201230085756_init_schema_down,
	})
}

func mig_20201230085756_init_schema_up(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE IF NOT EXISTS lists (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS list_items (
		id SERIAL PRIMARY KEY,
		data TEXT,
		list_id INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (list_id) REFERENCES lists(id)
	);`)

	if err != nil {
		fmt.Println("Unable to execute migration", err)
		return err
	}

	return nil
}

func mig_20201230085756_init_schema_down(tx *sql.Tx) error {
	_, err := tx.Exec("DROP TABLE lists; DROP TABLE list_items;")

	if err != nil {
		panic("migration failed")
	}

	return nil
}
