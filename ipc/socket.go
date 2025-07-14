package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
)

type MessageType string

const (
	MessageTypeRepoAdd      MessageType = "repo_add"
	MessageTypeRepoList     MessageType = "repo_list"
	MessageTypeRepoRemove   MessageType = "repo_remove"
	MessageTypeClaim        MessageType = "claim"
	MessageTypeRelease      MessageType = "release"
	MessageTypePoolStatus   MessageType = "pool_status"
	MessageTypeDaemonStatus MessageType = "daemon_status"
	MessageTypeWorktreeList MessageType = "worktree_list"
	MessageTypeRefresh      MessageType = "refresh"
	MessageTypeShow         MessageType = "show"
	MessageTypeResponse     MessageType = "response"
	MessageTypeError        MessageType = "error"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type RepoAddRequest struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	MaxWorktrees  int    `json:"max_worktrees"`
	DefaultBranch string `json:"default_branch"`
}

type ClaimRequest struct {
	RepoName string `json:"repo_name"`
	Branch   string `json:"branch"`
}

type ClaimResponse struct {
	WorktreeID string `json:"worktree_id"`
	Path       string `json:"path"`
}

type ReleaseRequest struct {
	WorktreeID string `json:"worktree_id"`
}

type PoolStatusRequest struct {
	RepoName string `json:"repo_name,omitempty"`
}

type RefreshRequest struct {
	RepoName string `json:"repo_name"`
}

type ShowRequest struct {
	WorktreeID string `json:"worktree_id"`
}

type Server struct {
	socketPath string
	listener   net.Listener
	handler    Handler
}

type Handler interface {
	HandleRepoAdd(req RepoAddRequest) Response
	HandleRepoList() Response
	HandleRepoRemove(name string) Response
	HandleClaim(req ClaimRequest) Response
	HandleRelease(req ReleaseRequest) Response
	HandlePoolStatus(req PoolStatusRequest) Response
	HandleDaemonStatus() Response
	HandleWorktreeList() Response
	HandleRefresh(req RefreshRequest) Response
	HandleShow(req ShowRequest) Response
}

func NewServer(socketPath string, handler Handler) (*Server, error) {
	// Remove existing socket if exists
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create unix socket: %w", err)
	}

	// Set socket permissions
	if err := os.Chmod(socketPath, 0600); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	return &Server{
		socketPath: socketPath,
		listener:   listener,
		handler:    handler,
	}, nil
}

func (s *Server) Serve() error {
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return fmt.Errorf("failed to accept connection: %w", err)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	var msg Message
	if err := decoder.Decode(&msg); err != nil {
		encoder.Encode(Response{Success: false, Error: "invalid message format"})
		return
	}

	var response Response

	switch msg.Type {
	case MessageTypeRepoAdd:
		var req RepoAddRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandleRepoAdd(req)
		}

	case MessageTypeRepoList:
		response = s.handler.HandleRepoList()

	case MessageTypeRepoRemove:
		var name string
		if err := json.Unmarshal(msg.Data, &name); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandleRepoRemove(name)
		}

	case MessageTypeClaim:
		var req ClaimRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandleClaim(req)
		}

	case MessageTypeRelease:
		var req ReleaseRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandleRelease(req)
		}

	case MessageTypePoolStatus:
		var req PoolStatusRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandlePoolStatus(req)
		}

	case MessageTypeDaemonStatus:
		response = s.handler.HandleDaemonStatus()

	case MessageTypeWorktreeList:
		response = s.handler.HandleWorktreeList()

	case MessageTypeRefresh:
		var req RefreshRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandleRefresh(req)
		}

	case MessageTypeShow:
		var req ShowRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			response = Response{Success: false, Error: "invalid request data"}
		} else {
			response = s.handler.HandleShow(req)
		}

	default:
		response = Response{Success: false, Error: "unknown message type"}
	}

	encoder.Encode(response)
}

type Client struct {
	socketPath string
}

func NewClient(socketPath string) *Client {
	return &Client{socketPath: socketPath}
}

func (c *Client) SendMessage(msg Message) (*Response, error) {
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	if err := encoder.Encode(msg); err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	var response Response
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &response, nil
}

func (c *Client) RepoAdd(req RepoAddRequest) (*Response, error) {
	data, _ := json.Marshal(req)
	return c.SendMessage(Message{Type: MessageTypeRepoAdd, Data: data})
}

func (c *Client) RepoList() (*Response, error) {
	return c.SendMessage(Message{Type: MessageTypeRepoList})
}

func (c *Client) RepoRemove(name string) (*Response, error) {
	data, _ := json.Marshal(name)
	return c.SendMessage(Message{Type: MessageTypeRepoRemove, Data: data})
}

func (c *Client) Claim(req ClaimRequest) (*Response, error) {
	data, _ := json.Marshal(req)
	return c.SendMessage(Message{Type: MessageTypeClaim, Data: data})
}

func (c *Client) Release(req ReleaseRequest) (*Response, error) {
	data, _ := json.Marshal(req)
	return c.SendMessage(Message{Type: MessageTypeRelease, Data: data})
}

func (c *Client) PoolStatus(req PoolStatusRequest) (*Response, error) {
	data, _ := json.Marshal(req)
	return c.SendMessage(Message{Type: MessageTypePoolStatus, Data: data})
}

func (c *Client) DaemonStatus() (*Response, error) {
	return c.SendMessage(Message{Type: MessageTypeDaemonStatus})
}

func (c *Client) WorktreeList() (*Response, error) {
	return c.SendMessage(Message{Type: MessageTypeWorktreeList})
}

func (c *Client) Refresh(req RefreshRequest) (*Response, error) {
	data, _ := json.Marshal(req)
	return c.SendMessage(Message{Type: MessageTypeRefresh, Data: data})
}

func (c *Client) Show(req ShowRequest) (*Response, error) {
	data, _ := json.Marshal(req)
	return c.SendMessage(Message{Type: MessageTypeShow, Data: data})
}
