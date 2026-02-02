package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type HealthHandler struct {
	db           *pgxpool.Pool
	kafkaBrokers []string
}

func NewHealthHandler(db *pgxpool.Pool, kafkaBrokers []string) *HealthHandler {
	return &HealthHandler{
		db:           db,
		kafkaBrokers: kafkaBrokers,
	}
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

func (h *HealthHandler) Health(c *gin.Context) {
	resp := HealthResponse{
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
	conn, err := kafka.Dial("tcp", h.kafkaBrokers[0])
	if err != nil {
		resp.Status = "degraded"
		resp.Services["kafka"] = "error: " + err.Error()
	} else {
		conn.Close()
		resp.Services["kafka"] = "ok"
	}

	statusCode := http.StatusOK
	if resp.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, resp)
}
