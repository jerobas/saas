package application

import (
	"errors"
	"testing"

	"github.com/jerobas/saas/internal/domain"
)

func TestEnsurePostingClockCompatibleAllowsIdempotentReplay(t *testing.T) {
	requested := mustInstant(2_000)

	if err := ensurePostingClockCompatible(mustInstant(1_000), requested); err != nil {
		t.Fatalf("past posted_at should be accepted as idempotent replay: %v", err)
	}
	if err := ensurePostingClockCompatible(requested, requested); err != nil {
		t.Fatalf("matching posted_at should be accepted as fresh post: %v", err)
	}
}

func TestEnsurePostingClockCompatibleRejectsFutureDocument(t *testing.T) {
	err := ensurePostingClockCompatible(mustInstant(3_000), mustInstant(2_000))
	if !errors.Is(err, domain.ErrInvariant) {
		t.Fatalf("future posted_at error = %v, want invariant", err)
	}
}
