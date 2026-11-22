package platform

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/booking"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const salesConsent = "sales-v1"

type groupRequestInput struct {
	Type         string          `json:"group_type"`
	Organization string          `json:"organization_name"`
	Name         string          `json:"contact_name"`
	Email        string          `json:"contact_email"`
	Phone        string          `json:"contact_phone"`
	Count        int             `json:"participant_count"`
	Destination  string          `json:"destination_text"`
	Start        *string         `json:"preferred_start_date"`
	End          *string         `json:"preferred_end_date"`
	Budget       *catalog.Money  `json:"budget_per_person"`
	Requirements string          `json:"requirements"`
	Consent      booking.Consent `json:"consent"`
	Token        string          `json:"turnstile_token,omitempty"`
}

type callbackInput struct {
	Name    string          `json:"name"`
	Phone   string          `json:"phone"`
	Time    string          `json:"preferred_time"`
	Tour    *string         `json:"tour_id"`
	Message string          `json:"message"`
	Consent booking.Consent `json:"consent"`
	Token   string          `json:"turnstile_token,omitempty"`
}

func boundedText(s string, min, max int) bool {
	n := utf8.RuneCountInString(s)
	return n >= min && n <= max && !strings.ContainsRune(s, 0)
}

func dateValue(s string) bool {
	d, err := time.Parse(time.DateOnly, s)
	return err == nil && d.Year() > 0
}

func (a *App) groupRequest(c *gin.Context) {
	var b groupRequestInput
	if !httpx.Body(c, &b) {
		return
	}
	contact := booking.Contact{Name: b.Name, Email: b.Email, Phone: b.Phone}
	contactValid := booking.NormalizeContact(&contact)
	b.Name, b.Email, b.Phone = contact.Name, contact.Email, contact.Phone
	b.Organization = strings.TrimSpace(b.Organization)
	b.Destination = strings.TrimSpace(b.Destination)
	b.Requirements = strings.TrimSpace(b.Requirements)
	if !b.Consent.Accepted || b.Consent.Version != salesConsent || !contactValid || !boundedText(b.Name, 2, 120) ||
		(b.Type != "school" && b.Type != "company") || !boundedText(b.Organization, 1, 200) ||
		!boundedText(b.Destination, 1, 200) || !boundedText(b.Requirements, 0, 5000) ||
		b.Count < 1 || b.Count > 10000 || b.Budget != nil && (*b.Budget < 0 || *b.Budget > 1_000_000_000_000) ||
		b.Start != nil && !dateValue(*b.Start) || b.End != nil && (b.Start == nil || !dateValue(*b.End) || *b.End < *b.Start) {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	token := b.Token
	b.Token = ""
	a.receiveLead(c, "group-request", "GR-", token, b, func(tx pgx.Tx, code string) error {
		ctx := c.Request.Context()
		if b.Start != nil {
			var today string
			if err := tx.QueryRow(ctx, `SELECT to_char(clock_timestamp() AT TIME ZONE 'Asia/Ho_Chi_Minh','YYYY-MM-DD')`).Scan(&today); err != nil {
				return err
			}
			if *b.Start < today {
				return errSalesValidation
			}
		}
		_, err := tx.Exec(ctx, `INSERT INTO group_requests(code,group_type,organization_name,contact_name,contact_email,contact_phone,participant_count,
			preferred_start_date,preferred_end_date,destination_text,budget_per_person,requirements,consent_version)
			VALUES($1,$2,$3,$4,$5,$6,$7,$8::text::date,$9::text::date,$10,$11,$12,$13)`,
			code, b.Type, b.Organization, b.Name, b.Email, b.Phone, b.Count, b.Start, b.End, b.Destination, b.Budget, b.Requirements, b.Consent.Version)
		return err
	})
}

func (a *App) callbackRequest(c *gin.Context) {
	var b callbackInput
	if !httpx.Body(c, &b) {
		return
	}
	b.Name, b.Time, b.Message = strings.TrimSpace(b.Name), strings.TrimSpace(b.Time), strings.TrimSpace(b.Message)
	b.Phone = booking.NormalizePhone(b.Phone)
	if !b.Consent.Accepted || b.Consent.Version != salesConsent || !booking.ValidPhone(b.Phone) ||
		!boundedText(b.Name, 0, 120) || !boundedText(b.Time, 0, 120) || !boundedText(b.Message, 0, 5000) ||
		b.Tour != nil && uuid.Validate(*b.Tour) != nil {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	token := b.Token
	b.Token = ""
	a.receiveLead(c, "callback-request", "CB-", token, b, func(tx pgx.Tx, code string) error {
		tag, err := tx.Exec(c.Request.Context(), `INSERT INTO callback_requests(code,name,phone,preferred_time,tour_id,message,consent_version)
			SELECT $1,$2,$3,$4,$5,$6,$7 WHERE $5::uuid IS NULL OR EXISTS(SELECT 1 FROM tours WHERE id=$5 AND status='published')`,
			code, b.Name, b.Phone, b.Time, b.Tour, b.Message, b.Consent.Version)
		if err == nil && tag.RowsAffected() == 0 {
			return errSalesValidation
		}
		return err
	})
}

func (a *App) receiveLead(c *gin.Context, action, prefix, token string, input any, insert func(pgx.Tx, string) error) {
	ctx := c.Request.Context()
	tx, err := a.DB.Begin(ctx)
	if err != nil {
		salesError(c, err)
		return
	}
	defer tx.Rollback(ctx)
	p := booking.Principal{ID: c.GetString("guest_id")}
	key := c.GetHeader("Idempotency-Key")
	cached, err := booking.Claim(ctx, tx, p, action, key, input)
	if err != nil {
		salesError(c, err)
		return
	}
	if len(cached) != 0 {
		httpx.OK(c, 201, cached)
		return
	}
	if !a.captcha(ctx, token, action) {
		httpx.Error(c, 403, "CAPTCHA_INVALID")
		return
	}

	code := prefix + strings.ToUpper(strings.ReplaceAll(uuid.NewString(), "-", "")[:20])
	if err = insert(tx, code); err != nil {
		salesError(c, err)
		return
	}
	var receipt json.RawMessage
	err = tx.QueryRow(ctx, `SELECT jsonb_build_object('code',$1::text,'status','new','received_at',now(),
		'response_sla_message',coalesce((SELECT value->>'message' FROM site_settings WHERE key='response_sla'),''),
		'is_demo',$2::boolean)`, code, a.Config.Env != "production").Scan(&receipt)
	if err == nil {
		err = booking.Remember(ctx, tx, p, action, key, receipt)
	}
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		salesError(c, err)
		return
	}
	httpx.OK(c, 201, receipt)
}
