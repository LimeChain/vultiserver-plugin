package syncer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
	"net/http"
	"time"
)

type Syncer struct {
	db     storage.DatabaseStorage
	logger *logrus.Logger
	client *http.Client
	config *config.Config
}

func NewSyncService(db storage.DatabaseStorage, logger *logrus.Logger, cfg *config.Config) *Syncer {
	return &Syncer{
		db:     db,
		logger: logger,
		config: cfg,
		client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

type SyncRequest struct {
	Operation   string             `json:"operation"`
	Policy      types.PluginPolicy `json:"policy"`
	TimeTrigger *types.TimeTrigger `json:"time_trigger,omitempty"`
}

type SyncResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (s *Syncer) SyncWithVerifier(ctx context.Context, req *SyncRequest) (*SyncResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sync request: %w", err)
	}

	verifierEndpoint := fmt.Sprintf("http://%s:%d/plugin/sync", s.config.Server.Host, s.config.Server.Port)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", verifierEndpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send sync request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sync request failed with status %s", resp.Status)
	}

	var syncResp SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, fmt.Errorf("failed to decode sync response: %w", err)
	}
	return &syncResp, nil
}

func (s *Syncer) CreatePolicySync(policy types.PluginPolicy) error {
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("fail to marshal policy, err: %w", err)
	}

	cfg, err := config.ReadConfig("config-server")
	if err != nil {
		return fmt.Errorf("fail to read plugin config, err: %w", err)
	}

	verifierPolicyEndpoint := fmt.Sprintf("http://%s:%d/plugin/policy", cfg.Server.Host, cfg.Server.Port)
	resp, err := http.Post(verifierPolicyEndpoint, "application/json", bytes.NewBuffer(policyBytes))
	if err != nil {
		return fmt.Errorf("fail to sync policy with verifier server, err: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to sync policy with verifier server, status: %d", resp.StatusCode)
	}

	return nil
}

func (s *Syncer) UpdatePolicySync(policy types.PluginPolicy) error {
	policyBytes, err := json.Marshal(policy)
	if err != nil {
		return fmt.Errorf("fail to marshal policy, err: %w", err)
	}

	cfg, err := config.ReadConfig("config-server")
	if err != nil {
		return fmt.Errorf("fail to read plugin config, err: %w", err)
	}

	verifierPolicyEndpoint := fmt.Sprintf("http://%s:%d/plugin/policy", cfg.Server.Host, cfg.Server.Port)

	req, err := http.NewRequest(http.MethodPut, verifierPolicyEndpoint, bytes.NewBuffer(policyBytes))
	if err != nil {
		return fmt.Errorf("fail to create request, err: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fail to sync policy with verifier server, err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fail to sync policy with verifier server, status: %d", resp.StatusCode)
	}

	return nil
}

func (s *Syncer) DeletePolicySync(policyID string) error {
	cfg, err := config.ReadConfig("config-server")
	if err != nil {
		return fmt.Errorf("fail to read plugin config, err: %w", err)
	}

	verifierPolicyEndpoint := fmt.Sprintf("http://%s:%d/plugin/policy/%s", cfg.Server.Host, cfg.Server.Port, policyID)

	req, err := http.NewRequest(http.MethodDelete, verifierPolicyEndpoint, nil)
	if err != nil {
		return fmt.Errorf("fail to create request, err: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("fail to delete policy on verifier server, err: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("fail to delete policy on verifier server, status: %d", resp.StatusCode)
	}

	return nil
}
