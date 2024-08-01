package transport

import (
	"log/slog"
	"sync"

	"github.com/GenesisEducationKyiv/software-engineering-school-4-0-Hukyl/pkg/broker/transport/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

var logger *slog.Logger

func getLogger() *slog.Logger {
	if logger == nil {
		logger = slog.Default().With(slog.Any("src", "transport"))
	}
	return logger
}

const (
	queueDurable    = false
	queueAutoDelete = false
	queueExclusive  = false
	queueNoWait     = false
)

const (
	consumeAutoAck   = true
	consumeNoLocal   = false
	consumeExclusive = false
	consumeNoWait    = false
)

type Listener func([]byte) error

type Consumer struct {
	config         config.Config
	conn           *amqp.Connection
	channel        *amqp.Channel
	messages       <-chan amqp.Delivery
	listeners      []Listener
	listenerAccess *sync.Mutex
}

func (c *Consumer) Subscribe(f Listener) {
	getLogger().Info(
		"adding subscriber",
		slog.Any("listener", f),
		slog.Any("totalListeners", len(c.listeners)+1),
	)
	c.listenerAccess.Lock()
	defer c.listenerAccess.Unlock()
	c.listeners = append(c.listeners, f)
}

func (c *Consumer) deliverMessage(msg amqp.Delivery) {
	c.listenerAccess.Lock()
	defer c.listenerAccess.Unlock()
	for _, listener := range c.listeners {
		getLogger().Debug("delivering message", slog.Any("listener", listener))
		err := listener(msg.Body)
		if err != nil {
			getLogger().Error(
				"error delivering message",
				slog.Any("error", err),
				slog.Any("listener", listener),
			)
			continue
		}
		getLogger().Debug("message delivered", slog.Any("listener", listener))
	}
}

func (c *Consumer) Listen(stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			getLogger().Info("received stop signal")
			return
		case msg, ok := <-c.messages:
			if !ok {
				getLogger().Error("failed to receive message")
				return
			}
			slog.Info("received message", slog.Any("queue", c.config.QueueName))
			c.deliverMessage(msg)
		}
	}
}

func (c *Consumer) Close() error {
	slog.Info("closing consumer", slog.Any("queue", c.config.QueueName))
	getLogger().Debug("closing channel")
	if err := c.channel.Close(); err != nil {
		return logAndWrap("closing channel", err)
	}
	getLogger().Debug("closing connection")
	if err := c.conn.Close(); err != nil {
		return logAndWrap("closing connection", err)
	}
	return nil
}

func NewConsumer(config config.Config) (*Consumer, error) {
	getLogger().Info("creating consumer", slog.Any("config", config))
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
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil, // arguments
	)
	if err != nil {
		return nil, logAndWrap("declaring queue", err)
	}
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		consumeAutoAck,
		consumeExclusive,
		consumeNoLocal,
		consumeNoWait,
		nil, // args
	)
	if err != nil {
		return nil, logAndWrap("delivery creating", err)
	}

	return &Consumer{
		config:         config,
		conn:           conn,
		channel:        ch,
		messages:       msgs,
		listeners:      make([]Listener, 0),
		listenerAccess: &sync.Mutex{},
	}, nil
}
