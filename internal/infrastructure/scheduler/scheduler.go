package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
	log  *slog.Logger
}

func New(log *slog.Logger) *Scheduler {
	c := cron.New(cron.WithSeconds(), cron.WithChain(cron.Recover(cronLog{log})))
	return &Scheduler{cron: c, log: log}
}

func (s *Scheduler) Every(d time.Duration, name string, fn func(context.Context)) error {
	_, err := s.cron.AddFunc("@every "+d.String(), func() {
		ctx, cancel := context.WithTimeout(context.Background(), d)
		defer cancel()
		s.log.Info("job tick", "job", name)
		fn(ctx)
	})
	return err
}

func (s *Scheduler) Start(ctx context.Context) {
	s.cron.Start()
	<-ctx.Done()
	stopCtx := s.cron.Stop()
	<-stopCtx.Done()
}

type cronLog struct{ *slog.Logger }

func (l cronLog) Info(msg string, kv ...any) { l.Logger.Info(msg, kv...) }
func (l cronLog) Error(err error, msg string, kv ...any) {
	l.Logger.Error(msg, append(kv, "err", err)...)
}
