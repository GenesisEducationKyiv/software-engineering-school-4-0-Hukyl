package transport

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	durable    = false
	autoDelete = false
	exclusive  = false
	noWait     = false
)

func logAndWrap(msg string, err error) error {
	slog.Error(msg, slog.Any("error", err))
	return fmt.Errorf("%s: %w", msg, err)
}

type Producer struct {
	config  config.Config
	conn    *amqp.Connection
	channel *amqp.Channel
}

func (p *Producer) Close() error {
	slog.Info("closing producer", slog.Any("queue", p.config.QueueName))
	if err := p.channel.Close(); err != nil {
		return logAndWrap("closing channel", err)
	}
	if err := p.conn.Close(); err != nil {
		return logAndWrap("closing connection", err)
	}
	return nil
}

func (p *Producer) Produce(ctx context.Context, msg []byte) error {
	err := p.channel.PublishWithContext(ctx,
		"",                 // exchange
		p.config.QueueName, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(msg),
		})
	slog.Info(
		"publishing message",
		slog.Any("queueName", p.config.QueueName),
		slog.Any("error", err),
	)
	if err != nil {
		return fmt.Errorf("publishing message: %w", err)
	}
	return nil
}

func NewProducer(config config.Config) (*Producer, error) {
	slog.Info("creating producer", slog.Any("config", config))
	conn, err := amqp.Dial(config.BrokerURI)
	if err != nil {
		return nil, logAndWrap("dialing broker", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, logAndWrap("creating channel", err)
	}
	_, err = ch.QueueDeclare(
		config.QueueName, // name
		durable,
		autoDelete,
		exclusive,
		noWait,
		nil, // arguments
	)
	if err != nil {
		return nil, logAndWrap("declaring queue", err)
	}

	return &Producer{
		config:  config,
		conn:    conn,
		channel: ch,
	}, nil
}
