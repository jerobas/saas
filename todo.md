# TODO

## Polish de frontend

- [ ] Adicionar UI minima para reversao exata de documentos.
- [ ] Refinar Ajustes: historico minimo de documentos e atalho para reversao.
- [ ] Melhorar Receitas: permitir multiplos componentes na UI.
- [ ] Melhorar Receitas: permitir escolher componente por embalagem, nao so unidade base.
- [x] Melhorar Receitas: polir historico/editor de revisoes sem quebrar imutabilidade.
- [ ] Refinar Vendas: permitir mais de uma linha/carrinho simples quando o detalhe/listagem estiver confortavel.

## Backlog adiado

- [ ] Phase 5.7: Dashboard/reporting real com queries read-only sobre vendas, estoque, baixo estoque, margem e produtos mais vendidos.
- [ ] Phase 5.8: Backup/restore funcional com validacao forte, safety backup, troca atomica e restart controlado.

## Phase 5.7 — Dashboard/reporting real

Objetivo: trocar o Dashboard demo por dados locais reais sem transformar reporting em fonte de verdade. Reporting le os documentos, linhas, lotes e projecoes existentes; nunca escreve e nunca substitui stores operacionais.

### 5.7.0 — Contrato e decisoes de numeros

- [ ] Definir contrato `DashboardReport` com `currencyCode`, `currencyMinorDigits`, periodo, cards, series e tabelas.
- [ ] Padronizar valores monetarios no contrato:
  - receita/comercial em `commercialTotalMinor`;
  - estoque/COGS/margem em `inventoryValueMicro`;
  - frontend formata cada escala explicitamente.
- [ ] Decidir regra de reversao para dashboard operacional: documentos revertidos por `REVERSAL` saem dos agregados; documentos `REVERSAL` ficam fora do dashboard e entram futuramente em auditoria.
- [ ] Comecar sem categorias reais; remover/ocultar grafico de categorias ate existir uma dimensao de catalogo equivalente.

### 5.7.1 — Queries read-only de reporting

- [ ] Criar `app/internal/infrastructure/sqlite/queries/reporting.sql`.
- [ ] Gerar sqlc para reporting.
- [ ] Implementar queries:
  - total de vendas no periodo;
  - receita no periodo;
  - COGS no periodo;
  - margem bruta no periodo;
  - serie por mes/dia usando `occurred_on`;
  - produtos mais vendidos por quantidade e receita;
  - valor total em estoque via `inventory_balances`;
  - itens abaixo do ponto de reposicao via `reorder_quantity_atomic`;
  - dashboard vazio sem documentos.
- [ ] Garantir que SALE revertida nao entra nos agregados.

### 5.7.2 — Store/application

- [ ] Criar `ReportingStore` read-only em SQLite.
- [ ] Criar testes de store com banco temporario:
  - dashboard vazio;
  - compras sem vendas;
  - venda normal;
  - venda revertida excluida;
  - estoque valorizado;
  - baixo estoque.
- [ ] Criar `ReportingService` fino, sem regra de escrita, apenas validacao de periodo e composicao do report.
- [ ] Adicionar teste de application para periodo invalido e periodo padrao.

### 5.7.3 — Wails/gateway

- [ ] Criar DTOs de reporting em `internal/presentation/wails/dto`.
- [ ] Criar `ReportingHandler.GetDashboardReport(request)`.
- [ ] Registrar handler no app Wails.
- [ ] Adicionar `reportingGateway` tipado no frontend.
- [ ] Cobrir gateway/handler em testes de superficie.

### 5.7.4 — Dashboard real minimo

- [ ] Converter `DashboardPage.jsx` para `DashboardPage.tsx`.
- [ ] Remover arrays demo (`salesdData`, `topProductsData`, `categoriesData`) ou isolar apenas como fallback visual claramente inativo.
- [ ] Carregar report real no mount.
- [ ] Implementar states loading, empty e error.
- [ ] Manter o layout visual antigo, mas trocar os cards para:
  - receita do periodo;
  - numero de vendas;
  - valor em estoque;
  - margem bruta quando houver venda.
- [ ] Trocar graficos para:
  - vendas/receita por periodo;
  - top produtos vendidos;
  - baixo estoque.

### 5.7.5 — Testes e docs

- [ ] Teste de componente para dashboard vazio.
- [ ] Teste de componente para dashboard com vendas/estoque.
- [ ] Documentar em `docs/domain` que dashboard e reporting sao read models derivados.
- [ ] Atualizar README/troubleshooting se o dashboard precisar de dados reais para aparecer preenchido.

## Lembretes do plano

- O trabalho atual ainda esta na Phase 5: vertical capabilities.
- A UI de receitas pertence a Phase 5.4: receitas e revisoes.
- A UI de vendas pertence a Phase 5.6: vendas, alocacao de saida, custo da mercadoria e saldos.
- Phase 5.7 e 5.8 existem, mas ficam adiadas por enquanto; o foco imediato e refinar as verticais ja implementadas.
- A Phase 7 vem depois e cobre durabilidade/release: validacao de backup/restore, replay de projecoes, smoke test empacotado e builds reproduziveis.
