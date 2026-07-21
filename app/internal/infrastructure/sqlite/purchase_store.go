package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jerobas/saas/database"
	"github.com/jerobas/saas/internal/domain"
)

const (
	purchaseDefaultPageSize = 50
	purchaseMaximumPageSize = 100
)

type PurchaseCursor struct {
	PostingSequence domain.PostingSequence
	ID              domain.StockDocumentID
}

type PurchaseListFilter struct {
	After    domain.Option[PurchaseCursor]
	PageSize int
}

type PurchasePage struct {
	items []PostedPurchaseDocument
	next  domain.Option[PurchaseCursor]
}

func (p PurchasePage) Items() []PostedPurchaseDocument {
	items := make([]PostedPurchaseDocument, len(p.items))
	copy(items, p.items)
	return items
}

func (p PurchasePage) Next() domain.Option[PurchaseCursor] { return p.next }

type PostPurchaseInput struct {
	IdempotencyKey domain.IdempotencyKey
	CounterpartyID domain.Option[domain.CounterpartyID]
	OccurredOn     domain.BusinessDate
	PostedAt       domain.UTCInstant
	Reason         domain.Option[domain.DocumentReason]
	Notes          domain.Option[domain.NonEmptyText]
	Lines          []PostPurchaseLineInput
}

type PostPurchaseLineInput struct {
	ItemID               domain.ItemID
	Quantity             domain.AtomicQuantity
	EnteredUnit          domain.UnitCode
	EnteredPackagingName domain.Option[domain.NonEmptyText]
	Conversion           domain.UnitConversion
	CommercialTotal      domain.MinorAmount
	LotCode              domain.Option[domain.NonEmptyText]
	ExpiresOn            domain.Option[domain.BusinessDate]
}

type PostedPurchaseDocument struct {
	id              domain.StockDocumentID
	idempotencyKey  domain.IdempotencyKey
	postingSequence domain.PostingSequence
	counterpartyID  domain.Option[domain.CounterpartyID]
	occurredOn      domain.BusinessDate
	postedAt        domain.UTCInstant
	currency        domain.Currency
	reason          domain.Option[domain.DocumentReason]
	notes           domain.Option[domain.NonEmptyText]
	lines           []PostedPurchaseLine
}

func NewPostedPurchaseDocument(
	id domain.StockDocumentID,
	idempotencyKey domain.IdempotencyKey,
	postingSequence domain.PostingSequence,
	counterpartyID domain.Option[domain.CounterpartyID],
	occurredOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	currency domain.Currency,
	reason domain.Option[domain.DocumentReason],
	notes domain.Option[domain.NonEmptyText],
	lines []PostedPurchaseLine,
) PostedPurchaseDocument {
	cloned := make([]PostedPurchaseLine, len(lines))
	copy(cloned, lines)
	return PostedPurchaseDocument{
		id: id, idempotencyKey: idempotencyKey, postingSequence: postingSequence,
		counterpartyID: counterpartyID, occurredOn: occurredOn, postedAt: postedAt,
		currency: currency, reason: reason, notes: notes, lines: cloned,
	}
}

func (d PostedPurchaseDocument) ID() domain.StockDocumentID              { return d.id }
func (d PostedPurchaseDocument) IdempotencyKey() domain.IdempotencyKey   { return d.idempotencyKey }
func (d PostedPurchaseDocument) PostingSequence() domain.PostingSequence { return d.postingSequence }
func (d PostedPurchaseDocument) CounterpartyID() domain.Option[domain.CounterpartyID] {
	return d.counterpartyID
}
func (d PostedPurchaseDocument) OccurredOn() domain.BusinessDate              { return d.occurredOn }
func (d PostedPurchaseDocument) PostedAt() domain.UTCInstant                  { return d.postedAt }
func (d PostedPurchaseDocument) Currency() domain.Currency                    { return d.currency }
func (d PostedPurchaseDocument) Reason() domain.Option[domain.DocumentReason] { return d.reason }
func (d PostedPurchaseDocument) Notes() domain.Option[domain.NonEmptyText]    { return d.notes }
func (d PostedPurchaseDocument) Lines() []PostedPurchaseLine {
	lines := make([]PostedPurchaseLine, len(d.lines))
	copy(lines, d.lines)
	return lines
}

