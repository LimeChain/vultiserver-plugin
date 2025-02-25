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
		schedule, err := createSchedule(trigger.CronExpression, trigger.Frequency, trigger.StartTime, trigger.Interval)
		if err != nil {
			s.logger.Errorf("Failed to create schedule: %v", err)
			continue
		}

		// Check if it's time to execute
		var nextTime time.Time
		if trigger.LastExecution != nil {
			nextTime = schedule.Next(*trigger.LastExecution)
		} else {
			nextTime = trigger.StartTime
		}

		nextTime = nextTime.UTC()

		s.logger.WithFields(logrus.Fields{
			"current_time":    time.Now().UTC(),
			"next_time":       nextTime,
			"start_time":      trigger.StartTime.UTC(),
			"policy_id":       trigger.PolicyID,
			"cron_expression": trigger.CronExpression,
			"last_exec":       trigger.LastExecution,
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

func (s *SchedulerService) CreateTimeTrigger(ctx context.Context, policy types.PluginPolicy, dbTx pgx.Tx) error {
	if s.db == nil {
		return fmt.Errorf("database backend is nil")
	}

	trigger, err := s.GetTriggerFromPolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to get trigger from policy: %w", err)
	}

	return s.db.CreateTimeTriggerTx(ctx, dbTx, *trigger)
}

func (s *SchedulerService) GetTriggerFromPolicy(policy types.PluginPolicy) (*types.TimeTrigger, error) {
	var policySchedule struct {
		Schedule struct {
			Frequency string     `json:"frequency"`
			StartTime time.Time  `json:"start_time"`
			Interval  string     `json:"interval"`
			EndTime   *time.Time `json:"end_time,omitempty"`
		} `json:"schedule"`
	}

	if err := json.Unmarshal(policy.Policy, &policySchedule); err != nil {
		return nil, fmt.Errorf("failed to parse policy schedule: %w", err)
	}

	interval, err := strconv.Atoi(policySchedule.Schedule.Interval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval: %w", err)
	}

	cronExpr := frequencyToCron(policySchedule.Schedule.Frequency, policySchedule.Schedule.StartTime, interval)
	trigger := types.TimeTrigger{
		PolicyID:       policy.ID,
		CronExpression: cronExpr,
		StartTime:      policySchedule.Schedule.StartTime,
		EndTime:        policySchedule.Schedule.EndTime,
		Frequency:      policySchedule.Schedule.Frequency,
		Interval:       interval,
	}

	return &trigger, nil
}

func createSchedule(cronExpr, frequency string, startTime time.Time, interval int) (cron.Schedule, error) {
	// Use our custom schedule implementation for intervals > 1 and when frequency is daily, weekly, monthly
	if interval > 1 && (frequency == "daily" || frequency == "weekly" || frequency == "monthly") {
		return NewIntervalSchedule(frequency, startTime, interval)
	}

	// For standard cron
	schedule, err := cron.ParseStandard(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cron expression: %w", err)
	}

	return schedule, nil
}

func frequencyToCron(frequency string, startTime time.Time, interval int) string {
	switch frequency {
	case "minutely":
		return fmt.Sprintf("*/%d * * * *", interval)
	case "hourly":
		if interval == 1 {
			return fmt.Sprintf("%d * * * *", startTime.Minute())
		}
		return fmt.Sprintf("%d */%d * * *", startTime.Minute(), interval)
	case "daily":
		return fmt.Sprintf("%d %d * * *", startTime.Minute(), startTime.Hour())
	case "weekly":
		return fmt.Sprintf("%d %d * * %d", startTime.Minute(), startTime.Hour(), startTime.Weekday())
	case "monthly":
		return fmt.Sprintf("%d %d %d * *", startTime.Minute(), startTime.Hour(), startTime.Day())
	default:
		return ""
	}
}

type IntervalSchedule struct {
	Frequency string
	Interval  int
	StartTime time.Time
	Minute    int
	Hour      int
	Day       int
	Weekday   time.Weekday
	Location  *time.Location
}

func NewIntervalSchedule(frequency string, startTime time.Time, interval int) (*IntervalSchedule, error) {
	if interval < 1 {
		return nil, fmt.Errorf("interval must be at least 1")
	}

	return &IntervalSchedule{
		Frequency: frequency,
		Interval:  interval,
		StartTime: startTime,
		Minute:    startTime.Minute(),
		Hour:      startTime.Hour(),
		Day:       startTime.Day(),
		Weekday:   startTime.Weekday(),
		Location:  startTime.Location(),
	}, nil
}

func (s *IntervalSchedule) Next(t time.Time) time.Time {
	t = t.In(s.Location)

	switch s.Frequency {
	case "daily":
		return s.nextDaily(t)
	case "weekly":
		return s.nextWeekly(t)
	case "monthly":
		return s.nextMonthly(t)
	default:
		return time.Time{}
	}
}

func (s *IntervalSchedule) nextDaily(t time.Time) time.Time {
	// First, find the next occurrence at the correct time of day
	nextTime := time.Date(t.Year(), t.Month(), t.Day(), s.Hour, s.Minute, 0, 0, s.Location)

	// If that time is in the past, move to the next day
	if !nextTime.After(t) {
		nextTime = nextTime.AddDate(0, 0, 1)
	}

	// Calculate days since start time
	daysSinceStart := int(nextTime.Sub(s.StartTime).Hours()) / 24

	// If we're not on a valid interval day, adjust forward
	if remainder := daysSinceStart % s.Interval; remainder != 0 {
		daysToAdd := s.Interval - remainder
		nextTime = nextTime.AddDate(0, 0, daysToAdd)
	}

	return nextTime
}

func (s *IntervalSchedule) nextWeekly(t time.Time) time.Time {
	// First, find the next occurrence of the correct weekday
	daysUntilWeekday := int(s.Weekday - t.Weekday())
	if daysUntilWeekday <= 0 {
		daysUntilWeekday += 7
	}

	// Calculate the candidate time (next occurrence of the correct weekday)
	nextTime := time.Date(
		t.Year(), t.Month(), t.Day()+daysUntilWeekday,
		s.Hour, s.Minute, 0, 0, s.Location,
	)

	// If that time is in the past, add a week
	if !nextTime.After(t) {
		nextTime = nextTime.AddDate(0, 0, 7)
	}

	// Calculate weeks since start time
	weeksSinceStart := int(nextTime.Sub(s.StartTime).Hours()) / (24 * 7)

	// If we're not on a valid interval week, adjust forward
	if remainder := weeksSinceStart % s.Interval; remainder != 0 {
		weeksToAdd := s.Interval - remainder
		nextTime = nextTime.AddDate(0, 0, 7*weeksToAdd)
	}

	return nextTime
}

func (s *IntervalSchedule) nextMonthly(t time.Time) time.Time {
	startMonth := s.StartTime.Month()
	startYear := s.StartTime.Year()

	// Calculate the candidate date
	candidateMonth := t.Month()
	candidateYear := t.Year()

	// If we're past the day of month in the current month, go to next month
	if t.Day() > s.Day || (t.Day() == s.Day && (t.Hour() > s.Hour || (t.Hour() == s.Hour && t.Minute() >= s.Minute))) {
		if candidateMonth == time.December {
			candidateMonth = time.January
			candidateYear++
		} else {
			candidateMonth++
		}
	}

	// Create the candidate time
	candidate := time.Date(candidateYear, candidateMonth, s.Day, s.Hour, s.Minute, 0, 0, s.Location)

	// Adjust for months with fewer days than our target day
	if candidate.Day() != s.Day {
		// We got bumped to the next month due to day overflow, go back to last day of previous month
		candidate = time.Date(candidateYear, candidateMonth, 0, s.Hour, s.Minute, 0, 0, s.Location)
	}

	// Calculate months since start
	monthsSinceStart := (candidateYear-startYear)*12 + int(candidateMonth-startMonth)

	// If we're not on a valid interval month, adjust forward
	if remainder := monthsSinceStart % s.Interval; remainder != 0 {
		monthsToAdd := s.Interval - remainder

		// Add the necessary months
		for i := 0; i < monthsToAdd; i++ {
			if candidateMonth == time.December {
				candidateMonth = time.January
				candidateYear++
			} else {
				candidateMonth++
			}
		}

		// Create the new candidate
		candidate = time.Date(candidateYear, candidateMonth, s.Day, s.Hour, s.Minute, 0, 0, s.Location)

		// Adjust for months with fewer days than our target day
		if candidate.Day() != s.Day {
			candidate = time.Date(candidateYear, candidateMonth, 0, s.Hour, s.Minute, 0, 0, s.Location)
		}
	}

	return candidate
}
