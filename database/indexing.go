package database

import (
	"database/sql"
	"fmt"
)

func createGetNextFreeUserIDFunction(db *sql.DB) error {
	query := `
    CREATE OR REPLACE FUNCTION get_next_free_user_id()
    RETURNS INTEGER AS '
    DECLARE
        next_id INTEGER;
    BEGIN
        -- Find the maximum ID and add 1
        SELECT COALESCE(MAX(id), 0) + 1 INTO next_id
        FROM users;

        RETURN next_id;
    END;
    ' LANGUAGE plpgsql;
    `
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating get_next_free_user_id function: %w", err)
	}
	return nil
}

func createGetNextFreeTaskIDFunction(db *sql.DB) error {
	query := `
    CREATE OR REPLACE FUNCTION get_next_free_task_id()
    RETURNS INTEGER AS '
    DECLARE
        next_id INTEGER;
    BEGIN
        -- Find the maximum ID and add 1
        SELECT COALESCE(MAX(id), 0) + 1 INTO next_id
        FROM tasks;

        RETURN next_id;
    END;
    ' LANGUAGE plpgsql;
    `
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating get_next_free_task_id function: %w", err)
	}
	return nil
}
