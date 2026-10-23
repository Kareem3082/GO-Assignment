package main

import (
"bufio"
"fmt"
"log"
"net/rpc"
"os"
"os/signal"
"strings"
"syscall"
"time"
"chat-rpc/shared"
)

func main() {
client, err := rpc.Dial("tcp", "localhost:1234")
if err != nil {
log.Fatalf("Failed to connect to chat server: %v", err)
}
defer client.Close()

reader := bufio.NewReader(os.Stdin)
var username string

for {
fmt.Print("Enter your unique username: ")
username, _ = reader.ReadString('\n')
username = strings.TrimSpace(username)
if username == "" {
fmt.Println("Username cannot be empty.")
continue
}

args := &shared.JoinArgs{Username: username}
var reply shared.JoinReply
err = client.Call(shared.ChatService+".Join", args, &reply)
if err != nil {
log.Fatalf("RPC error during join: %v", err)
}

if reply.Success {
fmt.Println(reply.Message)
break
} else {
fmt.Println("Error:", reply.Message)
}
}

sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
go func() {
<-sigChan
fmt.Println("\nDisconnecting...")
leaveArgs := &shared.LeaveArgs{Username: username}
var leaveReply shared.LeaveReply
_ = client.Call(shared.ChatService+".Leave", leaveArgs, &leaveReply)
os.Exit(0)
}()

go func() {
for {
time.Sleep(300 * time.Millisecond)
args := &shared.GetUpdatesArgs{Username: username}
var reply shared.GetUpdatesReply
err := client.Call(shared.ChatService+".GetUpdates", args, &reply)
if err != nil {
continue
}

for _, msg := range reply.Messages {
fmt.Printf("\r%s\n> ", msg)
}
}
}()

fmt.Println("\n--- Chat Room Active ---")
fmt.Println("Type your message and press Enter.")
fmt.Println("Commands: /users (list active users), /quit (leave chat)")
fmt.Println("------------------------")

for {
fmt.Print("> ")
text, _ := reader.ReadString('\n')
text = strings.TrimSpace(text)
if text == "" {
continue
}

if text == "/quit" {
leaveArgs := &shared.LeaveArgs{Username: username}
var leaveReply shared.LeaveReply
_ = client.Call(shared.ChatService+".Leave", leaveArgs, &leaveReply)
fmt.Println("You have left the chat.")
break
}

if text == "/users" {
args := &shared.GetUpdatesArgs{Username: username}
var reply shared.GetUpdatesReply
_ = client.Call(shared.ChatService+".GetUpdates", args, &reply)
fmt.Printf("Connected Users: %s\n", strings.Join(reply.Users, ", "))
continue
}

sendArgs := &shared.SendMessageArgs{
Username: username,
Text:     text,
}
var sendReply shared.SendMessageReply
err = client.Call(shared.ChatService+".SendMessage", sendArgs, &sendReply)
if err != nil {
fmt.Println("Failed to send message:", err)
}
}
}
