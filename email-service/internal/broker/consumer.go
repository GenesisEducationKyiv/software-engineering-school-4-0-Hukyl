package broker

import (
	"fmt"
	"log/slog"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/email-service/internal/broker/config"
)

func logAndWrap(msg string, err error) error {
	slog.Error(msg, slog.Any("error", err))
	return fmt.Errorf("%s: %w", msg, err)
}

var listenerAccessLock sync.Mutex = sync.Mutex{}

type Listener func([]byte) error

type Consumer struct {
	config    config.Config
	conn      *amqp.Connection
	channel   *amqp.Channel
	messages  <-chan amqp.Delivery
	listeners []Listener
}

func (c *Consumer) Subscribe(f Listener) {
	listenerAccessLock.Lock()
	defer listenerAccessLock.Unlock()
	c.listeners = append(c.listeners, f)
}

func (c *Consumer) Listen(stop <-chan struct{}) {
	deliverMessage := func(msg amqp.Delivery) {
		listenerAccessLock.Lock()
		defer listenerAccessLock.Unlock()
		for _, listener := range c.listeners {
			listener(msg.Body)
		}
	}

	for {
		select {
		case <-stop:
			return
		default:
			for msg := range c.messages {
				deliverMessage(msg)
			}
		}
	}
}

func (c *Consumer) Close() error {
	if err := c.channel.Close(); err != nil {
		return logAndWrap("closing channel", err)
	}
	if err := c.conn.Close(); err != nil {
		return logAndWrap("closing connection", err)
	}
	return nil
}

func NewConsumer(config config.Config) (*Consumer, error) {
	conn, err := amqp.Dial(config.BrokerURI)
	if err != nil {
		return nil, logAndWrap("dialing amqp", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, logAndWrap("getting channel", err)
	}
	q, err := ch.QueueDeclare(
		config.QueueName, // name
		false,            // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return nil, logAndWrap("declaring queue", err)
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil, logAndWrap("delivery creating", err)
	}

	return &Consumer{
		config:    config,
		conn:      conn,
		channel:   ch,
		messages:  msgs,
		listeners: make([]Listener, 0),
	}, nil
}