type PostedPurchaseLine struct {
	id                   domain.StockDocumentLineID
	lineOrder            domain.LineOrder
	itemID               domain.ItemID
	quantity             domain.AtomicQuantity
	enteredUnit          domain.UnitCode
	enteredPackagingName domain.Option[domain.NonEmptyText]
	conversion           domain.UnitConversion
	inventoryValue       domain.InventoryValue
	commercialTotal      domain.MinorAmount
	lotID                domain.InventoryLotID
	lotCode              domain.Option[domain.NonEmptyText]
	originatedOn         domain.BusinessDate
	expiresOn            domain.Option[domain.BusinessDate]
}

func NewPostedPurchaseLine(
	id domain.StockDocumentLineID,
	lineOrder domain.LineOrder,
	itemID domain.ItemID,
	quantity domain.AtomicQuantity,
	enteredUnit domain.UnitCode,
	enteredPackagingName domain.Option[domain.NonEmptyText],
	conversion domain.UnitConversion,
	inventoryValue domain.InventoryValue,
	commercialTotal domain.MinorAmount,
	lotID domain.InventoryLotID,
	lotCode domain.Option[domain.NonEmptyText],
	originatedOn domain.BusinessDate,
	expiresOn domain.Option[domain.BusinessDate],
) PostedPurchaseLine {
	return PostedPurchaseLine{
		id: id, lineOrder: lineOrder, itemID: itemID, quantity: quantity,
		enteredUnit: enteredUnit, enteredPackagingName: enteredPackagingName,
		conversion: conversion, inventoryValue: inventoryValue, commercialTotal: commercialTotal,
		lotID: lotID, lotCode: lotCode, originatedOn: originatedOn, expiresOn: expiresOn,
	}
}

func (l PostedPurchaseLine) ID() domain.StockDocumentLineID  { return l.id }
func (l PostedPurchaseLine) LineOrder() domain.LineOrder     { return l.lineOrder }
func (l PostedPurchaseLine) ItemID() domain.ItemID           { return l.itemID }
func (l PostedPurchaseLine) Quantity() domain.AtomicQuantity { return l.quantity }
func (l PostedPurchaseLine) EnteredUnit() domain.UnitCode    { return l.enteredUnit }
func (l PostedPurchaseLine) EnteredPackagingName() domain.Option[domain.NonEmptyText] {
	return l.enteredPackagingName
}
func (l PostedPurchaseLine) Conversion() domain.UnitConversion           { return l.conversion }
func (l PostedPurchaseLine) InventoryValue() domain.InventoryValue       { return l.inventoryValue }
func (l PostedPurchaseLine) CommercialTotal() domain.MinorAmount         { return l.commercialTotal }
func (l PostedPurchaseLine) LotID() domain.InventoryLotID                { return l.lotID }
func (l PostedPurchaseLine) LotCode() domain.Option[domain.NonEmptyText] { return l.lotCode }
func (l PostedPurchaseLine) OriginatedOn() domain.BusinessDate           { return l.originatedOn }
func (l PostedPurchaseLine) ExpiresOn() domain.Option[domain.BusinessDate] {
	return l.expiresOn
}

func (s *Store) GetPostedPurchase(ctx context.Context, id domain.StockDocumentID) (PostedPurchaseDocument, error) {
	if id.IsZero() {
		return PostedPurchaseDocument{}, domain.Invalid("document_id", domain.ViolationRequired, "DOC-001")
	}
	var document PostedPurchaseDocument
	err := s.database.Read(ctx, func(tx *database.ReadTx) error {
		value, err := loadPostedPurchaseDocument(ctx, tx, id.Int64())
		if err != nil {
			return err
		}
		document = value
		return nil
	})
	if err != nil {
		return PostedPurchaseDocument{}, classifyError("get posted purchase", err)
	}
	return document, nil
}

