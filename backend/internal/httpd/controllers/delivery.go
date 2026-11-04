package controllers

import (
	"context"
	"net/http"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apierr"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/envelope"
)

type deliveryService interface {
	RunDelivery(context.Context, domain.SessionID, domain.DeliveryRequest) (domain.DeliveryEvidence, error)
	DeliveryEvidence(context.Context, domain.SessionID) ([]domain.DeliveryEvidence, error)
}

// DeliveryEvidenceResponse lists persisted operations and their current validity.
type DeliveryEvidenceResponse struct {
	Evidence []domain.DeliveryEvidence `json:"evidence"`
}

func (c *SessionsController) runDelivery(w http.ResponseWriter, r *http.Request) {
	svc, ok := c.Svc.(deliveryService)
	if !ok {
		envelope.WriteError(w, r, apierr.Internal("DELIVERY_UNAVAILABLE", "Delivery evidence unavailable"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	var in domain.DeliveryRequest
	if err := decodeJSONStrict(r, &in); err != nil {
		envelope.WriteError(w, r, apierr.Invalid("INVALID_JSON", "Invalid delivery operation", nil))
		return
	}
	evidence, err := svc.RunDelivery(r.Context(), sessionID(r), in)
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, evidence)
}

func (c *SessionsController) deliveryEvidence(w http.ResponseWriter, r *http.Request) {
	svc, ok := c.Svc.(deliveryService)
	if !ok {
		envelope.WriteError(w, r, apierr.Internal("DELIVERY_UNAVAILABLE", "Delivery evidence unavailable"))
		return
	}
	rows, err := svc.DeliveryEvidence(r.Context(), sessionID(r))
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, DeliveryEvidenceResponse{Evidence: rows})
}
