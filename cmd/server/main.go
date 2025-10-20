package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/acai-travel/tech-challenge/internal/chat"
	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/httpx"
	"github.com/acai-travel/tech-challenge/internal/mongox"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/gorilla/mux"
	"github.com/twitchtv/twirp"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

func main() {
	meterProvider, traceProvider, reqCount, reqDuration, err := initProvider()
	if err != nil {
		slog.Error("failed to initialize OpenTelemetry provider", "error", err)
		return
	}

	defer func() {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down meter provider", "error", err)
		}
		if err := traceProvider.Shutdown(context.Background()); err != nil {
			slog.Error("Error shutting down trace provider", "error", err)
		}
	}()
	mongo := mongox.MustConnect()

	repo := model.New(mongo)
	assist := assistant.New()

	server := chat.NewServer(repo, assist)

	// Configure handler
	handler := mux.NewRouter()
	handler.Use(
		otelmux.Middleware("Clippy-chat"),
		httpx.Logger(),
		httpx.Recovery(),
		MetricsMiddleware(reqCount, reqDuration),
	)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hi, my name is Clippy!")
	})

	handler.PathPrefix("/twirp/").Handler(pb.NewChatServiceServer(server, twirp.WithServerJSONSkipDefaults(true)))

	// Start the server
	slog.Info("Starting the server...")
	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}
}
