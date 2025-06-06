package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// -----------------------------------------------------------------------------
// Types
// -----------------------------------------------------------------------------

type Index struct {
	Name      string
	TableName string
	SQL       string
}

type DB struct {
	Connection *sql.DB
	Schema     Schema
}

type Schema struct {
	Tables   map[string]Table
	Indicies []Index
}

type Table struct {
	Name    string
	SQL     string
	Columns map[string]TableColumn
}

type TableColumn struct {
	Name         string
	Type         string
	NotNull      bool
	DefaultValue any
	PrimaryKey   bool
}

type TableColumns map[string]TableColumn

// -----------------------------------------------------------------------------
// Public
// -----------------------------------------------------------------------------

// NewDB creates a new DB object
// params can be a string or a slice of strings
// if params is a string, it is treated as the DSN
// if params is a slice, the first element is treated as the DSN
// if params is a slice, the second element is treated as the schema file
func NewDB(params []string) (db *DB) {
	var dsn string

	if len(params) > 0 {
		dsn = params[0]
	}

	db = &DB{
		Connection: connectDB(dsn),
		Schema: Schema{
			Tables: make(map[string]Table),
		},
	}

	if len(params) > 1 {
		schemaFile := params[1]
		if err := db.Exec(ReadSchemaFile(schemaFile)); err != nil {
			log.Fatal(err)
		}
	}

	return db
}

func connectDB(dsn string) (db *sql.DB) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func (db *DB) Close() (err error) {
	err = db.Connection.Close()
	return err
}

func (db *DB) GetSchema() Schema {
	rows, err := db.Connection.Query(`SELECT type, name, tbl_name, sql FROM sqlite_schema`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var colType string
		var name string
		var tblName string
		var createSQL string

		err := rows.Scan(&colType, &name, &tblName, &createSQL)
		if err != nil {
			log.Fatal(err)
		}

		switch colType {
		case "table":
			db.Schema.Tables[tblName] = Table{
				Name:    name,
				SQL:     createSQL,
				Columns: make(map[string]TableColumn),
			}

		case "index":
			db.Schema.Indicies = append(db.Schema.Indicies, Index{Name: name, TableName: tblName, SQL: createSQL})
		}
	}

	return db.Schema
}

func (db *DB) Exec(sql string) (err error) {
	_, err = db.Connection.Exec(sql)
	if err != nil {
		log.Printf("%q: %s\n", err, sql)
	}
	return err
}

func (db *DB) Query(sql string) (rows *sql.Rows, err error) {
	rows, err = db.Connection.Query(sql)
	if err != nil {
		log.Printf("%q: %s\n", err, sql)
	}
	return rows, err
}

func (db *DB) RemoveTables(kv map[string]Table) (err error) {
	for name := range kv {
		_, err := db.Connection.Exec("DROP TABLE " + name)
		if err != nil {
			log.Printf("%q: %s\n", err, name)
			return err
		}
	}

	return nil
}

