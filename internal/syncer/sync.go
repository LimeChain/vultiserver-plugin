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
