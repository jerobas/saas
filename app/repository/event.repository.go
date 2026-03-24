package repository

import (
	"database/sql"
	"github.com/jerobas/saas/model"
)

type EventRepository struct {
	db *Database
}

func NewEventRepository(db *Database) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(evt *model.EventInsertDTO) (int64, error) {
	query := `
		INSERT INTO events
			(event_type, status, counterparty_entity_id, notes, occurred_at)
		VALUES
			(?, ?, ?, ?, ?)
	`

	res, err := r.db.Conn.Exec(
		query,
		evt.EventType,
		evt.Status,
		evt.CounterpartyEntityID, 
		evt.Notes,
		evt.OccurredAt,
	)

	if err != nil {
		return (-1, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return (-1, err)
	}

	return (id, nil)
}

func (r *EventRepository) GetByID(id int64) (*model.Event, error) {
	query := `
		SELECT
			id,
			event_type,
			status,
			counterparty_entity_id,
			notes,
			occurred_at,
			created_at
		FROM events
		WHERE id = ?
	`

	evt := &model.Event{}
	err := r.db.Conn.QueryRow(query, id).Scan(
		&evt.ID,
		&evt.EventType,
		&evt.Status,
		&evt.CounterpartyEntityID, 
		&evt.Notes,
		&evt.OccurredAt,
		&evt.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}

	return evt, nil
}

func (r *EventRepository) GetAll() ([]*model.Event, error) {
	query := `
		SELECT
			id,
			event_type,
			status,
			counterparty_entity_id,
			notes,
			occurred_at,
			created_at
		FROM events
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evts := []*model.Event{}
	for rows.Next() {
		evt := &model.Event{}
		if err := rows.Scan(
			&evt.ID,
			&evt.EventType,
			&evt.Status,
			&evt.CounterpartyEntityID, 
			&evt.Notes,
			&evt.OccurredAt,
			&evt.CreatedAt,
		); err != nil {
			return nil, err
		}
		evts = append(evts, evt)
	}

	return evts, rows.Err()
}

func (r *EventRepository) GetAllByEventType(eventType string) ([]*model.Event, error) {
	query := `
		SELECT
			id,
			event_type,
			status,
			counterparty_entity_id,
			notes,
			occurred_at,
			created_at
		FROM events
		WHERE event_type = ?
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evts := []*model.Event{}
	for rows.Next() {
		evt := &model.Event{}
		if err := rows.Scan(
			&evt.ID,
			&evt.EventType,
			&evt.Status,
			&evt.CounterpartyEntityID, 
			&evt.Notes,
			&evt.OccurredAt,
			&evt.CreatedAt,
		); err != nil {
			return nil, err
		}
		evts = append(evts, evt)
	}

	return evts, rows.Err()
}

func (r *EventRepository) GetAllByCounterpartyID(counterpartyID int64) ([]*model.Event, error) {
	query := `
		SELECT
			id,
			event_type,
			status,
			counterparty_entity_id,
			notes,
			occurred_at,
			created_at
		FROM events
		WHERE counterparty_entity_id = ?
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query, counterpartyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evts := []*model.Event{}
	for rows.Next() {
		evt := &model.Event{}
		if err := rows.Scan(
			&evt.ID,
			&evt.EventType,
			&evt.Status,
			&evt.CounterpartyEntityID, 
			&evt.Notes,
			&evt.OccurredAt,
			&evt.CreatedAt,
		); err != nil {
			return nil, err
		}
		evts = append(evts, evt)
	}

	return evts, rows.Err()
}

func (r *EventRepository) GetAllByEventTypeAndCounterpartyID(eventType string, counterpartyID int64) ([]*model.Event, error) {
	query := `
		SELECT
			id,
			event_type,
			status,
			counterparty_entity_id,
			notes,
			occurred_at,
			created_at
		FROM events
		WHERE event_type = ? 
		AND	counterparty_entity_id = ?
		ORDER BY occurred_at DESC
	`

	rows, err := r.db.Conn.Query(query, eventType, counterpartyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	evts := []*model.Event{}
	for rows.Next() {
		evt := &model.Event{}
		if err := rows.Scan(
			&evt.ID,
			&evt.EventType,
			&evt.Status,
			&evt.CounterpartyEntityID, 
			&evt.Notes,
			&evt.OccurredAt,
			&evt.CreatedAt,
		); err != nil {
			return nil, err
		}
		evts = append(evts, evt)
	}

	return evts, rows.Err()
}

func (r *EventRepository) Post(eventID int64) error {
	query := `
		UPDATE events
		SET status = 'POSTED'
		WHERE event_id = ?
	`

	res, err := r.db.Conn.Exec(query, eventID)

	if err != nil {
		return err
	}

	return nil
}

func (r *EventRepository) Cancel(eventID int64) error {
	query := `
		UPDATE events
		SET status = 'CANCELLED'
		WHERE event_id = ?
	`

	res, err := r.db.Conn.Exec(query, eventID)

	if err != nil {
		return err
	}

	return nil
}

// events cant be deleted

// func (r *EventRepository) Delete(id int64) error {
// 	query := `DELETE FROM events WHERE id = ?`
// 	_, err := r.db.Conn.Exec(query, id)
// 	return err
// }