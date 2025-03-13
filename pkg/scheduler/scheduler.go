package scheduler

import (
	"cmp"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/forest33/honeybee/pkg/logger"
)

type Scheduler struct {
	cfg    *Config
	log    *logger.Logger
	tasks  map[string][]*Task
	taskID atomic.Int64
	sync.Mutex
}

func New(cfg *Config, log *logger.Logger) *Scheduler {
	cfg.normalize()

	return &Scheduler{
		cfg:   cfg,
		log:   log,
		tasks: make(map[string][]*Task),
	}
}

func (s *Scheduler) AddTask(t *Task) {
	s.Lock()
	defer s.Unlock()

	t.Normalize()

	if s.cfg.MaxTasksPerSender > 0 && len(s.tasks[t.Sender]) >= s.cfg.MaxTasksPerSender {
		s.tasks[t.Sender][0].Stop()
		s.tasks[t.Sender] = slices.Delete(s.tasks[t.Sender], 0, 1)
	}

	if s.tasks[t.Sender] == nil {
		s.tasks[t.Sender] = make([]*Task, 0, 1)
	}

	s.tasks[t.Sender] = append(s.tasks[t.Sender], t)

	t.id = s.taskID.Add(1)
	t.createdAt = time.Now()
	t.maxDelay = s.cfg.MaxTaskDelay
	t.timer = time.AfterFunc(t.Delay, func() {
		s.task(t)
	})
}

func (s *Scheduler) task(t *Task) {
	s.Lock()
	defer s.Unlock()

	s.log.Debug().Str("sender", t.Sender).Float64("attempt", t.attempt).Msg("run scheduler task")

	if err := t.Handler(); err == nil {
		s.log.Debug().Str("sender", t.Sender).Float64("attempt", t.attempt).Msg("task completed")
		s.deleteTask(t)
		return
	}

	d, ok := t.GetDelay()
	if !ok {
		s.log.Debug().
			Str("sender", t.Sender).
			Float64("attempt", t.attempt).
			Dur("processing_time", time.Since(t.createdAt)).
			Msg("maximum number of attempts or processing time exceeded")
		return
	}

	t.timer = time.AfterFunc(d, func() {
		s.task(t)
	})
}

func (s *Scheduler) deleteTask(t *Task) {
	idx, ok := slices.BinarySearchFunc(s.tasks[t.Sender], t, func(a, b *Task) int {
		return cmp.Compare(a.id, b.id)
	})
	if !ok {
		return
	}
	s.tasks[t.Sender] = slices.Delete(s.tasks[t.Sender], idx, idx+1)
}
