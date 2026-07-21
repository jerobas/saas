import { useCallback, useEffect, useMemo, useState } from "react";
import { motion } from "motion/react";
import { ArrowsClockwise, Package, Warning } from "@phosphor-icons/react";
import { inventoryGateway } from "../../gateways/desktopBridge";

const integerFormatter = new Intl.NumberFormat("pt-BR");

const formatQuantity = (quantityAtomic, unitCode) =>
  `${integerFormatter.format(quantityAtomic)} ${unitCode}`;

const InventoryLotsView = ({ balances, loadingBalances }) => {
  const [selectedItemId, setSelectedItemId] = useState(null);
  const [lots, setLots] = useState([]);
  const [loadingLots, setLoadingLots] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setSelectedItemId((current) =>
      balances.some((item) => item.itemId === current) ? current : (balances[0]?.itemId ?? null),
    );
  }, [balances]);

  const selectedBalance = useMemo(
    () => balances.find((item) => item.itemId === selectedItemId) ?? null,
    [balances, selectedItemId],
  );

  const loadLots = useCallback(async () => {
    if (selectedItemId == null) {
      setLots([]);
      return;
    }

    try {
      setLoadingLots(true);
      setError(null);
      setLots(await inventoryGateway.listItemLotFacts(selectedItemId));
    } catch (err) {
      console.error("Erro ao carregar lotes:", err);
      setError(err.message || "Erro ao carregar lotes.");
    } finally {
      setLoadingLots(false);
    }
  }, [selectedItemId]);

  useEffect(() => {
    void loadLots();
  }, [loadLots]);

  const availableQuantity = lots.reduce((acc, lot) => acc + lot.availableQuantityAtomic, 0);
  const datedLots = lots.filter((lot) => lot.expiresOn != null).length;

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[320px_1fr]">
      <section className="rounded-2xl border border-slate-100 bg-white p-4 shadow-sm">
        <h2 className="mb-4 text-lg font-semibold text-slate-900">Itens em estoque</h2>
        {loadingBalances ? (
          <p className="py-8 text-center text-sm text-slate-500">Carregando itens...</p>
        ) : balances.length === 0 ? (
          <p className="py-8 text-center text-sm text-slate-500">
            Nenhum item com saldo encontrado.
          </p>
        ) : (
          <div className="space-y-2">
            {balances.map((item) => (
              <button
                key={item.itemId}
                type="button"
                onClick={() => setSelectedItemId(item.itemId)}
                className={`w-full rounded-xl border p-3 text-left transition ${
                  selectedItemId === item.itemId
                    ? "border-pink-300 bg-pink-50"
                    : "border-slate-100 hover:bg-slate-50"
                }`}
              >
                <div className="font-medium text-slate-900">{item.itemName}</div>
                <div className="mt-1 text-sm text-slate-500">
                  {formatQuantity(item.quantityAtomic, item.baseUnitCode)}
                </div>
              </button>
            ))}
          </div>
        )}
      </section>

      <section>
        <div className="mb-6 flex flex-col gap-4 rounded-2xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-900 sm:flex-row sm:items-center sm:justify-between">
          <p>
            Os lotes são criados por compras, produção e ajustes de entrada. Esta visão consulta o
            histórico físico real do estoque local.
          </p>
          <button
            type="button"
            onClick={() => void loadLots()}
            disabled={loadingLots || selectedItemId == null}
            className="inline-flex shrink-0 items-center justify-center gap-2 rounded-lg border border-blue-200 bg-white px-4 py-2 font-semibold text-blue-900 transition hover:bg-blue-100 disabled:cursor-not-allowed disabled:text-slate-400"
          >
            <ArrowsClockwise size={18} className={loadingLots ? "animate-spin" : ""} />
            Atualizar lotes
          </button>
        </div>

        {error && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 flex items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700"
          >
            <span>{error}</span>
            <button
              type="button"
              onClick={() => setError(null)}
              className="ml-auto text-red-600 hover:text-red-800"
              aria-label="Fechar erro"
            >
              ×
            </button>
          </motion.div>
        )}

        <div className="mb-6 grid grid-cols-1 gap-6 md:grid-cols-3">
          <div className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <h3 className="text-sm font-medium text-slate-600">Item selecionado</h3>
            <p className="mt-2 text-2xl font-bold text-slate-900">
              {selectedBalance?.itemName ?? "Nenhum"}
            </p>
          </div>
          <div className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <h3 className="text-sm font-medium text-slate-600">Disponível em lotes</h3>
            <p className="mt-2 text-2xl font-bold text-green-600">
              {selectedBalance == null
                ? "—"
                : formatQuantity(availableQuantity, selectedBalance.baseUnitCode)}
            </p>
          </div>
          <div className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <div className="flex items-center gap-2">
              <Warning size={20} className="text-orange-600" />
              <h3 className="text-sm font-medium text-slate-600">Lotes com validade</h3>
            </div>
            <p className="mt-2 text-2xl font-bold text-orange-600">{datedLots}</p>
          </div>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="overflow-hidden rounded-2xl border border-slate-100 bg-white shadow-sm"
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-slate-200 bg-slate-50">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Lote</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Disponível
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Inicial
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Origem
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Validade
                  </th>
                </tr>
              </thead>
              <tbody>
                {loadingLots ? (
                  <tr>
                    <td colSpan="5" className="px-6 py-12 text-center text-slate-500">
                      Carregando lotes...
                    </td>
                  </tr>
                ) : selectedBalance == null ? (
                  <tr>
                    <td colSpan="5" className="px-6 py-12 text-center text-slate-500">
                      Selecione um item para ver lotes.
                    </td>
                  </tr>
                ) : lots.length === 0 ? (
                  <tr>
                    <td colSpan="5" className="px-6 py-12 text-center text-slate-500">
                      <Package size={32} className="mx-auto mb-3 text-slate-400" />
                      Nenhum lote encontrado para este item.
                    </td>
                  </tr>
                ) : (
                  lots.map((lot) => (
                    <tr key={lot.id} className="border-b border-slate-100 hover:bg-slate-50">
                      <td className="px-6 py-4">
                        <div className="font-medium text-slate-900">
                          {lot.lotCode ?? `#${lot.id}`}
                        </div>
                        <div className="text-xs text-slate-500">ID {lot.id}</div>
                      </td>
                      <td className="px-6 py-4 text-sm font-semibold text-slate-900">
                        {formatQuantity(lot.availableQuantityAtomic, selectedBalance.baseUnitCode)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-700">
                        {formatQuantity(lot.initialQuantityAtomic, selectedBalance.baseUnitCode)}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-700">
                        {lot.sourceKind} #{lot.sourceDocumentId}
                        <div className="text-xs text-slate-500">{lot.sourceOccurredOn}</div>
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-700">
                        {lot.expiresOn ?? <span className="text-slate-400">Sem validade</span>}
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </motion.div>
      </section>
    </div>
  );
};

export default InventoryLotsView;
