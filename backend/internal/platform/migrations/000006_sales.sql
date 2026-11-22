CREATE TABLE group_requests (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), code varchar(24) NOT NULL UNIQUE,
 group_type text NOT NULL CHECK(group_type IN ('school','company')),
 organization_name varchar(200) NOT NULL, contact_name varchar(120) NOT NULL,
 contact_email varchar(254) NOT NULL, contact_phone varchar(20) NOT NULL,
 participant_count int NOT NULL CHECK(participant_count BETWEEN 1 AND 10000),
 preferred_start_date date, preferred_end_date date, destination_text varchar(200) NOT NULL,
 budget_per_person bigint CHECK(budget_per_person BETWEEN 0 AND 1000000000000),
 requirements text NOT NULL DEFAULT '',
 status text NOT NULL DEFAULT 'new' CHECK(status IN ('new','preparing','quoted','won','lost')),
 assigned_to uuid REFERENCES staff_users, follow_up_due_at timestamptz,
 consent_version text NOT NULL, consent_at timestamptz NOT NULL DEFAULT now(),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 CHECK(preferred_end_date IS NULL OR preferred_start_date IS NOT NULL AND preferred_end_date >= preferred_start_date)
);
CREATE INDEX group_requests_queue ON group_requests(status,assigned_to,follow_up_due_at);
CREATE TABLE callback_requests (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), code varchar(24) NOT NULL UNIQUE,
 name varchar(120) NOT NULL DEFAULT '', phone varchar(20) NOT NULL,
 preferred_time varchar(120) NOT NULL DEFAULT '', tour_id uuid REFERENCES tours,
 message text NOT NULL DEFAULT '',
 status text NOT NULL DEFAULT 'new' CHECK(status IN ('new','contacted','closed')),
 assigned_to uuid REFERENCES staff_users, follow_up_due_at timestamptz,
 consent_version text NOT NULL, consent_at timestamptz NOT NULL DEFAULT now(),
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX callback_requests_queue ON callback_requests(status,assigned_to,follow_up_due_at);
CREATE TABLE group_request_notes (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), group_request_id uuid NOT NULL REFERENCES group_requests,
 staff_id uuid NOT NULL REFERENCES staff_users, body text NOT NULL CHECK(length(body) BETWEEN 1 AND 5000),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX group_request_notes_history ON group_request_notes(group_request_id,created_at);
CREATE TABLE callback_notes (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), callback_request_id uuid NOT NULL REFERENCES callback_requests,
 staff_id uuid NOT NULL REFERENCES staff_users, body text NOT NULL CHECK(length(body) BETWEEN 1 AND 5000),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX callback_notes_history ON callback_notes(callback_request_id,created_at);
CREATE TABLE group_quotes (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), group_request_id uuid NOT NULL REFERENCES group_requests,
 version int NOT NULL CHECK(version>0), title varchar(200) NOT NULL, itinerary text NOT NULL,
 inclusions text NOT NULL, exclusions text NOT NULL,
 total_amount bigint NOT NULL CHECK(total_amount BETWEEN 1 AND 1000000000000),
 currency char(3) NOT NULL CHECK(currency='VND'), valid_until date NOT NULL,
 status text NOT NULL DEFAULT 'draft' CHECK(status IN ('draft','issued','accepted','rejected')),
 created_by_staff_id uuid NOT NULL REFERENCES staff_users,
 issued_at timestamptz, decided_at timestamptz, decided_by_staff_id uuid REFERENCES staff_users,
 decision_reason text,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now(),
 UNIQUE(group_request_id,version)
);
CREATE UNIQUE INDEX group_quotes_accepted ON group_quotes(group_request_id) WHERE status='accepted';
CREATE TABLE group_request_attachments (
 group_request_id uuid NOT NULL REFERENCES group_requests, media_id uuid NOT NULL REFERENCES media_assets,
 uploaded_by_staff_id uuid NOT NULL REFERENCES staff_users, created_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(group_request_id,media_id)
);
