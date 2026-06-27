package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/notification/domain"
)

var (
	ErrInvalidInput = errors.New("recipient and message are required")
	ErrMissingUser  = errors.New("user identity is required")
	ErrQueueFull    = errors.New("notification queue is full")
	ErrNotFound     = errors.New("notification not found")
	ErrStopped      = errors.New("notification service is stopped")
)

type SubmitRequest struct {
	Recipient string `json:"recipient"`
	Message   string `json:"message"`
}

type Stats struct {
	Workers     int `json:"workers"`
	QueueSize   int `json:"queue_size"`
	Queued      int `json:"queued"`
	Sending     int `json:"sending"`
	Delivered   int `json:"delivered"`
	Failed      int `json:"failed"`
	JobsWaiting int `json:"jobs_waiting"`
}

type Service struct {
	workers int

	// jobs is the input channel. HTTP handlers send jobs here and worker
	// goroutines receive from it. The buffer lets short traffic bursts wait
	// without blocking every caller immediately.
	jobs chan job

	// results is the output channel. Workers send completed jobs here. A single
	// collector goroutine receives results and updates the in-memory status map.
	results chan result

	mu            sync.RWMutex
	notifications map[string]domain.Notification

	stopOnce    sync.Once
	workerWG    sync.WaitGroup
	collectorWG sync.WaitGroup
	done        chan struct{}
	stopped     atomic.Bool
}

type job struct {
	notification domain.Notification
}

type result struct {
	id       string
	workerID int
	err      error
}

func NewService(workers int, queueSize int) *Service {
	if workers <= 0 {
		workers = 1
	}
	if queueSize <= 0 {
		queueSize = workers
	}

	return &Service{
		workers:       workers,
		jobs:          make(chan job, queueSize),
		results:       make(chan result, queueSize),
		notifications: make(map[string]domain.Notification),
		done:          make(chan struct{}),
	}
}

func (s *Service) Start(ctx context.Context) {
	// Start one collector. This goroutine is the only place that consumes the
	// results channel, so status updates stay ordered and easy to reason about.
	s.collectorWG.Add(1)
	go s.collectResults(ctx)

	// Start N workers. Each worker runs independently and competes for jobs from
	// the same channel. This is the common Go worker-pool pattern.
	for workerID := 1; workerID <= s.workers; workerID++ {
		s.workerWG.Add(1)
		go s.worker(ctx, workerID)
	}
}

func (s *Service) Stop() {
	s.stopOnce.Do(func() {
		s.stopped.Store(true)
		close(s.done)
		s.workerWG.Wait()
		close(s.results)
		s.collectorWG.Wait()
	})
}

func (s *Service) Submit(ctx context.Context, user domain.UserContext, req SubmitRequest) (domain.Notification, error) {
	if req.Recipient == "" || req.Message == "" {
		return domain.Notification{}, ErrInvalidInput
	}
	if user.UserID == "" || user.TenantID == "" || user.Role == "" {
		return domain.Notification{}, ErrMissingUser
	}
	if s.stopped.Load() {
		return domain.Notification{}, ErrStopped
	}

	now := time.Now().UTC()
	notification := domain.Notification{
		ID:          fmt.Sprintf("notif_%d", now.UnixNano()),
		UserID:      user.UserID,
		TenantID:    user.TenantID,
		Recipient:   req.Recipient,
		Message:     req.Message,
		Status:      domain.StatusQueued,
		SubmittedAt: now,
		UpdatedAt:   now,
	}

	s.mu.Lock()
	s.notifications[notification.ID] = notification
	s.mu.Unlock()

	// This select is the key handoff from HTTP to background work:
	// - if jobs has capacity, enqueue the job
	// - if the request context is canceled, stop waiting
	// - if the service is shutting down, reject the job
	// - after a short timeout, report queue pressure instead of hanging
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()

	select {
	case s.jobs <- job{notification: notification}:
		return notification, nil
	case <-ctx.Done():
		s.markFailed(notification.ID, ctx.Err())
		return domain.Notification{}, ctx.Err()
	case <-s.done:
		s.markFailed(notification.ID, ErrStopped)
		return domain.Notification{}, ErrStopped
	case <-timer.C:
		s.markFailed(notification.ID, ErrQueueFull)
		return domain.Notification{}, ErrQueueFull
	}
}

func (s *Service) Find(id string) (domain.Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	notification, ok := s.notifications[id]
	if !ok {
		return domain.Notification{}, ErrNotFound
	}
	return notification, nil
}

func (s *Service) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := Stats{
		Workers:     s.workers,
		QueueSize:   cap(s.jobs),
		JobsWaiting: len(s.jobs),
	}
	for _, notification := range s.notifications {
		switch notification.Status {
		case domain.StatusQueued:
			stats.Queued++
		case domain.StatusSending:
			stats.Sending++
		case domain.StatusDelivered:
			stats.Delivered++
		case domain.StatusFailed:
			stats.Failed++
		}
	}
	return stats
}

func (s *Service) worker(ctx context.Context, workerID int) {
	defer s.workerWG.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.done:
			return
		case item, ok := <-s.jobs:
			if !ok {
				return
			}
			s.markSending(item.notification.ID, workerID)
			err := sendNotification(ctx, item.notification)

			// Workers never update the status map directly. They report the
			// outcome through the results channel, then the collector owns the
			// final write. The done case prevents shutdown from blocking here.
			select {
			case s.results <- result{id: item.notification.ID, workerID: workerID, err: err}:
			case <-ctx.Done():
				return
			case <-s.done:
				return
			}
		}
	}
}

func (s *Service) collectResults(ctx context.Context) {
	defer s.collectorWG.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-s.results:
			if !ok {
				return
			}
			if item.err != nil {
				s.markFailed(item.id, item.err)
				continue
			}
			s.markDelivered(item.id, item.workerID)
		}
	}
}

func sendNotification(ctx context.Context, notification domain.Notification) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(600 * time.Millisecond):
		return nil
	}
}

func (s *Service) markSending(id string, workerID int) {
	s.update(id, func(notification *domain.Notification) {
		notification.Status = domain.StatusSending
		notification.WorkerID = workerID
	})
}

func (s *Service) markDelivered(id string, workerID int) {
	s.update(id, func(notification *domain.Notification) {
		notification.Status = domain.StatusDelivered
		notification.WorkerID = workerID
	})
}

func (s *Service) markFailed(id string, err error) {
	s.update(id, func(notification *domain.Notification) {
		notification.Status = domain.StatusFailed
		if err != nil {
			notification.Error = err.Error()
		}
	})
}

func (s *Service) update(id string, change func(*domain.Notification)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	notification, ok := s.notifications[id]
	if !ok {
		return
	}
	change(&notification)
	notification.UpdatedAt = time.Now().UTC()
	s.notifications[id] = notification
}
