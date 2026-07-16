# TODO

## Polish de frontend

- [x] Remover o card "Unidades" da tela de Produtos/Catalogo.
- [ ] Adicionar UI minima para reversao exata de documentos.
- [ ] Melhorar Receitas: renomear receita pela UI.
- [ ] Melhorar Receitas: permitir multiplos componentes na UI.
- [ ] Melhorar Receitas: permitir escolher componente por embalagem, nao so unidade base.
- [ ] Melhorar Receitas: polir historico/editor de revisoes sem quebrar imutabilidade.

## Lembretes do plano

- O trabalho atual ainda esta na Phase 5: vertical capabilities.
- A UI de receitas pertence a Phase 5.4: receitas e revisoes.
- A UI de vendas pertence a Phase 5.6: vendas, alocacao de saida, custo da mercadoria e saldos.
- A Phase 7 vem depois e cobre durabilidade/release: validacao de backup/restore, replay de projecoes, smoke test empacotado e builds reproduziveis.

## Durabilidade / idempotencia

- [ ] Revisar services de posting para permitir replay idempotente quando o documento existente foi postado em um instante anterior ao relogio atual.
