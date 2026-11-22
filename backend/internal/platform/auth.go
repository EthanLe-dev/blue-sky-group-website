package platform

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

func secret() string {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		panic(e)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func hash(v string) string { b := sha256.Sum256([]byte(v)); return hex.EncodeToString(b[:]) }
func (a *App) csrf(token string) string {
	h := hmac.New(sha256.New, []byte(a.Config.CSRFSecret))
	h.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
func Bootstrap(ctx context.Context, p *pgxpool.Pool, email, password string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if !strings.Contains(email, "@") || len(password) < 12 || len(password) > 72 {
		return errors.New("valid email and password of 12..72 bytes required")
	}
	b, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if e != nil {
		return e
	}
	tx, e := p.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	var id string
	e = tx.QueryRow(ctx, "INSERT INTO staff_users(email,password_hash,full_name) VALUES($1,$2,'Quản trị demo') ON CONFLICT (lower(email)) DO NOTHING RETURNING id", email, string(b)).Scan(&id)
	if e != nil {
		var exists bool
		if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM staff_users WHERE lower(email)=$1)", email).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return nil
		}
		return e
	}
	if _, e = tx.Exec(ctx, "INSERT INTO staff_roles(staff_id,role_id) SELECT $1,id FROM roles WHERE code='admin'", id); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

type User struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	FullName    string   `json:"full_name"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

func (a *App) user(ctx context.Context, id string) (User, error) {
	u := User{Roles: []string{}, Permissions: []string{}}
	e := a.DB.QueryRow(ctx, "SELECT id,email,full_name FROM staff_users WHERE id=$1 AND status='active'", id).Scan(&u.ID, &u.Email, &u.FullName)
	if e != nil {
		return u, e
	}
	rows, e := a.DB.Query(ctx, "SELECT DISTINCT r.code,p.code FROM staff_roles s JOIN roles r ON r.id=s.role_id JOIN role_permissions rp ON rp.role_id=r.id JOIN permissions p ON p.id=rp.permission_id WHERE s.staff_id=$1", id)
	if e != nil {
		return u, e
	}
	defer rows.Close()
	roles, perms := map[string]bool{}, map[string]bool{}
	for rows.Next() {
		var r, p string
		if e = rows.Scan(&r, &p); e != nil {
			return u, e
		}
		if !roles[r] {
			u.Roles = append(u.Roles, r)
			roles[r] = true
		}
		if !perms[p] {
			u.Permissions = append(u.Permissions, p)
			perms[p] = true
		}
	}
	return u, rows.Err()
}
func (a *App) access(userID, sessionID string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": userID, "sid": sessionID, "iss": "blue-sky-api", "aud": "blue-sky-admin", "iat": time.Now().Unix(), "exp": time.Now().Add(15 * time.Minute).Unix()}).SignedString([]byte(a.Config.JWTSecret))
}
func (a *App) createSession(ctx context.Context, userID string) (string, string, error) {
	token, id := secret(), uuid.NewString()
	_, e := a.DB.Exec(ctx, "INSERT INTO staff_sessions(id,staff_id,token_hash,family_id,expires_at) VALUES($1,$2,$3,$1,now()+interval '7 days')", id, userID, hash(token))
	return token, id, e
}
func (a *App) encrypt(token string) (string, error) {
	key := sha256.Sum256([]byte(a.Config.CSRFSecret))
	b, e := aes.NewCipher(key[:])
	if e != nil {
		return "", e
	}
	g, e := cipher.NewGCM(b)
	if e != nil {
		return "", e
	}
	n := make([]byte, g.NonceSize())
	if _, e = rand.Read(n); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(g.Seal(n, n, []byte(token), nil)), nil
}
func (a *App) decrypt(value string) (string, error) {
	key := sha256.Sum256([]byte(a.Config.CSRFSecret))
	b, _ := aes.NewCipher(key[:])
	g, _ := cipher.NewGCM(b)
	raw, e := base64.RawURLEncoding.DecodeString(value)
	if e != nil || len(raw) < g.NonceSize() {
		return "", fmt.Errorf("invalid preview")
	}
	v, e := g.Open(nil, raw[:g.NonceSize()], raw[g.NonceSize():], nil)
	return string(v), e
}
