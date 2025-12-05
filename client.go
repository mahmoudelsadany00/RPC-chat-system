package main

import (
	"bufio"
	"fmt"
	"net/rpc"
	"os"
	"strings"
	"time"
)

type JoinArgs struct{ ID string }
type JoinReply struct{ OK bool }

type SendArgs struct {
	ID   string
	Text string
}
type SendReply struct{ OK bool }

type Message struct {
	From   string
	Text   string
	Time   time.Time
	System bool
}

type PollArgs struct {
	ID        string
	TimeoutMs int
}
type PollReply struct {
	Messages []Message
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your ID: ")
	idRaw, _ := reader.ReadString('\n')
	id := strings.TrimSpace(idRaw)
	if id == "" {
		id = "anonymous"
	}

	client, err := rpc.Dial("tcp", "127.0.0.1:1234")
	if err != nil {
		fmt.Println("dial error:", err)
		return
	}
	defer client.Close()

	// Join
	var jReply JoinReply
	if err := client.Call("Chat.Join", JoinArgs{ID: id}, &jReply); err != nil {
		fmt.Println("join error:", err)
		return
	}
	fmt.Println("Joined. Type messages. Type 'exit' to quit.")
	fmt.Println("Note: no self-echo (your messages won't come back to you).")

	// Receiver goroutine (long-poll loop)
	go func() {
		for {
			var pReply PollReply
			err := client.Call("Chat.Poll", PollArgs{ID: id, TimeoutMs: 25000}, &pReply)
			if err != nil {
				fmt.Println("\nPoll error:", err)
				os.Exit(0)
			}
			for _, m := range pReply.Messages {
				if m.System {
					fmt.Printf("\n%s\n> ", m.Text)
				} else {
					fmt.Printf("\n[%s] %s: %s\n> ",
						m.Time.Format("15:04:05"), m.From, m.Text)
				}
			}
		}
	}()

	// Send loop
	for {
		fmt.Print("> ")
		line, _ := reader.ReadString('\n')
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if text == "exit" || text == "/exit" {
			fmt.Println("bye!")
			return
		}

		var sReply SendReply
		if err := client.Call("Chat.Send", SendArgs{ID: id, Text: text}, &sReply); err != nil {
			fmt.Println("send error:", err)
			continue
		}
	}
}
