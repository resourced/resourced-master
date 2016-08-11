package pubsub

import (
	"fmt"
	"time"

	"github.com/go-mangos/mangos"
	"github.com/go-mangos/mangos/protocol/pub"
	"github.com/go-mangos/mangos/protocol/sub"
	"github.com/go-mangos/mangos/transport/ipc"
	"github.com/go-mangos/mangos/transport/tcp"
)

// NewPubSocket creates a socket for publishing messages.
func NewPubSocket(url string) (mangos.Socket, error) {
	sock, err := pub.NewSocket()
	if err != nil {
		return nil, err
	}

	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())

	err = sock.Listen(url)
	if err != nil {
		return nil, err
	}

	return sock, nil
}

// NewSubSocket creates a socket for subscribing messages.
func NewSubSocket(url string) (mangos.Socket, error) {
	sock, err := sub.NewSocket()
	if err != nil {
		return nil, err
	}
	sock.AddTransport(ipc.NewTransport())
	sock.AddTransport(tcp.NewTransport())
	err = sock.Dial(url)
	if err != nil {
		return nil, err
	}
	return sock, nil
}

func NewPubSub(mode, url string) (*PubSub, error) {
	p := &PubSub{}
	p.Mode = mode

	if mode == "pub" {
		sock, err := NewPubSocket(url)
		if err != nil {
			return nil, err
		}
		p.Socket = sock

	} else if mode == "sub" {
		sock, err := NewSubSocket(url)
		if err != nil {
			return nil, err
		}
		p.Socket = sock
	}

	return p, nil
}

type PubSub struct {
	Mode   string
	Socket mangos.Socket
}

func (p *PubSub) Publish(topic, message string) error {
	if p.Mode != "pub" {
		return fmt.Errorf("Publish method cannot be called if Mode != pub")
	}

	payload := fmt.Sprintf("topic:%s|type:plain|created:%v|content:%s", topic, time.Now().UTC().Unix(), message)
	println("i am sending, correct? " + payload)

	return p.Socket.Send([]byte(payload))
}

func (p *PubSub) Subscribe(topic string) error {
	if p.Mode != "sub" {
		return fmt.Errorf("Subscribe method cannot be called if Mode != sub")
	}
	return p.Socket.SetOption(mangos.OptionSubscribe, []byte(fmt.Sprintf("topic:%v|", topic)))
}
