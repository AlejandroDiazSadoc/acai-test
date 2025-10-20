package chat

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/twitchtv/twirp"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ pb.ChatService = (*Server)(nil)

type Assistant interface {
	Title(ctx context.Context, conv *model.Conversation) (string, error)
	Reply(ctx context.Context, conv *model.Conversation) (string, error)
}

type Server struct {
	repo   *model.Repository
	assist Assistant
}

func NewServer(repo *model.Repository, assist Assistant) *Server {
	return &Server{repo: repo, assist: assist}
}

// Deprecated:  Kept for testing performance purposes
func (s *Server) StartConversationLegacy(ctx context.Context, req *pb.StartConversationRequest) (*pb.StartConversationResponse, error) {
	conversation := &model.Conversation{
		ID:        primitive.NewObjectID(),
		Title:     "Untitled conversation",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []*model.Message{{
			ID:        primitive.NewObjectID(),
			Role:      model.RoleUser,
			Content:   req.GetMessage(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}},
	}

	if strings.TrimSpace(req.GetMessage()) == "" {
		return nil, twirp.RequiredArgumentError("message")
	}

	// choose a title
	title, err := s.assist.Title(ctx, conversation)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to generate conversation title", "error", err)
	} else {
		conversation.Title = title
	}

	// generate a reply
	reply, err := s.assist.Reply(ctx, conversation)
	if err != nil {
		return nil, err
	}

	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleAssistant,
		Content:   reply,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	return &pb.StartConversationResponse{
		ConversationId: conversation.ID.Hex(),
		Title:          conversation.Title,
		Reply:          reply,
	}, nil
}

func (s *Server) StartConversation(ctx context.Context, req *pb.StartConversationRequest) (*pb.StartConversationResponse, error) {
	conversation := &model.Conversation{
		ID:        primitive.NewObjectID(),
		Title:     "Untitled conversation",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []*model.Message{{
			ID:        primitive.NewObjectID(),
			Role:      model.RoleUser,
			Content:   req.GetMessage(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}},
	}

	if strings.TrimSpace(req.GetMessage()) == "" {
		return nil, twirp.RequiredArgumentError("message")
	}

	// choose a title
	titleDone := make(chan struct{})
	titleChan := make(chan string, 1)
	titleErrChan := make(chan error, 1)
	go func() {
		defer close(titleDone)

		title, err := s.assist.Title(ctx, conversation)
		select {
		case <-titleDone: // Check if the main thread already finished/timed out
			return
		default:
			// Proceed with sending the result
			if err != nil {
				titleErrChan <- err
			} else {
				titleChan <- title
			}
		}
	}()

	// generate a reply
	reply, err := s.assist.Reply(ctx, conversation)
	if err != nil {
		return nil, err
	}

	defer func() {
		select {
		case <-titleDone: // Already finished
		default:
			close(titleDone)
		}
	}()

	select {
	case title := <-titleChan:
		conversation.Title = title
	case err := <-titleErrChan:
		slog.ErrorContext(ctx, "Failed to generate conversation title (Non-blocking)", "error", err)
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(3 * time.Second):
		// Timeout to wait, set as 3 after testing that sometimes was not waiting enough
		slog.WarnContext(ctx, "Title generation took too long, proceeding with default title.")
	}

	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleAssistant,
		Content:   reply,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	return &pb.StartConversationResponse{
		ConversationId: conversation.ID.Hex(),
		Title:          conversation.Title,
		Reply:          reply,
	}, nil
}

func (s *Server) ContinueConversation(ctx context.Context, req *pb.ContinueConversationRequest) (*pb.ContinueConversationResponse, error) {
	if req.GetConversationId() == "" {
		return nil, twirp.RequiredArgumentError("conversation_id")
	}

	if strings.TrimSpace(req.GetMessage()) == "" {
		return nil, twirp.RequiredArgumentError("message")
	}

	conversation, err := s.repo.DescribeConversation(ctx, req.GetConversationId())
	if err != nil {
		return nil, err
	}

	conversation.UpdatedAt = time.Now()
	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleUser,
		Content:   req.GetMessage(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	reply, err := s.assist.Reply(ctx, conversation)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleAssistant,
		Content:   reply,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err := s.repo.UpdateConversation(ctx, conversation); err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &pb.ContinueConversationResponse{Reply: reply}, nil
}

func (s *Server) ListConversations(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
	conversations, err := s.repo.ListConversations(ctx)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	resp := &pb.ListConversationsResponse{}
	for _, conv := range conversations {
		conv.Messages = nil // Clear messages to avoid sending large data
		resp.Conversations = append(resp.Conversations, conv.Proto())
	}

	return resp, nil
}

func (s *Server) DescribeConversation(ctx context.Context, req *pb.DescribeConversationRequest) (*pb.DescribeConversationResponse, error) {
	if req.GetConversationId() == "" {
		return nil, twirp.RequiredArgumentError("conversation_id")
	}

	conversation, err := s.repo.DescribeConversation(ctx, req.GetConversationId())
	if err != nil {
		return nil, err
	}

	if conversation == nil {
		return nil, twirp.NotFoundError("conversation not found")
	}

	return &pb.DescribeConversationResponse{Conversation: conversation.Proto()}, nil
}
