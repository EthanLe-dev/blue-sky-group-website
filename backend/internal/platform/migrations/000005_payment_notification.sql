ALTER TABLE bookings ADD COLUMN refunded_amount bigint NOT NULL DEFAULT 0 CHECK(refunded_amount BETWEEN 0 AND paid_amount),
 ADD COLUMN confirmed_at timestamptz;

CREATE TABLE payments (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 booking_id uuid NOT NULL REFERENCES bookings,
 provider text NOT NULL DEFAULT 'mock' CHECK(provider='mock'),
 purpose text NOT NULL CHECK(purpose IN ('full','deposit','balance')),
 amount bigint NOT NULL CHECK(amount>0),
 currency char(3) NOT NULL DEFAULT 'VND' CHECK(currency='VND'),
 status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','succeeded','cancelled')),
 is_mock boolean NOT NULL DEFAULT true CHECK(is_mock),
 initiated_by_staff_id uuid REFERENCES staff_users,
 expires_at timestamptz NOT NULL,
 paid_at timestamptz,
 created_at timestamptz NOT NULL DEFAULT now(),
 CHECK((status='succeeded')=(paid_at IS NOT NULL))
);
CREATE UNIQUE INDEX payments_one_pending ON payments(booking_id) WHERE status='pending';
CREATE INDEX payments_booking ON payments(booking_id,created_at);
CREATE INDEX payments_cash_in ON payments(paid_at) WHERE status='succeeded';

CREATE TABLE payment_events (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 payment_id uuid NOT NULL REFERENCES payments,
 event_key text NOT NULL UNIQUE,
 event_type text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE refunds (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 payment_id uuid NOT NULL REFERENCES payments,
 amount bigint NOT NULL CHECK(amount>0),
 reason text NOT NULL CHECK(length(reason) BETWEEN 2 AND 1000),
 requested_by_staff_id uuid NOT NULL REFERENCES staff_users,
 status text NOT NULL DEFAULT 'succeeded' CHECK(status='succeeded'),
 is_mock boolean NOT NULL DEFAULT true CHECK(is_mock),
 completed_at timestamptz NOT NULL DEFAULT clock_timestamp(),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX refunds_payment ON refunds(payment_id,created_at);
CREATE INDEX refunds_cash_out ON refunds(completed_at);

CREATE TABLE outbox_events (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 event_key text NOT NULL UNIQUE,
 booking_id uuid NOT NULL REFERENCES bookings,
 event_type text NOT NULL CHECK(event_type IN ('booking_confirmed','balance_paid')),
 status text NOT NULL DEFAULT 'pending' CHECK(status IN ('pending','processing','done','failed')),
 attempts int NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 5),
 next_attempt_at timestamptz NOT NULL DEFAULT now(),
 lease_until timestamptz,
 last_error text,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX outbox_due ON outbox_events(next_attempt_at,created_at) WHERE status IN ('pending','processing');
CREATE TABLE notification_logs (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 booking_id uuid NOT NULL REFERENCES bookings,
 channel text NOT NULL CHECK(channel IN ('email','sms')),
 template_code text NOT NULL,
 subject text,
 body text NOT NULL,
 status text NOT NULL DEFAULT 'simulated' CHECK(status='simulated'),
 is_mock boolean NOT NULL DEFAULT true CHECK(is_mock),
 dedupe_key text NOT NULL UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notification_booking ON notification_logs(booking_id,created_at);
