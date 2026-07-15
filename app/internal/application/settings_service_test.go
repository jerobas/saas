package application

import (
	"context"
	"testing"
	"time"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
	"github.com/jerobas/saas/internal/infrastructure/sqlite"
)

func TestSettingsServiceReadsAndUpdatesSettings(t *testing.T) {
	db, err := database.NewDatabaseWithOptions(":memory:", database.OpenOptions{BusyTimeout: time.Second})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	store := sqlite.NewStore(db)
	service := NewSettingsService(NewSQLiteSettingsStore(store))

	current, err := service.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if current.BusinessName().String() == "" {
		t.Fatalf("expected seeded settings to be present")
	}

	currency, err := domain.NewCurrency("EUR")
	if err != nil {
		t.Fatalf("new currency: %v", err)
	}
	locale, err := domain.NewLocale("en")
	if err != nil {
		t.Fatalf("new locale: %v", err)
	}
	timezone, err := domain.NewBusinessTimezone("Europe/London")
	if err != nil {
		t.Fatalf("new timezone: %v", err)
	}
	businessName, err := domain.NewDisplayName("Acme")
	if err != nil {
		t.Fatalf("new business name: %v", err)
	}
	updatedAt, err := domain.NewUTCInstant(time.Unix(1_700_000_200, 0).UTC())
	if err != nil {
		t.Fatalf("new updated timestamp: %v", err)
	}

	updated, err := service.UpdateSettings(context.Background(), SettingsUpdateInput{
		BusinessName:       businessName,
		Locale:             locale,
		Timezone:           timezone,
		Currency:           currency,
		HourlyLaborCost:    domain.None[domain.MinorAmount](),
		DefaultGrossMargin: domain.None[domain.BasisPoints](),
		ExpectedUpdatedAt:  current.UpdatedAt(),
		UpdatedAt:          updatedAt,
	})
	if err != nil {
		t.Fatalf("update settings: %v", err)
	}
	if updated.BusinessName().String() != "Acme" {
		t.Fatalf("updated business name = %q", updated.BusinessName().String())
	}
	if !updated.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("updated timestamp = %v, want %v", updated.UpdatedAt(), updatedAt)
	}
}
