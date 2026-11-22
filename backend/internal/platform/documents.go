package platform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/exportfile"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (a *App) tourPDF(c *gin.Context) {
	ctx := c.Request.Context()
	tx, err := a.DB.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	defer tx.Rollback(ctx)
	var id, title, included, excluded string
	var days, nights int
	err = tx.QueryRow(ctx, `SELECT id,title,duration_days,duration_nights,included,excluded
		FROM tours WHERE slug=$1 AND status='published'`, c.Param("slug")).Scan(&id, &title, &days, &nights, &included, &excluded)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	rows, err := tx.Query(ctx, `SELECT day_number,title,body,coalesce(meals,''),coalesce(accommodation,'')
		FROM tour_days WHERE tour_id=$1 ORDER BY day_number`, id)
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	defer rows.Close()
	var sections []exportfile.Section
	for rows.Next() {
		var number int
		var day, body, meal, hotel string
		if err = rows.Scan(&number, &day, &body, &meal, &hotel); err != nil {
			exportReply(c, exportfile.File{}, "", err)
			return
		}
		lines := []string{exportfile.HTMLText(body)}
		if meal != "" {
			lines = append(lines, "Bữa ăn: "+meal)
		}
		if hotel != "" {
			lines = append(lines, "Lưu trú: "+hotel)
		}
		sections = append(sections, exportfile.Section{Title: fmt.Sprintf("Ngày %d · %s", number, day), Body: strings.Join(lines, "\n")})
	}
	if err = rows.Err(); err == nil {
		err = tx.Commit(ctx)
	}
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	if len(sections) == 0 {
		sections = append(sections, exportfile.Section{Title: "Lịch trình", Body: "Lịch trình chi tiết đang được cập nhật."})
	}
	sections = append(sections, exportfile.Section{Title: "Bao gồm", Body: exportfile.HTMLText(included)},
		exportfile.Section{Title: "Không bao gồm", Body: exportfile.HTMLText(excluded)})
	file, err := exportfile.PDF(title, fmt.Sprintf("Lịch trình · %d ngày %d đêm", days, nights), sections)
	exportReply(c, file, "lich-trinh-"+c.Param("slug")+".pdf", err)
}

func (a *App) groupQuotePDF(c *gin.Context) {
	id := c.Param("id")
	if uuid.Validate(id) != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	var title, code, status, itinerary, inclusions, exclusions, amount, valid string
	var version int
	err := a.DB.QueryRow(c.Request.Context(), `SELECT q.title,g.code,q.version,q.status,q.itinerary,q.inclusions,q.exclusions,
		to_char(q.total_amount,'FM999G999G999G999G999'),q.valid_until::text
		FROM group_quotes q JOIN group_requests g ON g.id=q.group_request_id WHERE q.id=$1`, id).
		Scan(&title, &code, &version, &status, &itinerary, &inclusions, &exclusions, &amount, &valid)
	if errors.Is(err, pgx.ErrNoRows) {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if err != nil {
		exportReply(c, exportfile.File{}, "", err)
		return
	}
	states := map[string]string{"draft": "Bản nháp", "issued": "Đã phát hành", "accepted": "Đã chấp nhận ngoài hệ thống", "rejected": "Đã từ chối"}
	meta := fmt.Sprintf("Báo giá đoàn · %s · Phiên bản %d · %s\nTổng tiền: %s VND · Hiệu lực đến %s", code, version, states[status], strings.ReplaceAll(amount, ",", "."), valid)
	file, err := exportfile.PDF(title, meta, []exportfile.Section{
		{Title: "Lịch trình", Body: itinerary}, {Title: "Bao gồm", Body: inclusions}, {Title: "Không bao gồm", Body: exclusions},
	})
	exportReply(c, file, fmt.Sprintf("bao-gia-%s-v%d.pdf", code, version), err)
}
