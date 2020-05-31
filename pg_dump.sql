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

INSERT INTO public.accounts (account_id, name, domain, avatar, plan, mode, timezone, language, country, issued_at, expiry, created_at, updated_at) VALUES ('3cf27266-3473-4006-984f-9325122678b7', 'Wayplot', 'Wayplot', 'http://gravatar/vj', 0, 0, 'IST', 'EN', 'IN', '2019-11-20 00:00:00', '2020-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);


--
-- Data for Name: darwin_migrations; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (1, 1, 'Add accounts', 'ba26df90caa3990ed1aaa2cbcae7bf17', 1582383291, 3751371);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (2, 2, 'Add users', '09e2ff9530d82de1431d50320a44f463', 1582383291, 3801433);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (3, 3, 'Add teams', '19facb614f4eb95904388e1ef60cd53b', 1582383291, 3992615);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (4, 4, 'Add members', '06217fa84170fee36481295e2dc91c77', 1582383291, 4266935);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (5, 5, 'Add entities', '81fec8f3d8749fa8c7c9a9d11dfebd78', 1582383291, 3139380);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (6, 6, 'Add rules', 'a076bb5c72f4a04a1e36c16dc4404ae4', 1582383291, 2859884);
INSERT INTO public.darwin_migrations (id, version, description, checksum, applied_at, execution_time) VALUES (7, 7, 'Add items', 'c02619e51ecce335d6e5e77055795c47', 1582383291, 3098698);


--
-- Name: darwin_migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.darwin_migrations_id_seq', 7, true);


--
-- Data for Name: teams; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.teams (team_id, account_id, name, description, created_at, updated_at) VALUES (1, '3cf27266-3473-4006-984f-9325122678b7', 'MCR', NULL, '2020-02-22 15:03:57.416566', 1582383837);


--
-- Data for Name: entities; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES ('adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 1, 'Contacts', NULL, 1, 1, 1, 0, '[{"key": "2bf431f8-b2ae-467f-9c5b-e7216068ea40", "name": "Name", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "08320990-cc56-4809-801a-a937b62ec307", "name": "Email", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "bf3cfc1d-a170-473f-b52b-4fef7495a0e3", "name": "MRR", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "900d69bf-2fc7-4c34-95b1-ef9f79220810", "name": "Company Name", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}]', NULL, '2020-05-31 04:54:41.217704', 1590900881);
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES ('b8fb4ff2-d660-4846-b058-d27adfb10441', 1, 'Webhook Integration', NULL, 2, 1, 1, 0, '[{"key": "eb64eb6b-d95d-4942-8837-d5d2309a7277", "name": "Path", "value": "/actuator/info", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "470f9fff-7ee7-4954-8aa9-ab3144c1a18c", "name": "Host", "value": "https://stage.freshcontacts.io", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "84726e54-22f2-441a-aebc-1729e16ba957", "name": "Method", "value": "GET", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "f641b929-d91a-4ba2-898b-db4e7a9972ba", "name": "Headers", "value": "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}]', NULL, '2020-05-16 12:49:22.947029', 1589633362);
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES ('d9ccf588-e6eb-40b3-838f-f6d5262bac78', 1, 'Events', NULL, 3, 1, 1, 0, '[{"key": "d3e572e1-3950-46db-a230-d41b2f4cd8d0", "name": "StartTime", "value": "", "hidden": false, "unique": false, "data_type": "DT", "mandatory": false, "reference": ""}, {"key": "fa5e4f5e-b623-417b-a030-3c1b4385dbc0", "name": "EndTime", "value": "", "hidden": false, "unique": false, "data_type": "DT", "mandatory": false, "reference": ""}, {"key": "9f9ade37-9549-4d12-a82d-c69495e85980", "name": "Status", "value": "", "hidden": false, "unique": false, "data_type": "ST", "mandatory": false, "reference": ""}]', NULL, '2020-05-16 12:49:59.279275', 1589633399);
INSERT INTO public.entities (entity_id, team_id, name, description, category, state, mode, retry, attributes, tags, created_at, updated_at) VALUES ('033e8ce4-0cbf-4ee8-86be-306b583f618e', 1, 'MailGun Integration', NULL, 4, 1, 1, 0, '[{"key": "921ecaab-b3f0-42b6-a581-29239cc58e4b", "name": "Domain", "value": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "config": true, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "a8376197-b699-4f4b-b2dd-2bf5aa18ee16", "name": "API key", "value": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "config": true, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "4a68900c-5697-4f64-9a47-e49291ff9218", "name": "Sender", "value": "vijayasankar.jothi@wayplot.com", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "c2a0b583-cfb0-4f03-8b36-587548704b13", "name": "To", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "a27bb6d0-67df-4542-a806-e0974bff2e27", "name": "CC", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "e34c3e1e-62fe-44cb-8caa-23c4bbbfcefc", "name": "Subject", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}, {"key": "aaed7f03-291c-4276-a687-cbd80dc1eb52", "name": "Content", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": ""}]', NULL, '2020-05-31 05:16:27.059717', 1590902187);


--
-- Data for Name: items; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('9d53277d-bf0c-4baf-bb86-ce61259dab44', NULL, 'd9ccf588-e6eb-40b3-838f-f6d5262bac78', 0, '{"9f9ade37-9549-4d12-a82d-c69495e85980": "down", "d3e572e1-3950-46db-a230-d41b2f4cd8d0": "2020-05-16 12:49:59.279275", "fa5e4f5e-b623-417b-a030-3c1b4385dbc0": "2021-05-16 12:49:59.279275"}', '2020-05-30 07:44:05.760548', 1590824645);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('44e5918f-2cbe-4d62-92d2-86820adff0cd', NULL, 'adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "vijayasankarmail@gmail.com", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Vijay", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "FreshW", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "10000"}', '2020-05-31 04:55:22.480538', 1590900922);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('8670ef39-a38a-44c3-b8a2-684276a4e673', NULL, 'adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "saravanaprakas@gmail.com ", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Saravana", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "Zoho", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "200000"}', '2020-05-31 04:57:14.844344', 1590901034);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('7d9c4f94-890b-484c-8189-91c3d7e8e50b', NULL, 'adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "sksenthilkumaar@gmail.com ", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Senthil", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "Qatar Airways", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "500000"}', '2020-05-31 04:57:46.445474', 1590901066);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('7c766083-83af-4926-b980-37de0f9edde0', NULL, '033e8ce4-0cbf-4ee8-86be-306b583f618e', 0, '{"4a68900c-5697-4f64-9a47-e49291ff9218": "vijayasankar.jothi@wayplot.com", "921ecaab-b3f0-42b6-a581-29239cc58e4b": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "a27bb6d0-67df-4542-a806-e0974bff2e27": "vijayasankarmobile@gmail.com", "a8376197-b699-4f4b-b2dd-2bf5aa18ee16": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "aaed7f03-291c-4276-a687-cbd80dc1eb52": "This mail is sent you to tell that your MRR is {{adbd74c7-7add-4dcd-b2cf-6b05863b90e8.bf3cfc1d-a170-473f-b52b-4fef7495a0e3}}. We are very proud of you! ", "c2a0b583-cfb0-4f03-8b36-587548704b13": "{{adbd74c7-7add-4dcd-b2cf-6b05863b90e8.08320990-cc56-4809-801a-a937b62ec307}}", "e34c3e1e-62fe-44cb-8caa-23c4bbbfcefc": "Hello {{adbd74c7-7add-4dcd-b2cf-6b05863b90e8.2bf431f8-b2ae-467f-9c5b-e7216068ea40}}"}', '2020-05-31 05:26:54.805027', 1590902814);


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('5cf37266-3473-4006-984f-9325122678b7', '3cf27266-3473-4006-984f-9325122678b7', 'vijayasankar', 'http://gravatar/vj', 'vijayasankarmail@gmail.com', '9944293499', true, '{ADMIN,USER}', 'cfr07IBEBCfGxp9dxjBOGYdkjHG2', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);
INSERT INTO public.users (user_id, account_id, name, avatar, email, phone, verified, roles, password_hash, provider, issued_at, created_at, updated_at) VALUES ('45b5fbd3-755f-4379-8f07-a58d4a30fa2f', '3cf27266-3473-4006-984f-9325122678b7', 'vijay', 'http://gravatar/vj', 'vijayasankarj@gmail.com', '9940209164', true, '{USER}', 'ggOv3mMCqVZ6nFqaco4lD9qjxc63', 'firebase', '2019-11-20 00:00:00', '2019-11-20 00:00:00', 1574239364000);


--
-- Data for Name: members; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Name: members_member_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.members_member_id_seq', 1, false);


--
-- Data for Name: rules; Type: TABLE DATA; Schema: public; Owner: postgres
--



--
-- Name: teams_team_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.teams_team_id_seq', 1, true);


--
-- PostgreSQL database dump complete
--

