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

### 5.7.-1 — Selecao de metricas e endpoints

Antes de implementar 5.7.0, escolher quais metricas entram no contrato de reporting. Decisao atual: todas as metricas abaixo entram como endpoints/read-model fields de backend. Nesta subsecao, `[x]` significa "selecionado para a superficie de reporting", nao "ja implementado".

Regra visual: por enquanto o Dashboard real renderiza somente os cards/graficos marcados como `FAKE ATUAL`, mantendo o mesmo tipo de card/grafico e ligando em dados reais. Os demais endpoints ficam funcionando e documentados para uso futuro. Graficos novos, quando forem adicionados, devem usar a mesma stack visual atual: Recharts, Motion, Tailwind e o mesmo estilo de cards.

#### Cards/KPIs principais

- [x] Receita total do periodo. `FAKE ATUAL: card Receita Total e linha Receita`
- [x] Numero de vendas no periodo. `FAKE ATUAL: card Vendas e linha/barra Vendas`
- [x] Produtos ativos/cadastrados. `FAKE ATUAL: card Produtos`
- [x] Crescimento versus periodo anterior. `FAKE ATUAL: card Crescimento`
- [x] Ticket medio de venda.
- [x] COGS/custo da mercadoria vendida no periodo.
- [x] Margem bruta do periodo.
- [x] Percentual de margem bruta.
- [x] Valor total em estoque.
- [x] Quantidade de itens abaixo do ponto de reposicao.
- [x] Quantidade de itens zerados que ainda sao vendaveis.

#### Series e graficos de vendas

- [x] Serie vendas e receita por dia/mes. `FAKE ATUAL: grafico linha Vendas e Receita`
- [x] Receita mensal. `FAKE ATUAL: aba Receita`
- [x] Vendas mensais. `FAKE ATUAL: aba Vendas`
- [x] Top produtos vendidos por quantidade. `FAKE ATUAL: Produtos Mais Vendidos`
- [x] Top produtos vendidos por receita.
- [x] Vendas gratis/promocao/amostra por quantidade e valor comercial zerado.
- [x] Vendas por cliente, quando houver cliente informado.
- [x] Vendas sem cliente/anonimas.

#### Estoque, lotes e compras

- [x] Baixo estoque por item, com saldo atual e ponto de reposicao.
- [x] Lotes vencendo em 7/30 dias.
- [x] Lotes vencidos ainda com saldo.
- [x] Valor de estoque por item.
- [x] Compras/spend por periodo.
- [x] Top fornecedores por gasto.
- [x] Entradas gratis (`FREE_STOCK`) por periodo.

#### Producao, ajustes e qualidade operacional

- [x] Producao por receita/produto.
- [x] Custo direto de producao por periodo.
- [x] Variacao simples de yield: rendimento real versus rendimento padrao.
- [x] Ajustes negativos por motivo: perda, vencimento, dano, amostra, correcao.
- [x] Ajustes positivos por motivo: saldo inicial, brinde/estoque gratis, contagem fisica.
- [x] Documentos revertidos/correcoes exatas por periodo.

#### Futuro / depende de dimensao nova

- [x] Mix por categoria. `FAKE ATUAL: grafico pizza Categorias; endpoint placeholder ate existir categoria/tag de catalogo`

### 5.7.0 — Contratos e decisoes de numeros

- [x] Documentar a superficie completa de endpoints/read-model fields em `docs/domain/reporting.md`.
- [x] Definir contratos de reporting com `currencyCode`, `currencyMinorDigits`, periodo, cards, series e tabelas para todas as metricas selecionadas em 5.7.-1.
- [x] Manter `GetDashboardReport` como endpoint agregado para a tela atual e expor endpoints agrupados por dominio para uso futuro:
  - vendas;
  - estoque/lotes;
  - compras;
  - producao;
  - ajustes/correcoes;
  - categorias placeholder.
- [ ] Padronizar valores monetarios no contrato:
  - receita/comercial em `commercialTotalMinor`;
  - estoque/COGS/margem em `inventoryValueMicro`;
  - frontend formata cada escala explicitamente.
- [ ] Decidir regra de reversao para dashboard operacional: documentos revertidos por `REVERSAL` saem dos agregados; documentos `REVERSAL` ficam fora do dashboard e entram futuramente em auditoria.
- [x] Comecar sem categorias reais no dominio; o endpoint de categorias retorna vazio/indisponivel com motivo explicito, mas o grafico pizza pode continuar na tela como estado vazio.

