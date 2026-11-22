package platform

import (
	"encoding/json"
	"net/mail"
	"strings"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (a *App) registerCustomers(admin *gin.RouterGroup) {
	customers := admin.Group("/customers", a.permission("customer.write"))
	customers.GET("", a.listCustomers)
	customers.POST("", a.writeCustomer(true))
	customers.GET("/:id", a.customerDetail)
	customers.PATCH("/:id", a.writeCustomer(false))
	customers.GET("/:id/bookings", a.permission("booking.read"), a.customerBookings)
}

type customerWrite struct {
	Name    string  `json:"full_name"`
	Email   *string `json:"email"`
	Phone   *string `json:"phone"`
	Company *string `json:"company_name"`
	Notes   *string `json:"notes"`
}

func validateCustomer(b *customerWrite) bool {
	b.Name = strings.TrimSpace(b.Name)
	for _, field := range []**string{&b.Email, &b.Phone, &b.Company, &b.Notes} {
		if *field != nil {
			value := strings.TrimSpace(**field)
			*field = &value
			if value == "" {
				*field = nil
			}
		}
	}
	if !boundedText(b.Name, 2, 120) || b.Email == nil && b.Phone == nil ||
		b.Company != nil && !boundedText(*b.Company, 1, 200) || b.Notes != nil && !boundedText(*b.Notes, 1, 5000) {
		return false
	}
	if b.Email != nil {
		*b.Email = strings.ToLower(*b.Email)
		address, err := mail.ParseAddress(*b.Email)
		if err != nil || address.Address != *b.Email || len(*b.Email) > 254 {
			return false
		}
	}
	if b.Phone != nil {
		*b.Phone = booking.NormalizePhone(*b.Phone)
		if !booking.ValidPhone(*b.Phone) {
			return false
		}
	}
	return true
}

func (a *App) listCustomers(c *gin.Context) {
	p, n, valid := httpx.Page(c)
	q := strings.TrimSpace(c.Query("q"))
	if !valid || !boundedText(q, 0, 120) {
		salesError(c, errSalesValidation)
		return
	}
	ctx := c.Request.Context()
	const where = ` FROM customers g WHERE archived_at IS NULL AND ($1='' OR
		strpos(lower(full_name),lower($1))>0 OR strpos(coalesce(email,''),lower($1))>0 OR
		strpos(coalesce(phone,''),$1)>0 OR strpos(coalesce(phone,''),$2)>0 OR strpos(lower(coalesce(company_name,'')),lower($1))>0)`
	phone := booking.NormalizePhone(q)
	if phone == "" {
		phone = q
	}
	var count int
	if err := a.DB.QueryRow(ctx, "SELECT count(*)"+where, q, phone).Scan(&count); err != nil {
		salesError(c, err)
		return
	}
	out, err := httpx.JSONRows(ctx, a.DB, "SELECT to_jsonb(g)-'notes'-'archived_at'"+where+" ORDER BY created_at DESC,id DESC LIMIT $3 OFFSET $4", q, phone, n, (p-1)*n)
	if err != nil {
		salesError(c, err)
		return
	}
	httpx.List(c, out, p, n, count)
}

func (a *App) customerDetail(c *gin.Context) {
	id := c.Param("id")
	if uuid.Validate(id) != nil {
		salesError(c, errSalesValidation)
		return
	}
	var out json.RawMessage
	err := a.DB.QueryRow(c.Request.Context(), "SELECT to_jsonb(g)-'archived_at' FROM customers g WHERE id=$1 AND archived_at IS NULL", id).Scan(&out)
	if err != nil {
		salesError(c, err)
		return
	}
	httpx.OK(c, 200, out)
}

func (a *App) writeCustomer(create bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload json.RawMessage
		if !httpx.Body(c, &payload) {
			return
		}
		if len(payload) == 0 || payload[0] != '{' || !create && uuid.Validate(id) != nil {
			salesError(c, errSalesValidation)
			return
		}
		ctx := c.Request.Context()
		status := 200
		if create {
			status = 201
		}
		a.salesWrite(c, status, func(tx pgx.Tx) (json.RawMessage, error) {
			var b customerWrite
			if !create {
				err := tx.QueryRow(ctx, `SELECT full_name,email,phone,company_name,notes FROM customers
					WHERE id=$1 AND archived_at IS NULL FOR UPDATE`, id).Scan(&b.Name, &b.Email, &b.Phone, &b.Company, &b.Notes)
				if err != nil {
					return nil, err
				}
			}
			if !decode(payload, &b) || !validateCustomer(&b) {
				return nil, errSalesValidation
			}
			var out json.RawMessage
			var err error
			action := "update"
			if create {
				action = "create"
				err = tx.QueryRow(ctx, `INSERT INTO customers AS g(full_name,email,phone,company_name,notes)
					VALUES($1,$2,$3,$4,$5) RETURNING id,to_jsonb(g)-'archived_at'`, b.Name, b.Email, b.Phone, b.Company, b.Notes).Scan(&id, &out)
			} else {
				err = tx.QueryRow(ctx, `UPDATE customers g SET full_name=$2,email=$3,phone=$4,company_name=$5,notes=$6,
					updated_at=clock_timestamp() WHERE id=$1 RETURNING to_jsonb(g)-'archived_at'`, id, b.Name, b.Email, b.Phone, b.Company, b.Notes).Scan(&out)
			}
			if err != nil {
				return nil, err
			}
			return out, salesAudit(ctx, tx, c, action, "customers", id, gin.H{})
		})
	}
}

func (a *App) customerBookings(c *gin.Context) {
	id := c.Param("id")
	p, n, valid := httpx.Page(c)
	if !valid || uuid.Validate(id) != nil {
		salesError(c, errSalesValidation)
		return
	}
	ctx := c.Request.Context()
	var count int
	err := a.DB.QueryRow(ctx, `SELECT (SELECT count(*) FROM bookings WHERE customer_id=g.id AND customer_verified_at IS NOT NULL)
		FROM customers g WHERE g.id=$1 AND g.archived_at IS NULL`, id).Scan(&count)
	if err != nil {
		salesError(c, err)
		return
	}
	out, err := httpx.JSONRows(ctx, a.DB, `SELECT jsonb_build_object('id',id,'code',code,'status',status,
		'tour_title',tour_snapshot->>'title','created_at',created_at,'verified_at',customer_verified_at,'total_amount',total_amount::text)
		FROM bookings WHERE customer_id=$1 AND customer_verified_at IS NOT NULL ORDER BY created_at DESC,id DESC LIMIT $2 OFFSET $3`, id, n, (p-1)*n)
	if err != nil {
		salesError(c, err)
		return
	}
	httpx.List(c, out, p, n, count)
}
