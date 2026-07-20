# UI backlog

## Estoque e lotes

- [x] Unificar Lotes como uma visão/filtro da página de Estoque, em vez de manter como uma página principal separada.

## Compras

- [x] Redesenhar Compra como documento com cabeçalho compartilhado e múltiplas linhas:
  - fornecedor, data de ocorrência e observações no topo;
  - busca rápida de item + botão adicionar;
  - linhas editáveis para quantidade, unidade/embalagem, valor, lote e validade.
- [x] Adicionar um switch para informar lote/validade na compra:
  - desligado por padrão para manter a tela mais simples;
  - quando ligado, exibir campos de código do lote e validade;
  - manter lote/validade opcionais mesmo quando o switch estiver ligado;
  - lote interno continua existindo para estoque/custo, mas os dados visíveis de lote podem ficar em branco.

## Unidades e embalagens

- [x] Esconder conversão numerador/denominador da UI:
  - usuário escolhe unidade ou embalagem;
  - app mostra apenas um preview legível do equivalente na unidade base quando útil;
  - backend continua salvando a conversão exata historicamente.

## Ajustes e reversões

- [x] Adicionar UI mínima para reversão exata de documentos.
- [x] Refinar Ajustes: histórico mínimo de documentos e atalho para reversão.

## Receitas

- [x] Melhorar Receitas: permitir múltiplos componentes na UI.
- [x] Melhorar Receitas: permitir escolher componente por embalagem, não só unidade base.

## Vendas

- [x] Refinar Vendas: permitir mais de uma linha/carrinho simples quando o detalhe/listagem estiver confortável.

## Dashboard / reporting visual

O backend de reporting já tem endpoints reais separados por domínio e o dashboard
já faz o wiring oculto deles. A próxima decisão é visual: quando trocar os
cards/gráficos fake por esses dados reais.

- [x] Converter `DashboardPage.jsx` para `DashboardPage.tsx`.
- [x] Corrigir textos do dashboard que ainda dizem que as queries reais não existem; hoje o que falta é a visualização real.
- [x] Substituir os cards visíveis por dados reais:
  - receita do período;
  - número de vendas;
  - produtos ativos/cadastrados;
  - crescimento versus período anterior.
- [x] Substituir os gráficos visíveis por dados reais, mantendo o layout atual:
  - vendas/receita por período;
  - top produtos vendidos;
  - receita mensal;
  - vendas mensais;
  - mix por categoria como estado vazio/placeholder até existir categoria real no catálogo.
- [ ] Implementar states visíveis de loading, empty e error.
- [ ] Adicionar teste de componente para dashboard vazio.
- [ ] Adicionar teste de componente para dashboard com vendas/estoque.
- [ ] Atualizar README/troubleshooting quando o dashboard visível depender de dados reais para aparecer preenchido.

## Late game

- [ ] Permitir criar produto/componente rapidamente durante o lançamento de uma compra.
