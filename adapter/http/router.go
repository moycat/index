package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/moycat/index/data"
	"github.com/moycat/index/service"
	log "github.com/sirupsen/logrus"
)

type Dependencies struct {
	IndexService  *service.IndexService
	SearchService *service.SearchService
	AuthToken     string
	Debug         bool
	Logger        *log.Logger
}

type indexRequest struct {
	Posts []indexPostRecord `json:"posts"`
}

type indexPostRecord struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Content     string `json:"content"`
	PublishedAt int64  `json:"published_at"`
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewRouter(deps Dependencies) *gin.Engine {
	if deps.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery(), cors.Default())
	r.Use(requestLogger(deps.Logger))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	v1 := r.Group("/v1")
	v1.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		page := parseInt(c.DefaultQuery("page", "1"), 1)
		pageSize := parseInt(c.DefaultQuery("page_size", "10"), 10)

		hits, err := deps.SearchService.Search(c.Request.Context(), query, page, pageSize)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"query": query,
			"hits":  hits,
		})
	})

	v1.POST("/index", authMiddleware(deps.AuthToken), func(c *gin.Context) {
		var req indexRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			writeError(c, fmt.Errorf("%w: invalid json payload", service.ErrInvalidArgument))
			return
		}

		posts := make([]data.Post, 0, len(req.Posts))
		for _, item := range req.Posts {
			if item.PublishedAt <= 0 {
				writeError(c, fmt.Errorf("%w: published_at must be unix timestamp in seconds", service.ErrInvalidArgument))
				return
			}
			publishedAt := time.Unix(item.PublishedAt, 0).UTC()
			posts = append(posts, data.Post{
				Title:       item.Title,
				URL:         item.URL,
				Content:     item.Content,
				PublishedAt: publishedAt,
			})
		}

		indexReq := data.IndexRequest{
			Posts: posts,
		}
		if err := deps.IndexService.ReindexAllPosts(c.Request.Context(), indexReq); err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":     "indexed",
			"post_count": len(posts),
		})
	})

	return r
}

func requestLogger(logger *log.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = log.New()
	}
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		entry := logger.WithFields(log.Fields{
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status":       c.Writer.Status(),
			"latency_ms":   time.Since(start).Milliseconds(),
			"client_ip":    c.ClientIP(),
			"request_id":   c.GetHeader("X-Request-Id"),
			"user_agent":   c.Request.UserAgent(),
			"content_type": c.ContentType(),
		})
		entry.Info("http_request")
	}
}

func parseInt(raw string, fallback int) int {
	parsed, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return parsed
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrUnauthorized):
		c.JSON(http.StatusUnauthorized, errorBody{Error: errorDetail{Code: "unauthorized", Message: "unauthorized"}})
	case errors.Is(err, service.ErrInvalidArgument):
		c.JSON(http.StatusBadRequest, errorBody{Error: errorDetail{Code: "invalid_argument", Message: err.Error()}})
	default:
		c.JSON(http.StatusInternalServerError, errorBody{Error: errorDetail{Code: "internal", Message: "internal server error"}})
	}
}