func (db *DB) CreateTables(kv map[string]Table) (err error) {
	for _, table := range kv {
		err := db.Exec(table.SQL)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *DB) ApplySchemaChanges(CleanDB *DB) {
	newTables, tablesToDrop := Diff(CleanDB.GetSchema().Tables, db.GetSchema().Tables)

	err := db.RemoveTables(tablesToDrop)
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateTables(newTables)
	if err != nil {
		log.Fatal(err)
	}

	// New tables get new indicies
	for tableName := range newTables {
		newIndicies := CleanDB.GetSchema().GetTableIndices(tableName)
		for _, index := range newIndicies {
			err := db.Exec(index.SQL)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func (db *DB) DisableForeignKeys() {
	err := db.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		log.Fatal(err)
	}
}

func (db *DB) GetColumns(tableName string) TableColumns {
	// run the query to get the column
	rows, err := db.Query(`PRAGMA table_info(` + tableName + `)`)
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var id int
		var name string
		var coltype string
		var notnull int
		var dfltValue any
		var pk int

		err = rows.Scan(&id, &name, &coltype, &notnull, &dfltValue, &pk)
		if err != nil {
			log.Fatal(err)
		}

		db.Schema.Tables[tableName].Columns[name] = TableColumn{
			Name:         name,
			Type:         coltype,
			NotNull:      notnull == 1,
			DefaultValue: dfltValue,
			PrimaryKey:   pk == 1,
		}
	}

	return db.Schema.Tables[tableName].Columns
}

func (db *DB) findAlteredTables(CleanDB *DB) map[string]Table {
	alteredTables := make(map[string]Table)

	// Both schemas are cached before the tables were created/dropped
	// so we can compare the columns without filtering new ones out
	for name := range db.GetSchema().Tables {
		cleanTable, ok := CleanDB.GetSchema().Tables[name]
		if !ok {
			continue
		}

		CleanColumns := CleanDB.GetColumns(name)
		CurrentColumns := db.GetColumns(name)

		add, remove := Diff(CleanColumns, CurrentColumns)

		if len(add) > 0 || len(remove) > 0 {
			alteredTables[name] = cleanTable
		}
	}

	return alteredTables
}

func (schema Schema) GetTableIndices(tableName string) map[string]Index {
	tableIndicies := make(map[string]Index)

	for i := 0; i < len(schema.Indicies); i++ {
		if schema.Indicies[i].TableName == tableName {
			tableIndicies[schema.Indicies[i].Name] = schema.Indicies[i]
		}
	}

	return tableIndicies
}

func Migrate(db *sql.DB, schema string) {
	fmt.Println("migrating...")
	CurrentDB := DB{
		Connection: db,
		Schema: Schema{
			Tables: make(map[string]Table),
		},
	}

	// Temporary In Memory DB - Based on the schema.sql file
	CleanDB := NewDB([]string{"file:clean.sqlite3?mode=memory", schema})
	defer CleanDB.Close()

	// Apply schema changes (create tables/indices, drop tables/indices)
	CurrentDB.ApplySchemaChanges(CleanDB)

	// 1. Disable foreign keys
	CurrentDB.DisableForeignKeys()

	for tableName, table := range CurrentDB.findAlteredTables(CleanDB) {
		fmt.Println("found altered table " + tableName)

		// 2. Start transaction
		tx, err := CurrentDB.Connection.Begin()
		if err != nil {
			log.Fatal(err)
		}

		// 3. Define create table statement with new name
		tableNameNew := tableName + "_new"

		// 4. Create new tables
		createTable(tx, tableName, tableNameNew, table)

		// 5. Transfer table contents to new table
		migrateContent(tx, CleanDB, &CurrentDB, tableName, tableNameNew)

		// 6. Drop old table
		dropTable(tx, tableName)

		// 7. Rename new table to old table
		renameTable(tx, tableName, tableNameNew)

		// 8. Use CREATE INDEX, CREATE TRIGGER, and CREATE VIEW to reconstruct indexes, triggers, and views associated with table X. Perhaps use the old format of the triggers, indexes, and views saved from step 3 above as a guide, making changes as appropriate for the alteration.
		createIndicesOnTable(tx, tableName, CleanDB)

		// 9. If any views refer to table X in a way that is affected by the schema change, then drop those views using DROP VIEW and recreate them with whatever changes are necessary to accommodate the schema change using CREATE VIEW.

		// 10. If foreign key constraints were originally enabled then run PRAGMA foreign_key_check to verify that the schema change did not break any foreign key constraints.
		err = CurrentDB.Exec("PRAGMA foreign_key_check")
		if err != nil {
			log.Fatal(err)
		}

		// 11. End transaction
		err = tx.Commit()
	}

	// 12. Enable foreign keys again
	err := CurrentDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("migration complete.")
}

func createIndicesOnTable(tx *sql.Tx, tableName string, CleanDB *DB) {
	indices := CleanDB.GetSchema().GetTableIndices(tableName)
	for _, index := range indices {
		_, err := tx.Exec(index.SQL)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func renameTable(tx *sql.Tx, tableName string, tableNameNew string) {
	fmt.Println("renaming table " + tableName)
	_, err := tx.Exec("ALTER TABLE "+tableNameNew+" RENAME TO "+tableName, tableNameNew, tableName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("renamed " + tableNameNew + " to " + tableName)
}

func createTable(tx *sql.Tx, tableName string, tableNameNew string, table Table) {
	fmt.Println("creating table " + tableNameNew)
	stmt := strings.Replace(table.SQL, tableName, tableNameNew, 1)

	fmt.Println(stmt)
	_, err := tx.Exec(stmt)
	if err != nil {
		log.Fatal(err)
	}
}

func dropTable(tx *sql.Tx, tableName string) {
	fmt.Println("dropping " + tableName)
	_, err := tx.Exec(`DROP TABLE IF EXISTS ` + tableName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("dropped " + tableName)
}

func migrateContent(tx *sql.Tx, CleanDB *DB, CurrentDB *DB, tableName string, tableNameNew string) {
	fmt.Println("migrating content from " + tableName + " to " + tableNameNew + "...")
	intersection := Intersect(
		CleanDB.GetColumns(tableName),
		CurrentDB.GetColumns(tableName),
	)
	if len(intersection) == 0 {
		return
	}

	cols := strings.Join(intersection[:], ", ")
	_, err := tx.Exec("INSERT INTO " + tableNameNew + " (" + cols + ") SELECT " + cols + " FROM " + tableName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("inserted " + tableNameNew)
}

// -----------------------------------------------------------------------------
// Utilities
// -----------------------------------------------------------------------------

func Diff[T any](a, b map[string]T) (add, remove map[string]T) {
	add = make(map[string]T)
	remove = make(map[string]T)

	for k := range a {
		_, ok := b[k]
		if !ok {
			add[k] = a[k]
		}
	}

	for k := range b {
		_, ok := a[k]
		if !ok {
			remove[k] = b[k]
		}
	}

	return add, remove
}

func Intersect[T any](a, b map[string]T) []string {
	intersection := []string{}

	if len(a) > len(b) {
		a, b = b, a
	}

	for k := range a {
		_, ok := b[k]
		if ok {
			intersection = append(intersection, k)
		}
	}

	return intersection
}

func ReadSchemaFile(f string) string {
	b, err := os.ReadFile(f)
	if err != nil {
		log.Fatal(err)
	}

	return string(b)
}
