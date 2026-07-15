package application

import (
	"time"

	"github.com/jerobas/saas/internal/domain"
)

type Clock interface {
	Now() (domain.UTCInstant, error)
}

type SystemClock struct{}

func (SystemClock) Now() (domain.UTCInstant, error) {
	return domain.NewUTCInstant(time.Now().UTC())
}

func nextMutationInstant(clock Clock, expected domain.UTCInstant) (domain.UTCInstant, error) {
	now, err := clock.Now()
	if err != nil {
		return domain.UTCInstant{}, err
	}
	if !expected.IsZero() && now.Compare(expected) <= 0 {
		return domain.NewUTCInstant(expected.Time().Add(time.Millisecond))
	}
	return now, nil
}
