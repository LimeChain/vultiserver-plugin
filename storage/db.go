package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vultisig/vultisigner/internal/types"
)

type DatabaseStorage interface {
	Close() error

	InsertPluginPolicy(policyDoc types.PluginPolicy) (types.PluginPolicy, error)
	GetPluginPolicy(id string) (types.PluginPolicy, error)
	GetAllPluginPolicies(publicKey string, pluginType string) ([]types.PluginPolicy, error)
	UpdatePluginPolicy(policyDoc types.PluginPolicy) (types.PluginPolicy, error)
	DeletePluginPolicy(ctx context.Context, tx pgx.Tx, id string) error

	CreateTimeTrigger(trigger types.TimeTrigger) error
	GetPendingTriggers() ([]types.TimeTrigger, error)
	UpdateTriggerExecution(policyID string) error

	CreateTransactionHistory(tx types.TransactionHistory) (uuid.UUID, error)
	UpdateTransactionStatus(txID uuid.UUID, status types.TransactionStatus, metadata map[string]interface{}) error
	GetTransactionHistory(policyID uuid.UUID) ([]types.TransactionHistory, error)

	InsertPluginPolicyTx(ctx context.Context, tx pgx.Tx, policy types.PluginPolicy) error
	CreateTimeTriggerTx(ctx context.Context, tx pgx.Tx, trigger types.TimeTrigger) error
	UpdatePluginPolicyTx(ctx context.Context, tx pgx.Tx, policy types.PluginPolicy) error
	UpdateTriggerExecutionTx(ctx context.Context, tx pgx.Tx, policyID string) error
	Pool() *pgxpool.Pool
}