### 5.7.1 — Queries read-only de reporting

- [x] Criar `app/internal/infrastructure/sqlite/queries/reporting.sql`.
- [x] Gerar sqlc para reporting.
- [x] Implementar `GetSalesReport` real:
  - total de vendas;
  - receita;
  - COGS;
  - margem bruta;
  - percentual de margem bruta;
  - crescimento versus periodo anterior;
  - ticket medio;
  - series por dia/mes;
  - top produtos por quantidade e receita;
  - vendas gratis/promocao/amostra;
  - vendas por cliente;
  - vendas anonimas;
  - exclusao de vendas revertidas.
- [x] Implementar `GetInventoryReport` real:
  - valor total em estoque;
  - quantidade de itens abaixo do ponto de reposicao;
  - quantidade de itens vendaveis zerados;
  - baixo estoque por item;
  - lotes vencendo em 7/30 dias;
  - lotes vencidos ainda com saldo;
  - valor de estoque por item.
- [x] Implementar `GetPurchaseReport` real:
  - compras/spend por periodo;
  - top fornecedores por gasto;
  - entradas gratis `FREE_STOCK`;
  - exclusao de compras revertidas.
- [x] Implementar `GetProductionReport` real:
  - producao por receita/produto;
  - custo direto de producao por periodo;
  - variacao simples de yield;
  - exclusao de producoes revertidas.
- [x] Implementar `GetAdjustmentReport` real:
  - ajustes positivos/negativos por motivo;
  - documentos revertidos/correcoes exatas;
- [ ] Implementar queries restantes:
  - dashboard vazio sem documentos.
- [x] Garantir que SALE revertida nao entra nos agregados.

### 5.7.2 — Store/application

- [ ] Criar `ReportingStore` read-only em SQLite.
- [x] Criar primeiro `ReportingStore` read-only para vendas.
- [x] Expandir `ReportingStore` read-only para inventario/lotes.
- [x] Expandir `ReportingStore` read-only para compras.
- [x] Expandir `ReportingStore` read-only para producao.
- [x] Expandir `ReportingStore` read-only para ajustes/correcoes.
- [ ] Criar testes de store com banco temporario:
  - dashboard vazio;
  - compras sem vendas;
  - venda normal;
  - venda revertida excluida;
  - estoque valorizado;
  - baixo estoque.
- [x] Criar `ReportingService` fino, sem regra de escrita, para `GetSalesReport`.
- [x] Expandir `ReportingService` para ajustes/correcoes.
- [ ] Adicionar teste de application para periodo invalido e periodo padrao.

### 5.7.3 — Wails/gateway

- [x] Criar DTOs de reporting em `internal/presentation/wails/dto`.
- [x] Criar `ReportingHandler.GetDashboardReport(request)`.
- [x] Criar endpoints agrupados para as metricas completas:
  - `GetSalesReport`;
  - `GetInventoryReport`;
  - `GetPurchaseReport`;
  - `GetProductionReport`;
  - `GetAdjustmentReport`;
  - `GetCategoryMixReport`.
- [x] Registrar handler no app Wails.
- [x] Adicionar `reportingGateway` tipado no frontend.
- [x] Cobrir gateway/handler em testes de superficie para vendas, inventario, compras, producao, ajustes e placeholder de categorias.

### 5.7.4 — Dashboard real minimo

- [ ] Converter `DashboardPage.jsx` para `DashboardPage.tsx`.
- [ ] Remover arrays demo (`salesdData`, `topProductsData`, `categoriesData`) ou isolar apenas como fallback visual claramente inativo.
- [x] Carregar endpoints reais no mount em uma area oculta (`display: none`) para wiring:
  - vendas;
  - estoque;
  - compras;
  - producao;
  - ajustes;
  - mix de categorias placeholder.
- [ ] Implementar states loading, empty e error.
- [ ] Manter o layout visual antigo, mas trocar os cards para:
  - receita do periodo;
  - numero de vendas;
  - produtos ativos/cadastrados;
  - crescimento versus periodo anterior.
- [ ] Trocar graficos para:
  - vendas/receita por periodo;
  - top produtos vendidos;
  - receita mensal;
  - vendas mensais;
  - mix por categoria em estado vazio/placeholder ate existir categoria real.
- [ ] Nao renderizar ainda os endpoints novos sem grafico atual; eles ficam prontos no gateway para uso futuro.

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
