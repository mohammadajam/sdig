package server

import "strings"

// ClientRequest is requests by the client to a chat or to  the server manager.
type ClientRequest struct {
	// the type of the request.
	//	server manager related.
	//		"li": "login"
	//		"lo": "logout"
	//		"nu": "new user"
	//		"du": "delete user"
	//		"jo": "join chat"
	//		"le": "leave chat"
	//		"nc": "new chat"
	//		"dc": "delete chat"
	//		"gc": "get connected chats"		unimplemented
	//		"qu": "quit"
	//	chat related.
	//		"nm": "new message"
	//		"dm": "delete message"			unimplemented
	//		"gm": "get chat messages"		unimplemented
	//		"gu": "get connected users"		unimplemented
	string
	content string	// the content of the request.
	sender *User	// a pointer to the user who sent the request.
}

const (
	// A request to log in to an existing user.
	LoginRequestType  string		= "li"
	// A requset to log out from a user.
	LogoutRequestType string		= "lo"
	// A request to create a new user.
	NewUserRequestType string		= "nu"
	// A request to delete an existing user.
	DeleteUserRequestType string	= "du"
	// A request from a user to join an existing chat.
	JoinChatRequestType string   	= "jo"
	// A request from a user to leave an existing chat.
	LeaveChatRequestType string  	= "le"
	// A request from a user to create new chat.
	NewChatRequestType string		= "nc"
	// A request from a user to delete an existing chat.
	DeleteChatRequestType string  	= "dc"
	// A request from a user to get the chats they have joined.
	GetChatsRequestType string 		= "gc"
	// A request from a user to quit
	QuitRequestType string			= "qu"

	// A request to send a new message from a user in a chat to all members in that chat.
	NewMessageRequestType string 	= "nm"
	// A request from a user to delete an existing message in a chat.
	DeleteMessageRequestType string	= "dm"
	// A request from a user to send stored messages to the user.
	GetMessagesRequestType string  	= "gm"
	// A request from a user to send all other users who joined the chat.
	GetUsersRequestType string     	= "gu"
)

func NewClientRequest(request string, data string, user *User) ClientRequest {
	return ClientRequest{
		string: request,
		content: data,
		sender: user,
	}
}

// Creates a client request of the type LoginRequestType("li")
func LoginRequest(username string, password string, user *User) ClientRequest {
	req_content := strings.TrimSpace(username) + " " + strings.TrimSpace(password)
	return NewClientRequest(LoginRequestType, req_content, user)
}

// Creates a client request of the type LogoutRequestType("lo")
func LogoutRequest(user *User) ClientRequest {
	return NewClientRequest(LogoutRequestType, user.username, user)
}

// Creates a client request of the type NewUserRequestType("nu")
func NewUserRequest(username string, name string, password string, user *User) ClientRequest {
	reqContent := strings.Join([]string{username, name, password}, " ")
	return NewClientRequest(NewUserRequestType, reqContent, user)
}

// Creates a client request of the type DeleteUserRequestType("du")
func DeleteUserRequest(password string, user *User) ClientRequest {
	return NewClientRequest(DeleteUserRequestType, password, user)
}

// Creates a client request of the type JoinChatRequestType("jo")
func JoinChatRequest(chatId string, chatPassword string, user *User) ClientRequest {
	reqContent := strings.Join([]string{chatId, chatPassword}, " ")
	return NewClientRequest(JoinChatRequestType, reqContent, user)
}

// Creates a client request of the type LeaveChatRequestType("le")
func LeaveChatRequest(chatId string, user *User) ClientRequest {
	return NewClientRequest(LeaveChatRequestType, chatId, user)
}

// Creates a client request of the type NewChatRequestType("nc")
func NewChatRequest(chatId string, chatName string, chatPassword string, user *User) ClientRequest {
	reqContent := strings.Join([]string{chatId, chatName, chatPassword}, " ")
	return NewClientRequest(NewChatRequestType, reqContent, user)
}

// Creates a client request of the type DeleteChatRequestType("dc")
func DeleteChatRequest(chatId string, chatPassword string, user *User) ClientRequest {
	reqContent := strings.Join([]string{chatId, chatPassword}, " ")
	return NewClientRequest(DeleteChatRequestType, reqContent, user)
}

//Creates a client request of the type GetChatsRequestType("gc")
func GetChatsRequest(user *User) ClientRequest {
	return NewClientRequest(GetChatsRequestType, user.username, user)
}

// Creates a client request of the type QuitRequestType("qu")
func QuitRequest(user *User) ClientRequest {
	return NewClientRequest(QuitRequestType, user.username, user)
}


// Creates a client request of the type NewMessageRequestType("nm")
func NewMessageRequest(content string, user *User) ClientRequest {
	return NewClientRequest(NewMessageRequestType, content, user)
}

// Creates a client request of the type DeleteMessageRequestType("dm")
func DeleteMessageRequest(messageId string, user *User) ClientRequest {
	return NewClientRequest(DeleteMessageRequestType, messageId, user)
}

// Creates a client request of the type GetMessagesRequestType("gm")
func GetMessagesRequest(fromMessageId string, toMessageId string, user *User) ClientRequest {
	return NewClientRequest(GetMessagesRequestType, strings.Join([]string{fromMessageId, toMessageId}, " "), user)
}

// Creates a client request of the type GetUsersRequestType("gu")
func GetUsersRequest(user *User) ClientRequest {
	return NewClientRequest(GetUsersRequestType, user.username, user)
}
