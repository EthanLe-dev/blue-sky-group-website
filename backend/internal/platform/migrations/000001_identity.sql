CREATE TABLE staff_users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), email varchar(254) NOT NULL, password_hash text NOT NULL,
 full_name varchar(120) NOT NULL, status text NOT NULL DEFAULT 'active' CHECK(status IN ('active','disabled')),
 last_login_at timestamptz, created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX staff_email_unique ON staff_users(lower(email));
CREATE TABLE roles(id uuid PRIMARY KEY DEFAULT gen_random_uuid(), code text NOT NULL UNIQUE, name text NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE permissions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),code text NOT NULL UNIQUE, description text NOT NULL,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE staff_roles(staff_id uuid NOT NULL REFERENCES staff_users,role_id uuid NOT NULL REFERENCES roles,created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(staff_id,role_id));
CREATE TABLE role_permissions(role_id uuid NOT NULL REFERENCES roles,permission_id uuid NOT NULL REFERENCES permissions,created_at timestamptz NOT NULL DEFAULT now(),PRIMARY KEY(role_id,permission_id));
CREATE TABLE staff_sessions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),staff_id uuid NOT NULL REFERENCES staff_users,token_hash text NOT NULL UNIQUE,family_id uuid NOT NULL,expires_at timestamptz NOT NULL,revoked_at timestamptz,replaced_by uuid REFERENCES staff_sessions,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE INDEX staff_sessions_staff ON staff_sessions(staff_id,expires_at);
CREATE TABLE guest_sessions(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),token_hash text NOT NULL UNIQUE,expires_at timestamptz NOT NULL,revoked_at timestamptz,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE password_reset_tokens(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),staff_id uuid NOT NULL REFERENCES staff_users,token_hash text NOT NULL UNIQUE,expires_at timestamptz NOT NULL,used_at timestamptz,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE customers(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),full_name varchar(120) NOT NULL,email varchar(254),phone varchar(20),company_name varchar(200),notes text,archived_at timestamptz,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now(),CHECK(email IS NOT NULL OR phone IS NOT NULL));
CREATE TABLE audit_logs(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),actor_staff_id uuid REFERENCES staff_users,action text NOT NULL,entity_type text NOT NULL,entity_id uuid,request_id text NOT NULL,changed_fields jsonb NOT NULL DEFAULT '{}',reason text,created_at timestamptz NOT NULL DEFAULT now());
CREATE TABLE rate_limit_buckets(key_hash text PRIMARY KEY,window_start timestamptz NOT NULL,count int NOT NULL CHECK(count>=0),expires_at timestamptz NOT NULL);
CREATE TABLE mock_password_deliveries(id uuid PRIMARY KEY DEFAULT gen_random_uuid(),staff_id uuid NOT NULL REFERENCES staff_users,encrypted_token text NOT NULL,expires_at timestamptz NOT NULL,created_at timestamptz NOT NULL DEFAULT now());
INSERT INTO roles(code,name) VALUES('admin','Quản trị'),('product_manager','Sản phẩm'),('sales','Tư vấn'),('accountant','Kế toán'),('executive','Lãnh đạo');
INSERT INTO permissions(code,description) SELECT x,x FROM unnest(ARRAY['catalog.write','content.write','media.write','booking.read','booking.write','customer.write','lead.write','notification.read','media.private.write','payment.write','report.read','report.export','staff.write','audit.read','settings.write']) x;
INSERT INTO role_permissions SELECT r.id,p.id FROM roles r CROSS JOIN permissions p WHERE r.code='admin' OR
(r.code='product_manager' AND p.code IN('catalog.write','content.write','media.write')) OR
(r.code='sales' AND p.code IN('booking.read','booking.write','customer.write','lead.write','notification.read','media.private.write')) OR
(r.code='accountant' AND p.code IN('booking.read','payment.write','report.read','report.export')) OR
(r.code='executive' AND p.code='report.read');
