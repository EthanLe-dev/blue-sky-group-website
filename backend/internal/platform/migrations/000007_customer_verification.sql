ALTER TABLE bookings ADD COLUMN customer_verified_at timestamptz,
 ADD COLUMN customer_verified_by_staff_id uuid REFERENCES staff_users,
 ADD CONSTRAINT booking_customer_verification CHECK(
  (customer_verified_at IS NULL AND customer_verified_by_staff_id IS NULL)
  OR (customer_id IS NOT NULL AND customer_verified_at IS NOT NULL AND customer_verified_by_staff_id IS NOT NULL)
 );
CREATE INDEX bookings_verified_customer ON bookings(customer_id,created_at DESC,id DESC) WHERE customer_verified_at IS NOT NULL;
CREATE INDEX customers_email ON customers(lower(email)) WHERE archived_at IS NULL;
CREATE INDEX customers_phone ON customers(phone) WHERE archived_at IS NULL;
