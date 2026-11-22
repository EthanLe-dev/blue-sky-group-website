package platform

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errSalesValidation = errors.New("sales validation")
var errSalesTransition = errors.New("sales transition")

func salesError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		httpx.Error(c, 404, "NOT_FOUND")
	case errors.Is(err, errSalesValidation):
		httpx.Error(c, 422, "VALIDATION_ERROR")
	case errors.Is(err, errSalesTransition):
		httpx.Error(c, 409, "INVALID_STATE")
	case err.Error() == "IDEMPOTENCY_KEY_REQUIRED":
		httpx.Error(c, 422, "IDEMPOTENCY_KEY_REQUIRED")
	case err.Error() == "IDEMPOTENCY_CONFLICT", err.Error() == "REQUEST_IN_PROGRESS":
		httpx.Error(c, 409, err.Error())
	default:
		httpx.Error(c, 503, "SERVICE_UNAVAILABLE")
	}
}

func leadJSON(r leadResource) string {
	if r.table == groupLead.table {
		return "(to_jsonb(l) || jsonb_build_object('budget_per_person',l.budget_per_person::text))"
	}
	return "to_jsonb(l)"
}

type leadResource struct {
	path, table, notes, foreignKey string
	statuses                       []string
}

var groupLead = leadResource{"/group-requests", "group_requests", "group_request_notes", "group_request_id", []string{"new", "preparing", "quoted", "won", "lost"}}
var callbackLead = leadResource{"/callback-requests", "callback_requests", "callback_notes", "callback_request_id", []string{"new", "contacted", "closed"}}

func (a *App) registerSales(v, admin *gin.RouterGroup) {
	v.POST(groupLead.path, a.origin, a.guestAccess, a.limit("lead", 10, 600), a.groupRequest)
	v.POST(callbackLead.path, a.origin, a.guestAccess, a.limit("lead", 10, 600), a.callbackRequest)
	leads := admin.Group("", a.permission("lead.write"))
	leads.GET("/lead-assignees", a.leadAssignees)
	for _, resource := range []leadResource{groupLead, callbackLead} {
		leads.GET(resource.path, a.leadList(resource))
		leads.GET(resource.path+"/:id", a.leadDetail(resource))
		leads.PATCH(resource.path+"/:id", a.patchLead(resource))
		leads.POST(resource.path+"/:id/notes", a.addLeadNote(resource))
	}
	leads.POST("/group-requests/:id/quotes", a.writeGroupQuote(true))
	leads.PATCH("/group-quotes/:id", a.writeGroupQuote(false))
	leads.POST("/group-quotes/:id/issue", a.transitionQuote(true))
	leads.POST("/group-quotes/:id/decision", a.transitionQuote(false))
	leads.POST("/group-requests/:id/attachments", a.permission("media.private.write"), a.attachLead)
}

const eligibleStaff = `FROM staff_users u WHERE u.status='active' AND EXISTS(
	SELECT 1 FROM staff_roles r JOIN role_permissions rp USING(role_id)
	JOIN permissions p ON p.id=rp.permission_id WHERE r.staff_id=u.id AND p.code='lead.write')`

func (a *App) leadAssignees(c *gin.Context) {
	out, err := httpx.JSONRows(c.Request.Context(), a.DB, `SELECT jsonb_build_object('id',u.id,'name',u.full_name) `+eligibleStaff+` ORDER BY u.full_name,u.id`)
	if err != nil {
		salesError(c, err)
		return
	}
	httpx.OK(c, 200, out)
}

func optionalDate(s string) bool { return s == "" || dateValue(s) }

func (a *App) leadList(r leadResource) gin.HandlerFunc {
	return func(c *gin.Context) {
		p, n, valid := httpx.Page(c)
		status, assignee, from, to := c.Query("status"), c.Query("assigned_to"), c.Query("from"), c.Query("to")
		if !valid || status != "" && !hasStatus(r.statuses, status) || assignee != "" && uuid.Validate(assignee) != nil ||
			!optionalDate(from) || !optionalDate(to) || from != "" && to != "" && from > to {
			salesError(c, errSalesValidation)
			return
		}
		filter := ` FROM ` + r.table + ` l WHERE ($1='' OR status=$1) AND ($2='' OR assigned_to=nullif($2,'')::uuid)
			AND ($3='' OR (created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date>=nullif($3,'')::date)
			AND ($4='' OR (created_at AT TIME ZONE 'Asia/Ho_Chi_Minh')::date<=nullif($4,'')::date)`
		ctx := c.Request.Context()
		var count int
		if err := a.DB.QueryRow(ctx, "SELECT count(*)"+filter, status, assignee, from, to).Scan(&count); err != nil {
			salesError(c, err)
			return
		}
		out, err := httpx.JSONRows(ctx, a.DB, "SELECT "+leadJSON(r)+"-'consent_version'-'consent_at'"+filter+" ORDER BY created_at DESC,id DESC LIMIT $5 OFFSET $6", status, assignee, from, to, n, (p-1)*n)
		if err != nil {
			salesError(c, err)
			return
		}
		httpx.List(c, out, p, n, count)
	}
}

