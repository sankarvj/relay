--
-- PostgreSQL database dump
--

-- Dumped from database version 9.5.19
-- Dumped by pg_dump version 9.5.19

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: accounts; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.accounts (account_id, parent_account_id, name, domain, avatar, plan, mode, timezone, language, country, issued_at, expiry, created_at, updated_at) VALUES ('3cf27266-3473-4006-984f-9325122678b7', NULL, 'Wayplot', 'Wayplot', 'http://gravatar/vj', 0, 0, 'IST', 'EN', 'IN', '2019-11-20 00:00:00', '2020-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);


--
-- Data for Name: teams; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.teams (team_id, account_id, name, description, created_at, updated_at) VALUES ('8cf27268-3473-4006-984f-9325122678b7', '3cf27266-3473-4006-984f-9325122678b7', 'CRM', NULL, '2020-02-22 15:03:57.416566', 1582383837);


--
-- Data for Name: entities; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000001', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Status', NULL, 1, 0, 0, '[{"key": "uuid-00-name", "meta": null, "name": "name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-color", "meta": null, "name": "color", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}]', NULL, '2020-09-30 12:19:23.283371', 1601468363);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000002', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Contacts', NULL, 1, 0, 0, '[{"key": "uuid-00-fname", "meta": null, "name": "First Name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-email", "meta": null, "name": "Email", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-mobile-numbers", "meta": null, "name": "Mobile Numbers", "field": {"key": "", "meta": null, "name": "", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "", "choices": null, "dom_type": "MS", "data_type": "L", "display_name": ""}, {"key": "uuid-00-nps-score", "meta": null, "name": "NPS Score", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "", "data_type": "N", "display_name": ""}, {"key": "uuid-00-lf-stage", "meta": null, "name": "Lifecycle Stage", "field": null, "value": null, "ref_id": "", "choices": ["lead", "contact", "won"], "dom_type": "SE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-status", "meta": null, "name": "Status", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "00000000-0000-0000-0000-000000000001", "choices": null, "dom_type": "TE", "data_type": "R", "display_name": ""}]', NULL, '2020-09-30 12:19:23.291571', 1601468363);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000003', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Tasks', NULL, 1, 0, 0, '[{"key": "uuid-00-desc", "meta": null, "name": "desc", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-contact", "meta": null, "name": "Contact", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "00000000-0000-0000-0000-000000000002", "choices": null, "dom_type": "TE", "data_type": "R", "display_name": ""}]', NULL, '2020-09-30 12:19:23.299915', 1601468363);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000004', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Deals', NULL, 1, 0, 0, '[{"key": "uuid-00-deal-name", "meta": null, "name": "Deal Name", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-deal-amount", "meta": null, "name": "Deal Amount", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "N", "display_name": ""}, {"key": "uuid-00-contacts", "meta": null, "name": "Contacts", "field": {"key": "id", "meta": null, "name": "", "field": null, "value": "--", "ref_id": "", "choices": null, "dom_type": "", "data_type": "S", "display_name": ""}, "value": null, "ref_id": "00000000-0000-0000-0000-000000000002", "choices": null, "dom_type": "TE", "data_type": "R", "display_name": ""}]', NULL, '2020-09-30 12:19:23.347388', 1601468363);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000005', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'MailGun Intg', NULL, 4, 0, 0, '[{"key": "uuid-00-domain", "meta": {"config": "true"}, "name": "domain", "field": null, "value": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-api-key", "meta": {"config": "true"}, "name": "api_key", "field": null, "value": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-sender", "meta": null, "name": "sender", "field": null, "value": "vijayasankar.jothi@wayplot.com", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-to", "meta": null, "name": "to", "field": null, "value": "", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-cc", "meta": null, "name": "cc", "field": null, "value": "", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-subject", "meta": null, "name": "subject", "field": null, "value": "", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-body", "meta": null, "name": "body", "field": null, "value": "", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}]', NULL, '2020-09-30 12:19:23.364222', 1601468363);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000006', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'WebHook', NULL, 2, 0, 0, '[{"key": "uuid-00-path", "meta": null, "name": "path", "field": null, "value": "/actuator/info", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-host", "meta": null, "name": "host", "field": null, "value": "https://stage.freshcontacts.io", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-method", "meta": null, "name": "method", "field": null, "value": "GET", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}, {"key": "uuid-00-headers", "meta": null, "name": "headers", "field": null, "value": "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}]', NULL, '2020-09-30 12:19:23.373021', 1601468363);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000007', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Delay Timer', NULL, 7, 0, 0, '[{"key": "uuid-00-delay-by", "meta": null, "name": "delay_by", "field": null, "value": null, "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "T", "display_name": ""}, {"key": "uuid-00-repeat", "meta": null, "name": "repeat", "field": null, "value": "true", "ref_id": "", "choices": null, "dom_type": "TE", "data_type": "S", "display_name": ""}]', NULL, '2020-09-30 12:19:23.381294', 1601468363);


