package serf

import (
	"github.com/hashicorp/serf/client"

	"github.com/makerdao/gofer/pkg/transport"
)

const messageQueueSize = 1024

type Serf struct {
	client       *client.RPCClient
	streamHandle client.StreamHandle
	subscribers  map[string]chan []byte
	payloads     map[string]chan transport.Payload
	ch           chan map[string]interface{}
}

func NewSerf(rpc string) (*Serf, error) {
	conf := client.Config{Addr: rpc}

	serfClient, err := client.ClientFromConfig(&conf)
	if err != nil {
		return nil, err
	}

	return &Serf{
		client:      serfClient,
		subscribers: make(map[string]chan []byte, 0),
		payloads:    make(map[string]chan transport.Payload, 0),
		ch:          make(chan map[string]interface{}, messageQueueSize),
	}, nil
}

func (s *Serf) Broadcast(event *transport.Event) error {
	payload, err := event.Payload.PayloadMarshall()
	if err != nil {
		return err
	}

	return s.client.UserEvent(event.Name, payload, false)
}

func (s *Serf) WaitFor(eventName string, payload transport.Payload) chan transport.Payload {
	s.registerSubscriber(eventName)
	s.initStream()

	go func() {
		// TODO: log errors
		_ = payload.PayloadUnmarshall(<-s.subscribers[eventName])
		s.payloads[eventName] <- payload
	}()

	return s.payloads[eventName]
}

func (s *Serf) registerSubscriber(eventName string) {
	if _, ok := s.subscribers[eventName]; !ok {
		s.subscribers[eventName] = make(chan []byte, 0)
		s.payloads[eventName] = make(chan transport.Payload, 0)
	}
}

func (s *Serf) initStream() {
	if s.streamHandle != 0 {
		return
	}

	var err error
	s.streamHandle, err = s.client.Stream("user", s.ch)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		for {
			msg := <-s.ch
			if subscriber, ok := s.subscribers[msg["Name"].(string)]; ok {
				subscriber <- msg["Payload"].([]byte)
			}
		}
	}()
}
