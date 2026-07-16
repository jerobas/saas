# TODO

## Polish de frontend

- [ ] Adicionar UI minima para reversao exata de documentos.
- [ ] Refinar Compras: detalhe de documento postado com linhas/lotes/valores.
- [ ] Refinar Ajustes: historico minimo de documentos e atalho para reversao.
- [x] Melhorar Receitas: renomear receita pela UI.
- [ ] Melhorar Receitas: permitir multiplos componentes na UI.
- [ ] Melhorar Receitas: permitir escolher componente por embalagem, nao so unidade base.
- [ ] Melhorar Receitas: polir historico/editor de revisoes sem quebrar imutabilidade.
- [ ] Refinar Producao: preview simples antes de postar, mostrando consumo esperado, lotes FEFO e custo estimado.
- [ ] Refinar Vendas: permitir mais de uma linha/carrinho simples quando o detalhe/listagem estiver confortavel.

## Backlog adiado

- [ ] Phase 5.7: Dashboard/reporting real com queries read-only sobre vendas, estoque, baixo estoque, margem e produtos mais vendidos.
- [ ] Phase 5.8: Backup/restore funcional com validacao forte, safety backup, troca atomica e restart controlado.

## Lembretes do plano

- O trabalho atual ainda esta na Phase 5: vertical capabilities.
- A UI de receitas pertence a Phase 5.4: receitas e revisoes.
- A UI de vendas pertence a Phase 5.6: vendas, alocacao de saida, custo da mercadoria e saldos.
- Phase 5.7 e 5.8 existem, mas ficam adiadas por enquanto; o foco imediato e refinar as verticais ja implementadas.
- A Phase 7 vem depois e cobre durabilidade/release: validacao de backup/restore, replay de projecoes, smoke test empacotado e builds reproduziveis.
