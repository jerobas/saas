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
