package domain

import (
	"strings"
	"time"
	_ "time/tzdata"

	"golang.org/x/text/language"
)

type BusinessDate struct {
	year  int
	month time.Month
	day   int
}

func NewBusinessDate(year int, month time.Month, day int) (BusinessDate, error) {
	value := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if value.Year() != year || value.Month() != month || value.Day() != day || year < 1 || year > 9999 {
		return BusinessDate{}, Invalid("business_date", ViolationInvalidFormat, "SET-004")
	}
	return BusinessDate{year: year, month: month, day: day}, nil
}

func ParseBusinessDate(raw string) (BusinessDate, error) {
	if len(raw) != len("2006-01-02") {
		return BusinessDate{}, Invalid("business_date", ViolationInvalidFormat, "SET-004")
	}
	value, err := time.Parse("2006-01-02", raw)
	if err != nil || value.Format("2006-01-02") != raw {
		return BusinessDate{}, Invalid("business_date", ViolationInvalidFormat, "SET-004")
	}
	return NewBusinessDate(value.Year(), value.Month(), value.Day())
}

func (d BusinessDate) String() string {
	if d.year == 0 {
		return ""
	}
	return time.Date(d.year, d.month, d.day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
}
func (d BusinessDate) IsZero() bool                   { return d.year == 0 }
func (d BusinessDate) Equal(other BusinessDate) bool  { return d.Compare(other) == 0 }
func (d BusinessDate) Before(other BusinessDate) bool { return d.Compare(other) < 0 }
func (d BusinessDate) After(other BusinessDate) bool  { return d.Compare(other) > 0 }
func (d BusinessDate) Compare(other BusinessDate) int {
	left, right := d.String(), other.String()
	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

type UTCInstant struct{ value time.Time }

func NewUTCInstant(value time.Time) (UTCInstant, error) {
	if value.IsZero() || value.UnixMilli() < 0 {
		return UTCInstant{}, Invalid("instant", ViolationOutOfRange, "SET-004")
	}
	return UTCInstant{value: time.UnixMilli(value.UnixMilli()).UTC()}, nil
}

func UTCInstantFromUnixMilli(value int64) (UTCInstant, error) {
	if value < 0 {
		return UTCInstant{}, Invalid("instant_ms", ViolationOutOfRange, "SET-004")
	}
	return UTCInstant{value: time.UnixMilli(value).UTC()}, nil
}

func (i UTCInstant) IsZero() bool                 { return i.value.IsZero() }
func (i UTCInstant) Time() time.Time              { return i.value }
func (i UTCInstant) UnixMilli() int64             { return i.value.UnixMilli() }
func (i UTCInstant) Before(other UTCInstant) bool { return i.value.Before(other.value) }
func (i UTCInstant) Equal(other UTCInstant) bool  { return i.value.Equal(other.value) }
func (i UTCInstant) Compare(other UTCInstant) int { return i.value.Compare(other.value) }

type BusinessTimezone struct {
	name     string
	location *time.Location
}

func NewBusinessTimezone(raw string) (BusinessTimezone, error) {
	name := strings.TrimSpace(raw)
	if name == "" || name == "Local" {
		return BusinessTimezone{}, Invalid("timezone_name", ViolationInvalidFormat, "SET-002")
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return BusinessTimezone{}, Invalid("timezone_name", ViolationInvalidFormat, "SET-002")
	}
	return BusinessTimezone{name: name, location: location}, nil
}

func (z BusinessTimezone) Name() string             { return z.name }
func (z BusinessTimezone) Location() *time.Location { return z.location }
func (z BusinessTimezone) IsZero() bool             { return z.location == nil }

type Locale struct{ tag language.Tag }

func NewLocale(raw string) (Locale, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return Locale{}, Invalid("locale_code", ViolationRequired, "")
	}
	tag, err := language.Parse(value)
	if err != nil || tag == language.Und {
		return Locale{}, Invalid("locale_code", ViolationInvalidFormat, "")
	}
	return Locale{tag: tag}, nil
}

func (l Locale) String() string    { return l.tag.String() }
func (l Locale) Tag() language.Tag { return l.tag }
func (l Locale) IsZero() bool      { return l.tag == language.Und }

func ValidateTimestampOrder(createdAt, updatedAt UTCInstant, archivedAt Option[UTCInstant]) error {
	violations := make([]Violation, 0, 3)
	if createdAt.IsZero() {
		violations = append(violations, Violation{Field: "created_at", Code: ViolationRequired})
	}
	if updatedAt.IsZero() {
		violations = append(violations, Violation{Field: "updated_at", Code: ViolationRequired})
	} else if !createdAt.IsZero() && updatedAt.Before(createdAt) {
		violations = append(violations, Violation{Field: "updated_at", Code: ViolationInvariant, InvariantID: "SET-004"})
	}
	if archived, ok := archivedAt.Get(); ok {
		// Archiving is itself the master-data mutation represented by
		// updated_at. Keeping both instants identical makes that value a complete
		// optimistic version and prevents a restore from moving backward behind
		// a later archived_at timestamp.
		if archived.IsZero() || (!updatedAt.IsZero() && !archived.Equal(updatedAt)) {
			violations = append(violations, Violation{Field: "archived_at", Code: ViolationInvariant, InvariantID: "SET-004"})
		}
	}
	return NewValidationError(violations...)
}
