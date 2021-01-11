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
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '{3cf27266-3473-4006-984f-9325122678b7}', 'vijay', 'http://gravatar/vj', 'vijayasankarj@gmail.com', '9940209164', true, '{USER}', 'Zyg2U2ogVEafE7aaXXeQpYsI9G33', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('55b5fbd3-755f-4379-8f07-a58d4a30fa2f', '{3cf27266-3473-4006-984f-9325122678b7}', 'senthil', 'http://gravatar/vj', 'sksenthilkumaar@gmail.com', '9940209164', true, '{USER}', 'sk_replacetokenhere', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('65b5fbd3-755f-4379-8f07-a58d4a30fa2f', '{3cf27266-3473-4006-984f-9325122678b7}', 'saravana', 'http://gravatar/vj', 'saravanaprakas@gmail.com', '9940209164', true, '{USER}', 'sexy_replacetokenhere', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_ids, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('5cf37266-3473-4006-984f-9325122678b7', '{3cf27266-3473-4006-984f-9325122678b7}', 'vijayasankar', 'http://gravatar/vj', 'vijayasankarmail@gmail.com', '9944293499', true, '{ADMIN,USER}', 'MYmfEIgwFYWrlKaDNJ0O3UNJSPg2', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1609760492);
`

const accountSeeds = `
INSERT INTO public.accounts (account_id, parent_account_id, name, domain, avatar, plan, mode, timezone, language, country, issued_at, expiry, created_at, updated_at) VALUES ('3cf17266-3473-4006-984f-9325122678b7', NULL, 'Wayplot', 'wayplot.com', NULL, 0, 0, NULL, NULL, NULL, '2021-01-10 14:53:12.100372', '2021-01-10 14:53:12.100372', '2021-01-10 14:53:12.100372', 1610290392);
INSERT INTO public.teams (team_id, account_id, name, description, created_at, updated_at) VALUES ('8cf27268-3473-4006-984f-9325122678b7', '3cf17266-3473-4006-984f-9325122678b7', 'CRM', '', '2021-01-10 14:53:12.104292', 1610290392);
`

const entityItemSeeds = `
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('793cf01b-6efc-4c01-b374-a6ee77fbe96b', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'owners', 'Owners', 5, 0, 0, '[{"key": "64307bf9-dc8a-4931-a3ee-9e18006a4ae7", "meta": null, "name": "user_id", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": "User ID"}, {"key": "931001ef-569b-4714-b463-384782bf4dbd", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "2bd1b96a-1bc5-4234-9816-1ecbd5f55551", "meta": null, "name": "avatar", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Avatar"}, {"key": "f08b8e06-2193-46ff-a71c-ea959464904d", "meta": null, "name": "email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Email"}]', NULL, '2021-01-10 14:53:12.10663', 1610290392);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('57ee876a-1083-440b-a6de-f4c2a8500523', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'emails_config', 'Email Integrations', 10, 0, 0, '[{"key": "dfbf6da5-56d0-4f88-9348-bebfd6e44944", "meta": null, "name": "domain", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": "Domain"}, {"key": "ce3ea60b-5511-4f41-9bd7-67f0e480d883", "meta": null, "name": "api_key", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": "API Key"}, {"key": "6e98a2ca-a28e-420d-87d0-84260ec5fa20", "meta": null, "name": "email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "E-Mail"}, {"key": "bf231d69-3afd-46e6-9d9a-f62eaf6ffe6b", "meta": null, "name": "common", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "c668edcc-c24d-4ea6-9efa-9eeccb2d22ab", "meta": {"display_gex": "f08b8e06-2193-46ff-a71c-ea959464904d"}, "name": "owner", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "793cf01b-6efc-4c01-b374-a6ee77fbe96b", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "Associated To"}]', NULL, '2021-01-10 14:53:12.113862', 1610290392);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('447a9c03-ad0c-4543-9dcf-fce1a8fbceed', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'emails', 'Emails', 4, 0, 0, '[{"key": "467f2610-1468-4641-a5ae-e0e5e73907e0", "meta": {"config": "true"}, "name": "config", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "57ee876a-1083-440b-a6de-f4c2a8500523", "choices": null, "dom_type": "NA", "action_id": "", "data_type": "R", "display_name": ""}, {"key": "863a7bb2-5081-420b-8366-17b5197b0e11", "meta": {"display_gex": "c668edcc-c24d-4ea6-9efa-9eeccb2d22ab"}, "name": "from", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "57ee876a-1083-440b-a6de-f4c2a8500523", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "From"}, {"key": "6c8116ca-8962-47e2-b79e-11d09f6323e6", "meta": null, "name": "to", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "To"}, {"key": "fa6687cb-6f07-4d95-8145-f027a55a3bb5", "meta": null, "name": "cc", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Cc"}, {"key": "edb89249-83a5-4c3c-af94-7ee38de10893", "meta": null, "name": "bcc", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Bcc"}, {"key": "9a128d7c-7853-48c3-bb98-94a24e6e51df", "meta": null, "name": "subject", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Subject"}, {"key": "e95c0df4-2caa-4285-8bea-dba3c413ba1c", "meta": null, "name": "body", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Body"}]', NULL, '2021-01-10 14:53:12.116986', 1610290392);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('224e4a30-e2db-47c8-94ff-30186dbd3aa5', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Status', 8, 0, 0, '[{"key": "uuid-00-name", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "uuid-00-color", "meta": null, "name": "color", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Color"}]', NULL, '2021-01-10 14:53:15.948294', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Contacts', 1, 0, 0, '[{"key": "uuid-00-fname", "meta": {"layout": "title"}, "name": "first_name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "First Name"}, {"key": "uuid-00-email", "meta": {"layout": "sub-title"}, "name": "email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Email"}, {"key": "uuid-00-mobile-numbers", "meta": null, "name": "mobile_numbers", "field": {"key": "", "meta": null, "name": "", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "", "choices": null, "dom_type": "MS", "action_id": "", "data_type": "L", "display_name": "Mobile Numbers"}, {"key": "uuid-00-nps-score", "meta": null, "name": "nps_score", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "N", "display_name": "NPS Score"}, {"key": "uuid-00-lf-stage", "meta": null, "name": "lifecycle_stage", "field": null, "value": null, "ref_id": "", "choices": [{"id": "1", "expression": "", "display_value": "Lead"}, {"id": "2", "expression": "", "display_value": "Contact"}], "dom_type": "SE", "action_id": "", "data_type": "S", "display_name": "Lifecycle Stage"}, {"key": "uuid-00-status", "meta": {"display_gex": "uuid-00-name"}, "name": "status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "224e4a30-e2db-47c8-94ff-30186dbd3aa5", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "Status"}, {"key": "uuid-00-owner", "meta": {"display_gex": "name"}, "name": "owner", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "793cf01b-6efc-4c01-b374-a6ee77fbe96b", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "Owner"}]', NULL, '2021-01-10 14:53:15.953932', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('f24b9868-654f-46ca-b444-b15032cf613b', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Companies', 1, 0, 0, '[{"key": "uuid-00-name", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "uuid-00-website", "meta": null, "name": "website", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Website"}]', NULL, '2021-01-10 14:53:15.960859', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('86ae76d6-b62d-412b-96be-a02d55606560', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Tasks', 1, 0, 0, '[{"key": "uuid-00-desc", "meta": null, "name": "desc", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Notes"}, {"key": "uuid-00-contact", "meta": {"display_gex": "uuid-00-fname"}, "name": "contact", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44", "choices": null, "dom_type": "AC", "action_id": "", "data_type": "R", "display_name": "Associated To"}, {"key": "uuid-00-status", "meta": {"display_gex": "uuid-00-name"}, "name": "status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "224e4a30-e2db-47c8-94ff-30186dbd3aa5", "choices": [{"id": "42dc2c2a-c2cd-4b36-a0d1-4cd2478952d6", "expression": "{{self.uuid-00-due-by}} af {now}", "display_value": null}, {"id": "f1d8944e-dfdf-49b6-8fef-0d89b3a4f5e6", "expression": "{{self.uuid-00-due-by}} bf {now}", "display_value": null}], "dom_type": "AS", "action_id": "", "data_type": "R", "display_name": "Status"}, {"key": "uuid-00-due-by", "meta": null, "name": "due_by", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "T", "display_name": "Due By"}, {"key": "uuid-00-reminder", "meta": null, "name": "reminder", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "uuid-00-contact", "data_type": "T", "display_name": "Reminder"}]', NULL, '2021-01-10 14:53:15.962656', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('43c7ceb7-2f9c-4e23-a1d7-f0a7c9e7f545', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'WebHook', 2, 0, 0, '[{"key": "uuid-00-path", "meta": null, "name": "path", "field": null, "value": "/actuator/info", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "uuid-00-host", "meta": null, "name": "host", "field": null, "value": "https://stage.freshcontacts.io", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "uuid-00-method", "meta": null, "name": "method", "field": null, "value": "GET", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}, {"key": "uuid-00-headers", "meta": null, "name": "headers", "field": null, "value": "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}]', NULL, '2021-01-10 14:53:15.969105', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('74075e0f-57f7-4f99-b3e0-77690b83f117', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Delay Timer', 7, 0, 0, '[{"key": "uuid-00-delay-by", "meta": null, "name": "delay_by", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "T", "display_name": ""}, {"key": "uuid-00-repeat", "meta": null, "name": "repeat", "field": null, "value": "true", "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": ""}]', NULL, '2021-01-10 14:53:15.969931', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('294530fa-2689-4c12-b021-eef7745b7c3d', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Deals', 1, 0, 0, '[{"key": "uuid-00-deal-name", "meta": null, "name": "deal_name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Deal Name"}, {"key": "uuid-00-deal-amount", "meta": null, "name": "deal_amount", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "N", "display_name": "Deal Amount"}, {"key": "uuid-00-contacts", "meta": {"display_gex": "uuid-00-fname"}, "name": "contact", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44", "choices": null, "dom_type": "MS", "action_id": "", "data_type": "R", "display_name": "Associated Contacts"}, {"key": "uuid-00-pipe", "meta": {"display_gex": "uuid-00-fname"}, "name": "pipeline_stage", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "105c542b-d863-4036-8ac6-7f14d91fe4a2", "choices": null, "dom_type": "PL", "action_id": "", "data_type": "O", "display_name": "Pipeline Stage"}]', NULL, '2021-01-10 14:53:15.974667', 1610290395);
INSERT INTO public.entities (entity_id, account_id, team_id, name, display_name, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('db380265-2e73-455b-a057-58c6c9390681', '3cf17266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', '', 'Tickets', 1, 0, 0, '[{"key": "uuid-00-subject", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "action_id": "", "data_type": "S", "display_name": "Name"}, {"key": "uuid-00-status", "meta": {"display_gex": "uuid-00-name"}, "name": "status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "action_id": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "224e4a30-e2db-47c8-94ff-30186dbd3aa5", "choices": null, "dom_type": "SE", "action_id": "", "data_type": "R", "display_name": "Status"}]', NULL, '2021-01-10 14:53:15.977143', 1610290395);

INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('5cf37266-3473-4006-984f-9325122678b7', '3cf17266-3473-4006-984f-9325122678b7', '793cf01b-6efc-4c01-b374-a6ee77fbe96b', 0, '{"2bd1b96a-1bc5-4234-9816-1ecbd5f55551": "http://gravatar/vj", "64307bf9-dc8a-4931-a3ee-9e18006a4ae7": "5cf37266-3473-4006-984f-9325122678b7", "931001ef-569b-4714-b463-384782bf4dbd": "vijayasankar", "f08b8e06-2193-46ff-a71c-ea959464904d": "vijayasankarmail@gmail.com"}', '2021-01-10 14:53:12.109105', 1610290392);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('42dc2c2a-c2cd-4b36-a0d1-4cd2478952d6', '3cf17266-3473-4006-984f-9325122678b7', '224e4a30-e2db-47c8-94ff-30186dbd3aa5', 0, '{"uuid-00-name": "Open", "uuid-00-color": "#fb667e"}', '2021-01-10 14:53:15.95024', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('cee462a0-d9ca-441b-8078-f14e14860626', '3cf17266-3473-4006-984f-9325122678b7', '224e4a30-e2db-47c8-94ff-30186dbd3aa5', 0, '{"uuid-00-name": "Closed", "uuid-00-color": "#66fb99"}', '2021-01-10 14:53:15.951969', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('f1d8944e-dfdf-49b6-8fef-0d89b3a4f5e6', '3cf17266-3473-4006-984f-9325122678b7', '224e4a30-e2db-47c8-94ff-30186dbd3aa5', 0, '{"uuid-00-name": "OverDue", "uuid-00-color": "#66fb99"}', '2021-01-10 14:53:15.952959', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('9d9ab317-897d-4297-9818-088674ce497e', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', 0, '{"uuid-00-email": "vijayasankarmail@gmail.com", "uuid-00-fname": "Vijay", "uuid-00-owner": [], "uuid-00-status": ["42dc2c2a-c2cd-4b36-a0d1-4cd2478952d6"], "uuid-00-lf-stage": "lead", "uuid-00-nps-score": 100, "uuid-00-mobile-numbers": ["9944293499", "9940209164"]}', '2021-01-10 14:53:15.956451', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('19496230-0793-4669-b1a3-cef4c8bdbdad', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', 0, '{"uuid-00-email": "vijayasankarmail@gmail.com", "uuid-00-fname": "Senthil", "uuid-00-owner": [], "uuid-00-status": ["cee462a0-d9ca-441b-8078-f14e14860626"], "uuid-00-lf-stage": "lead", "uuid-00-nps-score": 100, "uuid-00-mobile-numbers": ["9944293499", "9940209164"]}', '2021-01-10 14:53:15.959214', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('9a722df4-78a1-4a5b-a9dd-b98d48702b08', '3cf17266-3473-4006-984f-9325122678b7', 'f24b9868-654f-46ca-b444-b15032cf613b', 0, '{"uuid-00-name": "Zoho", "uuid-00-website": "zoho.com"}', '2021-01-10 14:53:15.961697', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('dba4f6ef-3a57-4865-bf38-7b19a84b580b', '3cf17266-3473-4006-984f-9325122678b7', '86ae76d6-b62d-412b-96be-a02d55606560', 0, '{"uuid-00-desc": "make cake", "uuid-00-due-by": "2021-01-10 20:23:15 +0530", "uuid-00-status": [], "uuid-00-contact": ["9d9ab317-897d-4297-9818-088674ce497e"], "uuid-00-reminder": "2021-01-10 20:23:15 +0530"}', '2021-01-10 14:53:15.964698', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('f4626250-03d5-4c69-b44e-0b9d018206e6', '3cf17266-3473-4006-984f-9325122678b7', '86ae76d6-b62d-412b-96be-a02d55606560', 0, '{"uuid-00-desc": "make call", "uuid-00-due-by": "2021-01-10 20:23:15 +0530", "uuid-00-status": [], "uuid-00-contact": ["9d9ab317-897d-4297-9818-088674ce497e"], "uuid-00-reminder": "2021-01-10 20:23:15 +0530"}', '2021-01-10 14:53:15.966015', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('ef73833b-d52a-4fdb-b5b4-59e409508d44', '3cf17266-3473-4006-984f-9325122678b7', '57ee876a-1083-440b-a6de-f4c2a8500523', 0, '{"6e98a2ca-a28e-420d-87d0-84260ec5fa20": "vijayasankar.jothi@wayplot.com", "bf231d69-3afd-46e6-9d9a-f62eaf6ffe6b": "false", "c668edcc-c24d-4ea6-9efa-9eeccb2d22ab": "5cf37266-3473-4006-984f-9325122678b7", "ce3ea60b-5511-4f41-9bd7-67f0e480d883": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "dfbf6da5-56d0-4f88-9348-bebfd6e44944": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org"}', '2021-01-10 14:53:15.967869', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('ef068193-426d-4155-b00b-378c8fcc82be', '3cf17266-3473-4006-984f-9325122678b7', '447a9c03-ad0c-4543-9dcf-fce1a8fbceed', 0, '{"467f2610-1468-4641-a5ae-e0e5e73907e0": "ef73833b-d52a-4fdb-b5b4-59e409508d44", "6c8116ca-8962-47e2-b79e-11d09f6323e6": "{{17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44.uuid-00-email}}", "863a7bb2-5081-420b-8366-17b5197b0e11": "ef73833b-d52a-4fdb-b5b4-59e409508d44", "9a128d7c-7853-48c3-bb98-94a24e6e51df": "This mail is sent you to tell that your NPS scrore is {{17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44.uuid-00-nps-score}}. We are very proud of you!", "e95c0df4-2caa-4285-8bea-dba3c413ba1c": "Hello {{17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44.uuid-00-email}}", "edb89249-83a5-4c3c-af94-7ee38de10893": "", "fa6687cb-6f07-4d95-8145-f027a55a3bb5": "vijayasankarmobile@gmail.com"}', '2021-01-10 14:53:15.968756', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('d202dc5d-fa75-49ab-9ebb-b2609654322c', '3cf17266-3473-4006-984f-9325122678b7', '74075e0f-57f7-4f99-b3e0-77690b83f117', 0, '{"uuid-00-repeat": "true", "uuid-00-delay-by": "2"}', '2021-01-10 14:53:15.970495', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('7bbe2be3-6d0e-4f91-a37a-46158cdd0eb9', '3cf17266-3473-4006-984f-9325122678b7', '294530fa-2689-4c12-b021-eef7745b7c3d', 0, '{"uuid-00-pipe": ["6b5dc44e-98e3-4193-9cb3-faf4298fbe6e"], "uuid-00-contacts": ["9d9ab317-897d-4297-9818-088674ce497e", "19496230-0793-4669-b1a3-cef4c8bdbdad"], "uuid-00-deal-name": "Big Deal", "uuid-00-deal-amount": 1000}', '2021-01-10 14:53:15.975635', 1610290395);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('e7f5f01f-aa38-47a9-bbf1-b516f65b130b', '3cf17266-3473-4006-984f-9325122678b7', 'db380265-2e73-455b-a057-58c6c9390681', 0, '{"uuid-00-status": ["42dc2c2a-c2cd-4b36-a0d1-4cd2478952d6"], "uuid-00-subject": "My Laptop Is Not Working"}', '2021-01-10 14:53:15.978148', 1610290395);
`

const workflowSeeds = `
INSERT INTO public.flows (flow_id, account_id, entity_id, expression, name, description, mode, type, condition, status, created_at, updated_at) VALUES ('105c542b-d863-4036-8ac6-7f14d91fe4a2', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', '{{17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44.uuid-00-fname}} eq {Vijay} && {{17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44.uuid-00-nps-score}} gt {98}', 'Sales Pipeline', '', 1, 0, 1, 0, '2021-01-10 14:53:15.971076', 1610290395);

INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('6b5dc44e-98e3-4193-9cb3-faf4298fbe6e', '00000000-0000-0000-0000-000000000000', '3cf17266-3473-4006-984f-9325122678b7', '105c542b-d863-4036-8ac6-7f14d91fe4a2', '00000000-0000-0000-0000-000000000000', 'Opputunity', 0, 7, '', '{}', '2021-01-10 14:53:15.971852', 1610290395);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('df6ce502-c915-4f07-9f7b-7cdcce106820', '6b5dc44e-98e3-4193-9cb3-faf4298fbe6e', '3cf17266-3473-4006-984f-9325122678b7', '105c542b-d863-4036-8ac6-7f14d91fe4a2', '00000000-0000-0000-0000-000000000000', 'Deal Won', 0, 7, '{Vijay} eq {Vijay}', '{}', '2021-01-10 14:53:15.972564', 1610290395);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('b78879a6-3c52-4af4-bcfb-059b6fcd7403', '6b5dc44e-98e3-4193-9cb3-faf4298fbe6e', '3cf17266-3473-4006-984f-9325122678b7', '105c542b-d863-4036-8ac6-7f14d91fe4a2', 'ef73833b-d52a-4fdb-b5b4-59e409508d44', '', 0, 3, '', '{"ef73833b-d52a-4fdb-b5b4-59e409508d44": "ef068193-426d-4155-b00b-378c8fcc82be"}', '2021-01-10 14:53:15.972935', 1610290395);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('0fd76617-0f2e-4d6d-8bf6-6c7c5f447f54', '6b5dc44e-98e3-4193-9cb3-faf4298fbe6e', '3cf17266-3473-4006-984f-9325122678b7', '105c542b-d863-4036-8ac6-7f14d91fe4a2', '43c7ceb7-2f9c-4e23-a1d7-f0a7c9e7f545', '', 0, 4, '', '{}', '2021-01-10 14:53:15.973531', 1610290395);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('30b0dec6-0d3c-44e2-af48-6191aeb72097', 'df6ce502-c915-4f07-9f7b-7cdcce106820', '3cf17266-3473-4006-984f-9325122678b7', '105c542b-d863-4036-8ac6-7f14d91fe4a2', '74075e0f-57f7-4f99-b3e0-77690b83f117', '', 0, 6, '', '{"74075e0f-57f7-4f99-b3e0-77690b83f117": "d202dc5d-fa75-49ab-9ebb-b2609654322c"}', '2021-01-10 14:53:15.974244', 1610290395);
`

const pipelineSeeds = `
INSERT INTO public.flows (flow_id, account_id, entity_id, expression, name, description, mode, type, condition, status, created_at, updated_at) VALUES ('0dc63178-115b-4db2-a1ac-cf1550d82e67', '3cf27266-3473-4006-984f-9325122678b7', 'ded73f19-06e9-4229-a155-cdb7f04bd004', '{{ded73f19-06e9-4229-a155-cdb7f04bd004.uuid-00-fname}} eq {Vijay} && {{ded73f19-06e9-4229-a155-cdb7f04bd004.uuid-00-nps-score}} gt {98}', 'Sales Pipeline', '', 1, 0, 1, 0, '2021-01-04 11:41:36.150252', 1609760496);

INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('eff4262c-9b3f-4d29-91a0-3b4e37007996', '00000000-0000-0000-0000-000000000000', '3cf27266-3473-4006-984f-9325122678b7', '0dc63178-115b-4db2-a1ac-cf1550d82e67', '00000000-0000-0000-0000-000000000000', 'Opputunity', 0, 7, '', '{}', '2021-01-04 11:41:36.151199', 1609760496);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('2a8a97ef-b8a9-4a2d-878b-a1ad139b5117', 'eff4262c-9b3f-4d29-91a0-3b4e37007996', '3cf27266-3473-4006-984f-9325122678b7', '0dc63178-115b-4db2-a1ac-cf1550d82e67', '00000000-0000-0000-0000-000000000000', 'Deal Won', 0, 7, '{Vijay} eq {Vijay}', '{}', '2021-01-04 11:41:36.152101', 1609760496);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('c8b15cac-4cd1-41b9-a04c-01f2bbcbc7af', 'eff4262c-9b3f-4d29-91a0-3b4e37007996', '3cf27266-3473-4006-984f-9325122678b7', '0dc63178-115b-4db2-a1ac-cf1550d82e67', 'abfe25a3-102e-428a-98e0-ecd800090204', '', 0, 3, '', '{"abfe25a3-102e-428a-98e0-ecd800090204": "881291bd-8b72-433d-bb00-b944439a561c"}', '2021-01-04 11:41:36.152688', 1609760496);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('f3b374b0-ae8a-4b0d-ba7d-7e4fc582ac78', 'eff4262c-9b3f-4d29-91a0-3b4e37007996', '3cf27266-3473-4006-984f-9325122678b7', '0dc63178-115b-4db2-a1ac-cf1550d82e67', '1eeb2a4a-35e9-492b-bb84-79d0b866c02c', '', 0, 4, '', '{}', '2021-01-04 11:41:36.153264', 1609760496);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, name, weight, type, expression, actuals, created_at, updated_at) VALUES ('fc9aa2f4-d9c8-4f46-a327-ec6531700ea5', '2a8a97ef-b8a9-4a2d-878b-a1ad139b5117', '3cf27266-3473-4006-984f-9325122678b7', '0dc63178-115b-4db2-a1ac-cf1550d82e67', '39f5f6c3-21f6-4479-ba88-f8c3ed840c7b', '', 0, 6, '', '{"39f5f6c3-21f6-4479-ba88-f8c3ed840c7b": "bbb5cdf1-4e5b-47eb-ab7e-60e559432be7"}', '2021-01-04 11:41:36.153793', 1609760496);
`

const relationShipSeeds = `
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('0a8109d9-34de-45f5-bc08-50e5d459e649', '3cf17266-3473-4006-984f-9325122678b7', '57ee876a-1083-440b-a6de-f4c2a8500523', '793cf01b-6efc-4c01-b374-a6ee77fbe96b', 'c668edcc-c24d-4ea6-9efa-9eeccb2d22ab', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('f3574dd3-7e10-4dc0-a244-a30ee9f10a01', '3cf17266-3473-4006-984f-9325122678b7', '447a9c03-ad0c-4543-9dcf-fce1a8fbceed', '57ee876a-1083-440b-a6de-f4c2a8500523', '467f2610-1468-4641-a5ae-e0e5e73907e0', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('ab9fe036-7d04-445f-ac14-cb78ac75cc3a', '3cf17266-3473-4006-984f-9325122678b7', '447a9c03-ad0c-4543-9dcf-fce1a8fbceed', '57ee876a-1083-440b-a6de-f4c2a8500523', '863a7bb2-5081-420b-8366-17b5197b0e11', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('2c6dc29d-2522-4d98-9abf-bb57e344a962', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', '224e4a30-e2db-47c8-94ff-30186dbd3aa5', 'uuid-00-status', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('1a92f687-c12c-48ad-8e91-d18155b10aa1', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', '793cf01b-6efc-4c01-b374-a6ee77fbe96b', 'uuid-00-owner', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('485e8c3d-a4f3-42d2-90f1-62562861f7f5', '3cf17266-3473-4006-984f-9325122678b7', '86ae76d6-b62d-412b-96be-a02d55606560', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', 'uuid-00-contact', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('cb0a33c6-2e31-464c-9d9b-455fd444a2a7', '3cf17266-3473-4006-984f-9325122678b7', '86ae76d6-b62d-412b-96be-a02d55606560', '224e4a30-e2db-47c8-94ff-30186dbd3aa5', 'uuid-00-status', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('66089597-e335-416f-a534-05abba66bd66', '3cf17266-3473-4006-984f-9325122678b7', '294530fa-2689-4c12-b021-eef7745b7c3d', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', 'uuid-00-contacts', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('c5baa458-ea48-4adc-a97d-0c2f627816ca', '3cf17266-3473-4006-984f-9325122678b7', 'db380265-2e73-455b-a057-58c6c9390681', '224e4a30-e2db-47c8-94ff-30186dbd3aa5', 'uuid-00-status', 0);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('4e130aba-09aa-4c93-bd29-6ac3e820f745', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', 'f24b9868-654f-46ca-b444-b15032cf613b', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('4e130aba-09aa-4c93-bd29-6ac3e820f745', '3cf17266-3473-4006-984f-9325122678b7', 'f24b9868-654f-46ca-b444-b15032cf613b', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('a97921a4-7d5a-46e0-b38e-986578d46860', '3cf17266-3473-4006-984f-9325122678b7', 'db380265-2e73-455b-a057-58c6c9390681', '294530fa-2689-4c12-b021-eef7745b7c3d', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('a97921a4-7d5a-46e0-b38e-986578d46860', '3cf17266-3473-4006-984f-9325122678b7', '294530fa-2689-4c12-b021-eef7745b7c3d', 'db380265-2e73-455b-a057-58c6c9390681', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('67693211-e12b-48e3-a647-ea6aea3f6d37', '3cf17266-3473-4006-984f-9325122678b7', 'db380265-2e73-455b-a057-58c6c9390681', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', '00000000-0000-0000-0000-000000000000', 1);
INSERT INTO public.relationships (relationship_id, account_id, src_entity_id, dst_entity_id, field_id, type) VALUES ('67693211-e12b-48e3-a647-ea6aea3f6d37', '3cf17266-3473-4006-984f-9325122678b7', '17b61c5a-c6f9-4894-82b4-8b0e4b2d5d44', 'db380265-2e73-455b-a057-58c6c9390681', '00000000-0000-0000-0000-000000000000', 1);

INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '2c6dc29d-2522-4d98-9abf-bb57e344a962', '9d9ab317-897d-4297-9818-088674ce497e', '42dc2c2a-c2cd-4b36-a0d1-4cd2478952d6');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '2c6dc29d-2522-4d98-9abf-bb57e344a962', '19496230-0793-4669-b1a3-cef4c8bdbdad', 'cee462a0-d9ca-441b-8078-f14e14860626');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '485e8c3d-a4f3-42d2-90f1-62562861f7f5', 'dba4f6ef-3a57-4865-bf38-7b19a84b580b', '9d9ab317-897d-4297-9818-088674ce497e');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '485e8c3d-a4f3-42d2-90f1-62562861f7f5', 'f4626250-03d5-4c69-b44e-0b9d018206e6', '9d9ab317-897d-4297-9818-088674ce497e');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '66089597-e335-416f-a534-05abba66bd66', '7bbe2be3-6d0e-4f91-a37a-46158cdd0eb9', '9d9ab317-897d-4297-9818-088674ce497e');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '66089597-e335-416f-a534-05abba66bd66', '7bbe2be3-6d0e-4f91-a37a-46158cdd0eb9', '19496230-0793-4669-b1a3-cef4c8bdbdad');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', 'c5baa458-ea48-4adc-a97d-0c2f627816ca', 'e7f5f01f-aa38-47a9-bbf1-b516f65b130b', '42dc2c2a-c2cd-4b36-a0d1-4cd2478952d6');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '4e130aba-09aa-4c93-bd29-6ac3e820f745', '9d9ab317-897d-4297-9818-088674ce497e', '9a722df4-78a1-4a5b-a9dd-b98d48702b08');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', 'a97921a4-7d5a-46e0-b38e-986578d46860', 'e7f5f01f-aa38-47a9-bbf1-b516f65b130b', '7bbe2be3-6d0e-4f91-a37a-46158cdd0eb9');
INSERT INTO public.connections (account_id, relationship_id, src_item_id, dst_item_id) VALUES ('3cf17266-3473-4006-984f-9325122678b7', '67693211-e12b-48e3-a647-ea6aea3f6d37', 'e7f5f01f-aa38-47a9-bbf1-b516f65b130b', '9d9ab317-897d-4297-9818-088674ce497e');
`
