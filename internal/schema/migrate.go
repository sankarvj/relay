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
			account_id    		UUID,
			parent_account_id   UUID,
			name          		TEXT UNIQUE,
			domain        		TEXT,
			avatar        		TEXT,
			plan          		INTEGER DEFAULT 0,
			mode          		INTEGER DEFAULT 0,
			timezone      		TEXT,
			language      		TEXT,
			country       		TEXT,
			issued_at     		TIMESTAMP,
			expiry        		TIMESTAMP,
			created_at    		TIMESTAMP,
			updated_at    		BIGINT,
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
			accounts     JSONB,
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
			UNIQUE (email)
		);
		`,
	},
	{
		Version:     3,
		Description: "Add drafts",
		Script: `
		CREATE TABLE drafts (
			draft_id    		UUID,
			account_name        TEXT,
			business_email      TEXT,
			teams 				TEXT[],
			created_at    		TIMESTAMP,
			updated_at    		BIGINT,
			PRIMARY KEY (draft_id)
		);
		`,
	},
	{
		Version:     4,
		Description: "Add teams",
		Script: `
		CREATE TABLE teams (
			team_id       UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			name          TEXT,
			description   TEXT,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (team_id),
			UNIQUE (account_id,name)
		);
		`,
	},
	{
		Version:     5,
		Description: "Add entities",
		Script: `
		CREATE TABLE entities (
			entity_id     UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			team_id       UUID REFERENCES teams ON DELETE CASCADE,
			name          TEXT,
			display_name  TEXT,
			category      INTEGER DEFAULT 0,
			state         INTEGER DEFAULT 0,
			status        INTEGER DEFAULT 0,
			fieldsb       JSONB,
			metab         JSONB,
			tags          TEXT[],
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (entity_id),
			UNIQUE (team_id,display_name)
		);
		`,
	},
	{
		Version:     6,
		Description: "Add items",
		Script: `
		CREATE TABLE items (
			item_id          UUID,
			account_id       UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id        UUID REFERENCES entities ON DELETE CASCADE,
			genie_id         UUID,
			user_id          UUID,
			stage_id         UUID,
			state            INTEGER DEFAULT 0,
			type             INTEGER DEFAULT 0,
			name             TEXT,
			fieldsb          JSONB,
			created_at       TIMESTAMP,
			updated_at       BIGINT,
			PRIMARY KEY (item_id)
		);
		`,
	},
	{
		Version:     7,
		Description: "Add flows",
		Script: `
		CREATE TABLE flows (
			flow_id       UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id     UUID REFERENCES entities ON DELETE CASCADE,
			expression    TEXT,
			tokenb        JSONB,
			name    	  TEXT,
			description   TEXT,
			mode      	  INTEGER DEFAULT 0,
			type      	  INTEGER DEFAULT 0,
			condition     INTEGER DEFAULT 0,
			status        INTEGER DEFAULT 0,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (flow_id)
		);
		`,
	},
	// In nodes it seems we are not using the stage_id effectively.
	// We are using the parent_id + type logic to vet stage nodes inside a stage node.
	{
		Version:     8,
		Description: "Add nodes",
		Script: `
		CREATE TABLE nodes (
			node_id       	UUID,
			parent_node_id  UUID,
			account_id      UUID,
			flow_id         UUID REFERENCES flows ON DELETE CASCADE,
			actor_id 	    UUID,
			stage_id        UUID, 
			name            TEXT,
			description     TEXT,
			weight          INTEGER DEFAULT 0,
			type			INTEGER DEFAULT 0,
			expression    	TEXT,
			tokenb          JSONB,
			actuals         JSONB,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			PRIMARY KEY (node_id)
		);
		`,
	},
	{
		Version:     9,
		Description: "Add active_flows",
		Script: `
		CREATE TABLE active_flows (
			account_id  UUID,
			flow_id    	UUID REFERENCES flows ON DELETE CASCADE,
			item_id    	UUID REFERENCES items ON DELETE CASCADE,
			node_id    	UUID,
		    life 	   	INTEGER DEFAULT 0,
			is_active	BOOLEAN DEFAULT FALSE,
			UNIQUE (flow_id,item_id)
		);
		`,
	},
	{
		Version:     10,
		Description: "Add active_nodes",
		Script: `
		CREATE TABLE active_nodes (
			account_id  UUID,
			flow_id    	UUID REFERENCES flows ON DELETE CASCADE,
			entity_id   UUID REFERENCES entities ON DELETE CASCADE,
			item_id    	UUID REFERENCES items ON DELETE CASCADE,
			node_id    	UUID REFERENCES nodes ON DELETE CASCADE,
		    life 	   	INTEGER DEFAULT 0,
			is_active	BOOLEAN DEFAULT FALSE,
			UNIQUE (flow_id,entity_id,item_id,node_id)
		);
		`,
	},
	{
		Version:     11, //TODO delete relationship on the update/delete of the field or make field_id as ON DELETE CASCADE
		Description: "Add relationships",
		Script: `
		CREATE TABLE relationships (
			relationship_id UUID,
			parent_rel_id   TEXT,
			account_id  	UUID REFERENCES accounts ON DELETE CASCADE,
	    	src_entity_id	UUID REFERENCES entities ON DELETE CASCADE,
			dst_entity_id   UUID REFERENCES entities ON DELETE CASCADE,
			field_id        TEXT, 
			type 	   	    INTEGER DEFAULT 0,
			position 	   	BIGINT,
			UNIQUE (account_id,src_entity_id,dst_entity_id,field_id)
		);
		`,
	},
	{
		Version:     12,
		Description: "Add connections",
		Script: `
		CREATE TABLE connections (
			connection_id   UUID,
			account_id  	UUID REFERENCES accounts ON DELETE CASCADE,
			user_id         UUID,
			relationship_id UUID,
			entity_name     TEXT,
			src_entity_id 	UUID REFERENCES entities ON DELETE CASCADE,
			dst_entity_id 	UUID REFERENCES entities ON DELETE CASCADE,
			src_item_id 	UUID REFERENCES items ON DELETE CASCADE,
			dst_item_id 	UUID REFERENCES items ON DELETE CASCADE,
			title           TEXT,
			sub_title       TEXT,
			action          TEXT,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			PRIMARY KEY     (connection_id)
		);
		`,
	},
	{
		Version:     13,
		Description: "Add discoveries",
		Script: `
		CREATE TABLE discoveries (
			discovery_id    TEXT,
			discovery_type  TEXT,
			account_id  	UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id 	    UUID REFERENCES entities ON DELETE CASCADE,
			item_id 	    UUID REFERENCES items ON DELETE CASCADE,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			UNIQUE (account_id, entity_id, discovery_id, discovery_type)
		);
		`,
	},
	{
		Version:     14,
		Description: "Add layouts",
		Script: `
		CREATE TABLE layouts (
			name            TEXT,
			account_id  	UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id 	    UUID REFERENCES entities ON DELETE CASCADE,
			user_id         UUID,
			fieldsb         JSONB,
			type 	   	    INTEGER DEFAULT 0,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			UNIQUE (account_id, entity_id, user_id, name)
		);
		`,
	},
	{
		Version:     15,
		Description: "Add conversations",
		Script: `
		CREATE TABLE conversations (
			conversation_id  UUID,
			account_id       UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id        UUID REFERENCES entities ON DELETE CASCADE,
			item_id          UUID,
			user_id          UUID, 
			type             INTEGER DEFAULT 0,
			message          TEXT,
			payload          JSONB,
			created_at       TIMESTAMP,
			updated_at       BIGINT,
			PRIMARY KEY (conversation_id)
		);
		`,
	},
	{
		Version:     16,
		Description: "Add clients",
		Script: `
		CREATE TABLE clients (
			account_id      UUID REFERENCES accounts ON DELETE CASCADE,
			user_id    		UUID REFERENCES users ON DELETE CASCADE,
			device_token    TEXT,
			device_type 	TEXT,
			status          INTEGER DEFAULT 0,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			UNIQUE (account_id, user_id, device_token)
		);
		`,
	},
}
