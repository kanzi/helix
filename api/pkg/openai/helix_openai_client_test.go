package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/helixml/helix/api/pkg/config"
	"github.com/helixml/helix/api/pkg/model"
	"github.com/helixml/helix/api/pkg/pubsub"
	"github.com/helixml/helix/api/pkg/scheduler"
	"github.com/helixml/helix/api/pkg/types"
	openai "github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	gomock "go.uber.org/mock/gomock"
)

const (
	runnerID = "runner1"
)

func TestHelixClientTestSuite(t *testing.T) {
	suite.Run(t, new(HelixClientTestSuite))
}

type HelixClientTestSuite struct {
	ctx context.Context
	suite.Suite
	ctrl   *gomock.Controller
	pubsub pubsub.PubSub

	srv *InternalHelixServer
}

func (suite *HelixClientTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.ctrl = gomock.NewController(suite.T())

	pubsub, err := pubsub.NewInMemoryNats()
	suite.Require().NoError(err)

	suite.pubsub = pubsub

	cfg, _ := config.LoadServerConfig()
	scheduler := scheduler.NewScheduler(suite.ctx, &cfg, nil)
	scheduler.UpdateRunner(&types.RunnerState{
		ID:          runnerID,
		TotalMemory: 9999999999,
	})
	suite.Require().NoError(err)
	suite.srv = NewInternalHelixServer(&cfg, pubsub, scheduler)
}

func (suite *HelixClientTestSuite) Test_CreateChatCompletion_Response() {
	var (
		ownerID       = "owner1"
		sessionID     = "session1"
		interactionID = "interaction1"
	)

	// Fake running will pick up our request and send a response
	go startFakeRunner(suite.T(), suite.srv, []*types.RunnerLLMInferenceResponse{
		{
			OwnerID:       ownerID,
			SessionID:     sessionID,
			InteractionID: interactionID,
			Response: &openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "Hello, world!",
						},
					},
				},
				Usage: openai.Usage{
					PromptTokens:     5,
					CompletionTokens: 12,
					TotalTokens:      17,
				},
			},
		},
	})

	ctx := SetContextValues(suite.ctx, &ContextValues{
		OwnerID:       ownerID,
		SessionID:     sessionID,
		InteractionID: interactionID,
	})

	resp, err := suite.srv.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model.ModelOllamaLlama38b,
		Stream:   false,
		Messages: []openai.ChatCompletionMessage{},
	})
	suite.NoError(err)

	suite.Equal("Hello, world!", resp.Choices[0].Message.Content)
}

func (suite *HelixClientTestSuite) Test_CreateChatCompletion_ErrorResponse() {
	var (
		ownerID       = "owner1"
		sessionID     = "session1"
		interactionID = "interaction1"
	)

	// Fake running will pick up our request and send a response
	go startFakeRunner(suite.T(), suite.srv, []*types.RunnerLLMInferenceResponse{
		{
			OwnerID:       ownerID,
			SessionID:     sessionID,
			InteractionID: interactionID,
			Error:         "too many tokens",
		},
	})

	ctx := SetContextValues(suite.ctx, &ContextValues{
		OwnerID:       ownerID,
		SessionID:     sessionID,
		InteractionID: interactionID,
	})

	_, err := suite.srv.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model.ModelOllamaLlama38b,
		Stream:   false,
		Messages: []openai.ChatCompletionMessage{},
	})
	suite.Error(err)
	suite.Contains(err.Error(), "too many tokens")
}

func (suite *HelixClientTestSuite) Test_CreateChatCompletion_StreamingResponse() {
	var (
		ownerID       = "owner1"
		sessionID     = "session1"
		interactionID = "interaction1"
	)

	// Fake running will pick up our request and send a response
	go startFakeRunner(suite.T(), suite.srv, []*types.RunnerLLMInferenceResponse{
		{
			OwnerID:       ownerID,
			SessionID:     sessionID,
			InteractionID: interactionID,
			StreamResponse: &openai.ChatCompletionStreamResponse{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Content: "One,",
						},
					},
				},
			},
		},
		{
			OwnerID:       ownerID,
			SessionID:     sessionID,
			InteractionID: interactionID,
			StreamResponse: &openai.ChatCompletionStreamResponse{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Content: "Two,",
						},
					},
				},
			},
		},
		{
			OwnerID:       ownerID,
			SessionID:     sessionID,
			InteractionID: interactionID,
			StreamResponse: &openai.ChatCompletionStreamResponse{
				Choices: []openai.ChatCompletionStreamChoice{
					{
						Delta: openai.ChatCompletionStreamChoiceDelta{
							Content: "Three.",
						},
					},
				},
			},
			Done: true,
		},
	})

	ctx := SetContextValues(suite.ctx, &ContextValues{
		OwnerID:       ownerID,
		SessionID:     sessionID,
		InteractionID: interactionID,
	})

	stream, err := suite.srv.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    model.ModelOllamaLlama38b,
		Stream:   true,
		Messages: []openai.ChatCompletionMessage{},
	})
	suite.NoError(err)

	defer stream.Close()

	var resp string

	for {
		response, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		suite.NoError(err)

		resp += response.Choices[0].Delta.Content
	}

	suite.Equal("One,Two,Three.", resp)
}

// startFakeRunner starts polling the queue for requests and sends responses. Exits once context
// is done
func startFakeRunner(t *testing.T, srv *InternalHelixServer, responses []*types.RunnerLLMInferenceResponse) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())

	t.Cleanup(func() {
		cancel()
	})

	for {
		select {
		case <-ctx.Done():
			return
		default:
			req, err := srv.GetNextLLMInferenceRequest(ctx, types.InferenceRequestFilter{}, runnerID)
			require.NoError(t, err)

			if req == nil {
				time.Sleep(200 * time.Millisecond)
				continue
			}

			t.Logf("sending response for request %s (owner: %s | session: %s)", req.RequestID, req.OwnerID, req.SessionID)

			for _, resp := range responses {
				bts, err := json.Marshal(resp)
				require.NoError(t, err)

				err = srv.pubsub.Publish(ctx, pubsub.GetRunnerResponsesQueue(req.OwnerID, req.RequestID), bts)
				require.NoError(t, err)
			}
			fmt.Println("all responses sent")
			t.Log("all responses sent")

		}
	}
}
