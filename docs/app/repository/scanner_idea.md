```go
// scanSaleLineRow scans a single SaleLine from anything that implements Scan
// (both *sql.Row and *sql.Rows do).
type sqlScanner interface {
	Scan(dest ...any) error
}

func scanSaleLineRow(s sqlScanner) (*model.SaleLine, error) {
	sl := &model.SaleLine{}
	if err := s.Scan(
		&sl.ID,
		&sl.EventID,
		&sl.ItemID,
		&sl.Quantity,
		&sl.UnitPrice,
		&sl.TotalPrice,
		&sl.CreatedAt,
		&sl.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return sl, nil
}

row := r.db.Conn.QueryRow(query, id)
sl, err := scanSaleLineRow(row)
if err == sql.ErrNoRows {
	return nil, nil // ou ErrNotFound, se você padronizar isso
}
return sl, err

saleLines := []*model.SaleLine{}
for rows.Next() {
	sl, err := scanSaleLineRow(rows)
	if err != nil {
		return nil, err
	}
	saleLines = append(saleLines, sl)
}
return saleLines, rows.Err()
```