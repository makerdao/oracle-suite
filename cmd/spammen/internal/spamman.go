//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/makerdao/oracle-suite/pkg/transport/messages"

	"github.com/makerdao/oracle-suite/pkg/spire/config"
)

type Spamman struct {
	MessageRate      int
	ValidMessages    bool
	InvalidSignature bool
	instance         *config.Instances
	mg               *MessageGenerator
}

type Config struct {
	MessageRate      int
	ValidMessages    bool
	InvalidSignature bool
	Pairs            []string
	Instance         *config.Instances
}

func NewSpamman(cfg Config) *Spamman {
	return &Spamman{
		MessageRate:      cfg.MessageRate,
		ValidMessages:    cfg.ValidMessages,
		InvalidSignature: cfg.InvalidSignature,
		instance:         cfg.Instance,
		mg:               NewMessageGenerator(cfg.Pairs),
	}
}

func (s *Spamman) Start() error {
	s.instance.Logger.Info("Starting spamman")

	if s.MessageRate <= 0 {
		return fmt.Errorf("invalid message rate")
	}
	// Have to subscribe to topic before we will ba able to send something to it.
	return s.instance.Transport.Subscribe(messages.PriceMessageName, (*messages.Price)(nil))
}

func (s *Spamman) Stop() {
	s.instance.Logger.Info("Stopping spamman")

	err := s.instance.Transport.Close()
	if err != nil {
		s.instance.Logger.Errorf("Failed to stop transport layer: %w", err)
	}
}

func (s *Spamman) Run(ctx context.Context) error {
	go s.process(ctx)
	return nil
}

func (s *Spamman) process(ctx context.Context) {
	// message rate per minute in ms
	delay := int(60_000 / s.MessageRate)

	s.instance.Logger.
		WithField("delay", delay).
		Debug("Next message will be sent")

	select {
	case <-time.After(time.Duration(delay) * time.Millisecond):
		// TODO: sending msg in sync mode. will cause delay
		s.generateAndSendMessage()
		// scheduling next msg
		s.process(ctx)
	case <-ctx.Done():
		return
	}
}

func (s *Spamman) generateAndSendMessage() {
	msg, err := s.generateMessage()
	if err != nil {
		s.instance.Logger.Errorf("failed to generate message: %w", err)
	}

	err = s.sendMessage(msg)
	if err != nil {
		s.instance.Logger.Errorf("failed to send message: %w", err)
	}
}

func (s *Spamman) generateMessage() (*messages.Price, error) {
	if s.ValidMessages {
		msg := s.mg.ValidPriceMessage()
		// Signing message before sending it
		err := msg.Price.Sign(s.instance.Signer)
		if err != nil {
			return nil, err
		}
		return msg, nil
	}

	if s.InvalidSignature {
		return s.mg.ValidPriceMessage(), nil
	}

	return nil, fmt.Errorf("no massage type to generate")
}

func (s *Spamman) sendMessage(message *messages.Price) error {
	s.instance.Logger.
		WithField("wat", message.Price.Wat).
		WithField("val", message.Price.Val).
		WithField("age", message.Price.Age).
		Debug("Sending new valid message")

	err := s.instance.Transport.Broadcast(messages.PriceMessageName, message)
	if err != nil {
		return err
	}
	return nil
}
