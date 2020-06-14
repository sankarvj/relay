package schema

import (
	"github.com/jmoiron/sqlx"
)

// Seed runs the set of seed-data queries against db. The queries are ran in a
// transaction and rolled back if any fail.
func Seed(db *sqlx.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(seeds); err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return tx.Commit()
}

// const for seed data ids
const (
	SeedTeamID                 = "1"
	SeedAccountID              = "3cf27266-3473-4006-984f-9325122678b7"
	SeedUserID1                = "5cf37266-3473-4006-984f-9325122678b7"
	SeedUserID2                = "45b5fbd3-755f-4379-8f07-a58d4a30fa2f"
	SeedEntityTimeSeriesID     = "d9ccf588-e6eb-40b3-838f-f6d5262bac78"
	SeedEntityAPIID            = "b8fb4ff2-d660-4846-b058-d27adfb10441"
	SeedEntityContactID        = "adbd74c7-7add-4dcd-b2cf-6b05863b90e8"
	SeedEntityEmailID          = "033e8ce4-0cbf-4ee8-86be-306b583f618e"
	SeedEntityTaskID           = "fcf13a59-47a9-4661-8ed2-62947d572b31"
	SeedFieldKeyStTimeID       = "d3e572e1-3950-46db-a230-d41b2f4cd8d0"
	SeedFieldKeyEndTimeID      = "fa5e4f5e-b623-417b-a030-3c1b4385dbc0"
	SeedFieldKeyStatusID       = "9f9ade37-9549-4d12-a82d-c69495e85980"
	SeedFieldKeyPathID         = "eb64eb6b-d95d-4942-8837-d5d2309a7277"
	SeedFieldKeyHostID         = "470f9fff-7ee7-4954-8aa9-ab3144c1a18c"
	SeedFieldKeyMethodID       = "84726e54-22f2-441a-aebc-1729e16ba957"
	SeedFieldKeyHeaderID       = "f641b929-d91a-4ba2-898b-db4e7a9972ba"
	SeedFieldKeyContactName    = "2bf431f8-b2ae-467f-9c5b-e7216068ea40"
	SeedFieldKeyContactEmail   = "08320990-cc56-4809-801a-a937b62ec307"
	SeedFieldKeyContactMRR     = "bf3cfc1d-a170-473f-b52b-4fef7495a0e3"
	SeedFieldKeyContactCompany = "900d69bf-2fc7-4c34-95b1-ef9f79220810"
	SeedFieldKeyTaskDesc       = "a57f650c-211c-49cb-ae56-d141cb380342"
	SeedFieldKeyAssigned       = "be084c25-7f85-4a89-af21-a0dbaa49a7e8"
	SeedFieldKeyTaskForCon     = "dfb640a0-94cb-4218-b90b-1573d7ba3805"
	SeedItemEventID            = "9d53277d-bf0c-4baf-bb86-ce61259dab44"
	SeedItemContactID1         = "44e5918f-2cbe-4d62-92d2-86820adff0cd"
	SeedItemContactID2         = "8670ef39-a38a-44c3-b8a2-684276a4e673"
	SeedItemContactID3         = "7d9c4f94-890b-484c-8189-91c3d7e8e50b"
	SeedItemContactUpdatableID = "7d9c4f94-890b-484c-8189-91c3d7e8e501"
	SeedItemEmailID            = "7c766083-83af-4926-b980-37de0f9edde0"
	SeedItemTaskID             = "3d247443-b257-4b06-ba99-493cf9d83ce7"
)

