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
		Description: "Initial DB schema",
		Script: `

		CREATE TABLE accounts (
			account_id    		UUID,
			parent_account_id   UUID,
			name          		TEXT UNIQUE,
			domain        		TEXT,
			avatar        		TEXT,
			mode          		INTEGER DEFAULT 0,
			cus_mail 			TEXT,
			cus_id				TEXT,
			cus_seat 			INTEGER DEFAULT 0,
			cus_status 			TEXT,
			cus_plan          	INTEGER DEFAULT 0,
			trail_start     	BIGINT,
			trail_end        	BIGINT,
			timezone      		TEXT,
			language      		TEXT,
			country       		TEXT,
			created_at    		TIMESTAMP,
			updated_at    		BIGINT,
			PRIMARY KEY (account_id)
		);
		CREATE INDEX idx_accounts_parent_account_id 
		ON accounts(parent_account_id);
		
		CREATE TABLE users (
			user_id       UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			member_id     UUID,
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
			UNIQUE (account_id,email)
		);
		CREATE INDEX idx_users_email
		ON users(email);
		CREATE INDEX idx_users_account_id
		ON users(account_id);
		
		
		CREATE TABLE drafts (
			draft_id    		UUID,
			account_name        TEXT,
			business_email      TEXT,
			host                TEXT,
			teams 				TEXT[],
			created_at    		TIMESTAMP,
			updated_at    		BIGINT,
			PRIMARY KEY (draft_id)
		);
		
		CREATE TABLE teams (
			team_id       UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			look_up       TEXT,
			name          TEXT,
			description   TEXT,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (team_id),
			UNIQUE (account_id,look_up)
		);
		CREATE INDEX idx_teams_account_id
		ON teams(account_id);
		
		CREATE TABLE entities (
			entity_id     		UUID,
			account_id    		UUID REFERENCES accounts ON DELETE CASCADE,
			team_id       		UUID REFERENCES teams ON DELETE CASCADE,
			name          		TEXT,
			display_name  		TEXT,
			category      		INTEGER DEFAULT 0,
			state         		INTEGER DEFAULT 0,
			status        		INTEGER DEFAULT 0,
			fieldsb       		JSONB,
			metab         		JSONB,
			tags          		TEXT[],
			is_public	  		BOOLEAN DEFAULT FALSE,
			is_core	      		BOOLEAN DEFAULT FALSE,
			is_shared	  		BOOLEAN DEFAULT FALSE,
			shared_team_ids  	TEXT[],
			created_at    		TIMESTAMP,
			updated_at    		BIGINT,
			PRIMARY KEY (entity_id),
			UNIQUE (team_id,display_name)
		);
		CREATE INDEX idx_entities_account_id
		ON entities(account_id);
		CREATE INDEX idx_entities_team_id
		ON entities(team_id);
		
		CREATE TABLE items (
			item_id          UUID,
			account_id       UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id        UUID REFERENCES entities ON DELETE CASCADE,
			genie_id         TEXT,
			user_id          UUID,
			stage_id         UUID,
			state            INTEGER DEFAULT 0,
			type             INTEGER DEFAULT 0,
			name             TEXT,
			fieldsb          JSONB,
			metab            JSONB,
			is_public	     BOOLEAN DEFAULT FALSE,
			created_at       TIMESTAMP,
			updated_at       BIGINT,
			PRIMARY KEY (item_id)
		);
		CREATE INDEX idx_items_account_id
		ON items(account_id);
		CREATE INDEX idx_items_entity_id
		ON items(entity_id);
		CREATE INDEX idx_items_account_entity_ids
		ON items(account_id,entity_id);
		CREATE INDEX idx_items_genie_id
		ON items(genie_id);
		
		CREATE TABLE flows (
			flow_id       UUID,
			account_id    UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id     UUID REFERENCES entities ON DELETE CASCADE,
			expression    TEXT,
			tokenb        JSONB,
			name    	  TEXT,
			description   TEXT,
			mode      	  INTEGER DEFAULT 0,
			state      	  INTEGER DEFAULT 0,
			type      	  INTEGER DEFAULT 0,
			condition     INTEGER DEFAULT 0,
			status        INTEGER DEFAULT 0,
			created_at    TIMESTAMP,
			updated_at    BIGINT,
			PRIMARY KEY (flow_id)
		);
		CREATE INDEX idx_flows_account_id
		ON flows(account_id);
		CREATE INDEX idx_flows_entity_id
		ON flows(entity_id);
		CREATE INDEX idx_flows_account_entity_ids
		ON flows(account_id,entity_id);
		
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
		CREATE INDEX idx_nodes_account_id
		ON nodes(account_id);
		CREATE INDEX idx_nodes_flow_id
		ON nodes(flow_id);
		
		CREATE TABLE active_flows (
			account_id  UUID,
			flow_id    	UUID REFERENCES flows ON DELETE CASCADE,
			item_id    	UUID REFERENCES items ON DELETE CASCADE,
			node_id    	UUID,
		    life 	   	INTEGER DEFAULT 0,
			is_active	BOOLEAN DEFAULT FALSE,
			created_at  TIMESTAMP,
			updated_at  BIGINT,
			UNIQUE (flow_id, item_id)
		);
		CREATE INDEX idx_active_flows_flow_id
		ON active_flows(flow_id);
		
		CREATE TABLE active_nodes (
			account_id  UUID,
			flow_id    	UUID REFERENCES flows ON DELETE CASCADE,
			entity_id   UUID REFERENCES entities ON DELETE CASCADE,
			item_id    	UUID REFERENCES items ON DELETE CASCADE,
			node_id    	UUID REFERENCES nodes ON DELETE CASCADE,
		    life 	   	INTEGER DEFAULT 0,
			is_active	BOOLEAN DEFAULT FALSE,
			created_at  TIMESTAMP,
			updated_at  BIGINT,
			UNIQUE (flow_id,entity_id,item_id,node_id)
		);
		CREATE INDEX idx_active_nodes_flow_id
		ON active_nodes(flow_id);
		
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
		CREATE INDEX idx_relationships_src_entity_id
		ON relationships(src_entity_id);
		CREATE INDEX idx_relationships_dst_entity_id
		ON relationships(dst_entity_id);
		
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
		CREATE INDEX idx_connections_account_id
		ON connections(account_id);
		CREATE INDEX idx_connections_relationship_id
		ON connections(relationship_id);
		CREATE INDEX idx_connections_src_entity_id
		ON connections(src_entity_id);
		CREATE INDEX idx_connections_dst_item_id
		ON connections(dst_item_id);
		
		CREATE TABLE discoveries (
			discovery_id    TEXT,
			discovery_type  TEXT,
			account_id  	UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id 	    UUID REFERENCES entities ON DELETE CASCADE,
			item_id 	    UUID REFERENCES items ON DELETE CASCADE,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			PRIMARY KEY 	(discovery_id),
			UNIQUE (account_id, entity_id, discovery_id, discovery_type)
		);
		CREATE INDEX idx_discoveries_discovery_id
		ON discoveries(discovery_id);
		CREATE INDEX idx_discoveries_account_discovery_entity_ids
		ON discoveries(account_id,discovery_id,entity_id);
		
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
		
		CREATE TABLE conversations (
			conversation_id  TEXT,
			account_id       UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id        UUID REFERENCES entities ON DELETE CASCADE,
			item_id          UUID REFERENCES items ON DELETE CASCADE,
			user_id          UUID, 
			type             INTEGER DEFAULT 0,
			state            INTEGER DEFAULT 0,
			message          TEXT,
			payload          JSONB,
			created_at       TIMESTAMP,
			updated_at       BIGINT,
			PRIMARY KEY (conversation_id),
			UNIQUE (account_id, conversation_id)
		);
		CREATE INDEX idx_conversations_account_entity_item_ids
		ON conversations(account_id,entity_id,item_id);
		
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
		CREATE INDEX idx_clients_account_user_ids
		ON clients(account_id,user_id);
		
		CREATE TABLE visitors (
			visitor_id      UUID,
			account_id      UUID REFERENCES accounts ON DELETE CASCADE,
			team_id         UUID REFERENCES teams ON DELETE CASCADE,
			entity_id    	UUID REFERENCES entities ON DELETE CASCADE,
			item_id 		UUID REFERENCES items ON DELETE CASCADE,
			name 			TEXT,
			email 			TEXT,
			token 			TEXT,
			active		    BOOLEAN DEFAULT TRUE,
			signed_in		BOOLEAN DEFAULT FALSE,
			expire_at       TIMESTAMP,
			created_at    	TIMESTAMP,
			updated_at    	BIGINT,
			UNIQUE (account_id, team_id, entity_id, item_id)
		);
		CREATE INDEX idx_visitors_visitor_id
		ON visitors(visitor_id);
		CREATE INDEX idx_visitors_account_id
		ON visitors(account_id);
		CREATE INDEX idx_visitors_email
		ON visitors(email);
		
		CREATE TABLE user_settings (
			account_id      		UUID REFERENCES accounts ON DELETE CASCADE,
			user_id    				UUID REFERENCES users ON DELETE CASCADE,
			layout_style    		TEXT,
			selected_team 			TEXT,
			notification_setting    JSONB,
			UNIQUE (account_id, user_id)
		);
		CREATE INDEX idx_user_settings_account_user_ids
		ON user_settings(account_id,user_id);
		
		CREATE TABLE log_streams (
			account_id      		UUID REFERENCES accounts ON DELETE CASCADE,
			log_id      		    TEXT,
			comment      		    TEXT,
			state    		        INTEGER DEFAULT 0,
			created_at    	        TIMESTAMP,
			UNIQUE (account_id, log_id)
		);
		CREATE INDEX idx_log_streams_account_log_ids
		ON log_streams(account_id,log_id);
		
		CREATE TABLE tokens (
			token    				TEXT,
			account_id      		UUID REFERENCES accounts ON DELETE CASCADE,
			type    		        INTEGER DEFAULT 0,
			state    		        INTEGER DEFAULT 0,
			scope					TEXT[],
			issued_at     		    TIMESTAMP,
			expiry        		    TIMESTAMP,
			created_at    	        TIMESTAMP,
			PRIMARY KEY (token)
		);
		CREATE INDEX idx_tokens_account_id
		ON tokens(account_id);

		CREATE TABLE timeseries (
			timeseries_id           UUID,
			account_id      		UUID REFERENCES accounts ON DELETE CASCADE,
			entity_id      		    UUID REFERENCES entities ON DELETE CASCADE,
			type    		        INTEGER DEFAULT 0,
			identifier    		    TEXT,
			tags					TEXT[],
			event     		        TEXT,
			description     		TEXT,
			count                   INTEGER DEFAULT 0,
			start_time              TIMESTAMP NOT NULL,
			end_time                TIMESTAMP NOT NULL,
			fieldsb          		JSONB,
			PRIMARY KEY (timeseries_id)
		);
		CREATE INDEX idx_timeseries_start_time
		ON timeseries(start_time);
		CREATE INDEX idx_timeseries_end_time
		ON timeseries(end_time);
		CREATE INDEX idx_timeseries_entity_id
		ON items(entity_id);
		CREATE INDEX idx_timeseries_identifier
		ON timeseries(identifier);
		CREATE INDEX idx_timeseries_event
		ON timeseries(event);
		CREATE INDEX idx_timeseries_tags
		ON timeseries(tags);


		CREATE TABLE dashboards (
			dashboard_id    	    UUID,
			account_id      		UUID REFERENCES accounts ON DELETE CASCADE,
			team_id                 UUID REFERENCES teams ON DELETE CASCADE,
		    entity_id      	        UUID,
			user_id                 UUID,
			name                    TEXT,
			type    		        TEXT,
			metab          		    JSONB,
			created_at    	        TIMESTAMP,
			PRIMARY KEY (dashboard_id)
		);
		CREATE INDEX idx_dashboards_account_id
		ON dashboards(account_id);
		CREATE INDEX idx_dashboards_entity_id
		ON dashboards(entity_id);


		CREATE TABLE charts (
			chart_id    			UUID,
			account_id      		UUID REFERENCES accounts ON DELETE CASCADE,
			team_id                 UUID REFERENCES teams ON DELETE CASCADE,
			entity_id      		    UUID REFERENCES entities ON DELETE CASCADE,
			dashboard_id      	    UUID REFERENCES dashboards ON DELETE CASCADE,
			name                    TEXT,
			display_name            TEXT,
			type    		        TEXT,
			duration    		    TEXT,
			state    		        INTEGER DEFAULT 0,
			position    		    INTEGER DEFAULT 0,
			metab          		    JSONB,
			created_at    	        TIMESTAMP,
			PRIMARY KEY (chart_id)
		);
		CREATE INDEX idx_charts_account_id
		ON charts(account_id);
		CREATE INDEX idx_charts_dashboard_id
		ON charts(dashboard_id);
		`,
	},
}
