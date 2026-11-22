package platform

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/mail"
	"net/url"
	"sort"
	"strings"
	"time"
)

type AgeBand struct {
	Type  string `json:"passenger_type"`
	Min   int    `json:"min_age"`
	Max   *int   `json:"max_age_exclusive"`
	Seats int    `json:"seats_per_person"`
}
type BookingDefaults struct {
	Hold    int       `json:"hold_minutes"`
	Deposit int       `json:"deposit_bps"`
	Balance int       `json:"balance_due_days"`
	Max     int       `json:"max_travellers"`
	Ages    []AgeBand `json:"age_bands"`
	Policy  string    `json:"cancellation_policy"`
}

func decode(raw []byte, v any) bool {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	return d.Decode(v) == nil
}
func validateSetting(key string, raw []byte) bool {
	switch key {
	case "booking_defaults":
		var b BookingDefaults
		if !decode(raw, &b) || b.Hold < 15 || b.Hold > 30 || b.Deposit < 1 || b.Deposit > 10000 || b.Balance < 1 || b.Balance > 365 || b.Max < 1 || b.Max > 20 || len(b.Policy) == 0 || len(b.Policy) > 10000 || len(b.Ages) != 3 {
			return false
		}
		sort.Slice(b.Ages, func(i, j int) bool { return b.Ages[i].Min < b.Ages[j].Min })
		seen := map[string]bool{}
		end := 0
		for i, v := range b.Ages {
			if seen[v.Type] || v.Min != end || v.Seats < 0 || v.Seats > 1 {
				return false
			}
			seen[v.Type] = true
			if v.Type != "adult" && v.Type != "child" && v.Type != "infant" {
				return false
			}
			if i == 2 {
				if v.Type != "adult" || v.Max != nil || v.Seats != 1 {
					return false
				}
			} else {
				if v.Max == nil || *v.Max <= v.Min {
					return false
				}
				end = *v.Max
			}
		}
		return true
	case "response_sla":
		var b struct {
			Days    int    `json:"business_days"`
			Message string `json:"message"`
		}
		return decode(raw, &b) && b.Days >= 1 && b.Days <= 30 && len(b.Message) > 0 && len(b.Message) <= 500
	case "company_profile":
		var b struct {
			Legal   string  `json:"legal_name"`
			Display string  `json:"display_name"`
			License *string `json:"license_number"`
			Address *string `json:"address"`
			Phone   *string `json:"hotline"`
			Email   *string `json:"email"`
			Zalo    *string `json:"zalo_url"`
			Office  *struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"office_coordinates"`
			Documents []string `json:"document_media_ids"`
			Demo      bool     `json:"is_demo"`
		}
		if !decode(raw, &b) || strings.TrimSpace(b.Legal) == "" || strings.TrimSpace(b.Display) == "" || len(b.Legal) > 200 || len(b.Display) > 200 {
			return false
		}
		for _, s := range []*string{b.License, b.Address, b.Phone, b.Email, b.Zalo} {
			if s != nil && (len(*s) > 500 || strings.ContainsAny(*s, "<>")) {
				return false
			}
		}
		if b.Email != nil && *b.Email != "" {
			if _, e := mail.ParseAddress(*b.Email); e != nil {
				return false
			}
		}
		if b.Zalo != nil && *b.Zalo != "" {
			u, e := url.Parse(*b.Zalo)
			if e != nil || u.Scheme != "https" || u.Host != "zalo.me" {
				return false
			}
		}
		if b.Office != nil && (b.Office.Latitude < -90 || b.Office.Latitude > 90 || b.Office.Longitude < -180 || b.Office.Longitude > 180) {
			return false
		}
		if len(b.Documents) > 20 {
			return false
		}
		seen := map[string]bool{}
		for _, id := range b.Documents {
			if uuid.Validate(id) != nil || seen[id] {
				return false
			}
			seen[id] = true
		}
		return true
	}
	return false
}
func (a *App) getSetting(c *gin.Context) {
	var value json.RawMessage
	var version int
	var at time.Time
	e := a.DB.QueryRow(c.Request.Context(), "SELECT value,version,updated_at FROM site_settings WHERE key=$1", c.Param("key")).Scan(&value, &version, &at)
	if e != nil {
		fail(c, 404, "NOT_FOUND")
		return
	}
	ok(c, 200, gin.H{"key": c.Param("key"), "value": value, "version": version, "updated_at": at})
}
func (a *App) putSetting(c *gin.Context) {
	var b struct {
		Value   json.RawMessage `json:"value"`
		Version int             `json:"expected_version"`
	}
	if !body(c, &b) {
		return
	}
	key := c.Param("key")
	if !validateSetting(key, b.Value) {
		fail(c, 422, "VALIDATION_ERROR")
		return
	}
	ctx := c.Request.Context()
	tx, e := a.DB.Begin(ctx)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer tx.Rollback(ctx)
	if key == "company_profile" {
		var profile struct {
			Documents []string `json:"document_media_ids"`
		}
		json.Unmarshal(b.Value, &profile)
		for _, id := range profile.Documents {
			var valid bool
			if tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1 AND visibility='public' AND status='ready' AND kind IN('image','document'))", id).Scan(&valid) != nil || !valid {
				fail(c, 422, "VALIDATION_ERROR")
				return
			}
		}
	}
	u := c.MustGet("user").(User)
	tag, e := tx.Exec(ctx, "UPDATE site_settings SET value=$2,version=version+1,updated_by=$3,updated_at=now() WHERE key=$1 AND version=$4", key, b.Value, u.ID, b.Version)
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	if tag.RowsAffected() != 1 {
		fail(c, 409, "VERSION_CONFLICT")
		return
	}
	if _, e = tx.Exec(ctx, "INSERT INTO audit_logs(actor_staff_id,action,entity_type,request_id,changed_fields) VALUES($1,'settings.update','site_settings',$2,jsonb_build_object('key',$3::text))", u.ID, c.GetString("request_id"), key); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	if e = tx.Commit(ctx); e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	a.getSetting(c)
}
func (a *App) publicSettings(c *gin.Context) {
	rows, e := a.DB.Query(c.Request.Context(), "SELECT key,value FROM site_settings WHERE key IN ('company_profile','response_sla')")
	if e != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	defer rows.Close()
	out := gin.H{}
	for rows.Next() {
		var key string
		var raw json.RawMessage
		if rows.Scan(&key, &raw) != nil {
			fail(c, 503, "SERVICE_UNAVAILABLE")
			return
		}
		out[key] = raw
	}
	ok(c, 200, out)
}
func (a *App) publicDefaults(c *gin.Context) {
	var raw []byte
	if a.DB.QueryRow(c.Request.Context(), "SELECT value FROM site_settings WHERE key='booking_defaults'").Scan(&raw) != nil {
		fail(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	var b BookingDefaults
	json.Unmarshal(raw, &b)
	ok(c, 200, gin.H{"max_travellers": b.Max, "hold_minutes": b.Hold, "is_demo": true})
}
