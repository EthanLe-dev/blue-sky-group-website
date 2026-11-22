package booking

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/catalog"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

type Counts struct {
	Adult  int `json:"adult"`
	Child  int `json:"child"`
	Infant int `json:"infant"`
}

func (c Counts) Total() int { return c.Adult + c.Child + c.Infant }
func (c Counts) Get(t string) int {
	switch t {
	case "adult":
		return c.Adult
	case "child":
		return c.Child
	case "infant":
		return c.Infant
	}
	return 0
}

type QuoteInput struct {
	Departure string   `json:"departure_id"`
	Counts    Counts   `json:"counts"`
	Codes     []string `json:"surcharge_codes"`
}
type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
type Passenger struct {
	Type  string `json:"passenger_type"`
	Name  string `json:"full_name"`
	Birth string `json:"date_of_birth"`
}
type Consent struct {
	Accepted bool   `json:"accepted"`
	Version  string `json:"policy_version"`
}
type CreateInput struct {
	Quote      string      `json:"quote_id"`
	Contact    Contact     `json:"contact"`
	Passengers []Passenger `json:"passengers"`
	Option     string      `json:"payment_option"`
	Requests   string      `json:"special_requests"`
	Consent    Consent     `json:"consent"`
	Captcha    string      `json:"turnstile_token"`
	Source     string      `json:"source"`
	Customer   *string     `json:"customer_id"`
}
type Line struct {
	Code     string        `json:"code"`
	Label    string        `json:"label"`
	Quantity int           `json:"quantity"`
	Unit     catalog.Money `json:"unit_amount"`
	Total    catalog.Money `json:"total_amount"`
}
type Snapshot struct {
	Lines    []Line          `json:"items"`
	Prices   []catalog.Price `json:"age_bands"`
	Total    catalog.Money   `json:"total_amount"`
	Deposit  catalog.Money   `json:"deposit_amount"`
	Balance  catalog.Money   `json:"balance_after_deposit"`
	Discount catalog.Money   `json:"discount_amount"`
	Seats    int             `json:"seat_count"`
	Policy   string          `json:"cancellation_policy"`
	Tour     json.RawMessage `json:"tour"`
	Start    time.Time       `json:"starts_at"`
	Due      time.Time       `json:"balance_due_at"`
	Options  []string        `json:"allowed_payment_options"`
}
type Defaults struct {
	Hold int `json:"hold_minutes"`
	Max  int `json:"max_travellers"`
}
type Quote struct {
	ID              string    `json:"id"`
	Departure       string    `json:"departure_id"`
	Version         int       `json:"price_version"`
	Counts          Counts    `json:"counts"`
	Seats           int       `json:"seat_count"`
	Snapshot        Snapshot  `json:"price_breakdown"`
	Options         []string  `json:"allowed_payment_options"`
	Due             time.Time `json:"balance_due_at"`
	Expires         time.Time `json:"expires_at"`
	Now             time.Time `json:"server_time"`
	SettingsVersion int       `json:"settings_version"`
	Hold            int       `json:"hold_minutes"`
	Max             int       `json:"max_travellers"`
}
type Principal struct {
	ID    string
	Staff bool
}

func (p Principal) Key() string {
	if p.Staff {
		return "staff:" + p.ID
	}
	return "guest:" + p.ID
}
func (p Principal) Values() (any, any) {
	if p.Staff {
		return nil, p.ID
	}
	return p.ID, nil
}

type Server struct {
	DB        *pgxpool.Pool
	Principal func(*gin.Context, bool) (Principal, bool)
	Captcha   func(context.Context, string, string) bool
}

var ErrNotFound = errors.New("NOT_FOUND")

func problem(code string) error { return errors.New(code) }

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

func ValidPhone(p string) bool { return phonePattern.MatchString(p) }

func NormalizePhone(p string) string {
	p = strings.NewReplacer(" ", "", "-", "", "(", "", ")", "").Replace(p)
	if strings.HasPrefix(p, "0") {
		p = "+84" + p[1:]
	}
	return p
}
func NormalizeContact(c *Contact) bool {
	c.Name = strings.TrimSpace(c.Name)
	c.Email = strings.ToLower(strings.TrimSpace(c.Email))
	c.Phone = NormalizePhone(c.Phone)
	a, e := mail.ParseAddress(c.Email)
	return e == nil && a.Address == c.Email && len(c.Email) <= 254 && len([]rune(c.Name)) >= 2 && len([]rune(c.Name)) <= 120 && ValidPhone(c.Phone)
}
func validatePassengers(list []Passenger, q Quote) bool {
	if len(list) != q.Counts.Total() {
		return false
	}
	counts := Counts{}
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	start := q.Snapshot.Start.In(loc)
	for _, p := range list {
		dob, e := time.Parse("2006-01-02", p.Birth)
		if e != nil || dob.After(start) || dob.Year() < start.Year()-120 || len([]rune(strings.TrimSpace(p.Name))) < 2 || len([]rune(p.Name)) > 120 {
			return false
		}
		age := start.Year() - dob.Year()
		if int(start.Month()) < int(dob.Month()) || start.Month() == dob.Month() && start.Day() < dob.Day() {
			age--
		}
		valid := false
		for _, band := range q.Snapshot.Prices {
			if band.Type == p.Type && age >= band.Min && (band.Max == nil || age < *band.Max) {
				valid = true
			}
		}
		if !valid {
			return false
		}
		switch p.Type {
		case "adult":
			counts.Adult++
		case "child":
			counts.Child++
		case "infant":
			counts.Infant++
		default:
			return false
		}
	}
	return counts == q.Counts
}
