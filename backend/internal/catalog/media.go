package catalog

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/EthanLe-dev/blue-sky-group-website/backend/internal/httpx"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type MediaStore struct {
	Root, Key, PublicBucket, PrivateBucket string
	S3                                     *s3.Client
}

func NewMediaStore(key string) *MediaStore {
	root := os.Getenv("LOCAL_STORAGE_PATH")
	if root == "" {
		root = "./storage"
	}
	m := &MediaStore{Root: root, Key: key, PublicBucket: os.Getenv("R2_PUBLIC_BUCKET"), PrivateBucket: os.Getenv("R2_PRIVATE_BUCKET")}
	if os.Getenv("STORAGE_PROVIDER") == "r2" {
		endpoint := os.Getenv("R2_ENDPOINT")
		m.S3 = s3.NewFromConfig(aws.Config{Region: "auto", Credentials: credentials.NewStaticCredentialsProvider(os.Getenv("R2_ACCESS_KEY_ID"), os.Getenv("R2_SECRET_ACCESS_KEY"), ""), RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired}, func(o *s3.Options) { o.BaseEndpoint = &endpoint; o.UsePathStyle = true })
	}
	return m
}
func (m *MediaStore) bucket(visibility string) string {
	if visibility == "public" {
		return m.PublicBucket
	}
	return m.PrivateBucket
}
func (m *MediaStore) put(ctx context.Context, key, visibility, mime string, f *os.File, size int64) error {
	if m.S3 != nil {
		_, e := m.S3.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(m.bucket(visibility)), Key: aws.String(key), Body: f, ContentType: aws.String(mime), ContentLength: aws.Int64(size)})
		return e
	}
	if e := os.MkdirAll(m.Root, 0700); e != nil {
		return e
	}
	out, e := os.OpenFile(filepath.Join(m.Root, key), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if e != nil {
		return e
	}
	_, e = io.Copy(out, f)
	closeErr := out.Close()
	if e != nil {
		os.Remove(filepath.Join(m.Root, key))
		return e
	}
	return closeErr
}
func (m *MediaStore) remove(ctx context.Context, key, visibility string) {
	if m.S3 != nil {
		m.S3.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(m.bucket(visibility)), Key: aws.String(key)})
		return
	}
	os.Remove(filepath.Join(m.Root, key))
}
func (m *MediaStore) signature(id, expires string) string {
	h := hmac.New(sha256.New, []byte(m.Key))
	h.Write([]byte(id + ":" + expires))
	return hex.EncodeToString(h.Sum(nil))
}
func (s *Server) registerMedia(v, admin *gin.RouterGroup, media, private gin.HandlerFunc) {
	v.GET("/media/:id", s.serveMedia)
	admin.POST("/media", func(c *gin.Context) {
		visibility := c.DefaultQuery("visibility", "private")
		if visibility == "public" {
			media(c)
		} else {
			private(c)
		}
	}, s.uploadMedia)
	admin.GET("/media", media, s.listMedia)
	admin.PATCH("/media/:id", media, s.patchMedia)
	admin.DELETE("/media/:id", media, s.archiveMedia)
	admin.POST("/media/:id/download-url", private, s.downloadURL)
}
func detectMedia(head []byte) (mime, kind, ext string, max int64) {
	mime = http.DetectContentType(head)
	switch mime {
	case "image/jpeg":
		return mime, "image", ".jpg", 10 << 20
	case "image/png":
		return mime, "image", ".png", 10 << 20
	case "image/webp":
		return mime, "image", ".webp", 10 << 20
	case "application/pdf":
		return mime, "document", ".pdf", 20 << 20
	}
	if len(head) >= 12 && string(head[4:8]) == "ftyp" {
		return "video/mp4", "video", ".mp4", 50 << 20
	}
	if len(head) >= 4 && bytes.Equal(head[:4], []byte{0x1a, 0x45, 0xdf, 0xa3}) && bytes.Contains(head, []byte("webm")) {
		return "video/webm", "video", ".webm", 50 << 20
	}
	return "", "", "", 0
}
func (s *Server) uploadMedia(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 51<<20)
	if e := c.Request.ParseMultipartForm(1 << 20); e != nil {
		httpx.Error(c, 413, "PAYLOAD_TOO_LARGE")
		return
	}
	if c.Request.MultipartForm != nil {
		defer c.Request.MultipartForm.RemoveAll()
	}
	source, header, e := c.Request.FormFile("file")
	if e != nil {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	defer source.Close()
	head := make([]byte, 512)
	n, _ := io.ReadFull(source, head)
	head = head[:n]
	mime, kind, ext, max := detectMedia(head)
	if max == 0 {
		httpx.Error(c, 415, "UNSUPPORTED_MEDIA_TYPE")
		return
	}
	visibility := c.DefaultQuery("visibility", "private")
	if form := c.PostForm("visibility"); form != "" && form != visibility {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	if visibility != "public" && visibility != "private" {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	f, e := os.CreateTemp("", "blue-sky-upload-*")
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	defer os.Remove(f.Name())
	defer f.Close()
	reader := io.MultiReader(bytes.NewReader(head), source)
	size, e := io.Copy(f, io.LimitReader(reader, max+1))
	if e != nil || size > max || size == 0 {
		httpx.Error(c, 413, "PAYLOAD_TOO_LARGE")
		return
	}
	var width, height *int
	if kind == "image" {
		f.Seek(0, 0)
		conf, _, e := image.DecodeConfig(f)
		if e != nil || conf.Width > 20000 || conf.Height > 20000 || int64(conf.Width)*int64(conf.Height) > 100000000 {
			httpx.Error(c, 415, "UNSUPPORTED_MEDIA_TYPE")
			return
		}
		width = &conf.Width
		height = &conf.Height
	}
	f.Seek(0, 0)
	id := uuid.NewString()
	key := id + ext
	if s.Media.put(c.Request.Context(), key, visibility, mime, f, size) != nil {
		httpx.Error(c, 503, "INTERNAL_ERROR")
		return
	}
	_, e = s.DB.Exec(c.Request.Context(), "INSERT INTO media_assets(id,object_key,original_name,mime_type,size_bytes,width,height,kind,visibility,status,alt_text,collection,source_credit,uploaded_by_staff_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'ready',$10,$11,$12,$13)", id, key, filepath.Base(header.Filename), mime, size, width, height, kind, visibility, c.PostForm("alt_text"), c.PostForm("collection"), c.PostForm("source_credit"), s.UserID(c))
	if e != nil {
		s.Media.remove(c.Request.Context(), key, visibility)
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	httpx.OK(c, 201, gin.H{"id": id, "object_key": key, "kind": kind, "mime_type": mime, "size_bytes": size, "width": width, "height": height, "visibility": visibility, "status": "ready", "url": "/api/v1/media/" + id})
}
func (s *Server) serveMedia(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var key, visibility, mime, status string
	e := s.DB.QueryRow(c.Request.Context(), "SELECT object_key,visibility,mime_type,status FROM media_assets WHERE id=$1", c.Param("id")).Scan(&key, &visibility, &mime, &status)
	if e != nil || status != "ready" || filepath.Base(key) != key {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	if visibility != "public" {
		expires := c.Query("expires")
		seconds, e := strconv.ParseInt(expires, 10, 64)
		if e != nil || seconds <= time.Now().Unix() || seconds > time.Now().Add(5*time.Minute).Unix() || !hmac.Equal([]byte(s.Media.signature(c.Param("id"), expires)), []byte(c.Query("signature"))) {
			httpx.Error(c, 404, "NOT_FOUND")
			return
		}
		c.Header("Cache-Control", "no-store")
	} else {
		c.Header("Cache-Control", "public,max-age=300")
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Security-Policy", "default-src 'none'; sandbox")
	if mime == "application/pdf" {
		c.Header("Content-Disposition", `attachment; filename="document.pdf"`)
	}
	if s.Media.S3 != nil {
		out, e := s.Media.S3.GetObject(c.Request.Context(), &s3.GetObjectInput{Bucket: aws.String(s.Media.bucket(visibility)), Key: aws.String(key)})
		if e != nil {
			httpx.Error(c, 404, "NOT_FOUND")
			return
		}
		defer out.Body.Close()
		size := int64(-1)
		if out.ContentLength != nil {
			size = *out.ContentLength
		}
		c.DataFromReader(200, size, mime, out.Body, nil)
		return
	}
	f, e := os.Open(filepath.Join(s.Media.Root, key))
	if e != nil {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	defer f.Close()
	stat, _ := f.Stat()
	http.ServeContent(c.Writer, c.Request, key, stat.ModTime(), f)
}
func (s *Server) listMedia(c *gin.Context) {
	p, n, valid := httpx.Page(c)
	if !valid {
		httpx.Error(c, 422, "VALIDATION_ERROR")
		return
	}
	out, e := httpx.JSONRows(c.Request.Context(), s.DB, "SELECT to_jsonb(m)||jsonb_build_object('url',CASE WHEN visibility='public' THEN '/api/v1/media/'||id::text ELSE NULL END) FROM media_assets m WHERE status!='archived' AND (visibility='public' OR uploaded_by_staff_id=$1) ORDER BY created_at DESC LIMIT $2 OFFSET $3", s.UserID(c), n, (p-1)*n)
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	var total int
	s.DB.QueryRow(c.Request.Context(), "SELECT count(*) FROM media_assets WHERE status!='archived' AND (visibility='public' OR uploaded_by_staff_id=$1)", s.UserID(c)).Scan(&total)
	httpx.List(c, out, p, n, total)
}
func (s *Server) downloadURL(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var allowed bool
	e := s.DB.QueryRow(c.Request.Context(), "SELECT EXISTS(SELECT 1 FROM media_assets WHERE id=$1 AND status='ready' AND (visibility='public' OR uploaded_by_staff_id=$2))", c.Param("id"), s.UserID(c)).Scan(&allowed)
	if e != nil || !allowed {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	s.signedDownload(c, c.Param("id"))
}

func (s *Server) LeadAttachmentURL(c *gin.Context) {
	if !validID(c, "id") || !validID(c, "media_id") {
		return
	}
	var allowed bool
	id := c.Param("media_id")
	err := s.DB.QueryRow(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM group_request_attachments a
		JOIN media_assets m ON m.id=a.media_id WHERE a.group_request_id=$1 AND m.id=$2
		AND m.visibility='private' AND m.status='ready')`, c.Param("id"), id).Scan(&allowed)
	if err != nil {
		httpx.Error(c, 503, "SERVICE_UNAVAILABLE")
		return
	}
	if !allowed {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	s.signedDownload(c, id)
}

func (s *Server) signedDownload(c *gin.Context, id string) {
	expires := time.Now().Add(5 * time.Minute).Unix()
	url := fmt.Sprintf("/api/v1/media/%s?expires=%d&signature=%s", id, expires, s.Media.signature(id, strconv.FormatInt(expires, 10)))
	httpx.OK(c, 200, gin.H{"url": url, "expires_at": time.Unix(expires, 0).UTC()})
}
func (s *Server) patchMedia(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	var b struct {
		Alt        *string `json:"alt_text"`
		Caption    *string `json:"caption"`
		Collection *string `json:"collection"`
	}
	if !httpx.Body(c, &b) {
		return
	}
	tag, e := s.DB.Exec(c.Request.Context(), "UPDATE media_assets SET alt_text=COALESCE($3,alt_text),caption=COALESCE($4,caption),collection=COALESCE($5,collection),updated_at=now() WHERE id=$1 AND (visibility='public' OR uploaded_by_staff_id=$2)", c.Param("id"), s.UserID(c), b.Alt, b.Caption, b.Collection)
	if e != nil || tag.RowsAffected() != 1 {
		httpx.Error(c, 404, "NOT_FOUND")
		return
	}
	httpx.OK(c, 200, gin.H{"id": c.Param("id")})
}
func (s *Server) archiveMedia(c *gin.Context) {
	if !validID(c, "id") {
		return
	}
	tag, e := s.DB.Exec(c.Request.Context(), `UPDATE media_assets m SET status='archived' WHERE id=$1 AND (visibility='public' OR uploaded_by_staff_id=$2) AND NOT EXISTS(SELECT 1 FROM tours WHERE cover_media_id=m.id OR poster_media_id=m.id) AND NOT EXISTS(SELECT 1 FROM tour_media WHERE media_id=m.id) AND NOT EXISTS(SELECT 1 FROM posts WHERE cover_media_id=m.id) AND NOT EXISTS(SELECT 1 FROM destinations WHERE cover_media_id=m.id) AND NOT EXISTS(SELECT 1 FROM banners WHERE media_id=m.id) AND NOT EXISTS(SELECT 1 FROM reviews WHERE media_id=m.id) AND NOT EXISTS(SELECT 1 FROM group_request_attachments WHERE media_id=m.id) AND NOT EXISTS(SELECT 1 FROM site_settings WHERE key='company_profile' AND value->'document_media_ids' @> jsonb_build_array(m.id::text))`, c.Param("id"), s.UserID(c))
	if e != nil {
		httpx.Error(c, 500, "INTERNAL_ERROR")
		return
	}
	if tag.RowsAffected() != 1 {
		httpx.Error(c, 409, "CONFLICT")
		return
	}
	httpx.OK(c, 200, gin.H{"id": c.Param("id"), "archived": true})
}
