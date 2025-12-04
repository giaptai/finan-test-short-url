package handler

import (
	"github.com/giaptai/finan-test-short-url/bll"
	"github.com/gin-gonic/gin"
	"strconv"
)

type URLHandler struct {
	service *bll.URLService
	baseURL string
}

type CreateURLRequest struct {
	URL string `json:"url" binding:"required"`
}

// constructor
func NewURLHandler(service *bll.URLService, baseURL string) *URLHandler {
	return &URLHandler{
		service: service,
		baseURL: baseURL,
	}
}

func (h *URLHandler) CreateShortURL(c *gin.Context) {
	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	url, err := h.service.CreateShortURL(req.URL)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	res := gin.H{
		"short_code":   url.ShortCode,
		"short_url":    h.baseURL + "/" + url.ShortCode,
		"original_url": url.OriginalURL,
		"created_at":   url.CreatedAt,
	}
	c.JSON(201, res)
}

func (h *URLHandler) RedirectToOriginal(c *gin.Context) {
	shortCode := c.Param("shortCode")
	url, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	// asynchronously
	go h.service.IncreaseClick(shortCode)

	// return original url - Permanent Redirect
	c.Redirect(301, url.OriginalURL)
}

func (h *URLHandler) GetURLInfo(c *gin.Context) {
	shortCode := c.Param("shortCode")
	url, err := h.service.GetOriginalURL(shortCode)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	res := gin.H{
		"id":           url.ID,
		"short_code":   url.ShortCode,
		"short_url":    h.baseURL + "/" + url.ShortCode,
		"original_url": url.OriginalURL,
		"clicks":       url.Clicks,
		"created_at":   url.CreatedAt,
	}
	c.JSON(200, res)
}

func (h *URLHandler) ListURLs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	urls, err := h.service.ListURLs(limit, offset)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	res := gin.H{
		"urls":   urls,
		"limit":  limit,
		"offset": offset,
		"count":  len(urls),
	}
	c.JSON(200, res)
}
