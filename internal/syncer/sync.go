package syncer

import (
	"bytes"
<<<<<<< HEAD
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

=======
	"context"
	"encoding/json"
	"fmt"
>>>>>>> 48ccef6 (feat(syncer):basic setup in creation policy flow)
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/config"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
<<<<<<< HEAD
)

const (
	defaultTimeout = 10 * time.Second
	policyEndpoint = "/plugin/policy"

	// Retry configuration
	maxRetries     = 3
	initialBackoff = 100 * time.Millisecond
)

type PolicySyncer interface {
	CreatePolicySync(policy types.PluginPolicy) error
	UpdatePolicySync(policy types.PluginPolicy) error
	DeletePolicySync(policyID string) error
}

type Syncer struct {
	db         storage.DatabaseStorage
	logger     *logrus.Logger
	client     *http.Client
	config     *config.Config
	serverAddr string
}

func NewSyncService(db storage.DatabaseStorage, logger *logrus.Logger, cfg *config.Config) PolicySyncer {
	if db == nil {
		logger.Fatal("database connection is nil")
	}
=======
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
>>>>>>> 48ccef6 (feat(syncer):basic setup in creation policy flow)
	return &Syncer{
		db:     db,
		logger: logger,
		config: cfg,
		client: &http.Client{
<<<<<<< HEAD
			Timeout: defaultTimeout,
		},
		serverAddr: fmt.Sprintf("http://%s:%d", cfg.Server.Host, cfg.Server.Port),
	}
}

func (s *Syncer) CreatePolicySync(policy types.PluginPolicy) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"plugin_type": policy.PluginType,
	}).Info("Starting policy creation sync")

	return s.retryWithBackoff("CreatePolicySync", func() error {
		policyBytes, err := json.Marshal(policy)
		if err != nil {
			return fmt.Errorf("fail to marshal policy: %w", err)
		}

		url := s.serverAddr + policyEndpoint

		resp, err := s.client.Post(url, "application/json", bytes.NewBuffer(policyBytes))
		if err != nil {
			return fmt.Errorf("fail to sync policy with verifier server: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"policy_id":   policy.ID,
			}).Error("Failed to sync create policy")
			return fmt.Errorf("fail to sync policy with verifier server: status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"policy_id": policy.ID,
		}).Info("Successfully sync created policy")

		return nil
	})
}

func (s *Syncer) UpdatePolicySync(policy types.PluginPolicy) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id":   policy.ID,
		"plugin_type": policy.PluginType,
	}).Info("Starting policy update sync")

	return s.retryWithBackoff("UpdatePolicySync", func() error {
		policyBytes, err := json.Marshal(policy)
		if err != nil {
			return fmt.Errorf("fail to marshal policy: %w", err)
		}

		url := s.serverAddr + policyEndpoint

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(policyBytes))
		if err != nil {
			return fmt.Errorf("fail to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("fail to sync policy with verifier server: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"policy_id":   policy.ID,
			}).Error("Failed to sync update policy")
			return fmt.Errorf("fail to sync policy with verifier server, status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"policy_id": policy.ID,
		}).Info("Successfully sync updated policy")

		return nil
	})
}

func (s *Syncer) DeletePolicySync(policyID string) error {
	s.logger.WithFields(logrus.Fields{
		"policy_id": policyID,
	}).Info("Starting policy delete sync")

	return s.retryWithBackoff("DeletePolicySync", func() error {
		url := s.serverAddr + policyEndpoint + "/" + policyID

		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			return fmt.Errorf("fail to create request, err: %w", err)
		}

		resp, err := s.client.Do(req)
		if err != nil {
			return fmt.Errorf("fail to delete policy on verifier server, err: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusNoContent {
			s.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
				"policy_id":   policyID,
			}).Error("Failed to sync delete policy")
			return fmt.Errorf("fail to delete policy on verifier server, status: %d", resp.StatusCode)
		}

		s.logger.WithFields(logrus.Fields{
			"policy_id": policyID,
		}).Info("Successfully sync deleted policy")

		return nil
	})
}

// retryWithBackoff attempts to execute the given operation with exponential backoff
func (s *Syncer) retryWithBackoff(operation string, fn func() error) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			s.logger.WithFields(logrus.Fields{
				"attempt":   attempt,
				"backoff":   backoff.String(),
				"operation": operation,
			}).Debug("Retrying sync")

			time.Sleep(backoff)
			backoff *= 2
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		s.logger.WithFields(logrus.Fields{
			"attempt":   attempt,
			"error":     err.Error(),
			"operation": operation,
		}).Warn("Sync failed, will retry")
	}

	return fmt.Errorf("sync failed: %w", lastErr)
=======
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
>>>>>>> 48ccef6 (feat(syncer):basic setup in creation policy flow)
}