func (s *Store) ListPostedPurchases(ctx context.Context, filter PurchaseListFilter) (PurchasePage, error) {
	pageSize, err := purchasePageSize(filter.PageSize)
	if err != nil {
		return PurchasePage{}, err
	}
	var page PurchasePage
	err = s.database.Read(ctx, func(tx *database.ReadTx) error {
		documentIDs, err := listPostedPurchaseIDs(ctx, tx, filter.After, pageSize+1)
		if err != nil {
			return err
		}
		hasMore := len(documentIDs) > pageSize
		if hasMore {
			documentIDs = documentIDs[:pageSize]
		}
		items := make([]PostedPurchaseDocument, 0, len(documentIDs))
		for _, id := range documentIDs {
			document, err := loadPostedPurchaseDocument(ctx, tx, id)
			if err != nil {
				return err
			}
			items = append(items, document)
		}
		next := domain.None[PurchaseCursor]()
		if hasMore && len(items) > 0 {
			last := items[len(items)-1]
			next = domain.Some(PurchaseCursor{
				PostingSequence: last.PostingSequence(),
				ID:              last.ID(),
			})
		}
		page = PurchasePage{items: items, next: next}
		return nil
	})
	if err != nil {
		return PurchasePage{}, classifyError("list posted purchases", err)
	}
	return page, nil
}

func (s *Store) PostPurchase(ctx context.Context, input PostPurchaseInput) (PostedPurchaseDocument, error) {
	var posted PostedPurchaseDocument
	err := s.database.Write(ctx, func(tx *database.WriteTx) error {
		value, err := postPurchaseTx(ctx, tx, input)
		if err != nil {
			return err
		}
		posted = value
		return nil
	})
	if err != nil {
		return PostedPurchaseDocument{}, classifyError("post purchase", err)
	}
	return posted, nil
}

func purchasePageSize(requested int) (int, error) {
	if requested == 0 {
		return purchaseDefaultPageSize, nil
	}
	if requested < 1 || requested > purchaseMaximumPageSize {
		return 0, domain.Invalid("page_size", domain.ViolationOutOfRange, "")
	}
	return requested, nil
}

