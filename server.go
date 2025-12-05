package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"
)

type Message struct {
	From   string
	Text   string
	Time   time.Time
	System bool
}

// ===== RPC Args/Replies =====

type JoinArgs struct {
	ID string
}
type JoinReply struct {
	OK bool
}

type SendArgs struct {
	ID   string
	Text string
}
type SendReply struct {
	OK bool
}

type PollArgs struct {
	ID string
	// Long-poll timeout on the server side
	TimeoutMs int
}
type PollReply struct {
	Messages []Message // usually 1 message; could be >1 if buffered
}

// ===== Chat Service =====

type ClientState struct {
	id string
	ch chan Message // delivery channel for this client
}

type Chat struct {
	mu      sync.Mutex
	clients map[string]*ClientState

	// Central event stream for broadcast (optional but clean)
	events chan Message
}

func NewChat() *Chat {
	c := &Chat{
		clients: make(map[string]*ClientState),
		events:  make(chan Message, 128),
	}
	go c.runBroadcaster()
	return c
}

func (c *Chat) runBroadcaster() {
	for msg := range c.events {
		c.mu.Lock()
		for id, st := range c.clients {
			if id == msg.From && !msg.System {
				// no self-echo for normal messages
				continue
			}
			// non-blocking send حتى عميل بطيء ما يهنّجش السيرفر
			select {
			case st.ch <- msg:
			default:
				// لو القناة مليانة، بنسقط الرسالة (اختياري)
			}
		}
		c.mu.Unlock()

		// log على السيرفر
		if msg.System {
			fmt.Printf("[SERVER] %s\n", msg.Text)
		} else {
			fmt.Printf("[SERVER %s] %s: %s\n", msg.Time.Format("15:04:05"), msg.From, msg.Text)
		}
	}
}

func (c *Chat) Join(args JoinArgs, reply *JoinReply) error {
	id := args.ID
	if id == "" {
		return errors.New("empty id")
	}

	c.mu.Lock()
	if _, exists := c.clients[id]; exists {
		c.mu.Unlock()
		return errors.New("id already in use")
	}

	c.clients[id] = &ClientState{
		id: id,
		ch: make(chan Message, 32),
	}
	c.mu.Unlock()

	reply.OK = true

	// notify others (system join message)
	c.events <- Message{
		From:   id,
		Text:   fmt.Sprintf("User [%s] joined", id),
		Time:   time.Now(),
		System: true,
	}
	return nil
}

func (c *Chat) Send(args SendArgs, reply *SendReply) error {
	if args.Text == "" {
		return errors.New("empty message")
	}

	c.mu.Lock()
	_, ok := c.clients[args.ID]
	c.mu.Unlock()
	if !ok {
		return errors.New("not joined")
	}

	reply.OK = true
	c.events <- Message{
		From:   args.ID,
		Text:   args.Text,
		Time:   time.Now(),
		System: false,
	}
	return nil
}

// Poll: long-poll receive (real-time-ish over RPC)
func (c *Chat) Poll(args PollArgs, reply *PollReply) error {
	if args.ID == "" {
		return errors.New("empty id")
	}
	if args.TimeoutMs <= 0 {
		args.TimeoutMs = 25000 // default 25s
	}

	c.mu.Lock()
	st, ok := c.clients[args.ID]
	c.mu.Unlock()
	if !ok {
		return errors.New("not joined")
	}

	// انتظر رسالة أو timeout
	timeout := time.NewTimer(time.Duration(args.TimeoutMs) * time.Millisecond)
	defer timeout.Stop()

	select {
	case msg, ok := <-st.ch:
		if !ok {
			return errors.New("client channel closed")
		}
		reply.Messages = append(reply.Messages, msg)

		// لو فيه رسائل متخزنة زيادة، خد كام واحدة بسرعة (اختياري)
		for i := 0; i < 9; i++ { // max 10 messages per poll
			select {
			case m2 := <-st.ch:
				reply.Messages = append(reply.Messages, m2)
			default:
				return nil
			}
		}
		return nil

	case <-timeout.C:
		// لا رسائل جديدة
		reply.Messages = nil
		return nil
	}
}

func main() {
	chat := NewChat()
	if err := rpc.Register(chat); err != nil {
		panic(err)
	}

	l, err := net.Listen("tcp", ":1234")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	fmt.Println("RPC broadcast chat server listening on :1234")
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}
