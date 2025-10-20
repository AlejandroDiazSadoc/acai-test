package assistant

import (
	"context"
	"testing"

	. "github.com/acai-travel/tech-challenge/internal/chat/testing"
)

// Added testing for tittle method
func TestServer_AssistantTittle(t *testing.T) {
	ctx := context.Background()

	t.Run("test assistant title method", WithFixture(func(t *testing.T, f *Fixture) {
		c := f.CreateConversation()
		assist := New()

		out, err := assist.Title(ctx, c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if out == "Checking the Weather Today" {
			t.Fatalf("Output title does not match expected value")
		}

		t.Logf("Title: %s", out)

	}))

}
