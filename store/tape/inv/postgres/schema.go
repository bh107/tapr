package postgres

// resetSchema is a list of SQL statements that when executed in sequence
// will reset the postgres database.
var resetSchema = []string{
	// drop types
	`DROP TYPE IF EXISTS volume_category CASCADE`,
	`DROP TYPE IF EXISTS slot_category CASCADE`,
	`DROP TYPE IF EXISTS volume_location CASCADE`,

	// drop tables
	`DROP TABLE IF EXISTS volumes`,
	`DROP TABLE IF EXISTS tree`,

	// create types
	`CREATE TYPE volume_category AS ENUM (
		-- order 'filling' before 'scratch'
		'unknown', 'allocating', 'filling', 'scratch', 'full', 'missing', 'damaged', 'cleaning'
	)`,

	`CREATE TYPE slot_category AS ENUM (
		'transfer', 'storage', 'ix'
	)`,

	`CREATE TYPE volume_location AS (
		addr integer,
		category slot_category
	)`,

	// create tables
	`CREATE TABLE volumes (
		-- the unique volume serial
		serial text PRIMARY KEY,

		-- volume location (unique, only one volume can occupy a slot)
		location volume_location UNIQUE,

		-- volume home location
		home volume_location UNIQUE,

		-- volume status
		category volume_category DEFAULT 'scratch',

		-- volume status
		flags bit varying(10)
	)`,

	`CREATE TABLE tree (
		path text PRIMARY KEY,
		serial text
	)`,
	/*
		`CREATE TABLE dirtree_closure (
			parent text,
			child text,

			-- constraints
			PRIMARY KEY (parent, child),
			FOREIGN KEY (parent) REFERENCES dirtree_nodes (path),
			FOREIGN KEY (child)  REFERENCES dirtree_nodes (path)
		)`,
	*/
}
