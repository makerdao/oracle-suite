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
	payloadCh    map[string]chan []byte
	statusCh     map[string]chan transport.Status
	msgCh        chan map[string]interface{}
	doneCh       chan bool
}

func NewSerf(rpc string, queueSize int) (*Serf, error) {
	serfClient, err := client.ClientFromConfig(&client.Config{Addr: rpc})
	if err != nil {
		return nil, err
	}

	return &Serf{
		client:    serfClient,
		payloadCh: make(map[string]chan []byte, 0),
		statusCh:  make(map[string]chan transport.Status, 0),
		msgCh:     make(chan map[string]interface{}, queueSize),
		doneCh:    make(chan bool, 0),
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

	if _, ok := s.payloadCh[eventName]; ok {
		return errors.New("already subscribed")
	}

	err := s.stream()
	if err != nil {
		return err
	}

	s.payloadCh[eventName] = make(chan []byte, 0)
	s.statusCh[eventName] = make(chan transport.Status, 0)

	return nil
}

func (s *Serf) Unsubscribe(eventName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.payloadCh[eventName]; !ok {
		return errors.New("unable to unsubscribe")
	}

	close(s.payloadCh[eventName])
	close(s.statusCh[eventName])
	delete(s.payloadCh, eventName)
	delete(s.statusCh, eventName)

	return nil
}

func (s *Serf) WaitFor(eventName string, event transport.Event) chan transport.Status {
	go func() {
		select {
		case <-s.doneCh:
			s.send(eventName, transport.Status{
				Error: errors.New("closed"),
			})
		case payload := <-s.payloadCh[eventName]:
			s.send(eventName, transport.Status{
				Error: event.PayloadUnmarshall(payload),
			})
		}
	}()

	return s.statusCh[eventName]
}

func (s *Serf) Close() error {
	close(s.doneCh)
	for eventName, _ := range s.statusCh {
		err := s.Unsubscribe(eventName)
		if err != nil {
			return err
		}
	}

	return s.client.Close()
}

func (s *Serf) send(eventName string, status transport.Status) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.statusCh[eventName]; ok {
		ch <- status
	}
}

func (s *Serf) stream() error {
	if s.streamHandle != 0 {
		return nil
	}

	var err error
	s.streamHandle, err = s.client.Stream("user", s.msgCh)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-s.doneCh:
				return
			case msg := <-s.msgCh:
				if ch, ok := s.payloadCh[msg["Name"].(string)]; ok {
					ch <- msg["Payload"].([]byte)
				}
			}
		}
	}()

	return nil
}
