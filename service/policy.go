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
	UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) error
	DeletePolicyWithSync(ctx context.Context, policyID string) error
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
	// Sync if only syncer exists.
	if s.syncer != nil {
		//syncReq := syncer.SyncRequest{
		//	Operation: "CREATE",
		//	Policy:    policy,
		//}
		err := s.syncer.CreatePolicySync(policy)
		if err != nil {
			return fmt.Errorf("failed to sync create policy with verifier: %w", err)
		}

		//if !resp.Success {
		//	return fmt.Errorf("verifier sync failed: %w", resp.Error)
		//}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PolicyService) UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) error {
	// start transaction
	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Update policy with tx
	if err := s.db.UpdatePluginPolicyTx(ctx, tx, policy); err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	if s.scheduler != nil {
		if err := s.db.UpdateTriggerExecutionTx(ctx, tx, policy.ID); err != nil {
			return fmt.Errorf("failed to update trigger execution tx: %w", err)
		}
	}

	if s.syncer != nil {
		if err := s.syncer.UpdatePolicySync(policy); err != nil {
			return fmt.Errorf("failed to sync update policy with verifier: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PolicyService) DeletePolicyWithSync(ctx context.Context, policyID string) error {

	tx, err := s.db.Pool().Begin(ctx)
	defer tx.Rollback(ctx)

	err = s.db.DeletePluginPolicy(ctx, tx, policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if s.syncer != nil {
		if err := s.syncer.DeletePolicySync(policyID); err != nil {
			return fmt.Errorf("failed to sync delete policy with verifier: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
