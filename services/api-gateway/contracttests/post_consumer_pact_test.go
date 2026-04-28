package contracttests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/pact-foundation/pact-go/v2/consumer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	postpb "ouroboros/proto/generated/post"
)

const (
	postConsumerName = "graphql-api-gateway"
	postProviderName = "post-service"
	protobufPlugin   = "protobuf"

	uuidRegex    = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[1-5][0-9a-fA-F]{3}-[89abAB][0-9a-fA-F]{3}-[0-9a-fA-F]{12}$"
	iso8601Regex = "^(?:[0-9]{4})-(?:0[1-9]|1[0-2])-(?:0[1-9]|[12][0-9]|3[01])T(?:[01][0-9]|2[0-3]):[0-5][0-9]:[0-5][0-9]Z$"
)

func TestPostServiceConsumerPact(t *testing.T) {
	skipIfPactPluginMissing(t)

	pactDir := filepath.Join(contractTestDir(t), "pacts")
	logDir := filepath.Join(contractTestDir(t), "logs")
	if err := os.MkdirAll(pactDir, 0o755); err != nil {
		t.Fatalf("create pact dir: %v", err)
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		t.Fatalf("create log dir: %v", err)
	}

	mockProvider, err := consumer.NewV4Pact(consumer.MockHTTPProviderConfig{
		Consumer: postConsumerName,
		Provider: postProviderName,
		PactDir:  pactDir,
		LogDir:   logDir,
		Host:     "127.0.0.1",
	})
	if err != nil {
		t.Fatalf("create pact: %v", err)
	}

	protoPath := postProtoPath(t)
	getPostPayload := mustJSON(t, grpcServiceInteraction(protoPath, "post.PostService/GetPost", map[string]any{
		"request": map[string]any{
			"id": "matching(equalTo, fromProviderState('${postId}', 'post-123'))",
		},
		"response": postMessageExample(
			"matching(regex, '"+uuidRegex+"', '123e4567-e89b-42d3-a456-426614174000')",
			"notEmpty('user-1')",
			"notEmpty('Hello from Pact')",
			"matching(regex, '"+iso8601Regex+"', '2026-04-27T12:34:56Z')",
		),
	}))
	getPostsByIDsPayload := mustJSON(t, grpcServiceInteraction(protoPath, "post.PostService/GetPostsByIds", map[string]any{
		"request": map[string]any{
			"ids": []any{
				"matching(equalTo, fromProviderState('${postId1}', 'x'))",
				"matching(equalTo, fromProviderState('${postId2}', 'y'))",
				"matching(equalTo, fromProviderState('${postId3}', 'z'))",
			},
		},
		"response": map[string]any{
			"posts": map[string]any{
				"pact:match": "atLeast(1), eachValue(matching($'items'))",
				"items": postMessageExample(
					"matching(regex, '"+uuidRegex+"', '223e4567-e89b-42d3-a456-426614174001')",
					"notEmpty('user-2')",
					"notEmpty('A post in the feed')",
					"matching(regex, '"+iso8601Regex+"', '2026-04-27T15:04:05Z')",
				),
			},
		},
	}))

	// Interaction 1: the gateway asks for one post and expects a Post payload back.
	//
	// The request example is "post-123" as requested, but the `fromProviderState(...)`
	// wrapper lets provider verification swap in a real UUID-backed row later. That keeps
	// the consumer example human-readable without fighting the UUID response matcher.
	mockProvider.
		AddInteraction().
		Given("a post with ID post-123 exists").
		UponReceiving("a GetPost request for post-123").
		UsingPlugin(consumer.PluginConfig{Plugin: protobufPlugin, Version: "latest"}).
		WithRequest(consumer.POST, "/", func(b *consumer.V4InteractionWithPluginRequestBuilder) {
			b.PluginContents("application/grpc", getPostPayload)
		}).
		WillRespondWith(200, func(b *consumer.V4InteractionWithPluginResponseBuilder) {
			b.PluginContents("application/grpc", getPostPayload)
		})

	// Interaction 2: repeated response field with a minimum size of 1, and each post
	// item must match the same shape as Interaction 1.
	mockProvider.
		AddInteraction().
		Given("multiple posts exist for IDs x, y, z").
		UponReceiving("a GetPostsByIds request for a list of post IDs").
		UsingPlugin(consumer.PluginConfig{Plugin: protobufPlugin, Version: "latest"}).
		WithRequest(consumer.POST, "/", func(b *consumer.V4InteractionWithPluginRequestBuilder) {
			b.PluginContents("application/grpc", getPostsByIDsPayload)
		}).
		WillRespondWith(200, func(b *consumer.V4InteractionWithPluginResponseBuilder) {
			b.PluginContents("application/grpc", getPostsByIDsPayload)
		})

	err = mockProvider.ExecuteTest(t, func(config consumer.MockServerConfig) error {
		client, conn, err := newPostClient(config)
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		postRes, err := client.GetPost(ctx, &postpb.GetPostRequest{Id: "post-123"})
		if err != nil {
			return fmt.Errorf("GetPost: %w", err)
		}
		if postRes.GetId() == "" {
			return fmt.Errorf("GetPost returned an empty post.id")
		}
		if postRes.GetAuthorId() == "" {
			return fmt.Errorf("GetPost returned an empty post.author_id")
		}
		if postRes.GetContent() == "" {
			return fmt.Errorf("GetPost returned an empty post.content")
		}
		if postRes.GetCreatedAt() == "" {
			return fmt.Errorf("GetPost returned an empty post.created_at")
		}

		postsRes, err := client.GetPostsByIds(ctx, &postpb.GetPostsByIdsRequest{Ids: []string{"x", "y", "z"}})
		if err != nil {
			return fmt.Errorf("GetPostsByIds: %w", err)
		}
		if len(postsRes.GetPosts()) < 1 {
			return fmt.Errorf("GetPostsByIds returned %d posts, want at least 1", len(postsRes.GetPosts()))
		}
		for i, post := range postsRes.GetPosts() {
			if post.GetId() == "" || post.GetAuthorId() == "" || post.GetContent() == "" || post.GetCreatedAt() == "" {
				return fmt.Errorf("GetPostsByIds returned an incomplete post at index %d: %+v", i, post)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("verify consumer pact: %v", err)
	}
}

func grpcServiceInteraction(protoPath, service string, payload map[string]any) map[string]any {
	payload["pact:proto"] = protoPath
	payload["pact:content-type"] = "application/grpc"
	payload["pact:proto-service"] = service
	return payload
}

func postMessageExample(id, authorID, content, createdAt string) map[string]any {
	return map[string]any{
		"id":         id,
		"author_id":  authorID,
		"content":    content,
		"created_at": createdAt,
	}
}

func newPostClient(config consumer.MockServerConfig) (postpb.PostServiceClient, *grpc.ClientConn, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial pact gRPC mock server %s: %w", addr, err)
	}
	return postpb.NewPostServiceClient(conn), conn, nil
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	body, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal plugin contents: %v", err)
	}
	return string(body)
}

func postProtoPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "proto", "post", "post.proto")
}

func contractTestDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Dir(file)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	return filepath.Clean(filepath.Join(contractTestDir(t), "..", "..", ".."))
}

func skipIfPactPluginMissing(t *testing.T) {
	t.Helper()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("could not resolve home directory for Pact plugin lookup: %v", err)
	}

	matches, err := filepath.Glob(filepath.Join(home, ".pact", "plugins", "protobuf-*"))
	if err != nil || len(matches) == 0 {
		t.Skip("Pact protobuf plugin is not installed under ~/.pact/plugins; install it before running these tests")
	}
}
