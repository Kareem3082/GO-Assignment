package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"sync"
	"chat-rpc/shared"
)

type ServerState struct {
	mu       sync.Mutex
	users    map[string][]string
	userList []string
}

type ChatServer struct {
	state *ServerState
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		state: &ServerState{
			users:    make(map[string][]string),
			userList: make([]string, 0),
		},
	}
}

func (s *ChatServer) Join(args *shared.JoinArgs, reply *shared.JoinReply) error {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if _, exists := s.state.users[args.Username]; exists || args.Username == "" {
		reply.Success = false
		reply.Message = "Username is already taken or invalid."
		return nil
	}

	s.state.users[args.Username] = []string{}
	s.state.userList = append(s.state.userList, args.Username)

	notif := fmt.Sprintf("*** [%s joined the chat] ***", args.Username)
	for user, msgs := range s.state.users {
		if user != args.Username {
			s.state.users[user] = append(msgs, notif)
		}
	}

	reply.Success = true
	reply.Message = "Welcome to the chat!"
	log.Printf("User joined: %s", args.Username)
	return nil
}

func (s *ChatServer) Leave(args *shared.LeaveArgs, reply *shared.LeaveReply) error {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if _, exists := s.state.users[args.Username]; !exists {
		reply.Success = false
		return nil
	}

	delete(s.state.users, args.Username)
	for i, u := range s.state.userList {
		if u == args.Username {
			s.state.userList = append(s.state.userList[:i], s.state.userList[i+1:]...)
			break
		}
	}

	notif := fmt.Sprintf("*** [%s left the chat] ***", args.Username)
	for user, msgs := range s.state.users {
		s.state.users[user] = append(msgs, notif)
	}

	reply.Success = true
	log.Printf("User left: %s", args.Username)
	return nil
}

func (s *ChatServer) SendMessage(args *shared.SendMessageArgs, reply *shared.SendMessageReply) error {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	if _, exists := s.state.users[args.Username]; !exists {
		reply.Success = false
		return nil
	}

	formattedMsg := fmt.Sprintf("[%s]: %s", args.Username, args.Text)

	for user, msgs := range s.state.users {
		if user != args.Username {
			s.state.users[user] = append(msgs, formattedMsg)
		}
	}

	reply.Success = true
	return nil
}

func (s *ChatServer) GetUpdates(args *shared.GetUpdatesArgs, reply *shared.GetUpdatesReply) error {
	s.state.mu.Lock()
	defer s.state.mu.Unlock()

	msgs, exists := s.state.users[args.Username]
	if !exists {
		return nil
	}

	reply.Messages = msgs
	s.state.users[args.Username] = []string{}
	reply.Users = append([]string{}, s.state.userList...)
	return nil
}

func main() {
	chatServer := NewChatServer()
	rpc.RegisterName("ChatService", chatServer)

	listener, err := net.Listen("tcp", ":1234")
	if err != nil {
		log.Fatal("Listener error:", err)
	}
	defer listener.Close()

	fmt.Println("Chat Server is running on port 1234...")
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection accept error: %v", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}