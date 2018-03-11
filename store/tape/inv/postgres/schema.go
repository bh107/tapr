// Copyright 2018 Klaus Birkelund Abildgaard Jensen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
		'unknown', 'allocating', 'allocated', 'filling', 'scratch', 'full', 'missing', 'damaged', 'cleaning'
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

	`CREATE TABLE datasets (
		id integer PRIMARY KEY
	)`,

	`CREATE TABLE files (
		-- file path
		path text PRIMARY KEY,

		-- serial
		serial text,

		-- optional dataset relation
		dataset integer,

		-- constraints
		FOREIGN KEY (serial)  REFERENCES volumes  (serial),
		FOREIGN KEY (dataset) REFERENCES datasets (id)
	)`,
}
