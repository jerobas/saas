Beleza. Vou te dar um exemplo **bem concreto** de migrações “na unha” (sem Prisma), usando só `database/sql` + SQLite, com uma arquitetura simples e profissional.

A ideia é:

1. você mantém arquivos `.sql` versionados (`001_...sql`, `002_...sql`)
2. na inicialização do app, você:

   * cria uma tabela `schema_migrations`
   * lê os arquivos
   * executa os que ainda não foram aplicados
   * registra a versão aplicada numa transação

Isso dá o básico do Prisma migrate: “aplicar só o que falta”.

---

## Estrutura de projeto recomendada (simples)

```
myapp/
  cmd/
    myapp/
      main.go
  internal/
    db/
      open.go
      migrate/
        migrate.go
        fs.go
  migrations/
    001_init.sql
    002_items_products_recipes.sql
    003_triggers.sql
  go.mod
```

* `migrations/` fica na raiz (ou embedado no binário).
* `internal/db/open.go` abre conexão e configura PRAGMAs.
* `internal/db/migrate/migrate.go` aplica migrações.

---

## Como ficam os arquivos de migração

### `migrations/001_init.sql`

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

> Sim: eu crio `schema_migrations` no primeiro arquivo, mas no código eu também garanto que exista (redundância segura).

### `migrations/002_items_products_recipes.sql`

Você coloca seu schema aqui (CREATE TABLE, INDEX, VIEW…).

### `migrations/003_triggers.sql`

Só triggers, pra ficar organizado.

---

## Código: abrir DB + migrar

### `internal/db/open.go`

```go
package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite" // ou mattn/go-sqlite3
)

func Open(path string) (*sql.DB, error) {
	// SQLite DSN exemplos:
	// modernc sqlite: "file:app.db?_pragma=foreign_keys(1)"
	// mattn: "file:app.db?_foreign_keys=on"
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)", path)

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	// Ajuste de pool pra SQLite (geralmente 1 conexão ajuda a evitar surpresas com PRAGMA por-conn)
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Testa conexão
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}

	// Extra: PRAGMA por conexão (redundante, mas robusto)
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
```

---

## Ler migrações do filesystem (ou embed)

Pra começo, vamos ler do disco.

### `internal/db/migrate/fs.go`

```go
package migrate

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Migration struct {
	Version  string
	Filename string
	SQL      string
}

func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, name)
		}
	}

	sort.Strings(files)

	migs := make([]Migration, 0, len(files))
	for _, name := range files {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}

		// Versão = prefixo antes do primeiro "_"
		// ex: "002_items.sql" => "002"
		version := name
		if i := strings.Index(name, "_"); i > 0 {
			version = name[:i]
		}

		migs = append(migs, Migration{
			Version:  version,
			Filename: name,
			SQL:      string(b),
		})
	}

	return migs, nil
}
```

---

## Aplicar migrações

### `internal/db/migrate/migrate.go`

```go
package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

func EnsureMigrationsTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func AppliedVersions(ctx context.Context, db *sql.DB) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := map[string]bool{}
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func ApplyAll(ctx context.Context, db *sql.DB, migrationsDir string) error {
	if err := EnsureMigrationsTable(ctx, db); err != nil {
		return err
	}

	migs, err := LoadMigrations(migrationsDir)
	if err != nil {
		return err
	}

	applied, err := AppliedVersions(ctx, db)
	if err != nil {
		return err
	}

	for _, m := range migs {
		if applied[m.Version] {
			continue
		}

		// Uma migração = uma transação
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		// Executa o SQL do arquivo (múltiplas statements separadas por ;)
		if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s (%s) failed: %w", m.Version, m.Filename, err)
		}

		// Marca como aplicada
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(version) VALUES (?)`, m.Version,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s insert failed: %w", m.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}
```

---

## `main.go`: chamar isso tudo

### `cmd/myapp/main.go`

```go
package main

import (
	"context"
	"log"

	"myapp/internal/db"
	"myapp/internal/db/migrate"
)

func main() {
	ctx := context.Background()

	database, err := db.Open("app.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	if err := migrate.ApplyAll(ctx, database, "migrations"); err != nil {
		log.Fatal(err)
	}

	log.Println("DB ready ✅")
	// start server...
}
```

---

## Como você usa no dia a dia (fluxo concreto)

1. Você precisa mudar schema:

   * cria um novo arquivo:
     `migrations/004_add_items.sql`
2. Dentro dele você coloca o `ALTER TABLE` / `CREATE TABLE` / etc.
3. Rodou o app → ele aplica só o 004 e grava em `schema_migrations`.

### Exemplo real de migração “nova”

`migrations/004_add_is_sellable.sql`:

```sql
ALTER TABLE items ADD COLUMN can_sell INTEGER NOT NULL DEFAULT 0 CHECK(can_sell IN (0,1));
```

---

## Boas práticas rápidas (as que importam)

* Nomeia migração com prefixo incremental e descrição: `005_add_event_table.sql`
* Uma migração por mudança lógica.
* Nunca edite migrações antigas depois que o app já rodou em outro ambiente.
* Em SQLite, prefira `CREATE TABLE new + copy + drop` quando `ALTER TABLE` for limitado (tipo remover coluna).

---

## Quer deixar ainda mais profissional?

Dois upgrades pequenos:

1. **Embed migrations** no binário (Go `embed`) pra não depender de arquivo no deploy.
2. Guardar versão como “filename completo” (ex `002_items_products_recipes.sql`) e não só `002`, pra evitar colisão.

Se você quiser, eu adapto o exemplo pra `embed.FS` (fica bem bonito e é zero dependência).
