package notifications

import (
	"context"
	"strings"
	"testing"
	"time"

	"api-gateway/internal/auth"
	"api-gateway/internal/testutil"

	notifpb "ouroboros/proto/generated/notification"
)

func TestSubscribeEmitsOnlyNewNotificationsOldestFirstAndClosesOnCancel(t *testing.T) {
	all := []*notifpb.Notification{
		testutil.Notif("3", "2026-01-01T00:00:03Z"),
		testutil.Notif("2", "2026-01-01T00:00:02Z"),
		testutil.Notif("1", "2026-01-01T00:00:01Z"),
	}
	svc := Service{
		PollInterval: 10 * time.Millisecond,
		Client: &testutil.Notification{Pages: [][]*notifpb.Notification{
			{testutil.Notif("1", "2026-01-01T00:00:01Z")}, // priming poll: must NOT be replayed
			all, all,
		}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := svc.Subscribe(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}

	var got []string
	for len(got) < 2 {
		select {
		case ev := <-ch:
			got = append(got, ev.ID)
		case <-time.After(2 * time.Second):
			t.Fatalf("timed out, got %v", got)
		}
	}
	if strings.Join(got, ",") != "2,3" {
		t.Fatalf("want 2,3 (oldest first, no replay of 1), got %v", got)
	}

	select {
	case ev := <-ch:
		t.Fatalf("unexpected duplicate %s", ev.ID)
	case <-time.After(60 * time.Millisecond):
	}
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("channel should be closed after cancel")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("channel not closed after cancel")
	}
}

func TestSubscribeEndsWhenTokenExpires(t *testing.T) {
	svc := Service{
		PollInterval: 10 * time.Millisecond,
		Client:       &testutil.Notification{Pages: [][]*notifpb.Notification{{}}},
	}
	ctx := auth.WithExpiry(context.Background(), time.Now().Add(40*time.Millisecond))
	ch, err := svc.Subscribe(ctx, "alice")
	if err != nil {
		t.Fatal(err)
	}
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("unexpected event")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("subscription outlived its token")
	}
}