--
-- Data for Name: flows; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.flows (flow_id, account_id, entity_id, expression, name, description, type, condition, status, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000017', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000002', '{{00000000-0000-0000-0000-000000000002.uuid-00-fname}} eq {Vijay} && {{00000000-0000-0000-0000-000000000002.uuid-00-nps-score}} gt {98}', 'The Workflow', '', 1, 1, 0, '2020-09-30 12:19:23.471304', 1601468363);
INSERT INTO public.flows (flow_id, account_id, entity_id, expression, name, description, type, condition, status, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000023', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000002', '{{00000000-0000-0000-0000-000000000002.uuid-00-fname}} eq {Vijay} && {{00000000-0000-0000-0000-000000000002.uuid-00-nps-score}} gt {98}', 'The Pipeline', '', 3, 1, 0, '2020-09-30 12:19:23.513213', 1601468363);


--
-- Data for Name: items; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000008', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000001', 0, '{"uuid-00-name": "Open", "uuid-00-color": "#fb667e"}', '2020-09-30 12:19:23.388923', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000009', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000001', 0, '{"uuid-00-name": "Closed", "uuid-00-color": "#66fb99"}', '2020-09-30 12:19:23.396272', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000010', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000002', 0, '{"uuid-00-email": "vijayasankarmail@gmail.com", "uuid-00-fname": "Vijay", "uuid-00-status": ["00000000-0000-0000-0000-000000000008"], "uuid-00-lf-stage": "lead", "uuid-00-nps-score": 100, "uuid-00-mobile-numbers": ["9944293499", "9940209164"]}', '2020-09-30 12:19:23.404207', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000011', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000002', 0, '{"uuid-00-email": "senthil@gmail.com", "uuid-00-fname": "Senthil", "uuid-00-status": ["00000000-0000-0000-0000-000000000009"], "uuid-00-lf-stage": "lead", "uuid-00-nps-score": 100, "uuid-00-mobile-numbers": ["9944293499", "9940209164"]}', '2020-09-30 12:19:23.411772', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000012', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000003', 0, '{"uuid-00-desc": "add deal price", "uuid-00-contact": ["00000000-0000-0000-0000-000000000010"]}', '2020-09-30 12:19:23.419341', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000013', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000003', 0, '{"uuid-00-desc": "make call", "uuid-00-contact": ["00000000-0000-0000-0000-000000000010"]}', '2020-09-30 12:19:23.426806', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000014', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000004', 0, '{"uuid-00-contacts": ["00000000-0000-0000-0000-000000000010", "00000000-0000-0000-0000-000000000011"], "uuid-00-deal-name": "Big Deal", "uuid-00-deal-amount": 1000}', '2020-09-30 12:19:23.433685', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000015', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000005', 0, '{"uuid-00-cc": "vijayasankarmobile@gmail.com", "uuid-00-to": "{{00000000-0000-0000-0000-000000000002.uuid-00-email}}", "uuid-00-body": "Hello {{00000000-0000-0000-0000-000000000002.uuid-00-fname}}", "uuid-00-domain": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "uuid-00-sender": "vijayasankar.jothi@wayplot.com", "uuid-00-api-key": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "uuid-00-subject": "This mail is sent you to tell that your NPS scrore is {{00000000-0000-0000-0000-000000000002.uuid-00-nps-score}}. We are very proud of you! "}', '2020-09-30 12:19:23.444648', 1601468363);
INSERT INTO public.items (item_id, account_id, entity_id, state, fieldsb, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000016', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000007', 0, '{"uuid-00-repeat": "true", "uuid-00-delay-by": "2"}', '2020-09-30 12:19:23.459974', 1601468363);


