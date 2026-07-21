package catalog

import "github.com/jerobas/saas/internal/domain"

type ItemSummaryParams struct {
	ID               domain.ItemID
	Name             domain.UniqueName
	SKU              domain.Option[domain.SKU]
	Description      domain.Option[domain.NonEmptyText]
	BaseUnit         domain.UnitCode
	Capabilities     Capabilities
	DefaultSalePrice domain.Option[domain.MinorAmount]
	ReorderQuantity  domain.Option[domain.AtomicQuantity]
	CreatedAt        domain.UTCInstant
	UpdatedAt        domain.UTCInstant
	ArchivedAt       domain.Option[domain.UTCInstant]
}

// ItemSummary is the ListItems projection. It deliberately has no packaging
// accessor, so callers cannot confuse a list row with a loaded aggregate.
type ItemSummary struct{ item Item }

func NewItemSummary(params ItemSummaryParams) (ItemSummary, error) {
	item, err := NewItem(ItemParams{
		ID: params.ID, Name: params.Name, SKU: params.SKU,
		Description: params.Description, BaseUnit: params.BaseUnit,
		Capabilities: params.Capabilities, DefaultSalePrice: params.DefaultSalePrice,
		ReorderQuantity: params.ReorderQuantity, CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt, ArchivedAt: params.ArchivedAt,
	})
	if err != nil {
		return ItemSummary{}, err
	}
	return ItemSummary{item: item}, nil
}

func (s ItemSummary) ID() domain.ItemID                               { return s.item.ID() }
func (s ItemSummary) Name() domain.UniqueName                         { return s.item.Name() }
func (s ItemSummary) SKU() domain.Option[domain.SKU]                  { return s.item.SKU() }
func (s ItemSummary) Description() domain.Option[domain.NonEmptyText] { return s.item.Description() }
func (s ItemSummary) BaseUnit() domain.UnitCode                       { return s.item.BaseUnit() }
func (s ItemSummary) Capabilities() Capabilities                      { return s.item.Capabilities() }
func (s ItemSummary) DefaultSalePrice() domain.Option[domain.MinorAmount] {
	return s.item.DefaultSalePrice()
}
func (s ItemSummary) ReorderQuantity() domain.Option[domain.AtomicQuantity] {
	return s.item.ReorderQuantity()
}
func (s ItemSummary) CreatedAt() domain.UTCInstant                 { return s.item.CreatedAt() }
func (s ItemSummary) UpdatedAt() domain.UTCInstant                 { return s.item.UpdatedAt() }
func (s ItemSummary) ArchivedAt() domain.Option[domain.UTCInstant] { return s.item.ArchivedAt() }
func (s ItemSummary) IsArchived() bool                             { return s.item.IsArchived() }
