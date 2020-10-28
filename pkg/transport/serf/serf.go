package serf

import (
	"errors"
	"sync"

	"github.com/hashicorp/serf/client"

	"github.com/makerdao/gofer/pkg/transport"
)

type Serf struct {
	mu sync.Mutex

	client       *client.RPCClient
	streamHandle client.StreamHandle
	subscribers  map[string]subscriber
	rawMsgCh     chan map[string]interface{}
	doneCh       chan struct{}
}

type subscriber struct {
	payloadCh chan []byte
	statusCh  chan transport.Status
	doneCh    chan struct{}
}

func NewSerf(rpc string, queueSize int) (*Serf, error) {
	serfClient, err := client.ClientFromConfig(&client.Config{Addr: rpc})
	if err != nil {
		return nil, err
	}

	return &Serf{
		client:      serfClient,
		subscribers: make(map[string]subscriber, 0),
		rawMsgCh:    make(chan map[string]interface{}, queueSize),
		doneCh:      make(chan struct{}, 0),
	}, nil
}

func (s *Serf) Broadcast(eventName string, event transport.Event) error {
	payload, err := event.PayloadMarshall()
	if err != nil {
		return err
	}

	return s.client.UserEvent(eventName, payload, false)
}

func (s *Serf) Subscribe(eventName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[eventName]; ok {
		return errors.New("already subscribed")
	}

	err := s.startStream()
	if err != nil {
		return err
	}

	s.subscribers[eventName] = subscriber{
		payloadCh: make(chan []byte, 0),
		statusCh:  make(chan transport.Status, 0),
		doneCh:    make(chan struct{}, 0),
	}

	return nil
}

func (s *Serf) Unsubscribe(eventName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[eventName]; !ok {
		return errors.New("unable to unsubscribe")
	}

	subscriber := s.subscribers[eventName]
	close(subscriber.doneCh)
	close(subscriber.statusCh)
	close(subscriber.payloadCh)
	delete(s.subscribers, eventName)

	if len(s.subscribers) == 0 {
		return s.stopStream()
	}

	return nil
}

func (s *Serf) WaitFor(eventName string, event transport.Event) chan transport.Status {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscribers[eventName]; !ok {
		return nil
	}

	go func() {
		select {
		case <-s.subscribers[eventName].doneCh:
			s.handle(eventName, transport.Status{
				Error: errors.New("closed"),
			})
		case payload := <-s.subscribers[eventName].payloadCh:
			s.handle(eventName, transport.Status{
				Error: event.PayloadUnmarshall(payload),
			})
		}
	}()

	return s.subscribers[eventName].statusCh
}

func (s *Serf) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	close(s.doneCh)
	close(s.rawMsgCh)
	for _, subscriber := range s.subscribers {
		close(subscriber.doneCh)
		close(subscriber.statusCh)
		close(subscriber.payloadCh)
	}

	s.subscribers = nil
	return s.client.Close()
}

func (s *Serf) handle(eventName string, status transport.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if subscriber, ok := s.subscribers[eventName]; ok {
		select {
		case <-subscriber.doneCh:
			return
		case subscriber.statusCh <- status:
			return
		}
	}
}

func (s *Serf) startStream() error {
	if s.streamHandle != 0 {
		return nil
	}

	var err error
	s.streamHandle, err = s.client.Stream("user", s.rawMsgCh)
	if err != nil {
		return err
	}

	go s.listen()
	return nil
}

func (s *Serf) stopStream() error {
	if s.streamHandle == 0 {
		return nil
	}

	return s.client.Stop(s.streamHandle)
}

func (s *Serf) listen() {
	for {
		select {
		case <-s.doneCh:
			return
		case msg := <-s.rawMsgCh:
			if subscriber, ok := s.subscribers[msg["Name"].(string)]; ok {
				subscriber.payloadCh <- msg["Payload"].([]byte)
			}
		}
	}
}
