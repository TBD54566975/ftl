package pubsub

import "fmt"

// Control messages for the topic.
type control[T any] interface{ control() }

type subscribe[T any] chan T

func (subscribe[T]) control() {}

type unsubscribe[T any] chan T

func (unsubscribe[T]) control() {}

type stop struct{}

func (stop) control() {}

type Topic[T any] struct {
	publish chan T
	control chan control[T]
}

// New creates a new topic that can be used to publish and subscribe to messages.
func New[T any]() *Topic[T] {
	s := &Topic[T]{
		publish: make(chan T, 64),
		control: make(chan control[T]),
	}
	go s.run()
	return s
}

func (s *Topic[T]) Publish(t T) {
	s.publish <- t
}

// Subscribe a channel to the topic.
//
// The channel will be closed when the topic is closed.
func (s *Topic[T]) Subscribe(c chan T) chan T {
	s.control <- subscribe[T](c)
	return c
}

// Unsubscribe a channel from the topic, closing the channel.
func (s *Topic[T]) Unsubscribe(c chan T) {
	s.control <- unsubscribe[T](c)
}

// Close the topic, blocking until all subscribers have been closed.
func (s *Topic[T]) Close() error {
	s.control <- stop{}
	return nil
}

func (s *Topic[T]) run() {
	subscriptions := map[chan T]struct{}{}
	for {
		select {
		case msg := <-s.control:
			switch msg := msg.(type) {
			case subscribe[T]:
				subscriptions[msg] = struct{}{}

			case unsubscribe[T]:
				delete(subscriptions, msg)
				close(msg)

			case stop:
				for ch := range subscriptions {
					close(ch)
				}
				close(s.control)
				close(s.publish)
				return

			default:
				panic(fmt.Sprintf("unknown control message: %T", msg))
			}

		case msg := <-s.publish:
			for ch := range subscriptions {
				ch <- msg
			}
		}
	}
}
