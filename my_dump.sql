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
-- Data for Name: items; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('9d53277d-bf0c-4baf-bb86-ce61259dab44', NULL, 'd9ccf588-e6eb-40b3-838f-f6d5262bac78', 0, '{"9f9ade37-9549-4d12-a82d-c69495e85980": "down", "d3e572e1-3950-46db-a230-d41b2f4cd8d0": "2020-05-16 12:49:59.279275", "fa5e4f5e-b623-417b-a030-3c1b4385dbc0": "2021-05-16 12:49:59.279275"}', '2020-05-30 07:44:05.760548', 1590824645);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('44e5918f-2cbe-4d62-92d2-86820adff0cd', NULL, 'adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "vijayasankarmail@gmail.com", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Vijay", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "FreshW", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "10000"}', '2020-05-31 04:55:22.480538', 1590900922);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('8670ef39-a38a-44c3-b8a2-684276a4e673', NULL, 'adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "saravanaprakas@gmail.com ", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Saravana", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "Zoho", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "200000"}', '2020-05-31 04:57:14.844344', 1590901034);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('7d9c4f94-890b-484c-8189-91c3d7e8e50b', NULL, 'adbd74c7-7add-4dcd-b2cf-6b05863b90e8', 0, '{"08320990-cc56-4809-801a-a937b62ec307": "sksenthilkumaar@gmail.com ", "2bf431f8-b2ae-467f-9c5b-e7216068ea40": "Senthil", "900d69bf-2fc7-4c34-95b1-ef9f79220810": "Qatar Airways", "bf3cfc1d-a170-473f-b52b-4fef7495a0e3": "500000"}', '2020-05-31 04:57:46.445474', 1590901066);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('7c766083-83af-4926-b980-37de0f9edde0', NULL, '033e8ce4-0cbf-4ee8-86be-306b583f618e', 0, '{"4a68900c-5697-4f64-9a47-e49291ff9218": "vijayasankar.jothi@wayplot.com", "921ecaab-b3f0-42b6-a581-29239cc58e4b": "sandbox3ab4868d173f4391805389718914b89c.mailgun.org", "a27bb6d0-67df-4542-a806-e0974bff2e27": "vijayasankarmobile@gmail.com", "a8376197-b699-4f4b-b2dd-2bf5aa18ee16": "9c2d8fbbab5c0ca5de49089c1e9777b3-7fba8a4e-b5d71e35", "aaed7f03-291c-4276-a687-cbd80dc1eb52": "This mail is sent you to tell that your MRR is {{adbd74c7-7add-4dcd-b2cf-6b05863b90e8.bf3cfc1d-a170-473f-b52b-4fef7495a0e3}}. We are very proud of you! ", "c2a0b583-cfb0-4f03-8b36-587548704b13": "{{adbd74c7-7add-4dcd-b2cf-6b05863b90e8.08320990-cc56-4809-801a-a937b62ec307}}", "e34c3e1e-62fe-44cb-8caa-23c4bbbfcefc": "Hello {{adbd74c7-7add-4dcd-b2cf-6b05863b90e8.2bf431f8-b2ae-467f-9c5b-e7216068ea40}}"}', '2020-05-31 05:26:54.805027', 1590902814);
INSERT INTO public.items (item_id, parent_item_id, entity_id, state, input, created_at, updated_at) VALUES ('3d247443-b257-4b06-ba99-493cf9d83ce7', NULL, 'fcf13a59-47a9-4661-8ed2-62947d572b31', 0, '{"a57f650c-211c-49cb-ae56-d141cb380342": "Ummm. Prepare the research documents and make a call", "be084c25-7f85-4a89-af21-a0dbaa49a7e8": "agents.id", "dfb640a0-94cb-4218-b90b-1573d7ba3805": "44e5918f-2cbe-4d62-92d2-86820adff0cd"}', '2020-06-08 14:04:58.523412', 1591625098);


--
-- PostgreSQL database dump complete
--

