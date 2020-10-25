package serf

import (
	"sync"

	"github.com/hashicorp/serf/client"

	"github.com/makerdao/gofer/pkg/transport"
)

const messageQueueSize = 1024

type Serf struct {
	mu sync.Mutex

	client       *client.RPCClient
	streamHandle client.StreamHandle
	payloadCh    map[string]chan []byte
	eventCh      map[string]chan transport.Status
	msgCh        chan map[string]interface{}
	doneCh       chan bool
}

func NewSerf(rpc string) (*Serf, error) {
	conf := client.Config{Addr: rpc}

	serfClient, err := client.ClientFromConfig(&conf)
	if err != nil {
		return nil, err
	}

	return &Serf{
		client:    serfClient,
		payloadCh: make(map[string]chan []byte, 0),
		eventCh:   make(map[string]chan transport.Status, 0),
		msgCh:     make(chan map[string]interface{}, messageQueueSize),
		doneCh:    make(chan bool, 0),
	}, nil
}

func (s *Serf) Broadcast(event transport.Event) error {
	payload, err := event.PayloadMarshall()
	if err != nil {
		return err
	}

	return s.client.UserEvent(event.Name(), payload, false)
}

func (s *Serf) WaitFor(event transport.Event) chan transport.Status {
	s.mu.Lock()
	s.registerSubscriber(event.Name())
	s.initStream()
	s.mu.Unlock()

	go func() {
		err := event.PayloadUnmarshall(<-s.payloadCh[event.Name()])
		s.eventCh[event.Name()] <- transport.Status{Error: err}
	}()

	return s.eventCh[event.Name()]
}

func (s *Serf) Close() error {
	s.doneCh <- true
	return s.client.Close()
}

func (s *Serf) registerSubscriber(eventName string) {
	if _, ok := s.payloadCh[eventName]; !ok {
		s.payloadCh[eventName] = make(chan []byte, 0)
		s.eventCh[eventName] = make(chan transport.Status, 0)
	}
}

func (s *Serf) initStream() {
	if s.streamHandle != 0 {
		return
	}

	var err error
	s.streamHandle, err = s.client.Stream("user", s.msgCh)
	if err != nil {
		panic(err.Error())
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
}
