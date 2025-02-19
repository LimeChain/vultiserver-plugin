package service

import (
	"context"
	"fmt"
<<<<<<< HEAD

=======
>>>>>>> 24b0d62 (added policyService)
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/scheduler"
	"github.com/vultisig/vultisigner/internal/syncer"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type Policy interface {
<<<<<<< HEAD
	CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error)
	UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error)
	DeletePolicyWithSync(ctx context.Context, policyID string) error
	GetPluginPolicies(ctx context.Context, pluginType, publicKey string) ([]types.PluginPolicy, error)
	GetPluginPolicy(ctx context.Context, policyID string) (types.PluginPolicy, error)
=======
	CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) error
>>>>>>> 24b0d62 (added policyService)
}

type PolicyService struct {
	db        storage.DatabaseStorage
<<<<<<< HEAD
	syncer    syncer.PolicySyncer
=======
	syncer    *syncer.Syncer
>>>>>>> 24b0d62 (added policyService)
	scheduler *scheduler.SchedulerService
	logger    *logrus.Logger
}

<<<<<<< HEAD
func NewPolicyService(db storage.DatabaseStorage, syncer syncer.PolicySyncer, scheduler *scheduler.SchedulerService, logger *logrus.Logger) (*PolicyService, error) {
	if db == nil {
		return nil, fmt.Errorf("database storage cannot be nil")
	}
=======
func NewPolicyService(db storage.DatabaseStorage, syncer *syncer.Syncer, scheduler *scheduler.SchedulerService, logger *logrus.Logger) *PolicyService {
>>>>>>> 24b0d62 (added policyService)
	return &PolicyService{
		db:        db,
		syncer:    syncer,
		scheduler: scheduler,
		logger:    logger,
<<<<<<< HEAD
	}, nil
}

func (s *PolicyService) CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	// Start transaction
	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
=======
	}
}

func (s *PolicyService) CreatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) error {
	// Start transaction

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return err
>>>>>>> 24b0d62 (added policyService)
	}
	defer tx.Rollback(ctx)

	// Insert policy
<<<<<<< HEAD
	newPolicy, err := s.db.InsertPluginPolicyTx(ctx, tx, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to insert policy: %w", err)
=======
	if err := s.db.InsertPluginPolicyTx(ctx, tx, policy); err != nil {
		return fmt.Errorf("failed to insert policy: %w", err)
>>>>>>> 24b0d62 (added policyService)
	}

	// Handle trigger if scheduler exists
	if s.scheduler != nil {
		if err := s.scheduler.CreateTimeTrigger(ctx, policy, tx); err != nil {
<<<<<<< HEAD
			return nil, fmt.Errorf("failed to create time trigger: %w", err)
		}
	}
	// Sync if only syncer exists.
	if s.syncer != nil {
		err := s.syncer.CreatePolicySync(policy)
		if err != nil {
			return nil, fmt.Errorf("failed to sync create policy with verifier: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return newPolicy, nil
}

func (s *PolicyService) UpdatePolicyWithSync(ctx context.Context, policy types.PluginPolicy) (*types.PluginPolicy, error) {
	// start transaction
	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update policy with tx
	updatedPolicy, err := s.db.UpdatePluginPolicyTx(ctx, tx, policy)
	if err != nil {
		return nil, fmt.Errorf("failed to update policy: %w", err)
	}

	if s.scheduler != nil {
		trigger, err := s.scheduler.GetTriggerFromPolicy(policy)
		if err != nil {
			return nil, fmt.Errorf("failed to get trigger from policy: %w", err)
		}

		if err := s.db.UpdateTriggerTx(ctx, policy.ID, *trigger, tx); err != nil {
			return nil, fmt.Errorf("failed to update trigger execution tx: %w", err)
		}
	}

	if s.syncer != nil {
		if err := s.syncer.UpdatePolicySync(policy); err != nil {
			return nil, fmt.Errorf("failed to sync update policy with verifier: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return updatedPolicy, nil
}

func (s *PolicyService) DeletePolicyWithSync(ctx context.Context, policyID string) error {

	tx, err := s.db.Pool().Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = s.db.DeletePluginPolicyTx(ctx, tx, policyID)
	if err != nil {
		return fmt.Errorf("failed to delete policy: %w", err)
	}

	if s.syncer != nil {
		if err := s.syncer.DeletePolicySync(policyID); err != nil {
			return fmt.Errorf("failed to sync delete policy with verifier: %w", err)
		}
=======
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
>>>>>>> 24b0d62 (added policyService)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
<<<<<<< HEAD

func (s *PolicyService) GetPluginPolicies(ctx context.Context, pluginType, publicKey string) ([]types.PluginPolicy, error) {
	policies, err := s.db.GetAllPluginPolicies(ctx, pluginType, publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}
	return policies, nil
}

func (s *PolicyService) GetPluginPolicy(ctx context.Context, policyID string) (types.PluginPolicy, error) {
	policy, err := s.db.GetPluginPolicy(ctx, policyID)
	if err != nil {
		return types.PluginPolicy{}, fmt.Errorf("failed to get policy: %w", err)
	}
	return policy, nil
}
=======
>>>>>>> 24b0d62 (added policyService)
