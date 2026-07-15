package domain

import "strconv"

type positiveID struct{ value int64 }

func newPositiveID(field string, value int64) (positiveID, error) {
	if value <= 0 {
		return positiveID{}, Invalid(field, ViolationNotPositive, "")
	}
	return positiveID{value: value}, nil
}

func (id positiveID) Int64() int64   { return id.value }
func (id positiveID) IsZero() bool   { return id.value == 0 }
func (id positiveID) String() string { return strconv.FormatInt(id.value, 10) }

type ItemID struct{ positiveID }
type PackagingID struct{ positiveID }
type CounterpartyID struct{ positiveID }
type RecipeID struct{ positiveID }
type RecipeRevisionID struct{ positiveID }
type RecipeComponentID struct{ positiveID }
type StockDocumentID struct{ positiveID }
type StockDocumentLineID struct{ positiveID }
type InventoryLotID struct{ positiveID }
type LotAllocationID struct{ positiveID }

func NewItemID(value int64) (ItemID, error) {
	id, err := newPositiveID("item_id", value)
	return ItemID{id}, err
}
func NewPackagingID(value int64) (PackagingID, error) {
	id, err := newPositiveID("packaging_id", value)
	return PackagingID{id}, err
}
func NewCounterpartyID(value int64) (CounterpartyID, error) {
	id, err := newPositiveID("counterparty_id", value)
	return CounterpartyID{id}, err
}
func NewRecipeID(value int64) (RecipeID, error) {
	id, err := newPositiveID("recipe_id", value)
	return RecipeID{id}, err
}
func NewRecipeRevisionID(value int64) (RecipeRevisionID, error) {
	id, err := newPositiveID("recipe_revision_id", value)
	return RecipeRevisionID{id}, err
}
func NewRecipeComponentID(value int64) (RecipeComponentID, error) {
	id, err := newPositiveID("recipe_component_id", value)
	return RecipeComponentID{id}, err
}
func NewStockDocumentID(value int64) (StockDocumentID, error) {
	id, err := newPositiveID("stock_document_id", value)
	return StockDocumentID{id}, err
}
func NewStockDocumentLineID(value int64) (StockDocumentLineID, error) {
	id, err := newPositiveID("stock_document_line_id", value)
	return StockDocumentLineID{id}, err
}
func NewInventoryLotID(value int64) (InventoryLotID, error) {
	id, err := newPositiveID("inventory_lot_id", value)
	return InventoryLotID{id}, err
}
func NewLotAllocationID(value int64) (LotAllocationID, error) {
	id, err := newPositiveID("lot_allocation_id", value)
	return LotAllocationID{id}, err
}

type PostingSequence struct{ positiveID }
type RevisionNumber struct{ positiveID }
type LineOrder struct{ positiveID }
type ComponentOrder struct{ positiveID }

func NewPostingSequence(value int64) (PostingSequence, error) {
	v, err := newPositiveID("posting_sequence", value)
	return PostingSequence{v}, err
}
func NewRevisionNumber(value int64) (RevisionNumber, error) {
	v, err := newPositiveID("revision_number", value)
	return RevisionNumber{v}, err
}
func NewLineOrder(value int64) (LineOrder, error) {
	v, err := newPositiveID("line_order", value)
	return LineOrder{v}, err
}
func NewComponentOrder(value int64) (ComponentOrder, error) {
	v, err := newPositiveID("component_order", value)
	return ComponentOrder{v}, err
}

type PreparationMinutes struct{ value int64 }

func NewPreparationMinutes(value int64) (PreparationMinutes, error) {
	if value < 0 {
		return PreparationMinutes{}, Invalid("preparation_minutes", ViolationOutOfRange, "")
	}
	return PreparationMinutes{value: value}, nil
}

func (m PreparationMinutes) Int64() int64 { return m.value }

type MinorDigits struct{ value uint8 }

func NewMinorDigits(value int) (MinorDigits, error) {
	if value < 0 || value > 6 {
		return MinorDigits{}, Invalid("currency_minor_digits", ViolationOutOfRange, "INV-001")
	}
	return MinorDigits{value: uint8(value)}, nil
}

func (d MinorDigits) Int() int { return int(d.value) }

type BasisPoints struct{ value int64 }

func NewBasisPoints(value int64) (BasisPoints, error) {
	if value < 0 || value > 9999 {
		return BasisPoints{}, Invalid("basis_points", ViolationOutOfRange, "")
	}
	return BasisPoints{value: value}, nil
}

func (b BasisPoints) Int64() int64 { return b.value }
