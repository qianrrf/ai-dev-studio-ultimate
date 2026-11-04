package domain

import (
	"errors"
	"time"
)

// ErrProjectControlConflict also fences retries from a retired controller.
var ErrProjectControlConflict = errors.New("project controller revision conflict")

// ErrProjectWorkerCapacity preserves the product's two-worker execution limit.
var ErrProjectWorkerCapacity = errors.New("project worker capacity reached")

// ProjectControl is cooperative ownership between trusted local clients.
// Releasing ownership retains the row and advances Revision.
type ProjectControl struct {
	ProjectID ProjectID `json:"projectId"`
	OwnerKind string    `json:"ownerKind" enum:"internal,external,unowned"`
	OwnerID   string    `json:"ownerId"`
	Revision  int64     `json:"revision"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ControlClaim accompanies a controller-authored mutation, never a user intake.
type ControlClaim struct {
	OwnerID  string `json:"ownerId"`
	Revision int64  `json:"revision"`
}

// ProjectRequest is a durable brief, not a second task lifecycle.
type ProjectRequest struct {
	ID          string              `json:"id"`
	ProjectID   ProjectID           `json:"projectId"`
	Brief       string              `json:"brief"`
	Attachments []RequestAttachment `json:"attachments,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
}

// RequestAttachment retains a requirement attachment in the project inbox.
type RequestAttachment struct {
	Ext  string `json:"ext"`
	Data []byte `json:"data"`
}
