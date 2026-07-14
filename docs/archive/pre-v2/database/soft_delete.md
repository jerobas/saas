  active INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1))
  deleted INTEGER NOT NULL DEFAULT 0
    CHECK (deleted IN (0, 1))
    CHECK (NOT (active = 1 AND deleted = 1))