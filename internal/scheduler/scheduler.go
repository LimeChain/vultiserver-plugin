package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/vultisig/vultisigner/internal/tasks"
	"github.com/vultisig/vultisigner/internal/types"
	"github.com/vultisig/vultisigner/storage"
)

type SchedulerService struct {
	db        storage.DatabaseStorage
	logger    *logrus.Logger
	client    *asynq.Client
	inspector *asynq.Inspector
	done      chan struct{}
}

func NewSchedulerService(db storage.DatabaseStorage, logger *logrus.Logger, client *asynq.Client, redisOpts asynq.RedisClientOpt) *SchedulerService {
	if db == nil {
		logger.Fatal("database connection is nil")
	}

	// create inspector using the same Redis configuration as the client
	inspector := asynq.NewInspector(redisOpts)

	return &SchedulerService{
		db:        db,
		logger:    logger,
		client:    client,
		inspector: inspector,
		done:      make(chan struct{}),
	}
}

func (s *SchedulerService) Start() {
	go s.run()
}

func (s *SchedulerService) Stop() {
	close(s.done)
}

func (s *SchedulerService) run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.logger.Info("Checking and enqueuing tasks")
			if err := s.checkAndEnqueueTasks(); err != nil {
				s.logger.Errorf("Failed to check and enqueue tasks: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

func (s *SchedulerService) checkAndEnqueueTasks() error {
	triggers, err := s.db.GetPendingTriggers()
	if err != nil {
		return fmt.Errorf("failed to get pending triggers: %w", err)
	}
	s.logger.Info("Triggers: ", triggers)

	for _, trigger := range triggers {
		s.logger.WithFields(logrus.Fields{
			"policy_id": trigger.PolicyID,
			"last_exec": trigger.LastExecution,
		}).Info("Processing trigger")
		// Parse cron expression
		schedule, err := cron.ParseStandard(trigger.CronExpression)
		if err != nil {
			s.logger.Errorf("Failed to parse cron expression: %v", err)
			continue
		}

		// Check if it's time to execute
		var nextTime time.Time
		if trigger.LastExecution != nil {
			nextTime = schedule.Next(*trigger.LastExecution)
		} else {
			nextTime = schedule.Next(time.Now().Add(-24 * time.Hour))
		}

		nextTime = nextTime.UTC()

		isAfter := time.Now().UTC().After(nextTime)
		fmt.Println("IS AFTER: ", isAfter)

		s.logger.WithFields(logrus.Fields{
			"current_time": time.Now().UTC(),
			"next_time":    nextTime,
			"policy_id":    trigger.PolicyID,
			"last_exec":    trigger.LastExecution,
		}).Info("Checking execution time")

		if time.Now().UTC().After(nextTime) {
			triggerEvent := types.PluginTriggerEvent{
				PolicyID: trigger.PolicyID,
			}

			buf, err := json.Marshal(triggerEvent)
			if err != nil {
				s.logger.Errorf("Failed to marshal trigger event: %v", err)
				continue
			}
			ti, err := s.client.Enqueue(
				asynq.NewTask(tasks.TypePluginTransaction, buf),
				asynq.MaxRetry(-1),
				asynq.Timeout(5*time.Minute),
				asynq.Retention(10*time.Minute),
				asynq.Queue(tasks.QUEUE_NAME),
			)
			if err != nil {
				s.logger.Errorf("Failed to enqueue trigger task: %v", err)
				continue
			}

			s.logger.WithFields(logrus.Fields{
				"task_id":   ti.ID,
				"policy_id": trigger.PolicyID,
			}).Info("Enqueued trigger task")

			// TODO: quick hack to prevent multiple executions
			time.Sleep(1 * time.Minute)
		}
	}

	return nil
}

func (s *SchedulerService) CreateTimeTrigger(ctx context.Context, policy types.PluginPolicy, tx pgx.Tx) error {
	if s.db == nil {
		return fmt.Errorf("database backend is nil")
	}

	s.logger.Info("Attempting to parse policy schedule")

	var policySchedule struct {
		Schedule struct {
			Frequency string     `json:"frequency"`
			StartTime time.Time  `json:"start_time"`
			Interval  string     `json:"interval"`
			EndTime   *time.Time `json:"end_time,omitempty"`
		} `json:"schedule"`
	}

	if err := json.Unmarshal(policy.Policy, &policySchedule); err != nil {
		return fmt.Errorf("failed to parse policy schedule: %w", err)
	}

	s.logger.Info("Frequency to cron")
	interval, err := strconv.Atoi(policySchedule.Schedule.Interval)
	if err != nil {
		return fmt.Errorf("failed to parse interval: %w", err)
	}

	cronExpr := frequencyToCron(policySchedule.Schedule.Frequency, policySchedule.Schedule.StartTime, interval)
	if cronExpr == "" {
		return fmt.Errorf("invalid cron expression")
	}

	trigger := types.TimeTrigger{
		PolicyID:       policy.ID,
		CronExpression: cronExpr,
		StartTime:      policySchedule.Schedule.StartTime,
		EndTime:        policySchedule.Schedule.EndTime,
		Frequency:      policySchedule.Schedule.Frequency,
	}

	fmt.Println("CRON EXPRESSION: ", cronExpr)
	fmt.Println("START TIMEEE:  ", policySchedule.Schedule.StartTime)
	fmt.Println("END TIMEEE:  ", policySchedule.Schedule.EndTime)
	fmt.Println("FREQUENCYY: ", policySchedule.Schedule.Frequency)
	fmt.Println("INTERVALLL: ", policySchedule.Schedule.Interval)

	return s.db.CreateTimeTriggerTx(ctx, tx, trigger)
}

func frequencyToCron(frequency string, startTime time.Time, interval int) string {
	switch frequency {
	case "minutely":
		if interval < 15 {
			return ""
		}
		return fmt.Sprintf("*/%d * * * *", interval)
	case "hourly":
		if interval == 1 {
			return fmt.Sprintf("%d * * * *", startTime.Minute())
		}
		return fmt.Sprintf("%d */%d * * *", startTime.Minute(), interval)
	case "daily":
		if interval == 1 {
			return fmt.Sprintf("%d %d * * *", startTime.Minute(), startTime.Hour())
		}
		return fmt.Sprintf("%d %d */%d * *", startTime.Minute(), startTime.Hour(), interval)
	case "weekly":
		if interval == 1 {
			return fmt.Sprintf("%d %d * * %d", startTime.Minute(), startTime.Hour(), startTime.Weekday())
		}
		// For weekly intervals > 1, we need to use the day of month instead
		// This is a limitation of cron expressions
		return fmt.Sprintf("%d %d 1-31/%d * %d",
			startTime.Minute(),
			startTime.Hour(),
			7*interval,
			startTime.Weekday())
	case "monthly":
		if interval == 1 {
			return fmt.Sprintf("%d %d %d * *", startTime.Minute(), startTime.Hour(), startTime.Day())
		}
		return fmt.Sprintf("%d %d %d */%d *",
			startTime.Minute(),
			startTime.Hour(),
			startTime.Day(),
			interval)
	default:
		return ""
	}
}
