package domain

import (
	"context"
	"time"
)

// DeliveryRequest identifies an immutable controller operation. Existing session
// and conversation facts still own execution, cancellation and recovery.
type DeliveryRequest struct {
	ID       string       `json:"id"`
	Control  ControlClaim `json:"control"`
	Kind     string       `json:"kind" enum:"test,review,verdict,rework,accept"`
	BaseSHA  string       `json:"baseSha"`
	HeadSHA  string       `json:"headSha"`
	TestID   string       `json:"testId,omitempty"`
	ReviewID string       `json:"reviewId,omitempty"`
	Command  []string     `json:"command,omitempty"`
	Message  string       `json:"message,omitempty"`
}

// DeliveryResult retains observed execution output or an explicit review conclusion.
type DeliveryResult struct {
	ReviewerID   SessionID `json:"reviewerId,omitempty"`
	ReviewerPath string    `json:"reviewerPath,omitempty"`
	ExitCode     *int      `json:"exitCode,omitempty"`
	Output       string    `json:"output,omitempty"`
	Truncated    bool      `json:"truncated,omitempty"`
	Error        string    `json:"error,omitempty"`
	MessageID    string    `json:"messageId,omitempty"`
	Verdict      string    `json:"verdict,omitempty"`
	Delivered    bool      `json:"delivered,omitempty"`
}

// DeliveryEvidence binds an immutable request and its result to a worker commit.
type DeliveryEvidence struct {
	// Current is derived from live workspace and session facts, never stored.
	Current     bool            `json:"current"`
	ID          string          `json:"id"`
	ProjectID   ProjectID       `json:"projectId"`
	WorkerID    SessionID       `json:"workerId"`
	WorkerPath  string          `json:"workerPath"`
	Request     DeliveryRequest `json:"request"`
	Result      *DeliveryResult `json:"result,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
	CompletedAt *time.Time      `json:"completedAt,omitempty"`
}

type deliveryContextKey struct{}

// WithDeliveryReceipt carries a daemon-registered operation into the existing
// Chat delivery path. Public request bodies cannot set this context value.
func WithDeliveryReceipt(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, deliveryContextKey{}, id)
}

// DeliveryReceipt returns the registered operation carried by the daemon context.
func DeliveryReceipt(ctx context.Context) string {
	id, _ := ctx.Value(deliveryContextKey{}).(string)
	return id
}