// seeds is a string constant containing all of the queries needed to get the
// db seeded to a useful state for development.
//
// Note that database servers besides PostgreSQL may not support running
// multiple queries as part of the same execution so this single large constant
// may need to be broken up.
const seeds = `
-- Create a demo account wayplot
INSERT INTO public.accounts (account_id, name, domain, avatar, plan, mode, timezone, language, country, issued_at, expiry, created_at, updated_at) VALUES
	('` + SeedAccountID + `', 'Wayplot', 'Wayplot', 'http://gravatar/vj', 0, 0, 'IST', 'EN', 'IN', '2019-11-20 00:00:00', '2020-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000)
	ON CONFLICT DO NOTHING;
-- Create admin and regular User with password "gophers"
INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES
	('` + SeedUserID1 + `', '` + SeedAccountID + `', 'vijayasankar', 'http://gravatar/vj', 'vijayasankarmail@gmail.com', '9944293499', true, '{ADMIN,USER}', 'cfr07IBEBCfGxp9dxjBOGYdkjHG2', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000),
	('` + SeedUserID2 + `', '` + SeedAccountID + `', 'vijay', 'http://gravatar/vj', 'vijayasankarj@gmail.com', '9940209164', true, '{USER}', 'ggOv3mMCqVZ6nFqaco4lD9qjxc63', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000)
	ON CONFLICT DO NOTHING;
-- Create a demo team wayplot
INSERT INTO public.teams (team_id, account_id, name, description, created_at, updated_at) VALUES 
(` + SeedTeamID + `, '` + SeedAccountID + `', 'wayplot-team-A', NULL, '2020-02-22 15:03:57.416566', 1582383837);
-- Create A Timeseries Entity Called Events 
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES 
	('` + SeedEntityTimeSeriesID + `', ` + SeedTeamID + `, 'Events', NULL, 3, 1, 1, 0, 
	'[{"key": "` + SeedFieldKeyStTimeID + `", "name": "StartTime","display_name": "StartTime", "value": "", "hidden": false, "unique": false, "data_type": "DT", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyEndTimeID + `", "display_name": "EndTime","name": "EndTime", "value": "", "hidden": false, "unique": false, "data_type": "DT", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyStatusID + `", "name": "Status","display_name": "Status", "value": "", "hidden": false, "unique": false, "data_type": "ST", "mandatory": false, "reference": ""}]', 
	NULL, '2020-05-16 12:49:59.279275', 1589633399);
-- Create A API Entity Called Webhook Integration 
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES 
	('` + SeedEntityAPIID + `', ` + SeedTeamID + `, 'Webhook Integration', NULL, 2, 1, 1, 0, 
	'[{"key": "` + SeedFieldKeyPathID + `", "name": "path","display_name": "Path", "value": "/actuator/info", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyHostID + `", "name": "host","display_name": "Host", "value": "https://stage.freshcontacts.io", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyMethodID + `", "name": "method","display_name": "Method", "value": "GET", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyHeaderID + `", "name": "headers","display_name": "Headers", "value": "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}]',
	 NULL, '2020-05-16 12:49:22.947029', 1589633362);
-- Create A Data Entity Called Contacts
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES
 	('` + SeedEntityContactID + `', ` + SeedTeamID + `, 'Contacts', NULL, 1, 1, 1, 0, 
	 '[{"key": "` + SeedFieldKeyContactName + `", "display_name": "Name","name": "Name", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyContactEmail + `", "name": "email","display_name": "Email", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyContactMRR + `", "name": "mrr", "display_name": "MRR", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "` + SeedFieldKeyContactCompany + `", "name": "company_name" ,"display_name": "Company Name", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}]', 
	 NULL, '2020-05-31 04:54:41.217704', 1590900881);