func hasStatus(statuses []string, value string) bool {
	for _, status := range statuses {
		if status == value {
			return true
		}
	}
	return false
}

func (a *App) leadDetail(r leadResource) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if uuid.Validate(id) != nil {
			salesError(c, errSalesValidation)
			return
		}
		fields := `jsonb_build_object('notes',coalesce((SELECT jsonb_agg(to_jsonb(n) ORDER BY n.created_at,n.id)
			FROM ` + r.notes + ` n WHERE n.` + r.foreignKey + `=l.id),'[]'::jsonb))`
		if r.table == groupLead.table {
			fields += ` || jsonb_build_object(
				'quotes',coalesce((SELECT jsonb_agg(to_jsonb(q)||jsonb_build_object('total_amount',q.total_amount::text) ORDER BY q.version DESC) FROM group_quotes q WHERE q.group_request_id=l.id),'[]'::jsonb),
				'attachments',coalesce((SELECT jsonb_agg(jsonb_build_object('media_id',m.id,'name',m.original_name) ORDER BY m.created_at,m.id)
				FROM group_request_attachments a JOIN media_assets m ON m.id=a.media_id
				WHERE a.group_request_id=l.id AND m.visibility='private' AND m.status='ready'),'[]'::jsonb))`
		}
		var out json.RawMessage
		err := a.DB.QueryRow(c.Request.Context(), "SELECT "+leadJSON(r)+" || "+fields+" FROM "+r.table+" l WHERE id=$1", id).Scan(&out)
		if err != nil {
			salesError(c, err)
			return
		}
		httpx.OK(c, 200, out)
	}
}

func salesAudit(ctx context.Context, tx pgx.Tx, c *gin.Context, action, entity, id string, fields gin.H) error {
	_, err := tx.Exec(ctx, `INSERT INTO audit_logs(actor_staff_id,action,entity_type,entity_id,request_id,changed_fields)
		VALUES($1,$2,$3,$4,$5,$6)`, c.MustGet("user").(User).ID, action, entity, id, c.GetString("request_id"), fields)
	return err
}

func (a *App) salesWrite(c *gin.Context, status int, work func(pgx.Tx) (json.RawMessage, error)) {
	ctx := c.Request.Context()
	tx, err := a.DB.Begin(ctx)
	if err != nil {
		salesError(c, err)
		return
	}
	defer tx.Rollback(ctx)
	out, err := work(tx)
	if err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		salesError(c, err)
		return
	}
	httpx.OK(c, status, out)
}

type leadPatch struct {
	Status   string          `json:"status"`
	Assignee json.RawMessage `json:"assigned_to"`
	Due      json.RawMessage `json:"follow_up_due_at"`
	Reason   string          `json:"reason"`
}

func canMoveLead(group bool, from, to, reason string) bool {
	if to == "" || from == to {
		return true
	}
	if !group {
		return from == "new" && to == "contacted" || from == "contacted" && to == "closed"
	}
	return from == "new" && to == "preparing" || from == "preparing" && to == "lost" ||
		from == "quoted" && to == "lost" || from == "lost" && to == "preparing" && reason != ""
}

