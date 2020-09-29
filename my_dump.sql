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
-- Data for Name: entities; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('adbd74c7-7add-4dcd-b2cf-6b05863b90e8', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Contacts', NULL, 1, 1, 1, '[{"key": "2bf431f8-b2ae-467f-9c5b-e7216068ea40", "name": "Name", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Name"}, {"key": "08320990-cc56-4809-801a-a937b62ec307", "name": "email", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Email"}, {"key": "bf3cfc1d-a170-473f-b52b-4fef7495a0e3", "name": "mrr", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "MRR"}, {"key": "bd45fc1d-a170-473f-b52b-4fef7495a0e3", "name": "amount", "field": {"key": "element", "data_type": "N"}, "value": "", "hidden": false, "unique": false, "data_type": "L", "mandatory": false, "reference": "", "display_name": "Deal Amount"}, {"key": "900d69bf-2fc7-4c34-95b1-ef9f79220810", "name": "company_name", "value": "", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Company Name"}]', NULL, '2020-05-31 04:54:41.217704', 1590900881);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('b8fb4ff2-d660-4846-b058-d27adfb10441', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Webhook Integration', NULL, 2, 1, 1, '[{"key": "eb64eb6b-d95d-4942-8837-d5d2309a7277", "name": "path", "value": "/actuator/info", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Path"}, {"key": "470f9fff-7ee7-4954-8aa9-ab3144c1a18c", "name": "host", "value": "https://stage.freshcontacts.io", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Host"}, {"key": "84726e54-22f2-441a-aebc-1729e16ba957", "name": "method", "value": "GET", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Method"}, {"key": "f641b929-d91a-4ba2-898b-db4e7a9972ba", "name": "headers", "value": "{\"X-ClientToken\":\"mcr eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjExNTc2NTAwMjk2fQ.1KtXw_YgxbJW8ibv_v2hfpInjQKC6enCh9IO1ziV2RA\"}", "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": "Headers"}]', NULL, '2020-05-16 12:49:22.947029', 1589633362);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('fcf13a59-47a9-4661-8ed2-62947d572b31', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Task', NULL, 1, 1, 1, '[{"key": "a57f650c-211c-49cb-ae56-d141cb380342", "name": "Desc", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": ""}, {"key": "be084c25-7f85-4a89-af21-a0dbaa49a7e8", "name": "AssignedTo", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": ""}, {"key": "dfb640a0-94cb-4218-b90b-1573d7ba3805", "name": "Contact", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "S", "mandatory": false, "reference": "", "display_name": ""}]', NULL, '2020-06-08 08:25:49.617813', 1591604749);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('5cf37266-3473-1006-986f-9325122678b7', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Schedule', NULL, 6, 1, 1, '[{"key": "cfb640a0-84cb-4218-b90b-1573d7ba3805", "name": "schedule_at", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "DT", "mandatory": false, "reference": "", "display_name": ""}, {"key": "srb640a0-04cb-4218-b90b-1573d7ba3805", "name": "repeat", "value": "true", "config": false, "hidden": false, "unique": false, "data_type": "B", "mandatory": false, "reference": "", "display_name": ""}]', NULL, '2020-06-08 08:25:49.617813', 1591604749);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('ddd13a59-47a9-4661-8ed2-62947d572b31', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Delay', NULL, 7, 1, 1, '[{"key": "vfb640a0-04cb-4218-b90b-1573d7ba3805", "name": "delay_by", "value": "", "config": false, "hidden": false, "unique": false, "data_type": "M", "mandatory": false, "reference": "", "display_name": "Delay By"}, {"key": "rrb640a0-04cb-4218-b90b-1573d7ba3805", "name": "repeat", "value": "true", "config": false, "hidden": false, "unique": false, "data_type": "B", "mandatory": false, "reference": "", "display_name": "Repeat"}]', NULL, '2020-06-08 08:25:49.617813', 1591604749);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('033e8ce4-0cbf-4ee8-86be-306b583f618e', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'MailGun Integration', NULL, 4, 1, 1, '[{"key": "4a68900c-5697-4f64-9a47-e49291ff9218", "link": "", "meta": null, "name": "sender", "field": null, "value": "vijayasankar.jothi@wayplot.com", "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": "Sender"}, {"key": "c2a0b583-cfb0-4f03-8b36-587548704b13", "link": "", "meta": null, "name": "to", "field": null, "value": "", "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": "To"}, {"key": "a27bb6d0-67df-4542-a806-e0974bff2e27", "link": "", "meta": null, "name": "cc", "field": null, "value": "vijayasankarmail@gmail.com", "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": "CC"}, {"key": "e34c3e1e-62fe-44cb-8caa-23c4bbbfcefc", "link": "", "meta": null, "name": "subject", "field": null, "value": "", "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": "Subject"}, {"key": "aaed7f03-291c-4276-a687-cbd80dc1eb52", "link": "", "meta": null, "name": "body", "field": null, "value": "", "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": "Body"}]', NULL, '2020-05-31 05:16:27.059717', 1601373769);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('52c3b7d8-89c8-416f-8bff-d6fd72a9a2e2', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Status', NULL, 1, 1, 0, '[{"key": "3f68d326-963b-45f5-b863-3d8d32fd296a", "link": "", "meta": null, "name": "Name", "field": null, "value": "", "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": ""}, {"key": "209d81c0-0857-4b9b-b41d-eb7627ca40ca", "link": "", "meta": null, "name": "Color", "field": null, "value": null, "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": ""}]', NULL, '2020-09-29 11:14:10.248898', 1601378050);
INSERT INTO public.entities (entity_id, account_id, team_id, name, description, category, state, status, fieldsb, tags, created_at, updated_at) VALUES ('5fab8ae6-a334-4c29-9d5d-15f9117ac780', '3cf27266-3473-4006-984f-9325122678b7', '8cf27268-3473-4006-984f-9325122678b7', 'Customer', NULL, 1, 1, 0, '[{"key": "e247b7d1-d8ab-4ace-8b00-f0ef72bc8ce6", "link": "", "meta": null, "name": "First Name", "field": null, "value": null, "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": ""}, {"key": "829e8cf6-794b-49de-a6d1-1c1bb100d44a", "link": "", "meta": null, "name": "Email", "field": null, "value": null, "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": ""}, {"key": "b01ff3f7-0750-449c-8a88-ab0800c9a5f7", "link": "", "meta": null, "name": "Mobile Numbers", "field": null, "value": null, "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "L", "mandatory": false, "expression": "", "display_name": ""}, {"key": "f6e3c292-e705-4102-afec-3d5f50ac889b", "link": "", "meta": null, "name": "NPS Score", "field": null, "value": null, "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "N", "mandatory": false, "expression": "", "display_name": ""}, {"key": "dadc9197-100e-4c05-b556-741cc39294e8", "link": "", "meta": null, "name": "Status", "field": {"key": "id", "link": "", "meta": null, "name": "", "field": null, "value": null, "config": false, "hidden": false, "ref_id": "", "unique": false, "choices": null, "dom_type": "", "data_type": "S", "mandatory": false, "expression": "", "display_name": ""}, "value": null, "config": false, "hidden": false, "ref_id": "52c3b7d8-89c8-416f-8bff-d6fd72a9a2e2", "unique": false, "choices": null, "dom_type": "", "data_type": "R", "mandatory": false, "expression": "", "display_name": ""}]', NULL, '2020-09-29 11:42:59.996032', 1601379779);


--
-- PostgreSQL database dump complete
--

