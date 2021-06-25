package stdio

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/makerdao/oracle-suite/pkg/log"
	"github.com/makerdao/oracle-suite/pkg/transport"
	"github.com/makerdao/oracle-suite/pkg/transport/local"
)

type Stdio struct {
	ctx       context.Context
	ctxCancel context.CancelFunc

	dataCh chan []byte
	writer io.Writer
	local  *local.Local
}

type wrappedMessage struct {
	Topic   string            `json:"topic"`
	Message transport.Message `json:"message"`
}

func NewStdio(ctx context.Context, r io.Reader, w io.Writer, buffer int, logger log.Logger) *Stdio {
	ctx, ctxCancel := context.WithCancel(ctx)
	s := &Stdio{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		writer:    w,
		local:     local.New(buffer),
	}

	go func() {
		br := bufio.NewReader(r)
		for {
			line, err := br.ReadBytes('\n')
			if err != nil || ctx.Err() != nil {
				return
			}
			wrapped := &wrappedMessage{}
			err = json.Unmarshal(line, wrapped)
			if err != nil {
				logger.WithError(err).Error("Unable to parse message")
				continue
			}
			err = s.local.Broadcast(wrapped.Topic, wrapped.Message)
			if err != nil {
				logger.WithError(err).Error("Unable to parse message")
				continue
			}
		}
	}()

	return s
}

func (s *Stdio) Subscribe(topic string, typ transport.Message) error {
	return s.local.Subscribe(topic, typ)
}

func (s *Stdio) Unsubscribe(topic string) error {
	return s.local.Unsubscribe(topic)
}

func (s *Stdio) Broadcast(topic string, message transport.Message) error {
	j, err := json.Marshal(&wrappedMessage{
		Topic:   topic,
		Message: message,
	})
	if err != nil {
		return err
	}
	_, err = s.writer.Write(append(j, '\n'))
	if err != nil {
		return err
	}
	return s.local.Broadcast(topic, message)
}

func (s *Stdio) WaitFor(topic string) chan transport.ReceivedMessage {
	return s.local.WaitFor(topic)
}

func (s *Stdio) Close() error {
	s.ctxCancel()
	return s.local.Close()
}
