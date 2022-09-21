--
-- PostgreSQL database dump
--

-- Dumped from database version 14.4
-- Dumped by pg_dump version 14.4

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: tokens; Type: TABLE DATA; Schema: public; Owner: postgres
--

INSERT INTO public.tokens (token, account_id, type, state, scope, issued_at, expiry, created_at) VALUES ('eyJhbGciOiJSUzI1NiIsImtpZCI6IjEiLCJ0eXAiOiJKV1QifQ.eyJyb2xlcyI6W10sImV4cCI6MjI2ODU4MzgxMiwiaWF0IjoxNjYzNzgzODEyLCJzdWIiOiJjYTMxMmJjNi0zMGE1LTQ4YTgtYTU0ZC1iMTk3YjQyZjBjODEifQ.IrOr3cvLPr-rXxsbMsTffX0eRBVwtPLv4gFrz9bqTnI4dQF6x0DC255wfPWfK60jqv6n5CkzGSPok80Ltyp7RPg-pfjCQfvAQslXxDZozT90qS7VXksFc7b-TdHkpmEXlL0ffqiqmqAIOU5DUQ2jI8qRBKKdTr3W6w91pyq1uCq4JN9CpJ0M8N_W5aiVviFwm4luJ--QFxlGD5UsRwozE08GM5hUCUwLhSy85oW2bkWFQT22NGvpkUiAe8kWgU3ZuNBw-kS2s5j3kyJdPzzwlAxpsXQ7MHBpslW9x_M0Yt88_pKPcABUGORYv0WRUC4yObjOlrSPN4YLAB9VIfFaNQ', 'ca312bc6-30a5-48a8-a54d-b197b42f0c81', 0, 0, '{}', '2022-09-21 18:10:12.084639', '2041-11-20 18:10:12.084639', '2022-09-21 18:10:12.084639');


--
-- PostgreSQL database dump complete
--

