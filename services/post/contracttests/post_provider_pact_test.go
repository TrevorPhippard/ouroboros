package contracttests

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	postpb "ouroboros/proto/generated/post"
	postmodels "post/internal/models"
	postservice "post/internal/service"

	"github.com/pact-foundation/pact-go/v2/models"
	"github.com/pact-foundation/pact-go/v2/provider"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	postProviderName = "post-service"

	providerStateSingle   = "a post with ID post-123 exists"
	providerStateMultiple = "multiple posts exist for IDs x, y, z"
)

func TestPostServiceProviderVerification(t *testing.T) {
	skipIfPactPluginMissing(t)

	pactFile := filepath.Join(repoRoot(t), "services", "api-gateway", "contracttests", "pacts", "graphql-api-gateway-post-service.json")
	if _, err := os.Stat(pactFile); err != nil {
		t.Skipf("consumer pact file %s not found; run the consumer Pact test first", pactFile)
	}

	db := newProviderDB(t)
	server := &postservice.PostServiceServer{DB: db}
	port, shutdown := startPostGRPCServer(t, server)
	defer shutdown()

	verifier := provider.NewVerifier()
	err := verifier.VerifyProvider(t, provider.VerifyRequest{
		Provider:        postProviderName,
		ProviderBaseURL: fmt.Sprintf("http://127.0.0.1:%d", port),
		PactFiles:       []string{pactFile},
		Transports: []provider.Transport{
			{
				Protocol: "grpc",
				Scheme:   "http",
				Port:     uint16(port),
				Path:     "/",
			},
		},
		StateHandlers: models.StateHandlers{
			providerStateSingle: func(setup bool, state models.ProviderState) (models.ProviderStateResponse, error) {
				if !setup {
					return models.ProviderStateResponse{}, nil
				}

				postID := "123e4567-e89b-42d3-a456-426614174000"
				if err := resetPosts(db); err != nil {
					return nil, err
				}
				if err := seedPosts(db, []postmodels.DBPost{
					{
						ID:        postID,
						AuthorID:  "user-1",
						Content:   "Hello from Pact",
						CreatedAt: time.Date(2026, 4, 27, 12, 34, 56, 0, time.UTC),
					},
				}); err != nil {
					return nil, err
				}

				return models.ProviderStateResponse{
					"postId": postID,
				}, nil
			},
			providerStateMultiple: func(setup bool, state models.ProviderState) (models.ProviderStateResponse, error) {
				if !setup {
					return models.ProviderStateResponse{}, nil
				}

				postIDs := []string{
					"223e4567-e89b-42d3-a456-426614174001",
					"323e4567-e89b-42d3-a456-426614174002",
					"423e4567-e89b-42d3-a456-426614174003",
				}
				if err := resetPosts(db); err != nil {
					return nil, err
				}
				if err := seedPosts(db, []postmodels.DBPost{
					{
						ID:        postIDs[0],
						AuthorID:  "user-2",
						Content:   "A post in the feed",
						CreatedAt: time.Date(2026, 4, 27, 15, 4, 5, 0, time.UTC),
					},
					{
						ID:        postIDs[1],
						AuthorID:  "user-3",
						Content:   "Another post in the feed",
						CreatedAt: time.Date(2026, 4, 27, 15, 5, 5, 0, time.UTC),
					},
					{
						ID:        postIDs[2],
						AuthorID:  "user-4",
						Content:   "A third post in the feed",
						CreatedAt: time.Date(2026, 4, 27, 15, 6, 5, 0, time.UTC),
					},
				}); err != nil {
					return nil, err
				}

				return models.ProviderStateResponse{
					"postId1": postIDs[0],
					"postId2": postIDs[1],
					"postId3": postIDs[2],
				}, nil
			},
		},
	})
	if err != nil {
		t.Fatalf("verify provider pact: %v", err)
	}
}

func newProviderDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&postmodels.DBPost{}, &postmodels.DBComment{}); err != nil {
		t.Fatalf("migrate sqlite db: %v", err)
	}
	return db
}

func startPostGRPCServer(t *testing.T, service postpb.PostServiceServer) (int, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen for gRPC server: %v", err)
	}

	srv := grpc.NewServer()
	postpb.RegisterPostServiceServer(srv, service)

	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil {
			t.Logf("gRPC server stopped: %v", serveErr)
		}
	}()

	return lis.Addr().(*net.TCPAddr).Port, func() {
		srv.GracefulStop()
		_ = lis.Close()
	}
}

func resetPosts(db *gorm.DB) error {
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postmodels.DBComment{}).Error; err != nil {
		return fmt.Errorf("clear comments: %w", err)
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&postmodels.DBPost{}).Error; err != nil {
		return fmt.Errorf("clear posts: %w", err)
	}
	return nil
}

func seedPosts(db *gorm.DB, posts []postmodels.DBPost) error {
	if err := db.Create(&posts).Error; err != nil {
		return fmt.Errorf("seed posts: %w", err)
	}
	return nil
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
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
