package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/app/config"
	"github.com/segmentio/kafka-go"
)

// Handler handles health check requests
type Handler struct {
	db           *pgxpool.Pool
	kafkaBrokers []string
}

// NewHandler creates a new health handler
func NewHandler(db *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{
		db:           db,
		kafkaBrokers: cfg.KafkaBrokers,
	}
}

// Response represents the health check response
type Response struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// Check handles GET /health
func (h *Handler) Check(c *gin.Context) {
	resp := Response{
		Status:   "ok",
		Services: make(map[string]string),
	}

	// Check database
	if err := h.db.Ping(c.Request.Context()); err != nil {
		resp.Status = "degraded"
		resp.Services["postgres"] = "error: " + err.Error()
	} else {
		resp.Services["postgres"] = "ok"
	}

	// Check Kafka connectivity
	if len(h.kafkaBrokers) > 0 {
		conn, err := kafka.Dial("tcp", h.kafkaBrokers[0])
		if err != nil {
			resp.Status = "degraded"
			resp.Services["kafka"] = "error: " + err.Error()
		} else {
			conn.Close()
			resp.Services["kafka"] = "ok"
		}
	}

	statusCode := http.StatusOK
	if resp.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, resp)
}
