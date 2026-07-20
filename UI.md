# UI backlog

## Polimento das verticais já implementadas

- [ ] Adicionar UI mínima para reversão exata de documentos.
- [ ] Refinar Ajustes: histórico mínimo de documentos e atalho para reversão.
- [ ] Melhorar Receitas: permitir múltiplos componentes na UI.
- [ ] Melhorar Receitas: permitir escolher componente por embalagem, não só unidade base.
- [ ] Refinar Vendas: permitir mais de uma linha/carrinho simples quando o detalhe/listagem estiver confortável.

## Dashboard / reporting visual

O backend de reporting já tem endpoints reais separados por domínio e o dashboard
já faz o wiring oculto deles. A próxima decisão é visual: quando trocar os
cards/gráficos fake por esses dados reais.

- [ ] Converter `DashboardPage.jsx` para `DashboardPage.tsx`.
- [ ] Corrigir textos do dashboard que ainda dizem que as queries reais não existem; hoje o que falta é a visualização real.
- [ ] Substituir os cards visíveis por dados reais:
  - receita do período;
  - número de vendas;
  - produtos ativos/cadastrados;
  - crescimento versus período anterior.
- [ ] Substituir os gráficos visíveis por dados reais, mantendo o layout atual:
  - vendas/receita por período;
  - top produtos vendidos;
  - receita mensal;
  - vendas mensais;
  - mix por categoria como estado vazio/placeholder até existir categoria real no catálogo.
- [ ] Implementar states visíveis de loading, empty e error.
- [ ] Adicionar teste de componente para dashboard vazio.
- [ ] Adicionar teste de componente para dashboard com vendas/estoque.
- [ ] Atualizar README/troubleshooting quando o dashboard visível depender de dados reais para aparecer preenchido.
