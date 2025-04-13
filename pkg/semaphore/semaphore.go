package semaphore

import (
	"context"
)

type Logger interface {
	Debug(format string, args ...any)
	Info(format string, args ...any)
	Fail(format string, args ...any)
	Error(format string, args ...any)
}

type Semaphore struct {
	logger Logger
	sem    chan struct{}
}

func New(logger Logger) *Semaphore {
	return &Semaphore{
		logger: logger,
		sem:    make(chan struct{}, 1),
	}
}

func (s *Semaphore) TryAcquire(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	case s.sem <- struct{}{}:
		s.logger.Debug("Semaphore acquired")
		return true
	default:
		return false
	}
}

func (s *Semaphore) Release() {
	s.logger.Debug("Semaphore released")
	<-s.sem
}
