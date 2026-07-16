package application

import "github.com/jerobas/saas/internal/domain"

func ensurePostingClockCompatible(documentPostedAt, requestedPostedAt domain.UTCInstant) error {
	if documentPostedAt.IsZero() || requestedPostedAt.IsZero() {
		return domain.ErrInvariant
	}
	if requestedPostedAt.Before(documentPostedAt) {
		return domain.ErrInvariant
	}
	return nil
}
