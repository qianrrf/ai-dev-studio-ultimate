package controllers

import (
	"context"
	"net/http"
	"strings"
	"unicode"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/apierr"
	"github.com/aoagents/agent-orchestrator/backend/internal/httpd/envelope"
)

type projectControlService interface {
	ProjectControl(context.Context, domain.ProjectID) (domain.ProjectControl, error)
	SetProjectControl(context.Context, domain.ProjectID, string, string, int64) (domain.ProjectControl, error)
	SubmitProjectRequest(context.Context, domain.ProjectRequest) (domain.ProjectRequest, error)
	ProjectRequests(context.Context, domain.ProjectID) ([]domain.ProjectRequest, error)
}

// ProjectControlRequest changes ownership only at the expected revision.
type ProjectControlRequest struct {
	OwnerKind        string `json:"ownerKind" enum:"internal,external,unowned"`
	OwnerID          string `json:"ownerId" maxLength:"128"`
	ExpectedRevision int64  `json:"expectedRevision" minimum:"0"`
}

// ProjectIntakeRequest submits a requirement independently of worker dispatch.
type ProjectIntakeRequest struct {
	ID          string            `json:"id" maxLength:"128"`
	Brief       string            `json:"brief" maxLength:"16384"`
	Attachments []AttachmentInput `json:"attachments,omitempty"`
}

// ProjectRequestsResponse contains the durable project inbox.
type ProjectRequestsResponse struct {
	Requests []domain.ProjectRequest `json:"requests"`
}

func (c *SessionsController) projectControlService(w http.ResponseWriter, r *http.Request) projectControlService {
	svc, ok := c.Svc.(projectControlService)
	if !ok {
		envelope.WriteError(w, r, apierr.Internal("PROJECT_CONTROL_UNAVAILABLE", "Project control unavailable"))
		return nil
	}
	return svc
}

func (c *SessionsController) getProjectControl(w http.ResponseWriter, r *http.Request) {
	svc := c.projectControlService(w, r)
	if svc == nil {
		return
	}
	control, err := svc.ProjectControl(r.Context(), projectID(r))
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, control)
}

func (c *SessionsController) setProjectControl(w http.ResponseWriter, r *http.Request) {
	svc := c.projectControlService(w, r)
	if svc == nil {
		return
	}
	var in ProjectControlRequest
	if err := decodeJSONStrict(r, &in); err != nil {
		envelope.WriteError(w, r, apierr.Invalid("INVALID_JSON", "Invalid control request", nil))
		return
	}
	if in.ExpectedRevision < 0 || len(in.OwnerID) > 128 || strings.TrimSpace(in.OwnerID) != in.OwnerID || strings.IndexFunc(in.OwnerID, unicode.IsControl) >= 0 || (in.OwnerKind != "internal" && in.OwnerKind != "external" && in.OwnerKind != "unowned") || (in.OwnerKind == "unowned") != (in.OwnerID == "") {
		envelope.WriteError(w, r, apierr.Invalid("INVALID_CONTROLLER", "Specify a valid owner and expected revision; use unowned with an empty owner to release", nil))
		return
	}
	control, err := svc.SetProjectControl(r.Context(), projectID(r), in.OwnerKind, in.OwnerID, in.ExpectedRevision)
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, control)
}

func (c *SessionsController) submitProjectRequest(w http.ResponseWriter, r *http.Request) {
	svc := c.projectControlService(w, r)
	if svc == nil {
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxSpawnBodyBytes)
	var in ProjectIntakeRequest
	if err := decodeJSONStrict(r, &in); err != nil {
		envelope.WriteError(w, r, apierr.Invalid("INVALID_JSON", "Invalid requirement", nil))
		return
	}
	if in.ID == "" || !validClientRequestID(in.ID) || strings.TrimSpace(in.Brief) == "" || len(in.Brief) > maxPromptLen {
		envelope.WriteError(w, r, apierr.Invalid("INVALID_REQUIREMENT", "A request ID and a brief of at most 16 KiB are required", nil))
		return
	}
	attachments, err := decodeSpawnAttachments(in.Attachments)
	if err != nil {
		envelope.WriteError(w, r, apierr.Invalid(err.code, err.message, nil))
		return
	}
	request := domain.ProjectRequest{ID: in.ID, ProjectID: projectID(r), Brief: in.Brief}
	for _, a := range attachments {
		request.Attachments = append(request.Attachments, domain.RequestAttachment{Ext: a.Ext, Data: a.Data})
	}
	result, submitErr := svc.SubmitProjectRequest(r.Context(), request)
	if submitErr != nil {
		envelope.WriteError(w, r, submitErr)
		return
	}
	envelope.WriteJSON(w, http.StatusAccepted, result)
}

func (c *SessionsController) listProjectRequests(w http.ResponseWriter, r *http.Request) {
	svc := c.projectControlService(w, r)
	if svc == nil {
		return
	}
	requests, err := svc.ProjectRequests(r.Context(), projectID(r))
	if err != nil {
		envelope.WriteError(w, r, err)
		return
	}
	envelope.WriteJSON(w, http.StatusOK, ProjectRequestsResponse{Requests: requests})
}
