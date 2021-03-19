package schema

import (
	"github.com/jmoiron/sqlx"
)

const (
	SeedAccountID = "3cf17266-3473-4006-984f-9325122678b7"
	SeedTeamID    = "8cf27268-3473-4006-984f-9325122678b7"
	SeedUserID1   = "5cf37266-3473-4006-984f-9325122678b7"
	SeedUserID2   = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
	SeedUserID3   = "55b5fbd3-755f-4379-8f07-a58d4a30fa2f"
	SeedUserID4   = "65b5fbd3-755f-4379-8f07-a58d4a30fa2f"

	SeedFlowEntityName      = "flows"
	SeedNodeEntityName      = "nodes"
	SeedStatusEntityName    = "status"
	SeedContactsEntityName  = "contacts"
	SeedCompaniesEntityName = "companies"
	SeedTasksEntityName     = "tasks"
	SeedTicketsEntityName   = "tickets"
	SeedDealsEntityName     = "deals"
	SeedWebHookEntityName   = "hook"
	SeedDelayEntityName     = "delay"

	SeedFieldFNameKey = "uuid-00-fname"
	SeedFieldNPSKey   = "uuid-00-nps-score"
)

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(userSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	if _, err := tx.Exec(accountSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

func SeedUsers(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(userSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

//SeedEntity runs entity data
func SeedEntity(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(entityItemSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

//SeedWorkFlows runs workflows
func SeedWorkFlows(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(workflowSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

//SeedRelationShips runs pipeline flows
func SeedRelationShips(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(relationShipSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

//SeedPipelines runs pipeline flows
func SeedPipelines(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(pipelineSeeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// seeds is a string constant containing all of the queries needed to get the
// db seeded to a useful state for development.
//
// Note that database servers besides PostgreSQL may not support running
// multiple queries as part of the same execution so this single large constant
// may need to be broken up.

//TODO: this seed needs to get removed in the main project
const userSeeds = `
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '{3cf17266-3473-4006-984f-9325122678b7}', 'vijayasankarj', 'http://gravatar/vj', 'vijayasankarj@gmail.com', '9940209164', true, '{USER}', 'Zyg2U2ogVEafE7aaXXeQpYsI9G33', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('55b5fbd3-755f-4379-8f07-a58d4a30fa2f', '{3cf17266-3473-4006-984f-9325122678b7}', 'senthil', 'http://gravatar/vj', 'sksenthilkumaar@gmail.com', '9940209164', true, '{USER}', 'sk_replacetokenhere', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('65b5fbd3-755f-4379-8f07-a58d4a30fa2f', '{3cf17266-3473-4006-984f-9325122678b7}', 'saravana', 'http://gravatar/vj', 'saravanaprakas@gmail.com', '9940209164', true, '{USER}', 'sexy_replacetokenhere', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('5cf37266-3473-4006-984f-9325122678b7', '{3cf27266-3473-4006-984f-9325122678b7,3cf17266-3473-4006-984f-9325122678b7}', 'vijayasankarmail', 'http://gravatar/vj', 'vijayasankarmail@gmail.com', '9944293499', true, '{ADMIN,USER}', 'MYmfEIgwFYWrlKaDNJ0O3UNJSPg2', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1612675165);
`

const accountSeeds = `
INSERT INTO public.accounts (account_id, parent_account_id, name, domain, avatar, plan, mode, timezone, language, country, issued_at, expiry, created_at, updated_at) VALUES ('3cf17266-3473-4006-984f-9325122678b7', NULL, 'Wayplot', 'wayplot.com', NULL, 0, 0, NULL, NULL, NULL, '2021-01-10 14:53:12.100372', '2021-01-10 14:53:12.100372', '2021-01-10 14:53:12.100372', 1610290392);
INSERT INTO public.teams (team_id, account_id, name, description, created_at, updated_at) VALUES ('8cf27268-3473-4006-984f-9325122678b7', '3cf17266-3473-4006-984f-9325122678b7', 'CRM', 'The CRM App', '2021-01-10 14:53:12.104292', 1610290392);
`

const entityItemSeeds = `
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('f536c267-927b-4dda-9589-090ad436fb5b', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'deals', 'Deals', 1, 0, 0, '[{"key": "uuid-00-deal-name", "meta": null, "name": "deal_name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Deal Name"}, {"key": "uuid-00-deal-amount", "meta": null, "name": "deal_amount", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "N", "display_name": "Deal Amount"}, {"key": "uuid-00-contacts", "meta": {"display_gex": "uuid-00-fname"}, "name": "contact", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "907ff994-ca03-44b9-ab2b-4e96a691de2d", "choices": null, "dom_type": "MS", "action_id": "", "data_type": "R", "display_name": "Associated Contacts"}, {"key": "uuid-00-pipe", "meta": {"display_gex": "uuid-00-fname"}, "name": "pipeline_stage", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "437834ca-2dc3-4bdf-8d6f-27efb73d41f7", "choices": null, "dom_type": "PL", "action_id": "", "data_type": "O", "display_name": "Pipeline Stage"}]', NULL, '2021-02-07 05:19:30.150168', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('230e18f3-a7bf-4791-9434-df47f19f88aa', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'tasks', 'Tasks', 1, 0, 0, '[{"key": "uuid-00-desc", "meta": null, "name": "desc", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Notes"}, {"key": "uuid-00-contact", "meta": {"display_gex": "uuid-00-fname"}, "name": "contact", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "907ff994-ca03-44b9-ab2b-4e96a691de2d", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "Associated To"}, {"key": "uuid-00-status", "meta": {"display_gex": "uuid-00-name"}, "name": "status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "d240586e-0056-4ade-b331-9470a4a0706d", "choices": [{"id": "2cc69512-8f02-40ec-b506-61694b050f48", "expression": "{{self.uuid-00-due-by}} af {now}", "display_value": null}, {"id": "13b1676c-f752-4912-8c7f-da492e004583", "expression": "{{self.uuid-00-due-by}} bf {now}", "display_value": null}], "dom_type": "AS", "action_id": "", "data_type": "R", "display_name": "Status"}, {"key": "uuid-00-due-by", "meta": null, "name": "due_by", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "T", "display_name": "Due By"}, {"key": "uuid-00-reminder", "meta": null, "name": "reminder", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "uuid-00-contact", "data_type": "T", "display_name": "Reminder"}]', NULL, '2021-02-07 05:19:30.137735', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('0b0c0cf5-047b-4d6b-8d83-0d01ec9eb2c3', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'hook', 'WebHook', 2, 0, 0, '[{"key": "uuid-00-path", "meta": null, "name": "path", "field": null, "value": "/actuator/info", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "uuid-00-host", "meta": null, "name": "host", "field": null, "value": "https://stage.freshcontacts.io", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "uuid-00-method", "meta": null, "name": "method", "field": null, "value": "GET", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "uuid-00-headers", "meta": null, "name": "headers", "field": null, "value": "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}]', NULL, '2021-02-07 05:19:30.144479', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('57c3c70a-e5e7-4c2f-a390-2f55b8939c82', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'delay', 'Delay Timer', 7, 0, 0, '[{"key": "uuid-00-delay-by", "meta": null, "name": "delay_by", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "T", "display_name": ""}, {"key": "uuid-00-repeat", "meta": null, "name": "repeat", "field": null, "value": "true", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}]', NULL, '2021-02-07 05:19:30.14502', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('863e4d70-0a4a-4c36-b783-6ac0005abddf', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'tickets', 'Tickets', 1, 0, 0, '[{"key": "uuid-00-subject", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "uuid-00-status", "meta": {"display_gex": "uuid-00-name"}, "name": "status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "d240586e-0056-4ade-b331-9470a4a0706d", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "Status"}]', NULL, '2021-02-07 05:19:30.153114', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('6033a55d-fb80-4c60-af02-aeef898b2a09', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'owners', 'Owners', 5, 0, 0, '[{"key": "2da1c963-d5b9-416e-a9e0-c7380269c7e1", "meta": null, "name": "user_id", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": "User ID"}, {"key": "2d218795-7354-45c9-afd2-e6400c20a9a8", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "4919933f-abcd-4915-bd0c-60ce645b30cf", "meta": null, "name": "avatar", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Avatar"}, {"key": "a7f5bca1-0e44-4782-baee-f34908bd5bba", "meta": null, "name": "email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Email"}]', NULL, '2021-02-07 05:19:25.172694', 1612675165);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('f43655df-6199-445e-ac77-9fbda1a9a4a8', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'email_config', 'Email Integrations', 10, 0, 0, '[{"key": "17a18d8b-28f6-47c8-bfb7-10a30cff420f", "meta": {"config": "true"}, "name": "domain", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": "Domain"}, {"key": "e99d87c2-0b81-454c-b5f0-902577eb2a4c", "meta": {"config": "true"}, "name": "api_key", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": "API Key"}, {"key": "d5902718-61a4-480a-b243-a409b132833c", "meta": null, "name": "email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "E-Mail"}, {"key": "0878af23-4fe3-4759-b8c6-72919dbe007d", "meta": {"config": "true"}, "name": "common", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "db61cdb7-a151-44af-936a-49c3c56750e5", "meta": {"display_gex": "a7f5bca1-0e44-4782-baee-f34908bd5bba"}, "name": "owner", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "6033a55d-fb80-4c60-af02-aeef898b2a09", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "Associated To"}]', NULL, '2021-02-07 05:19:25.181269', 1612675165);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('82bb6968-8dcf-49cc-8aa9-b10c89b2e872', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'emails', 'Emails', 4, 0, 0, '[{"key": "80edd1ba-23b8-47e3-9e4a-bac2f4cd9f4e", "meta": {"display_gex": "d5902718-61a4-480a-b243-a409b132833c"}, "name": "from", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "f43655df-6199-445e-ac77-9fbda1a9a4a8", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "From"}, {"key": "fd3f209e-4b46-4029-8669-ba7498748d8a", "meta": {"display_gex": "uuid-00-email"}, "name": "to", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "907ff994-ca03-44b9-ab2b-4e96a691de2d", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "To"}, {"key": "34eb0205-9b9c-4adb-94ba-45e96c6c0423", "meta": {"display_gex": "d5902718-61a4-480a-b243-a409b132833c"}, "name": "cc", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "f43655df-6199-445e-ac77-9fbda1a9a4a8", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "Cc"}, {"key": "94076d73-d0f6-43fa-92ba-a2480b97fb11", "meta": {"display_gex": "d5902718-61a4-480a-b243-a409b132833c"}, "name": "bcc", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "f43655df-6199-445e-ac77-9fbda1a9a4a8", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "Bcc"}, {"key": "d2a93c91-7fd9-480c-a80b-dab5208b7792", "meta": null, "name": "subject", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Subject"}, {"key": "82c62226-7dfe-4c04-b207-2f9fe89f232e", "meta": null, "name": "body", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TA", "action_id": "", "data_type": "S", "display_name": "Body"}]', NULL, '2021-02-07 05:19:25.185425', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('d240586e-0056-4ade-b331-9470a4a0706d', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'status', 'Status', 8, 0, 0, '[{"key": "uuid-00-name", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "uuid-00-color", "meta": null, "name": "color", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Color"}]', NULL, '2021-02-07 05:19:30.122491', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('907ff994-ca03-44b9-ab2b-4e96a691de2d', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'contacts', 'Contacts', 1, 0, 0, '[{"key": "uuid-00-fname", "meta": {"layout": "title"}, "name": "first_name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "First Name"}, {"key": "uuid-00-email", "meta": {"layout": "sub-title"}, "name": "email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Email"}, {"key": "uuid-00-mobile-numbers", "meta": null, "name": "mobile_numbers", "field": {"key": "", "meta": null, "name": "", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "", "choices": null, "dom_type": "MS", "action_id": "", "data_type": "L", "display_name": "Mobile Numbers"}, {"key": "uuid-00-nps-score", "meta": null, "name": "nps_score", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "N", "display_name": "NPS Score"}, {"key": "uuid-00-lf-stage", "meta": null, "name": "lifecycle_stage", "field": null, "value": null, "ref_id": "", "choices": [{"id": "1", "expression": "", "display_value": "Lead"}, {"id": "2", "expression": "", "display_value": "Contact"}], "dom_type": "SE", "action_id": "", "data_type": "S", "display_name": "Lifecycle Stage"}, {"key": "uuid-00-status", "meta": {"display_gex": "uuid-00-name"}, "name": "status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "d240586e-0056-4ade-b331-9470a4a0706d", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "Status"}, {"key": "uuid-00-owner", "meta": {"display_gex": "a7f5bca1-0e44-4782-baee-f34908bd5bba"}, "name": "owner", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "6033a55d-fb80-4c60-af02-aeef898b2a09", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "Owner"}]', NULL, '2021-02-07 05:19:30.129476', 1612675170);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('dc7b871f-03a9-4171-a073-fe8537558ad8', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'companies', 'Companies', 1, 0, 0, '[{"key": "uuid-00-name", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "uuid-00-website", "meta": null, "name": "website", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Website"}]', NULL, '2021-02-07 05:19:30.136034', 1612675170);

INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('5cf37266-3473-4006-984f-9325122678b7', '3cf17266-3473-4006-984f-9325122678b7', '6033a55d-fb80-4c60-af02-aeef898b2a09', 0, '{"2d218795-7354-45c9-afd2-e6400c20a9a8": "vijayasankar", "2da1c963-d5b9-416e-a9e0-c7380269c7e1": "5cf37266-3473-4006-984f-9325122678b7", "4919933f-abcd-4915-bd0c-60ce645b30cf": "http://gravatar/vj", "a7f5bca1-0e44-4782-baee-f34908bd5bba": "vijayasankarmail@gmail.com"}', '2021-02-07 05:19:25.175346', 1612675165);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('2cc69512-8f02-40ec-b506-61694b050f48', '3cf17266-3473-4006-984f-9325122678b7', 'd240586e-0056-4ade-b331-9470a4a0706d', 0, '{"uuid-00-name": "Open", "uuid-00-color": "#fb667e"}', '2021-02-07 05:19:30.12462', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('74e94783-c4f1-478e-8611-d2586e16b1c0', '3cf17266-3473-4006-984f-9325122678b7', 'd240586e-0056-4ade-b331-9470a4a0706d', 0, '{"uuid-00-name": "Closed", "uuid-00-color": "#66fb99"}', '2021-02-07 05:19:30.12678', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('13b1676c-f752-4912-8c7f-da492e004583', '3cf17266-3473-4006-984f-9325122678b7', 'd240586e-0056-4ade-b331-9470a4a0706d', 0, '{"uuid-00-name": "OverDue", "uuid-00-color": "#66fb99"}', '2021-02-07 05:19:30.128091', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('1e0fb544-dae6-4c32-aa5a-ba48fdb096c6', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 0, '{"uuid-00-email": "vijayasankarmail@gmail.com", "uuid-00-fname": "Vijay", "uuid-00-owner": [], "uuid-00-status": ["2cc69512-8f02-40ec-b506-61694b050f48"], "uuid-00-lf-stage": "lead", "uuid-00-nps-score": 100, "uuid-00-mobile-numbers": ["9944293499", "9940209164"]}', '2021-02-07 05:19:30.132015', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('4da4e296-64b4-45f1-94b3-940d4df23b80', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 0, '{"uuid-00-email": "vijayasankarmail@gmail.com", "uuid-00-fname": "Senthil", "uuid-00-owner": [], "uuid-00-status": ["74e94783-c4f1-478e-8611-d2586e16b1c0"], "uuid-00-lf-stage": "lead", "uuid-00-nps-score": 100, "uuid-00-mobile-numbers": ["9944293499", "9940209164"]}', '2021-02-07 05:19:30.134316', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('84bdd778-9738-4e28-a1e4-72f6b983ecdf', '3cf17266-3473-4006-984f-9325122678b7', 'dc7b871f-03a9-4171-a073-fe8537558ad8', 0, '{"uuid-00-name": "Zoho", "uuid-00-website": "zoho.com"}', '2021-02-07 05:19:30.136617', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('d235476d-fd4c-4bdb-884e-14fbb1322043', '3cf17266-3473-4006-984f-9325122678b7', '230e18f3-a7bf-4791-9434-df47f19f88aa', 0, '{"uuid-00-desc": "make cake", "uuid-00-due-by": "2021-02-07 10:49:30 +0530", "uuid-00-status": [], "uuid-00-contact": ["1e0fb544-dae6-4c32-aa5a-ba48fdb096c6"], "uuid-00-reminder": "2021-02-07 10:49:30 +0530"}', '2021-02-07 05:19:30.139309', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('44e29003-26cf-4948-82cd-4b9fc973f505', '3cf17266-3473-4006-984f-9325122678b7', '230e18f3-a7bf-4791-9434-df47f19f88aa', 0, '{"uuid-00-desc": "make call", "uuid-00-due-by": "2021-02-07 10:49:30 +0530", "uuid-00-status": [], "uuid-00-contact": ["1e0fb544-dae6-4c32-aa5a-ba48fdb096c6"], "uuid-00-reminder": "2021-02-07 10:49:30 +0530"}', '2021-02-07 05:19:30.140929', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('5531c28a-d63a-42f5-a221-cb427796eaaa', '3cf17266-3473-4006-984f-9325122678b7', 'f43655df-6199-445e-ac77-9fbda1a9a4a8', 0, '{"0878af23-4fe3-4759-b8c6-72919dbe007d": "false", "17a18d8b-28f6-47c8-bfb7-10a30cff420f": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "d5902718-61a4-480a-b243-a409b132833c": "vijayasankar.jothi@wayplot.com", "db61cdb7-a151-44af-936a-49c3c56750e5": ["5cf37266-3473-4006-984f-9325122678b7"], "e99d87c2-0b81-454c-b5f0-902577eb2a4c": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35"}', '2021-02-07 05:19:30.14303', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('dba64e57-57a5-47e5-98f0-6d96d8fcd509', '3cf17266-3473-4006-984f-9325122678b7', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', 0, '{"34eb0205-9b9c-4adb-94ba-45e96c6c0423": ["vijayasankarmobile@gmail.com"], "80edd1ba-23b8-47e3-9e4a-bac2f4cd9f4e": ["5531c28a-d63a-42f5-a221-cb427796eaaa"], "82c62226-7dfe-4c04-b207-2f9fe89f232e": "Hello {{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-email}}", "94076d73-d0f6-43fa-92ba-a2480b97fb11": [], "d2a93c91-7fd9-480c-a80b-dab5208b7792": "This mail is sent you to tell that your NPS scrore is {{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-nps-score}}. We are very proud of you!", "fd3f209e-4b46-4029-8669-ba7498748d8a": ["{{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-email}}"]}', '2021-02-07 05:19:30.144121', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('bf96dd76-4bc5-4ed3-8708-a4062f9dbb0c', '3cf17266-3473-4006-984f-9325122678b7', '57c3c70a-e5e7-4c2f-a390-2f55b8939c82', 0, '{"uuid-00-repeat": "true", "uuid-00-delay-by": "2"}', '2021-02-07 05:19:30.145638', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('adc7ec48-8d10-482d-8212-558ee45fbb1b', '3cf17266-3473-4006-984f-9325122678b7', 'f536c267-927b-4dda-9589-090ad436fb5b', 0, '{"uuid-00-pipe": ["21e8f295-70a5-48c8-9b2d-8cad3c5d8310"], "uuid-00-contacts": ["1e0fb544-dae6-4c32-aa5a-ba48fdb096c6", "4da4e296-64b4-45f1-94b3-940d4df23b80"], "uuid-00-deal-name": "Big Deal", "uuid-00-deal-amount": 1000}', '2021-02-07 05:19:30.151207', 1612675170);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('37a287a8-ace2-4668-849c-049b515ec85b', '3cf17266-3473-4006-984f-9325122678b7', '863e4d70-0a4a-4c36-b783-6ac0005abddf', 0, '{"uuid-00-status": ["2cc69512-8f02-40ec-b506-61694b050f48"], "uuid-00-subject": "My Laptop Is Not Working"}', '2021-02-07 05:19:30.154061', 1612675170);
`

const workflowSeeds = `
INSERT INTO public.flows (flow_id, account_id, entity_id, expression, name, description, mode, type, condition, status, created_at, updated_at) VALUES ('437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', '{{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-fname}} eq {Vijay} && {{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-nps-score}} gt {98}', 'Sales Pipeline', '', 1, 0, 1, 0, '2021-02-07 05:19:30.146555', 1612675170);

INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '00000000-0000-0000-0000-000000000000', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000', 'opportunity','', 0, 7, '', '{}', '2021-02-07 05:19:30.1475', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('82adc579-b4df-48cc-a22e-dd42178d962c', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000','Deal Won','', 0, 7, '{Vijay} eq {Vijay}', '{}', '2021-02-07 05:19:30.148252', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('01eafec3-9a7e-4ba0-a56b-05eca43c8ba6', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '5531c28a-d63a-42f5-a221-cb427796eaaa', '00000000-0000-0000-0000-000000000000','','', 0, 3, '', '{"5531c28a-d63a-42f5-a221-cb427796eaaa": "dba64e57-57a5-47e5-98f0-6d96d8fcd509"}', '2021-02-07 05:19:30.148718', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('3214a000-a1dc-407a-a716-fda382f1c6d3', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '0b0c0cf5-047b-4d6b-8d83-0d01ec9eb2c3', '00000000-0000-0000-0000-000000000000','','', 0, 4, '', '{}', '2021-02-07 05:19:30.149221', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('07b3daab-2b2b-4d01-bafa-c5a05fcdc37b', '82adc579-b4df-48cc-a22e-dd42178d962c', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '57c3c70a-e5e7-4c2f-a390-2f55b8939c82', '00000000-0000-0000-0000-000000000000','','', 0, 6, '', '{"57c3c70a-e5e7-4c2f-a390-2f55b8939c82": "bf96dd76-4bc5-4ed3-8708-a4062f9dbb0c"}', '2021-02-07 05:19:30.149709', 1612675170);
`

const pipelineSeeds = `
INSERT INTO public.flows (flow_id, account_id, entity_id, expression, name, description, mode, type, condition, status, created_at, updated_at) VALUES ('437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', '{{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-fname}} eq {Vijay} && {{907ff994-ca03-44b9-ab2b-4e96a691de2d.uuid-00-nps-score}} gt {98}', 'Sales Pipeline', '', 1, 0, 1, 0, '2021-02-07 05:19:30.146555', 1612675170);

INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '00000000-0000-0000-0000-000000000000', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000','opportunity','', 0, 7, '', '{}', '2021-02-07 05:19:30.1475', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('82adc579-b4df-48cc-a22e-dd42178d962c', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000','Deal Won','', 0, 7, '{Vijay} eq {Vijay}', '{}', '2021-02-07 05:19:30.148252', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('01eafec3-9a7e-4ba0-a56b-05eca43c8ba6', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '5531c28a-d63a-42f5-a221-cb427796eaaa', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310','','', 0, 3, '', '{"5531c28a-d63a-42f5-a221-cb427796eaaa": "dba64e57-57a5-47e5-98f0-6d96d8fcd509"}', '2021-02-07 05:19:30.148718', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('3214a000-a1dc-407a-a716-fda382f1c6d3', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '0b0c0cf5-047b-4d6b-8d83-0d01ec9eb2c3', '21e8f295-70a5-48c8-9b2d-8cad3c5d8310','','', 0, 4, '', '{}', '2021-02-07 05:19:30.149221', 1612675170);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, stage_id, name, description, weight, type, expression, actuals, created_at, updated_at) VALUES ('07b3daab-2b2b-4d01-bafa-c5a05fcdc37b', '82adc579-b4df-48cc-a22e-dd42178d962c', '3cf17266-3473-4006-984f-9325122678b7', '437834ca-2dc3-4bdf-8d6f-27efb73d41f7', '57c3c70a-e5e7-4c2f-a390-2f55b8939c82', '82adc579-b4df-48cc-a22e-dd42178d962c','','', 0, 6, '', '{"57c3c70a-e5e7-4c2f-a390-2f55b8939c82": "bf96dd76-4bc5-4ed3-8708-a4062f9dbb0c"}', '2021-02-07 05:19:30.149709', 1612675170);
`

const relationShipSeeds = `
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('fcf3a068-6ad5-4229-a8b6-372e68d5790a', '3cf17266-3473-4006-984f-9325122678b7', 'f43655df-6199-445e-ac77-9fbda1a9a4a8', '6033a55d-fb80-4c60-af02-aeef898b2a09', 'db61cdb7-a151-44af-936a-49c3c56750e5', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('a1a25d9c-731d-4419-81b2-2f4e6521026c', '3cf17266-3473-4006-984f-9325122678b7', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', 'f43655df-6199-445e-ac77-9fbda1a9a4a8', '34eb0205-9b9c-4adb-94ba-45e96c6c0423', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('0520857e-97eb-41f6-9c15-3b61a178660e', '3cf17266-3473-4006-984f-9325122678b7', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', 'f43655df-6199-445e-ac77-9fbda1a9a4a8', '94076d73-d0f6-43fa-92ba-a2480b97fb11', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('ef23c29f-533e-4d06-a5ff-7d1aefbd7ac0', '3cf17266-3473-4006-984f-9325122678b7', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', 'f43655df-6199-445e-ac77-9fbda1a9a4a8', '80edd1ba-23b8-47e3-9e4a-bac2f4cd9f4e', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('970b68f9-2fb4-4f59-9ba8-31b747acb939', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 'd240586e-0056-4ade-b331-9470a4a0706d', 'uuid-00-status', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('ebb76c75-f7bd-4369-9008-7a466f512291', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', '6033a55d-fb80-4c60-af02-aeef898b2a09', 'uuid-00-owner', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('40fd58e0-82b0-4467-a218-3489f2a76a10', '3cf17266-3473-4006-984f-9325122678b7', '230e18f3-a7bf-4791-9434-df47f19f88aa', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 'uuid-00-contact', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('f6a32b19-1f3c-4309-8a07-8519c844d62e', '3cf17266-3473-4006-984f-9325122678b7', '230e18f3-a7bf-4791-9434-df47f19f88aa', 'd240586e-0056-4ade-b331-9470a4a0706d', 'uuid-00-status', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('09e4816b-bd3a-41ae-85b9-dd6175472095', '3cf17266-3473-4006-984f-9325122678b7', 'f536c267-927b-4dda-9589-090ad436fb5b', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 'uuid-00-contacts', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('711adb1e-b227-4808-800c-8da35920c5ec', '3cf17266-3473-4006-984f-9325122678b7', '863e4d70-0a4a-4c36-b783-6ac0005abddf', 'd240586e-0056-4ade-b331-9470a4a0706d', 'uuid-00-status', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('f9f2fac4-774c-4f50-b008-e5b23349596a', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 'dc7b871f-03a9-4171-a073-fe8537558ad8', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('f9f2fac4-774c-4f50-b008-e5b23349596a', '3cf17266-3473-4006-984f-9325122678b7', 'dc7b871f-03a9-4171-a073-fe8537558ad8', '907ff994-ca03-44b9-ab2b-4e96a691de2d', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('f8f856cd-1549-470a-9c99-800fadee7cbe', '3cf17266-3473-4006-984f-9325122678b7', '863e4d70-0a4a-4c36-b783-6ac0005abddf', 'f536c267-927b-4dda-9589-090ad436fb5b', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('f8f856cd-1549-470a-9c99-800fadee7cbe', '3cf17266-3473-4006-984f-9325122678b7', 'f536c267-927b-4dda-9589-090ad436fb5b', '863e4d70-0a4a-4c36-b783-6ac0005abddf', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('54e1e5e8-b9c0-4def-a16d-acde722f2407', '3cf17266-3473-4006-984f-9325122678b7', '863e4d70-0a4a-4c36-b783-6ac0005abddf', '907ff994-ca03-44b9-ab2b-4e96a691de2d', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('54e1e5e8-b9c0-4def-a16d-acde722f2407', '3cf17266-3473-4006-984f-9325122678b7', '907ff994-ca03-44b9-ab2b-4e96a691de2d', '863e4d70-0a4a-4c36-b783-6ac0005abddf', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('e538a97d-e59f-4805-8e94-2ed54fee2afc', '3cf17266-3473-4006-984f-9325122678b7', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', '907ff994-ca03-44b9-ab2b-4e96a691de2d', 'fd3f209e-4b46-4029-8669-ba7498748d8a', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('5a09faae-557e-4f9a-9721-8eb3c0af45e4', '3cf17266-3473-4006-984f-9325122678b7', 'f536c267-927b-4dda-9589-090ad436fb5b', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('5a09faae-557e-4f9a-9721-8eb3c0af45e4', '3cf17266-3473-4006-984f-9325122678b7', '82bb6968-8dcf-49cc-8aa9-b10c89b2e872', 'f536c267-927b-4dda-9589-090ad436fb5b', '00000000-0000-0000-0000-000000000000', 1);

INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '970b68f9-2fb4-4f59-9ba8-31b747acb939', '1e0fb544-dae6-4c32-aa5a-ba48fdb096c6', '2cc69512-8f02-40ec-b506-61694b050f48');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '970b68f9-2fb4-4f59-9ba8-31b747acb939', '4da4e296-64b4-45f1-94b3-940d4df23b80', '74e94783-c4f1-478e-8611-d2586e16b1c0');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '40fd58e0-82b0-4467-a218-3489f2a76a10', 'd235476d-fd4c-4bdb-884e-14fbb1322043', '1e0fb544-dae6-4c32-aa5a-ba48fdb096c6');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '40fd58e0-82b0-4467-a218-3489f2a76a10', '44e29003-26cf-4948-82cd-4b9fc973f505', '1e0fb544-dae6-4c32-aa5a-ba48fdb096c6');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '09e4816b-bd3a-41ae-85b9-dd6175472095', 'adc7ec48-8d10-482d-8212-558ee45fbb1b', '1e0fb544-dae6-4c32-aa5a-ba48fdb096c6');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '09e4816b-bd3a-41ae-85b9-dd6175472095', 'adc7ec48-8d10-482d-8212-558ee45fbb1b', '4da4e296-64b4-45f1-94b3-940d4df23b80');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '711adb1e-b227-4808-800c-8da35920c5ec', '37a287a8-ace2-4668-849c-049b515ec85b', '2cc69512-8f02-40ec-b506-61694b050f48');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', 'f9f2fac4-774c-4f50-b008-e5b23349596a', '1e0fb544-dae6-4c32-aa5a-ba48fdb096c6', '84bdd778-9738-4e28-a1e4-72f6b983ecdf');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', 'f8f856cd-1549-470a-9c99-800fadee7cbe', '37a287a8-ace2-4668-849c-049b515ec85b', 'adc7ec48-8d10-482d-8212-558ee45fbb1b');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '54e1e5e8-b9c0-4def-a16d-acde722f2407', '37a287a8-ace2-4668-849c-049b515ec85b', '1e0fb544-dae6-4c32-aa5a-ba48fdb096c6');
`