--
-- Data for Name: active_flows; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: nodes; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000018', NULL, '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000003', 1, '', '{}', '2020-09-30 12:19:23.47689', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000019', '00000000-0000-0000-0000-000000000018', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000000', 0, '{Vijay} eq {Vijay}', '{}', '2020-09-30 12:19:23.484252', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000019', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000005', 3, '{{xyz.result}} eq {true}', '{"00000000-0000-0000-0000-000000000005": "00000000-0000-0000-0000-000000000015"}', '2020-09-30 12:19:23.492288', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000019', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000006', 4, '{{xyz.result}} eq {false}', '{}', '2020-09-30 12:19:23.499152', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000022', '00000000-0000-0000-0000-000000000020', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000017', '00000000-0000-0000-0000-000000000007', 6, '', '{"00000000-0000-0000-0000-000000000007": "00000000-0000-0000-0000-000000000016"}', '2020-09-30 12:19:23.505702', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000024', NULL, '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000000', 7, '', '{}', '2020-09-30 12:19:23.519607', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000025', '00000000-0000-0000-0000-000000000024', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000000', 7, '{Vijay} eq {Vijay}', '{}', '2020-09-30 12:19:23.526991', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000026', '00000000-0000-0000-0000-000000000024', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000005', 3, '', '{"00000000-0000-0000-0000-000000000005": "00000000-0000-0000-0000-000000000015"}', '2020-09-30 12:19:23.536253', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000027', '00000000-0000-0000-0000-000000000024', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000006', 4, '', '{}', '2020-09-30 12:19:23.549398', 1601468363);
INSERT INTO public.nodes (node_id, parent_node_id, account_id, flow_id, actor_id, type, expression, actuals, created_at, updated_at) VALUES ('00000000-0000-0000-0000-000000000028', '00000000-0000-0000-0000-000000000025', '3cf27266-3473-4006-984f-9325122678b7', '00000000-0000-0000-0000-000000000023', '00000000-0000-0000-0000-000000000007', 6, '', '{"00000000-0000-0000-0000-000000000007": "00000000-0000-0000-0000-000000000016"}', '2020-09-30 12:19:23.565378', 1601468363);


--
-- Data for Name: active_nodes; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: darwin_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (17, 1, 'Add accounts', '68f14cc08160e9d40407026861cf604b', 1599891027, 22289788);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (18, 2, 'Add users', '09e2ff9530d82de1431d50320a44f463', 1599891027, 6155131);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (19, 3, 'Add teams', '942f94590f2c9652b02eadd180e23029', 1599891027, 5073913);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (20, 4, 'Add members', 'ea3e297c027f76a8d41a7c06dc8baa5e', 1599891027, 3682010);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (21, 5, 'Add entities', 'f33e47268128232957b598cf920a129d', 1599891027, 3844979);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (22, 6, 'Add items', '06037ce47e65f177a4219c203e572a97', 1599891027, 2999368);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (23, 7, 'Add flows', 'fc6b4eeb5081cf166512372dfcca78f0', 1599891027, 3967019);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (24, 8, 'Add nodes', '2cf930177c7f5c37dc59d1e46bdfb492', 1599891027, 3315767);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (25, 9, 'Add active_flows', '4150da24edf33db515d623428fb3fc26', 1599891027, 2588378);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (26, 10, 'Add active_nodes', '64df0f3070f977e633413c91626b62e3', 1599891027, 2662460);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (27, 11, 'Add relationships', '9a99d239e8c0539e11f4ed859d664ae6', 1599891027, 2186140);


--
-- Name: darwin_migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.darwin_migrations_id_seq', 27, true);


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('5cf37266-3473-4006-984f-9325122678b7', '3cf27266-3473-4006-984f-9325122678b7', 'vijayasankar', 'http://gravatar/vj', 'vijayasankarmail@gmail.com', '9944293499', true, '{ADMIN,USER}', 'cfr07IBEBCfGxp9dxjBOGYdkjHG2', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '3cf27266-3473-4006-984f-9325122678b7', 'vijay', 'http://gravatar/vj', 'vijayasankarj@gmail.com', '9940209164', true, '{USER}', 'ggOv3mMCqVZ6nFqaco4lD9qjxc63', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('55b5fbd3-755f-4379-8f07-a58d4a30fa2f', '3cf27266-3473-4006-984f-9325122678b7', 'senthil', 'http://gravatar/vj', 'sksenthilkumaar@gmail.com', '9940209164', true, '{USER}', 'sk_replacetokenhere', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('65b5fbd3-755f-4379-8f07-a58d4a30fa2f', '3cf27266-3473-4006-984f-9325122678b7', 'saravana', 'http://gravatar/vj', 'saravanaprakas@gmail.com', '9940209164', true, '{USER}', 'sexy_replacetokenhere', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);


--
-- Data for Name: members; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Data for Name: relationships; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- PostgreSQL database dump complete
--

