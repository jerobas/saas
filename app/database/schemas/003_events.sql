CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,

  event_type TEXT NOT NULL CHECK(event_type IN (
    'PURCHASE',
    'PRODUCTION',
    'SALE',
    'ADJUSTMENT',
    'CONVERSION'
  )),

  status TEXT NOT NULL DEFAULT 'DRAFT'
    CHECK(status IN ('DRAFT', 'POSTED', 'CANCELLED')),

  counterparty_entity_id INTEGER,
  notes TEXT,

  occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

  FOREIGN KEY (counterparty_entity_id) REFERENCES entities(id)
);

CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);
CREATE INDEX IF NOT EXISTS idx_events_occurred_at ON events(occurred_at);

CREATE TRIGGER IF NOT EXISTS trg_events_no_insert_unless_draft
BEFORE INSERT ON events
WHEN NEW.status <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'events can only be inserted if status=DRAFT');
END;

-- events: only mutable while DRAFT
CREATE TRIGGER IF NOT EXISTS trg_events_no_update_unless_draft
BEFORE UPDATE ON events
WHEN OLD.status <> 'DRAFT'
BEGIN
  SELECT RAISE(FAIL, 'events can only be updated while status=DRAFT');
END;

CREATE TRIGGER IF NOT EXISTS trg_events_no_delete
BEFORE DELETE ON events
BEGIN
  SELECT RAISE(FAIL, 'events can''t be deleted');
END;

CREATE TRIGGER IF NOT EXISTS trg_events_require_lines_on_post
BEFORE UPDATE OF status ON events
WHEN OLD.status = 'DRAFT' AND NEW.status = 'POSTED'
BEGIN
  -- SALE must have at least one sale line
  SELECT RAISE(FAIL, 'Cannot POST SALE event without sale_lines')
  WHERE NEW.event_type = 'SALE'
    AND NOT EXISTS (
      SELECT 1
      FROM sale_lines sl
      WHERE sl.event_id = NEW.id
      LIMIT 1
    );

  -- PURCHASE must have at least one purchase line
  SELECT RAISE(FAIL, 'Cannot POST PURCHASE event without purchase_lines')
  WHERE NEW.event_type = 'PURCHASE'
    AND NOT EXISTS (
      SELECT 1
      FROM purchase_lines pl
      WHERE pl.event_id = NEW.id
      LIMIT 1
    );
END;