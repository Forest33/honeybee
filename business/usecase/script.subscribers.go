package usecase

import (
	"sync"

	"github.com/forest33/honeybee/business/entity"
	"github.com/forest33/honeybee/pkg/structs"
)

type subscribers struct {
	data map[string]map[string]struct{}
	sync.RWMutex
}

func newSubscribers() *subscribers {
	return &subscribers{
		data: make(map[string]map[string]struct{}),
	}
}

func (s *subscribers) add(topic string, script entity.Script, handler func()) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.data[topic]; !ok {
		s.data[topic] = make(map[string]struct{}, 1)
	}

	s.data[topic][script.Path()] = struct{}{}

	handler()
}

func (s *subscribers) getScriptsByTopic(topic string) []string {
	s.RLock()
	defer s.RUnlock()

	if _, ok := s.data[topic]; !ok {
		return nil
	}

	return structs.Keys(s.data[topic])
}

func (s *subscribers) getTopics() []string {
	s.RLock()
	defer s.RUnlock()

	return structs.Keys(s.data)
}
