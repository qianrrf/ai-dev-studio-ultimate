package chat

import (
	"context"

	"github.com/aoagents/agent-orchestrator/backend/internal/domain"
)

func (s *Service) checkProjectDelivery(ctx context.Context, id domain.SessionID) error {
	if gate, ok := s.sessions.(interface {
		CheckProjectChatDelivery(context.Context, domain.SessionID) error
	}); ok {
		return gate.CheckProjectChatDelivery(ctx, id)
	}
	return nil
}

func (s *Service) checkIndependentSide(ctx context.Context, id domain.SessionID) error {
	if err := s.checkProjectDelivery(ctx, id); err != nil {
		return err
	}
	record, found, err := s.sessions.GetSession(ctx, id)
	if err != nil {
		return err
	}
	if found && record.ReadOnly {
		return domain.ErrProjectControlConflict
	}
	return nil
}