func (a *App) patchLead(r leadResource) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b leadPatch
		if !httpx.Body(c, &b) {
			return
		}
		id := c.Param("id")
		b.Reason = strings.TrimSpace(b.Reason)
		var assignee *string
		var due *time.Time
		if uuid.Validate(id) != nil || b.Status != "" && !hasStatus(r.statuses, b.Status) ||
			!boundedText(b.Reason, 0, 500) ||
			len(b.Assignee) > 0 && (json.Unmarshal(b.Assignee, &assignee) != nil || assignee != nil && uuid.Validate(*assignee) != nil) ||
			len(b.Due) > 0 && json.Unmarshal(b.Due, &due) != nil {
			salesError(c, errSalesValidation)
			return
		}
		ctx := c.Request.Context()
		a.salesWrite(c, 200, func(tx pgx.Tx) (json.RawMessage, error) {
			var status string
			if err := tx.QueryRow(ctx, "SELECT status FROM "+r.table+" WHERE id=$1 FOR UPDATE", id).Scan(&status); err != nil {
				return nil, err
			}
			if !canMoveLead(r.table == groupLead.table, status, b.Status, b.Reason) {
				return nil, errSalesTransition
			}
			if assignee != nil {
				var valid bool
				if err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 "+eligibleStaff+" AND u.id=$1)", assignee).Scan(&valid); err != nil {
					return nil, err
				}
				if !valid {
					return nil, errSalesValidation
				}
			}
			var out json.RawMessage
			err := tx.QueryRow(ctx, `UPDATE `+r.table+` l SET status=coalesce(nullif($2,''),status),
				assigned_to=CASE WHEN $3 THEN $4 ELSE assigned_to END,
				follow_up_due_at=CASE WHEN $5 THEN $6 ELSE follow_up_due_at END,updated_at=clock_timestamp()
				WHERE id=$1 RETURNING `+leadJSON(r), id, b.Status, len(b.Assignee) > 0, assignee, len(b.Due) > 0, due).Scan(&out)
			if err != nil {
				return nil, err
			}
			if status == "lost" && b.Status == "preparing" {
				_, err = tx.Exec(ctx, `INSERT INTO group_request_notes(group_request_id,staff_id,body) VALUES($1,$2,$3)`, id, c.MustGet("user").(User).ID, b.Reason)
				if err != nil {
					return nil, err
				}
			}
			return out, salesAudit(ctx, tx, c, "update", r.table, id, gin.H{"status": b.Status, "assignment_changed": len(b.Assignee) > 0, "due_changed": len(b.Due) > 0, "reason_recorded": b.Reason != ""})
		})
	}
}

func (a *App) addLeadNote(r leadResource) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b struct {
			Body string `json:"body"`
		}
		if !httpx.Body(c, &b) {
			return
		}
		id := c.Param("id")
		b.Body = strings.TrimSpace(b.Body)
		if uuid.Validate(id) != nil || !boundedText(b.Body, 1, 5000) {
			salesError(c, errSalesValidation)
			return
		}
		ctx := c.Request.Context()
		a.salesWrite(c, 201, func(tx pgx.Tx) (json.RawMessage, error) {
			var out json.RawMessage
			err := tx.QueryRow(ctx, `INSERT INTO `+r.notes+` AS n (`+r.foreignKey+`,staff_id,body)
				SELECT id,$2,$3 FROM `+r.table+` WHERE id=$1 RETURNING to_jsonb(n)`, id, c.MustGet("user").(User).ID, b.Body).Scan(&out)
			if err != nil {
				return nil, err
			}
			return out, salesAudit(ctx, tx, c, "add_note", r.table, id, gin.H{})
		})
	}
}

func (a *App) attachLead(c *gin.Context) {
	var b struct {
		Media string `json:"media_id"`
	}
	if !httpx.Body(c, &b) {
		return
	}
	id := c.Param("id")
	if uuid.Validate(id) != nil || uuid.Validate(b.Media) != nil {
		salesError(c, errSalesValidation)
		return
	}
	ctx := c.Request.Context()
	a.salesWrite(c, 201, func(tx pgx.Tx) (json.RawMessage, error) {
		var out json.RawMessage
		staff := c.MustGet("user").(User).ID
		err := tx.QueryRow(ctx, `WITH file AS (SELECT id FROM media_assets WHERE id=$2 AND visibility='private' AND status='ready' AND uploaded_by_staff_id=$3 FOR UPDATE)
			INSERT INTO group_request_attachments AS a(group_request_id,media_id,uploaded_by_staff_id)
			SELECT g.id,file.id,$3 FROM group_requests g CROSS JOIN file WHERE g.id=$1
			ON CONFLICT(group_request_id,media_id) DO UPDATE SET media_id=excluded.media_id RETURNING to_jsonb(a)`, id, b.Media, staff).Scan(&out)
		if err != nil {
			return nil, err
		}
		return out, salesAudit(ctx, tx, c, "attach_media", "group_requests", id, gin.H{})
	})
}