-- Create A Email Entity Called Mailgun Integration
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES 
	('` + SeedEntityEmailID + `', ` + SeedTeamID + `, 'MailGun Integration', NULL, 4, 1, 1, 0, 
	'[{"key": "921ecaab-b3f0-42b6-a581-29239cc58e4b", "name": "domain","display_name": "domain", "value": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "config": true, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "a8376197-b699-4f4b-b2dd-2bf5aa18ee16", "display_name": "API key","name": "api_key", "value": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "config": true, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "4a68900c-5697-4f64-9a47-e49291ff9218","name": "sender", "display_name": "Sender", "value": "vijayasankar.jothi@wayplot.com", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "c2a0b583-cfb0-4f03-8b36-587548704b13", "name": "to","display_name": "To", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "a27bb6d0-67df-4542-a806-e0974bff2e27", "name": "cc","display_name": "CC", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "e34c3e1e-62fe-44cb-8caa-23c4bbbfcefc", "name": "subject","display_name": "Subject", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "aaed7f03-291c-4276-a687-cbd80dc1eb52", "name": "body","display_name": "Body", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}]', 
	NULL, '2020-05-31 05:16:27.059717', 1590902187);
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) 
	VALUES ('` + SeedEntityTaskID + `', ` + SeedTeamID + `, 'Task', NULL, 1, 1, 1, 0, '[{"key": "` + SeedFieldKeyTaskDesc + `", "name": "Desc", "value": "Prepare the research documents and make a call", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": ""}, {"key": "` + SeedFieldKeyAssigned + `", "name": "AssignedTo", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": ""}, {"key": "` + SeedFieldKeyTaskForCon + `", "name": "Contact", "value": "{{` + SeedEntityContactID + `.id}}", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": ""}]', NULL, '2020-06-08 08:25:49.617813', 1591604749);
-- Create a demo items for Events
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemEventID + `', NULL, '` + SeedEntityTimeSeriesID + `', 0, '{"` + SeedFieldKeyStatusID + `": "down", "` + SeedFieldKeyStTimeID + `": "2020-05-16 12:49:59.279275", "` + SeedFieldKeyEndTimeID + `": "2021-05-16 12:49:59.279275"}', '2020-05-30 07:44:05.760548', 1590824645);
-- Create a demo items for Contacts
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemContactID1 + `', NULL, '` + SeedEntityContactID + `', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "vijayasankarmail@gmail.com", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Vijay", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "FreshW", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "10000"}', '2020-05-31 04:55:22.480538', 1590900922);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemContactID2 + `', NULL, '` + SeedEntityContactID + `', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "saravanaprakas@gmail.com ", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Saravana", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "Zoho", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "200000"}', '2020-05-31 04:57:14.844344', 1590901034);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemContactID3 + `', NULL, '` + SeedEntityContactID + `', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "sksenthilkumaar@gmail.com ", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Senthil", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "Qatar Airways", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "500000"}', '2020-05-31 04:57:46.445474', 1590901066);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemContactUpdatableID + `', NULL, '` + SeedEntityContactID + `', 0, '{"bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "123456"}', '2020-05-31 04:57:46.445474', 1590901066);
-- Create a demo items for Emails
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemEmailID + `', NULL, '` + SeedEntityEmailID + `', 0, '{"4a68900c-5697-4f64-9a47-e49291ff9218": "vijayasankar.jothi@wayplot.com", "921ecaab-b3f0-42b6-a581-29239cc58e4b": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "a27bb6d0-67df-4542-a806-e0974bff2e27": "vijayasankarmobile@gmail.com", "a8376197-b699-4f4b-b2dd-2bf5aa18ee16": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "aaed7f03-291c-4276-a687-cbd80dc1eb52": "This mail is sent you to tell that your MRR is {{` + SeedEntityContactID + `.` + SeedFieldKeyContactMRR + `}}. We are very proud of you! ", "c2a0b583-cfb0-4f03-8b36-587548704b13": "{{` + SeedEntityContactID + `.` + SeedFieldKeyContactEmail + `}}", "e34c3e1e-62fe-44cb-8caa-23c4bbbfcefc": "Hello {{` + SeedEntityContactID + `.2bf431f8-b2ae-467f-9c5b-e7216068ea40}}"}', '2020-05-31 05:26:54.805027', 1590902814);
-- Create a demo items for Tasks
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('` + SeedItemTaskID + `', NULL, '` + SeedEntityTaskID + `', 0, '{"` + SeedFieldKeyTaskDesc + `": "Ummm. Prepare the research documents and make a call", "` + SeedFieldKeyAssigned + `": "agents.id"}', '2020-06-08 14:04:58.523412', 1591625098);
`
