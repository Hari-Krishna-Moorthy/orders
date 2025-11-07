// internals/app/pubsub/pubsub.go
package pubsub

import (
	"container/ring"
	"sync"
	"time"

	error_types "github.com/Hari-Krishna-Moorthy/orders/internals/app/types"
)

type Message struct {
	ID      string      `json:"id"`
	Payload interface{} `json:"payload"`
	TS      time.Time   `json:"ts"`
}

type Subscriber struct {
	ID   string
	Ch   chan Message
	Quit chan struct{}
}

type Topic struct {
	Name        string
	Subs        map[string]*Subscriber
	Mu          sync.RWMutex
	Ring        *ring.Ring
	RingSize    int
	TotalEvents uint64
}

type Manager struct {
	Mu     sync.RWMutex
	Topics map[string]*Topic
}

func NewManager() *Manager { return &Manager{Topics: map[string]*Topic{}} }

// --- Topic management ---

func (m *Manager) CreateTopic(name string, ringSize int) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	if _, ok := m.Topics[name]; ok {
		return error_types.ALREADY_EXISTS_ERROR
	}
	if ringSize <= 0 {
		ringSize = 1
	}
	m.Topics[name] = &Topic{
		Name:     name,
		Subs:     map[string]*Subscriber{},
		Ring:     ring.New(ringSize),
		RingSize: ringSize,
	}
	return nil
}

func (m *Manager) DeleteTopic(name string) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	t, ok := m.Topics[name]
	if !ok {
		return error_types.NOT_FOUND_ERROR
	}
	for _, s := range t.Subs {
		close(s.Quit)
	}
	delete(m.Topics, name)
	return nil
}

func (m *Manager) List() []map[string]interface{} {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	out := make([]map[string]interface{}, 0, len(m.Topics))
	for _, t := range m.Topics {
		t.Mu.RLock()
		out = append(out, map[string]interface{}{
			"name":        t.Name,
			"subscribers": len(t.Subs),
		})
		t.Mu.RUnlock()
	}
	return out
}

func (m *Manager) Stats() map[string]map[string]interface{} {
	m.Mu.RLock()
	defer m.Mu.RUnlock()
	out := make(map[string]map[string]interface{}, len(m.Topics))
	for name, t := range m.Topics {
		t.Mu.RLock()
		out[name] = map[string]interface{}{
			"messages":    t.TotalEvents,
			"subscribers": len(t.Subs),
		}
		t.Mu.RUnlock()
	}
	return out
}

func (m *Manager) Subscribe(topic, subID string, buf int) (*Subscriber, error) {
	m.Mu.RLock()
	t, ok := m.Topics[topic]
	m.Mu.RUnlock()
	if !ok {
		return nil, error_types.TOPIC_NOT_FOUND_ERROR
	}
	if buf <= 0 {
		buf = 1
	}
	s := &Subscriber{
		ID:   subID,
		Ch:   make(chan Message, buf),
		Quit: make(chan struct{}),
	}
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if _, exists := t.Subs[subID]; exists {
		return nil, error_types.ALREADY_EXISTS_ERROR
	}
	t.Subs[subID] = s
	return s, nil
}

func (m *Manager) Unsubscribe(topic, subID string) error {
	m.Mu.RLock()
	t, ok := m.Topics[topic]
	m.Mu.RUnlock()
	if !ok {
		return error_types.TOPIC_NOT_FOUND_ERROR
	}
	t.Mu.Lock()
	defer t.Mu.Unlock()
	if s, ok := t.Subs[subID]; ok {
		close(s.Quit)
		delete(t.Subs, subID)
	}
	return nil
}

func (m *Manager) Publish(topic string, msg Message) error {
	m.Mu.RLock()
	t, ok := m.Topics[topic]
	m.Mu.RUnlock()
	if !ok {
		return error_types.TOPIC_NOT_FOUND_ERROR
	}
	// Save and fan out
	t.Mu.Lock()
	t.Ring.Value = msg
	t.Ring = t.Ring.Next()
	t.TotalEvents++
	for _, s := range t.Subs {
		// Backpressure policy: drop oldest in subscriber queue
		select {
		case s.Ch <- msg:
		default:
			<-s.Ch
			s.Ch <- msg
		}
	}
	t.Mu.Unlock()
	return nil
}

func (m *Manager) Replay(topic string, lastN int) ([]Message, error) {
	m.Mu.RLock()
	t, ok := m.Topics[topic]
	m.Mu.RUnlock()
	if !ok {
		return nil, error_types.TOPIC_NOT_FOUND_ERROR
	}
	if lastN <= 0 {
		return nil, nil
	}
	if lastN > t.RingSize {
		lastN = t.RingSize
	}

	msgs := make([]Message, 0, lastN)
	t.Mu.RLock()
	defer t.Mu.RUnlock()

	// iterate oldest -> newest
	r := t.Ring
	for i := 0; i < t.RingSize; i++ {
		r = r.Prev()
	}
	for i := 0; i < lastN; i++ {
		if r.Value != nil {
			msgs = append(msgs, r.Value.(Message))
		}
		r = r.Next()
	}
	return msgs, nil
}
