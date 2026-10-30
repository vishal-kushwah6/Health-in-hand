// Package uploads exposes POST /api/uploads/images. It reuses the existing
// Cloudinary storage (media domain) and the image compressor.
package uploads

import (
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	imageutil "artGallery/internal/infrastructure/image"
	media "artGallery/internal/media/domain"
	"artGallery/internal/shared/middleware"

	"github.com/gin-gonic/gin"
	_ "golang.org/x/image/webp" // registers the WebP decoder for image.Decode
)

const (
	maxUploadBytes = 10 << 20 // 10 MB before compression
	maxDimension   = 1600     // longest side after compression
	jpegQuality    = 80
)

type Handler struct {
	store media.Storage // nil when CLOUDINARY_URL is not configured
}

func NewHandler(store media.Storage) *Handler {
	return &Handler{store: store}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/uploads/images", h.UploadImage)
}

// POST /api/uploads/images   multipart/form-data, field name: "file"
// 201 -> {"url": "...", "key": "..."}
func (h *Handler) UploadImage(c *gin.Context) {
	uid, ok := middleware.ClerkUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Uploads are slower than normal requests, so lift the server-wide 5s
	// read/write limits for THIS request only (needs Go 1.20+; ignored otherwise).
	rc := http.NewResponseController(c.Writer)
	_ = rc.SetReadDeadline(time.Now().Add(60 * time.Second))
	_ = rc.SetWriteDeadline(time.Now().Add(90 * time.Second))

	// Refuse oversized bodies while reading, not after buffering everything.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes+(1<<20))

	if h.store == nil {
		// Read the upload BEFORE answering. If the server replies and closes while the
		// phone is still sending, the app only sees "network error" instead of this message.
		_, _ = io.Copy(io.Discard, c.Request.Body)
		log.Printf("upload rejected: CLOUDINARY_URL is not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image uploads are not configured on this server"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 10 MB)"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "send the image as multipart field \"file\""})
		return
	}
	defer file.Close()

	if header.Size > maxUploadBytes {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image is too large (max 10 MB)"})
		return
	}

	// Decoding doubles as validation: anything that is not a real image fails here.
	compressed, err := imageutil.CompressAvatar(file, imageutil.CompressOptions{
		MaxWidth: maxDimension, MaxHeight: maxDimension, Quality: jpegQuality,
	})
	if err != nil {
		log.Printf("upload rejected: cannot decode image %q (content-type %q): %v",
			header.Filename, header.Header.Get("Content-Type"), err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported or corrupt image (use JPG or PNG)"})
		return
	}

	res, err := h.store.Upload(c.Request.Context(), compressed, "posts/"+uid)
	if err != nil {
		log.Printf("cloudinary upload failed: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "could not store the image, try again"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"url": res.URL, "key": res.Key})
}
