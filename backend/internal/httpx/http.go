package httpx

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"strconv"
)

func OK(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{"success": true, "data": data, "error": nil, "message": nil})
}
func Error(c *gin.Context, status int, code string) {
	c.AbortWithStatusJSON(status, gin.H{"success": false, "data": nil, "error": gin.H{"code": code, "fields": gin.H{}}, "message": message(code)})
}

var Messages = map[string]string{"VALIDATION_ERROR": "Thông tin chưa hợp lệ. Vui lòng kiểm tra lại.", "NOT_FOUND": "Không tìm thấy nội dung.", "CONFLICT": "Dữ liệu đang được sử dụng hoặc bị trùng.", "VERSION_CONFLICT": "Dữ liệu đã thay đổi. Vui lòng tải lại trước khi lưu.", "FORBIDDEN": "Bạn không có quyền thực hiện thao tác này.", "INTERNAL_ERROR": "Có lỗi hệ thống. Vui lòng thử lại.", "SOLD_OUT": "Lịch khởi hành không còn đủ chỗ. Hãy chọn ngày khác.", "SALES_CLOSED": "Lịch khởi hành đã đóng bán.", "PRICE_CHANGED": "Giá đã thay đổi. Vui lòng xem và xác nhận báo giá mới.", "QUOTE_EXPIRED": "Báo giá đã hết hạn. Vui lòng tính lại giá.", "HOLD_EXPIRED": "Thời gian giữ chỗ đã hết. Vui lòng đặt lại.", "INVALID_STATE": "Trạng thái hiện tại không cho phép thao tác này.", "IDEMPOTENCY_CONFLICT": "Yêu cầu này đã được dùng cho dữ liệu khác.", "REQUEST_IN_PROGRESS": "Yêu cầu đang được xử lý. Vui lòng đợi.", "BOOKING_NOT_FOUND": "Không tìm thấy đơn với thông tin đã nhập.", "CAPTCHA_INVALID": "Vui lòng xác minh lại trước khi gửi.", "AUTH_REQUIRED": "Phiên truy cập đã hết hạn.", "PAYLOAD_TOO_LARGE": "Tệp hoặc dữ liệu vượt giới hạn.", "UNSUPPORTED_MEDIA_TYPE": "Định dạng tệp không được hỗ trợ."}

func Body(c *gin.Context, v any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 256*1024)
	d := json.NewDecoder(c.Request.Body)
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(&struct{}{}) != io.EOF {
		Error(c, 422, "VALIDATION_ERROR")
		return false
	}
	return true
}
func Page(c *gin.Context) (int, int, bool) {
	p, n := 1, 20
	var e error
	if v := c.Query("page"); v != "" {
		p, e = strconv.Atoi(v)
		if e != nil || p < 1 || p > 100000 {
			return 0, 0, false
		}
	}
	if v := c.Query("page_size"); v != "" {
		n, e = strconv.Atoi(v)
		if e != nil || n < 1 || n > 100 {
			return 0, 0, false
		}
	}
	return p, n, true
}
func List(c *gin.Context, data any, p, n, total int) {
	c.JSON(200, gin.H{"success": true, "data": data, "error": nil, "message": nil, "meta": gin.H{"page": p, "page_size": n, "total": total, "total_pages": (total + n - 1) / n}})
}
func JSONRows(ctx context.Context, db *pgxpool.Pool, query string, args ...any) ([]json.RawMessage, error) {
	rows, e := db.Query(ctx, query, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []json.RawMessage{}
	for rows.Next() {
		var raw json.RawMessage
		if e = rows.Scan(&raw); e != nil {
			return nil, e
		}
		out = append(out, raw)
	}
	return out, rows.Err()
}

func message(code string) string {
	if v := Messages[code]; v != "" {
		return v
	}
	switch code {
	case "CAPTCHA_FAILED":
		return "Xác minh bảo mật chưa thành công. Vui lòng thử lại."
	case "PASSENGER_MISMATCH":
		return "Thông tin hoặc tuổi hành khách chưa khớp báo giá."
	case "UNAUTHORIZED":
		return "Phiên truy cập không hợp lệ hoặc đã hết hạn."
	case "QUOTE_ALREADY_USED":
		return "Báo giá đã được sử dụng để tạo đơn."
	case "PAYMENT_OPTION_UNAVAILABLE":
		return "Phương án thanh toán không còn khả dụng."
	case "IDEMPOTENCY_KEY_REQUIRED":
		return "Yêu cầu cần mã chống gửi trùng hợp lệ."
	case "INVALID_TRANSITION":
		return "Trạng thái đơn không cho phép thao tác này."
	}
	return "Không thực hiện được yêu cầu. Vui lòng kiểm tra rồi thử lại."
}
