package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   amqp091.Queue
}

func NewRabbitMQClient(url string) (*RabbitMQClient, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("connection couldnt be estb to rmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("couldnt open channel: %w", err)
	}
	q, err := ch.QueueDeclare("user_events", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("coudnt declare the queue: %w", err)
	}
	return &RabbitMQClient{conn: conn, channel: ch, queue: q}, nil
}

// here i am publishing with context bcz currently the amqp lib doesn't impleemt syscall lv; interrrupts but has some preflight optimization in-place. it is so bcz amqp091, a fork of streadway/amqp, inherited the signature from its parent fut not yet fully implemeted it. In casein future, if it ever gets implemented, the code will already be ready with the nneded tools and can be updated hassle free
func (r *RabbitMQClient) Publish(ctx context.Context, body []byte) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return r.channel.PublishWithContext(ctx, "", r.queue.Name, false, false, amqp091.Publishing{ContentType: "application/body", Body: body})

}

func (r *RabbitMQClient) Consume() (<-chan amqp091.Delivery, error) {
	return r.channel.Consume(r.queue.Name, "", false, false, false, false, nil)
}

func (r *RabbitMQClient) Close() {
	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
}
