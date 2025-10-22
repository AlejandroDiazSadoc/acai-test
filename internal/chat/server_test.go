package chat

import (
	"context"
	"strings"
	"testing"

	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	. "github.com/acai-travel/tech-challenge/internal/chat/testing"
	"github.com/acai-travel/tech-challenge/internal/chat/tool"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestServer_DescribeConversation(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), nil)

	t.Run("describe existing conversation", WithFixture(func(t *testing.T, f *Fixture) {
		c := f.CreateConversation()

		out, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: c.ID.Hex()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, want := out.GetConversation(), c.Proto()
		if !cmp.Equal(got, want, protocmp.Transform()) {
			t.Errorf("DescribeConversation() mismatch (-got +want):\n%s", cmp.Diff(got, want, protocmp.Transform()))
		}
	}))

	t.Run("describe non existing conversation should return 404", WithFixture(func(t *testing.T, f *Fixture) {
		_, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: "08a59244257c872c5943e2a2"})
		if err == nil {
			t.Fatal("expected error for non-existing conversation, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.NotFound {
			t.Fatalf("expected twirp.NotFound error, got %v", err)
		}
	}))
}

// Added testing for StartConversation
func TestServer_StartConversation(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), assistant.New(tool.SetupTools()))

	t.Run("Test Start Conversation", WithFixture(func(t *testing.T, f *Fixture) {
		out, err := srv.StartConversation(ctx, &pb.StartConversationRequest{Message: "Hello, I am testing the Start Conversation API call"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(out.GetConversationId()) == "" {
			t.Fatalf("Conversation does not have conversation ID")
		}
		if strings.TrimSpace(out.GetTitle()) == "" {
			t.Fatalf("Didn't populate title")
		}
		if strings.TrimSpace(out.GetReply()) == "" {
			t.Fatalf("Didn't provide assistant answer")
		}

		t.Logf("ID: %s", out.GetConversationId())
		t.Logf("Title: %s", out.GetTitle())
		t.Logf("Assistant: %s", out.GetReply())
	}))
}

// Added benchmark testings for comparison of StartConversation methods perf ↓

func BenchmarkStartConversationLegacy(b *testing.B) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), assistant.New(tool.SetupTools()))

	for i := 0; i < b.N; i++ {
		out, err := srv.StartConversationLegacy(ctx, &pb.StartConversationRequest{Message: "Hello, I am testing the Start Conversation API call"})
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(out.GetConversationId()) == "" {
			b.Fatalf("Conversation does not have conversation ID")
		}
		if strings.TrimSpace(out.GetTitle()) == "" {
			b.Fatalf("Didn't populate title")
		}
		if strings.TrimSpace(out.GetReply()) == "" {
			b.Fatalf("Didn't provide assistant answer")
		}
	}
}

func BenchmarkStartConversation(b *testing.B) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), assistant.New(tool.SetupTools()))

	for i := 0; i < b.N; i++ {
		out, err := srv.StartConversation(ctx, &pb.StartConversationRequest{Message: "Hello, I am testing the Start Conversation API call"})
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
		if strings.TrimSpace(out.GetConversationId()) == "" {
			b.Fatalf("Conversation does not have conversation ID")
		}
		if strings.TrimSpace(out.GetTitle()) == "" {
			b.Fatalf("Didn't populate title")
		}
		if strings.TrimSpace(out.GetReply()) == "" {
			b.Fatalf("Didn't provide assistant answer")
		}
	}
}
