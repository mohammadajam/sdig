package server

import "strings"

// ClientRequest is requests by the client to a chat or to  the chat manager.
type ClientRequest struct {
	// the type of the request.
	//	chat manager related.
	//		"lo": "login"
	//		"nu": "new user"				unimplemented
	//		"du": "delete user"				unimplemented
	//		"jo": "join chat"				unimplemented
	//		"le": "leave chat"				unimplemented
	//		"nc": "new chat"				unimplemented
	//		"dc": "delete chat"				unimplemented
	//		"gc": "get connected chats"
	//	chat related.
	//		"nm": "new message"
	//		"dm": "delete message"			unimplemented
	//		"gm": "get chat messages"		unimplemented
	//		"gu": "get connected users"		unimplemented
	string
	content string	// the content of the request.
	sender *User	// a pointer to the user who sent the request.
}


func NewClientRequest(request string, data string, user *User) ClientRequest {
	return ClientRequest{
		string: request,
		content: data,
		sender: user,
	}
}

func NewLoginRequest(username string, password string, user *User) ClientRequest {
	req_content := strings.TrimSpace(username) + " " + strings.TrimSpace(password)
	return NewClientRequest("lo", req_content, user)
}

// TODO: add functions for all request types specified in the ClientRequest description.