func listPostedPurchaseIDs(
	ctx context.Context,
	tx databaseWriteTx,
	after domain.Option[PurchaseCursor],
	limit int,
) ([]int64, error) {
	args := []any{limit}
	query := `
		SELECT id
		FROM stock_documents
		WHERE kind = 'PURCHASE'
		ORDER BY posting_sequence DESC, id DESC
		LIMIT ?
	`
	if cursor, ok := after.Get(); ok {
		if cursor.PostingSequence.IsZero() || cursor.ID.IsZero() {
			return nil, domain.Invalid("cursor", domain.ViolationInvalidFormat, "")
		}
		args = []any{cursor.PostingSequence.Int64(), cursor.PostingSequence.Int64(), cursor.ID.Int64(), limit}
		query = `
			SELECT id
			FROM stock_documents
			WHERE kind = 'PURCHASE'
			  AND (posting_sequence < ? OR (posting_sequence = ? AND id < ?))
			ORDER BY posting_sequence DESC, id DESC
			LIMIT ?
		`
	}
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

type databaseWriteTx interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func postPurchaseTx(ctx context.Context, tx databaseWriteTx, input PostPurchaseInput) (PostedPurchaseDocument, error) {
	if input.IdempotencyKey.String() == "" {
		return PostedPurchaseDocument{}, domain.Invalid("idempotency_key", domain.ViolationRequired, "DOC-003")
	}
	if input.OccurredOn.IsZero() {
		return PostedPurchaseDocument{}, domain.Invalid("occurred_on", domain.ViolationRequired, "DOC-004")
	}
	if input.PostedAt.IsZero() {
		return PostedPurchaseDocument{}, domain.Invalid("posted_at", domain.ViolationRequired, "DOC-004")
	}
	if len(input.Lines) == 0 {
		return PostedPurchaseDocument{}, domain.Invalid("lines", domain.ViolationRequired, "DOC-002")
	}

	var existingID int64
	var existingKind string
	err := tx.QueryRowContext(ctx, `
		SELECT id, kind FROM stock_documents WHERE idempotency_key = ?
	`, input.IdempotencyKey.String()).Scan(&existingID, &existingKind)
	if err == nil {
		if existingKind != domain.DocumentPurchase.String() {
			return PostedPurchaseDocument{}, fmt.Errorf("%w: idempotency key belongs to %s", domain.ErrConflict, existingKind)
		}
		return loadPostedPurchaseDocument(ctx, tx, existingID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return PostedPurchaseDocument{}, err
	}

	currency, err := loadDocumentCurrency(ctx, tx)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}

	var postingSequence int64
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(posting_sequence), 0) + 1 FROM stock_documents
	`).Scan(&postingSequence); err != nil {
		return PostedPurchaseDocument{}, err
	}

	counterpartyID := nullableCounterpartyID(input.CounterpartyID)
	reason := nullableDocumentReason(input.Reason)
	notes := nullableText(input.Notes)

	var documentID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_documents (
			kind, idempotency_key, posting_sequence, counterparty_id, occurred_on,
			posted_at_ms, currency_code, currency_minor_digits, reason_code, notes
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		domain.DocumentPurchase.String(),
		input.IdempotencyKey.String(),
		postingSequence,
		counterpartyID,
		input.OccurredOn.String(),
		input.PostedAt.UnixMilli(),
		currency.Code().String(),
		int64(currency.MinorDigits().Int()),
		reason,
		notes,
	).Scan(&documentID); err != nil {
		return PostedPurchaseDocument{}, err
	}

	for index, line := range input.Lines {
		if err := insertPostedPurchaseLine(ctx, tx, documentID, int64(index+1), currency, input.OccurredOn, input.PostedAt, line); err != nil {
			return PostedPurchaseDocument{}, fmt.Errorf("line %d: %w", index+1, err)
		}
	}

	return loadPostedPurchaseDocument(ctx, tx, documentID)
}

func insertPostedPurchaseLine(
	ctx context.Context,
	tx databaseWriteTx,
	documentID int64,
	lineOrder int64,
	currency domain.Currency,
	originatedOn domain.BusinessDate,
	postedAt domain.UTCInstant,
	line PostPurchaseLineInput,
) error {
	if line.ItemID.IsZero() {
		return domain.Invalid("item_id", domain.ViolationRequired, "DOC-005")
	}
	if line.Quantity.Int64() <= 0 {
		return domain.Invalid("quantity_atomic", domain.ViolationNotPositive, "DOC-005")
	}
	if line.EnteredUnit.String() == "" {
		return domain.Invalid("entered_unit_code", domain.ViolationRequired, "DOC-005")
	}
	if line.Conversion.IsZero() {
		return domain.Invalid("conversion", domain.ViolationRequired, "DOC-005")
	}
	inventoryValue, err := line.CommercialTotal.ToInventoryValue(currency)
	if err != nil {
		return err
	}

	var lineID int64
	if err := tx.QueryRowContext(ctx, `
		INSERT INTO stock_document_lines (
			document_id, line_order, item_id, direction, quantity_atomic,
			entered_unit_code, entered_packaging_name, conversion_numerator_atomic,
			conversion_denominator, inventory_value_micro, commercial_total_minor
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`,
		documentID,
		lineOrder,
		line.ItemID.Int64(),
		domain.DirectionIn.String(),
		line.Quantity.Int64(),
		line.EnteredUnit.String(),
		nullableText(line.EnteredPackagingName),
		line.Conversion.NumeratorAtomic(),
		line.Conversion.Denominator(),
		inventoryValue.Int64(),
		line.CommercialTotal.Int64(),
	).Scan(&lineID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO inventory_lots (
			item_id, source_line_id, initial_quantity_atomic, lot_code,
			originated_on, expires_on, created_at_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		line.ItemID.Int64(),
		lineID,
		line.Quantity.Int64(),
		nullableText(line.LotCode),
		originatedOn.String(),
		nullableBusinessDate(line.ExpiresOn),
		postedAt.UnixMilli(),
	); err != nil {
		return err
	}

	result, err := tx.ExecContext(ctx, `
		UPDATE inventory_balances
		SET quantity_atomic = quantity_atomic + ?,
		    inventory_value_micro = inventory_value_micro + ?,
		    last_document_id = ?,
		    updated_at_ms = ?
		WHERE item_id = ?
	`,
		line.Quantity.Int64(),
		inventoryValue.Int64(),
		documentID,
		postedAt.UnixMilli(),
		line.ItemID.Int64(),
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return fmt.Errorf("%w: inventory balance missing", domain.ErrInvalidReference)
	}
	return nil
}

func loadDocumentCurrency(ctx context.Context, tx databaseWriteTx) (domain.Currency, error) {
	var code string
	var minorDigits int64
	if err := tx.QueryRowContext(ctx, `
		SELECT currency_code, currency_minor_digits FROM app_settings WHERE id = 1
	`).Scan(&code, &minorDigits); err != nil {
		return domain.Currency{}, err
	}
	return domain.RestoreCurrency(code, int(minorDigits))
}

func loadPostedPurchaseDocument(ctx context.Context, tx databaseWriteTx, id int64) (PostedPurchaseDocument, error) {
	var row postedPurchaseDocumentRow
	err := tx.QueryRowContext(ctx, `
		SELECT id, idempotency_key, posting_sequence, counterparty_id, occurred_on,
		       posted_at_ms, currency_code, currency_minor_digits, reason_code, notes
		FROM stock_documents
		WHERE id = ? AND kind = 'PURCHASE'
	`, id).Scan(
		&row.id,
		&row.idempotencyKey,
		&row.postingSequence,
		&row.counterpartyID,
		&row.occurredOn,
		&row.postedAtMS,
		&row.currencyCode,
		&row.currencyMinorDigits,
		&row.reasonCode,
		&row.notes,
	)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}

	lines, err := loadPostedPurchaseLines(ctx, tx, id)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	return mapPostedPurchaseDocument(row, lines)
}

type postedPurchaseDocumentRow struct {
	id, postingSequence, postedAtMS, currencyMinorDigits int64
	idempotencyKey, occurredOn, currencyCode             string
	counterpartyID                                       sql.NullInt64
	reasonCode, notes                                    sql.NullString
}

func loadPostedPurchaseLines(ctx context.Context, tx databaseWriteTx, documentID int64) ([]PostedPurchaseLine, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT line.id, line.line_order, line.item_id, line.quantity_atomic,
		       line.entered_unit_code, line.entered_packaging_name,
		       line.conversion_numerator_atomic, line.conversion_denominator,
		       line.inventory_value_micro, line.commercial_total_minor,
		       lot.id, lot.lot_code, lot.originated_on, lot.expires_on
		FROM stock_document_lines line
		JOIN inventory_lots lot ON lot.source_line_id = line.id
		WHERE line.document_id = ?
		ORDER BY line.line_order, line.id
	`, documentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []PostedPurchaseLine
	for rows.Next() {
		var row postedPurchaseLineRow
		if err := rows.Scan(
			&row.id,
			&row.lineOrder,
			&row.itemID,
			&row.quantityAtomic,
			&row.enteredUnitCode,
			&row.enteredPackagingName,
			&row.conversionNumeratorAtomic,
			&row.conversionDenominator,
			&row.inventoryValueMicro,
			&row.commercialTotalMinor,
			&row.lotID,
			&row.lotCode,
			&row.originatedOn,
			&row.expiresOn,
		); err != nil {
			return nil, err
		}
		line, err := mapPostedPurchaseLine(row)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

type postedPurchaseLineRow struct {
	id, lineOrder, itemID, quantityAtomic            int64
	conversionNumeratorAtomic, conversionDenominator int64
	inventoryValueMicro, commercialTotalMinor, lotID int64
	enteredUnitCode, originatedOn                    string
	enteredPackagingName, lotCode, expiresOn         sql.NullString
}

func mapPostedPurchaseDocument(row postedPurchaseDocumentRow, lines []PostedPurchaseLine) (PostedPurchaseDocument, error) {
	id, err := domain.NewStockDocumentID(row.id)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	idempotencyKey, err := domain.NewIdempotencyKey(row.idempotencyKey)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	postingSequence, err := domain.NewPostingSequence(row.postingSequence)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	counterpartyID, err := optionalCounterpartyID(row.counterpartyID)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	occurredOn, err := domain.ParseBusinessDate(row.occurredOn)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	postedAt, err := domain.UTCInstantFromUnixMilli(row.postedAtMS)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	currency, err := domain.RestoreCurrency(row.currencyCode, int(row.currencyMinorDigits))
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	reason, err := optionalDocumentReason(domain.DocumentPurchase, row.reasonCode)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	notes, err := optionalNonEmptyText(row.notes)
	if err != nil {
		return PostedPurchaseDocument{}, err
	}
	return NewPostedPurchaseDocument(
		id, idempotencyKey, postingSequence, counterpartyID, occurredOn, postedAt,
		currency, reason, notes, lines,
	), nil
}

func mapPostedPurchaseLine(row postedPurchaseLineRow) (PostedPurchaseLine, error) {
	id, err := domain.NewStockDocumentLineID(row.id)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	lineOrder, err := domain.NewLineOrder(row.lineOrder)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	itemID, err := domain.NewItemID(row.itemID)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	quantity, err := domain.NewPositiveAtomicQuantity(row.quantityAtomic)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	enteredUnit, err := domain.NewUnitCode(row.enteredUnitCode)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	enteredPackagingName, err := optionalNonEmptyText(row.enteredPackagingName)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	conversion, err := domain.NewUnitConversion(row.conversionNumeratorAtomic, row.conversionDenominator)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	inventoryValue, err := domain.NewInventoryValue(row.inventoryValueMicro)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	commercialTotal, err := domain.NewMinorAmount(row.commercialTotalMinor)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	lotID, err := domain.NewInventoryLotID(row.lotID)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	lotCode, err := optionalNonEmptyText(row.lotCode)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	originatedOn, err := domain.ParseBusinessDate(row.originatedOn)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	expiresOn, err := optionalBusinessDate(row.expiresOn)
	if err != nil {
		return PostedPurchaseLine{}, err
	}
	return NewPostedPurchaseLine(
		id, lineOrder, itemID, quantity, enteredUnit, enteredPackagingName,
		conversion, inventoryValue, commercialTotal, lotID, lotCode, originatedOn, expiresOn,
	), nil
}

func nullableCounterpartyID(value domain.Option[domain.CounterpartyID]) any {
	if id, ok := value.Get(); ok {
		return id.Int64()
	}
	return nil
}

func nullableDocumentReason(value domain.Option[domain.DocumentReason]) any {
	if reason, ok := value.Get(); ok {
		return reason.String()
	}
	return nil
}

func nullableBusinessDate(value domain.Option[domain.BusinessDate]) any {
	if date, ok := value.Get(); ok {
		return date.String()
	}
	return nil
}
