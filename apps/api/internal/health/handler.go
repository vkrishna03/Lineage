package health

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lineage/api/internal/app/config"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// Handler handles health check requests
type Handler struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

// NewHandler creates a new health handler
func NewHandler(db *pgxpool.Pool, cfg *config.Config) *Handler {
	return &Handler{
		db:  db,
		cfg: cfg,
	}
}

// Response represents the health check response
type Response struct {
	Status   string            `json:"status" example:"ok"`
	Services map[string]string `json:"services"`
}

// Check handles GET /health
// @Summary      Health check
// @Description  Check the health of the API and its dependencies (Postgres, Kafka)
// @Tags         health
// @Produce      json
// @Success      200 {object} Response
// @Failure      503 {object} Response
// @Router       /health [get]
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
	if len(h.cfg.KafkaBrokers) > 0 {
		if err := h.checkKafka(c.Request.Context()); err != nil {
			resp.Status = "degraded"
			resp.Services["kafka"] = "error: " + err.Error()
		} else {
			resp.Services["kafka"] = "ok"
		}
	}

	statusCode := http.StatusOK
	if resp.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, resp)
}

// checkKafka verifies Kafka connectivity with SASL if configured
func (h *Handler) checkKafka(ctx context.Context) error {
	dialer := &kafka.Dialer{
		Timeout: 5 * time.Second,
	}

	// Configure SASL if enabled
	if h.cfg.KafkaSASLEnabled {
		mechanism, err := scram.Mechanism(scram.SHA256, h.cfg.KafkaSASLUsername, h.cfg.KafkaSASLPassword)
		if err != nil {
			return err
		}
		dialer.SASLMechanism = mechanism

		tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

		// Load CA certificate if provided
		if h.cfg.KafkaCAPath != "" {
			caCert, err := os.ReadFile(h.cfg.KafkaCAPath)
			if err != nil {
				return err
			}
			caCertPool := x509.NewCertPool()
			if !caCertPool.AppendCertsFromPEM(caCert) {
				return err
			}
			tlsConfig.RootCAs = caCertPool
		}

		dialer.TLS = tlsConfig
	}

	conn, err := dialer.DialContext(ctx, "tcp", h.cfg.KafkaBrokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
