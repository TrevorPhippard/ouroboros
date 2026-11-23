package args

import (
	"strings"
	"testing"
)

func TestLimit(t *testing.T) {
	if n, err := Limit(nil, 10); err != nil || n != 10 {
		t.Fatalf("nil limit should default, got %d, %v", n, err)
	}
	for _, bad := range []int32{0, -5, MaxLimit + 1} {
		bad := bad
		if _, err := Limit(&bad, 10); err == nil {
			t.Errorf("limit %d must be rejected", bad)
		}
	}
	ok := int32(MaxLimit)
	if n, err := Limit(&ok, 10); err != nil || n != MaxLimit {
		t.Fatalf("max limit should pass, got %d, %v", n, err)
	}
}

func TestCheckIDCount(t *testing.T) {
	if err := CheckIDCount("x", make([]string, MaxIDs)); err != nil {
		t.Fatal(err)
	}
	if err := CheckIDCount("x", make([]string, MaxIDs+1)); err == nil || !strings.Contains(err.Error(), "x") {
		t.Fatalf("oversized list must be rejected, got %v", err)
	}
}

func TestUniqueIDs(t *testing.T) {
	got := UniqueIDs([]string{" a ", "b", "a", "", "  ", "b", "c"})
	if strings.Join(got, ",") != "a,b,c" {
		t.Fatalf("got %v", got)
	}
}
