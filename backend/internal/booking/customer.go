package booking

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func authorizeCustomerLink(ctx context.Context, tx pgx.Tx, staff Principal, customer string) error {
	var allowed bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM staff_users u JOIN staff_roles sr ON sr.staff_id=u.id
		JOIN role_permissions rp USING(role_id) JOIN permissions p ON p.id=rp.permission_id
		WHERE u.id=$1 AND u.status='active' AND p.code='customer.write')`, staff.ID).Scan(&allowed)
	if err != nil {
		return err
	}
	if !staff.Staff || !allowed {
		return problem("FORBIDDEN")
	}
	var id string
	err = tx.QueryRow(ctx, "SELECT id FROM customers WHERE id=$1 AND archived_at IS NULL FOR SHARE", customer).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func recordCustomerLink(ctx context.Context, tx pgx.Tx, booking, customer, staff string) error {
	tag, err := tx.Exec(ctx, `UPDATE bookings SET customer_id=$2,customer_verified_at=clock_timestamp(),customer_verified_by_staff_id=$3
		WHERE id=$1 AND (customer_verified_at IS NULL OR customer_id=$2)`, booking, customer, staff)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return problem("INVALID_STATE")
	}
	_, err = tx.Exec(ctx, "INSERT INTO booking_events(booking_id,event_type,actor_staff_id) VALUES($1,'customer_link_verified',$2)", booking, staff)
	return err
}
