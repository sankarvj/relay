package schema

import (
	"github.com/dimiro1/darwin"
	"github.com/jmoiron/sqlx"
)

// Migrate attempts to bring the schema for db up to date with the migrations
// defined in this package.
func Migrate(db *sqlx.DB) error {
	driver := darwin.NewGenericDriver(db.DB, darwin.PostgresDialect{})

	d := darwin.New(driver, migrations, nil)

	return d.Migrate()
}

// migrations contains the queries needed to construct the database schema.
// Entries should never be removed from this slice once they have been ran in
// production.
//
// Using constants in a .go file is an easy way to ensure the queries are part
// of the compiled executable and avoids pathing issues with the working
// directory. It has the downside that it lacks syntax highlighting and may be
// harder to read for some cases compared to using .sql files. You may also
// consider a combined approach using a tool like packr or go-bindata.
var migrations = []darwin.Migration{
	{
		Version:     1,
		Description: "Add accounts",
		Script: `
		CREATE TABLE accounts (
			account_id    UUID,
			name          TEXT,
			domain        TEXT UNIQUE,
			avatar        TEXT,
			plan          INTEGER DEFAULT 0,
			mode          INTEGER DEFAULT 0,
			timezone      TEXT,
			language      TEXT,
			country       TEXT,
			issued_at     TIMESTAMP,
			expiry        TIMESTAMP,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (account_id)
		);
		`,
	},
	{
		Version:     2,
		Description: "Add users",
		Script: `
		CREATE TABLE users (
			user_id       UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			name          TEXT,
			avatar 		  TEXT,
			email         TEXT,
			phone         TEXT,
			verified      BOOLEAN DEFAULT FALSE,
			roles         TEXT[],
			password_hash TEXT,
			provider      TEXT,
			issued_at     TIMESTAMP,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (user_id),
			UNIQUE (account_id, email)
		);
		`,
	},
	{
		Version:     3,
		Description: "Add teams",
		Script: `
		CREATE TABLE teams (
			team_id       BIGSERIAL PRIMARY KEY,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			name          TEXT,
			description   TEXT,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			UNIQUE (name)
		);
		`,
	},
	{
		Version:     4,
		Description: "Add members",
		Script: `
		CREATE TABLE members (
			member_id     BIGSERIAL PRIMARY KEY,
			team_id       BIGINT REFERENCES teams ON DELETE CASCADE,
			user_id       UUID REFERENCES users ON DELETE CASCADE,
			roles         TEXT[],
			created_at    TIMESTAMP,
			updated_at    BIGINT
		);
		`,
	},
	{
		Version:     5,
		Description: "Add entities",
		Script: `
		CREATE TABLE entities (
			entity_id     UUID,
			team_id       BIGINT REFERENCES teams ON DELETE CASCADE,
			name          TEXT,
			description   TEXT,
			category      INTEGER DEFAULT 0,
			state         INTEGER DEFAULT 0,
			mode          INTEGER DEFAULT 0,
			retry         INTEGER DEFAULT 0,
			attributes    JSONB,
			tags          TEXT[],
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (entity_id)
		);
		`,
	},
	{
		Version:     6,
		Description: "Add rules",
		Script: `
		CREATE TABLE rules (
			rule_id       UUID,
			entity_id     UUID REFERENCES entities ON DELETE CASCADE,
			expression    TEXT,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (rule_id)
		);
		`,
	},
	{
		Version:     7,
		Description: "Add items",
		Script: `
		CREATE TABLE items (
			item_id          UUID,
			parent_item_id   UUID,
			entity_id        UUID REFERENCES entities ON DELETE CASCADE,
			state            INTEGER DEFAULT 0,
			input            JSONB,
			created_at       TIMESTAMP,
			updated_at       BIGINT,
			PRIMARY KEY  (item_id)
		);
		`,
	},
}
