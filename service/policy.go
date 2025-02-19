package service

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/scheduler"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type Policy interface {
	CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) error
}

type PolicyService struct {
	db        storage.DatabaseStorage
	syncer    *syncer.Syncer
	scheduler *scheduler.SchedulerService
	logger    *logrus.Logger
}

func NewPolicyService(db storage.DatabaseStorage, syncer *syncer.Syncer, scheduler *scheduler.SchedulerService, logger *logrus.Logger) *PolicyService {
	return &PolicyService{
		db:        db,
		syncer:    syncer,
		scheduler: scheduler,
		logger:    logger,
	}
}

func (s *PolicyService) CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) error {
	// Start transaction

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Insert policy
	if err := s.db.InsertPluginPolicyTx(ctx, tx, policy); err != nil {
		return fmt.Errorf("failed to insert policy: %w", err)
	}

	// Handle trigger if scheduler exists
	if s.scheduler != nil {
		if err := s.scheduler.CreateTimeTrigger(ctx, policy, tx); err != nil {
			return fmt.Errorf("failed to create time trigger: %w", err)
		}
	}

	// Sync with verifier
	syncReq := syncer.SyncRequest{
		Operation: "CREATE",
		Policy:    policy,
	}

	resp, err := s.syncer.SyncWithVerifier(ctx, &syncReq)
	if err != nil {
		return fmt.Errorf("failed to sync with verifier: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("verifier sync failed: %w", resp.Error)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
