package shared

const ChatService = "ChatService"

type JoinArgs struct {
Username string
}

type JoinReply struct {
Success bool
Message string
}

type LeaveArgs struct {
Username string
}

type LeaveReply struct {
Success bool
}

type SendMessageArgs struct {
Username string
Text     string
}

type SendMessageReply struct {
Success bool
}

type GetUpdatesArgs struct {
Username string
}

type GetUpdatesReply struct {
Messages []string
Users    []string
}
