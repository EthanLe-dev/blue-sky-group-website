package platform

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type quoteWrite struct {
	Title      string        `json:"title"`
	Itinerary  string        `json:"itinerary"`
	Inclusions string        `json:"inclusions"`
	Exclusions string        `json:"exclusions"`
	Amount     catalog.Money `json:"total_amount"`
	Currency   string        `json:"currency"`
	ValidUntil string        `json:"valid_until"`
}

func lockQuoteLead(ctx context.Context, tx pgx.Tx, id string, create bool) (string, string, error) {
	if !create {
		err := tx.QueryRow(ctx, "SELECT group_request_id FROM group_quotes WHERE id=$1", id).Scan(&id)
		if err != nil {
			return "", "", err
		}
	}
	var status string
	err := tx.QueryRow(ctx, "SELECT status FROM group_requests WHERE id=$1 FOR UPDATE", id).Scan(&status)
	return id, status, err
}

func (a *App) writeGroupQuote(create bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload json.RawMessage
		if !httpx.Body(c, &payload) {
			return
		}
		if uuid.Validate(id) != nil || len(payload) == 0 || payload[0] != '{' {
			salesError(c, errSalesValidation)
			return
		}
		ctx := c.Request.Context()
		code := 200
		if create {
			code = 201
		}
		a.salesWrite(c, code, func(tx pgx.Tx) (json.RawMessage, error) {
			leadID, status, err := lockQuoteLead(ctx, tx, id, create)
			if err != nil {
				return nil, err
			}
			if status != "preparing" && status != "quoted" {
				return nil, errSalesTransition
			}
			var b quoteWrite
			if !create {
				var quoteStatus string
				err = tx.QueryRow(ctx, `SELECT status,title,itinerary,inclusions,exclusions,total_amount,currency,valid_until::text
					FROM group_quotes WHERE id=$1`, id).Scan(&quoteStatus, &b.Title, &b.Itinerary, &b.Inclusions, &b.Exclusions, &b.Amount, &b.Currency, &b.ValidUntil)
				if err != nil {
					return nil, err
				}
				if quoteStatus != "draft" {
					return nil, errSalesTransition
				}
			}
			if !decode(payload, &b) {
				return nil, errSalesValidation
			}
			b.Title, b.Itinerary = strings.TrimSpace(b.Title), strings.TrimSpace(b.Itinerary)
			b.Inclusions, b.Exclusions = strings.TrimSpace(b.Inclusions), strings.TrimSpace(b.Exclusions)
			if !boundedText(b.Title, 1, 200) || !boundedText(b.Itinerary, 1, 20000) || !boundedText(b.Inclusions, 0, 5000) ||
				!boundedText(b.Exclusions, 0, 5000) || b.Amount < 1 || b.Amount > 1_000_000_000_000 || b.Currency != "VND" || !dateValue(b.ValidUntil) {
				return nil, errSalesValidation
			}
			var dateOK bool
			err = tx.QueryRow(ctx, `SELECT $1::text::date >= (clock_timestamp() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date`, b.ValidUntil).Scan(&dateOK)
			if err != nil {
				return nil, err
			}
			if !dateOK {
				return nil, errSalesValidation
			}
			var out json.RawMessage
			action := "update"
			quoteID := id
			if create {
				action = "create"
				err = tx.QueryRow(ctx, `INSERT INTO group_quotes AS q (group_request_id,version,title,itinerary,inclusions,exclusions,total_amount,currency,valid_until,created_by_staff_id)
					VALUES($1,1+(SELECT coalesce(max(version),0) FROM group_quotes WHERE group_request_id=$1),$2,$3,$4,$5,$6,$7,$8::text::date,$9)
					RETURNING q.id,to_jsonb(q)||jsonb_build_object('total_amount',q.total_amount::text)`, leadID, b.Title, b.Itinerary, b.Inclusions, b.Exclusions, b.Amount, b.Currency, b.ValidUntil, c.MustGet("user").(User).ID).Scan(&quoteID, &out)
			} else {
				err = tx.QueryRow(ctx, `UPDATE group_quotes q SET title=$2,itinerary=$3,inclusions=$4,exclusions=$5,total_amount=$6,
					currency=$7,valid_until=$8::text::date,updated_at=clock_timestamp() WHERE id=$1 RETURNING to_jsonb(q)||jsonb_build_object('total_amount',q.total_amount::text)`,
					id, b.Title, b.Itinerary, b.Inclusions, b.Exclusions, b.Amount, b.Currency, b.ValidUntil).Scan(&out)
			}
			if err != nil {
				return nil, err
			}
			return out, salesAudit(ctx, tx, c, action, "group_quotes", quoteID, gin.H{"group_request_id": leadID})
		})
	}
}

func (a *App) transitionQuote(issue bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var b struct {
			Status string `json:"status"`
			Reason string `json:"reason"`
		}
		if !httpx.Body(c, &b) {
			return
		}
		b.Reason = strings.TrimSpace(b.Reason)
		if uuid.Validate(id) != nil || issue && (b.Status != "" || b.Reason != "") ||
			!issue && ((b.Status != "accepted" && b.Status != "rejected") || !boundedText(b.Reason, 1, 500)) {
			salesError(c, errSalesValidation)
			return
		}
		ctx := c.Request.Context()
		a.salesWrite(c, 200, func(tx pgx.Tx) (json.RawMessage, error) {
			leadID, leadStatus, err := lockQuoteLead(ctx, tx, id, false)
			if err != nil {
				return nil, err
			}
			var status string
			var current bool
			err = tx.QueryRow(ctx, `SELECT status,valid_until >= (clock_timestamp() AT TIME ZONE 'Asia/Ho_Chi_Minh')::date
				FROM group_quotes WHERE id=$1`, id).Scan(&status, &current)
			if err != nil {
				return nil, err
			}
			var out json.RawMessage
			switch {
			case issue && status == "draft" && current && (leadStatus == "preparing" || leadStatus == "quoted"):
				b.Status = "issued"
				leadStatus = "quoted"
				err = tx.QueryRow(ctx, `UPDATE group_quotes q SET status='issued',issued_at=clock_timestamp(),updated_at=clock_timestamp()
					WHERE id=$1 RETURNING to_jsonb(q)||jsonb_build_object('total_amount',q.total_amount::text)`, id).Scan(&out)
			case !issue && status == "issued" && leadStatus == "quoted" && (b.Status == "rejected" || current):
				if b.Status == "accepted" {
					leadStatus = "won"
				}
				err = tx.QueryRow(ctx, `UPDATE group_quotes q SET status=$2,decision_reason=$3,decided_by_staff_id=$4,
					decided_at=clock_timestamp(),updated_at=clock_timestamp() WHERE id=$1 RETURNING to_jsonb(q)||jsonb_build_object('total_amount',q.total_amount::text)`, id, b.Status, b.Reason, c.MustGet("user").(User).ID).Scan(&out)
			default:
				return nil, errSalesTransition
			}
			if err != nil {
				return nil, err
			}
			if _, err = tx.Exec(ctx, "UPDATE group_requests SET status=$2,updated_at=clock_timestamp() WHERE id=$1", leadID, leadStatus); err != nil {
				return nil, err
			}
			return out, salesAudit(ctx, tx, c, b.Status, "group_quotes", id, gin.H{"group_request_id": leadID, "lead_status": leadStatus})
		})
	}
}
