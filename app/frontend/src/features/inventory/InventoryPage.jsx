import { useCallback, useEffect, useMemo, useState } from "react";
import { motion } from "motion/react";
import { ArrowsClockwise, Package, Warning } from "@phosphor-icons/react";
import {
  adjustmentGateway,
  inventoryGateway,
  referenceDataGateway,
} from "../../gateways/desktopBridge";

const currencyFormatter = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

const integerFormatter = new Intl.NumberFormat("pt-BR");

const formatQuantity = (quantityAtomic, unitCode) =>
  `${integerFormatter.format(quantityAtomic)} ${unitCode}`;

const formatInventoryValue = (valueMicro) => currencyFormatter.format(valueMicro / 1_000_000);

const parseInteger = (value) => {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const parseInventoryValueMicro = (value) => {
  const normalized = value.trim().replace(",", ".");
  if (normalized.length === 0) return undefined;
  const parsed = Number.parseFloat(normalized);
  if (!Number.isFinite(parsed)) return undefined;
  return Math.round(parsed * 1_000_000);
};

const adjustmentReasons = {
  IN: [
    ["OPENING_BALANCE", "Saldo inicial"],
    ["FREE_STOCK", "Entrada gratuita"],
    ["PHYSICAL_COUNT", "Contagem fisica"],
    ["DOCUMENTED_CORRECTION", "Correcao documentada"],
  ],
  OUT: [
    ["WASTE", "Perda"],
    ["EXPIRY", "Validade"],
    ["DAMAGE", "Dano"],
    ["SAMPLE", "Amostra"],
    ["PHYSICAL_COUNT", "Contagem fisica"],
    ["DOCUMENTED_CORRECTION", "Correcao documentada"],
  ],
};

const newAdjustmentForm = () => ({
  itemId: "",
  direction: "OUT",
  reasonCode: "WASTE",
  quantityAtomic: "",
  inventoryValue: "",
  lotCode: "",
  expiresOn: "",
});

const capabilityLabels = {
  purchasable: "Compra",
  producible: "Produção",
  sellable: "Venda",
};

const activeCapabilities = (capabilities) =>
  Object.entries(capabilityLabels).filter(([key]) => capabilities?.[key]);

const InventoryPage = () => {
  const [balances, setBalances] = useState([]);
  const [units, setUnits] = useState([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(false);
  const [postingAdjustment, setPostingAdjustment] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);
  const [adjustmentForm, setAdjustmentForm] = useState(newAdjustmentForm);

  const loadBalances = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const [page, loadedUnits] = await Promise.all([
        inventoryGateway.listInventoryBalances({
          includeArchived: false,
          search: search.trim() ? search.trim() : null,
          pageSize: 100,
        }),
        referenceDataGateway.listMeasurementUnits(),
      ]);
      setBalances(page.items ?? []);
      setUnits(loadedUnits);
      setAdjustmentForm((current) => ({
        ...current,
        itemId: current.itemId || String(page.items?.[0]?.itemId ?? ""),
      }));
    } catch (err) {
      console.error("Erro ao carregar saldos de estoque:", err);
      setError(err.message || "Erro ao carregar saldos de estoque.");
    } finally {
      setLoading(false);
    }
  }, [search]);

  useEffect(() => {
    void loadBalances();
  }, [loadBalances]);

  const totals = useMemo(() => {
    const inventoryValueMicro = balances.reduce((acc, item) => acc + item.inventoryValueMicro, 0);
    const lowStockItems = balances.filter(
      (item) =>
        item.reorderQuantityAtomic != null && item.quantityAtomic <= item.reorderQuantityAtomic,
    ).length;

    return {
      inventoryValueMicro,
      lowStockItems,
    };
  }, [balances]);

  const selectedAdjustmentBalance = useMemo(
    () => balances.find((item) => String(item.itemId) === adjustmentForm.itemId) ?? null,
    [adjustmentForm.itemId, balances],
  );

  const selectedAdjustmentUnit = useMemo(
    () => units.find((unit) => unit.code === selectedAdjustmentBalance?.baseUnitCode) ?? null,
    [selectedAdjustmentBalance?.baseUnitCode, units],
  );

  const updateAdjustmentDirection = (direction) => {
    setAdjustmentForm((current) => ({
      ...current,
      direction,
      reasonCode: adjustmentReasons[direction][0][0],
      inventoryValue: direction === "OUT" ? "" : current.inventoryValue,
      lotCode: direction === "OUT" ? "" : current.lotCode,
      expiresOn: direction === "OUT" ? "" : current.expiresOn,
    }));
  };

  const postAdjustment = async () => {
    if (postingAdjustment) return;
    if (!selectedAdjustmentBalance || !selectedAdjustmentUnit) {
      setError("Selecione um item valido para ajustar.");
      return;
    }
    const quantityAtomic = parseInteger(adjustmentForm.quantityAtomic);
    if (!quantityAtomic || quantityAtomic <= 0) {
      setError("Informe uma quantidade atomica positiva.");
      return;
    }
    const inventoryValueMicro =
      adjustmentForm.direction === "IN"
        ? parseInventoryValueMicro(adjustmentForm.inventoryValue)
        : undefined;
    if (
      adjustmentForm.direction === "IN" &&
      inventoryValueMicro === undefined &&
      selectedAdjustmentBalance.quantityAtomic === 0
    ) {
      setError("Informe o valor de estoque para entrada quando o item nao tem saldo.");
      return;
    }

    setPostingAdjustment(true);
    setError(null);
    setSuccess(null);
    try {
      const posted = await adjustmentGateway.postAdjustment({
        idempotencyKey: `adjustment-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`,
        occurredOn: new Date().toISOString().slice(0, 10),
        reasonCode: adjustmentForm.reasonCode,
        lines: [
          {
            itemId: selectedAdjustmentBalance.itemId,
            direction: adjustmentForm.direction,
            quantityAtomic,
            enteredUnitCode: selectedAdjustmentBalance.baseUnitCode,
            conversionNumeratorAtomic: selectedAdjustmentUnit.numeratorAtomic,
            conversionDenominator: selectedAdjustmentUnit.denominator,
            inventoryValueMicro,
            lotCode:
              adjustmentForm.direction === "IN" ? adjustmentForm.lotCode.trim() || null : null,
            expiresOn:
              adjustmentForm.direction === "IN" && adjustmentForm.expiresOn
                ? adjustmentForm.expiresOn
                : null,
          },
        ],
      });
      setSuccess(`Ajuste #${posted.id} postado.`);
      setAdjustmentForm((current) => ({
        ...newAdjustmentForm(),
        itemId: current.itemId,
        direction: current.direction,
        reasonCode: adjustmentReasons[current.direction][0][0],
      }));
      await loadBalances();
    } catch (err) {
      console.error("Erro ao postar ajuste:", err);
      setError(err.message || "Erro ao postar ajuste.");
    } finally {
      setPostingAdjustment(false);
    }
  };

  return (
    <>
      <header className="bg-white border-b border-slate-200">
        <div className="max-w-7xl mx-auto px-6 py-8">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
            <div>
              <h1 className="text-3xl font-bold text-slate-900">Estoque</h1>
              <p className="text-slate-600 mt-2">
                Saldos reais do inventário V2, calculados a partir do ledger local.
              </p>
            </div>
            <button
              onClick={() => void loadBalances()}
              disabled={loading}
              className="inline-flex items-center justify-center gap-2 rounded-lg bg-pink-600 px-5 py-3 text-white transition-all hover:bg-pink-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              <ArrowsClockwise size={20} className={loading ? "animate-spin" : ""} />
              Atualizar
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-6 py-8">
        <div className="mb-6 rounded-2xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-900">
          O cadastro de itens fica em <strong>Produtos</strong>. Entrada de estoque será feita pelo
          fluxo de <strong>compras/postagem</strong>; esta tela agora apenas lê os saldos V2 reais.
        </div>

        {error && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 flex items-start gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700"
          >
            <span>{error}</span>
            <button
              onClick={() => setError(null)}
              className="ml-auto text-red-600 hover:text-red-800"
            >
              ×
            </button>
          </motion.div>
        )}

        {success && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            className="mb-6 flex items-start gap-3 rounded-lg border border-green-200 bg-green-50 p-4 text-sm text-green-700"
          >
            <span>{success}</span>
            <button
              onClick={() => setSuccess(null)}
              className="ml-auto text-green-600 hover:text-green-800"
            >
              ×
            </button>
          </motion.div>
        )}

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8 grid grid-cols-1 gap-6 md:grid-cols-3"
        >
          <div className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <h3 className="text-sm font-medium text-slate-600">Itens com saldo</h3>
            <p className="mt-2 text-3xl font-bold text-slate-900">{balances.length}</p>
          </div>
          <div className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <h3 className="text-sm font-medium text-slate-600">Valor em estoque</h3>
            <p className="mt-2 text-3xl font-bold text-green-600">
              {formatInventoryValue(totals.inventoryValueMicro)}
            </p>
          </div>
          <div className="rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
            <div className="flex items-center gap-2">
              <Warning size={20} className="text-orange-600" />
              <h3 className="text-sm font-medium text-slate-600">Abaixo do ponto de reposição</h3>
            </div>
            <p className="mt-2 text-3xl font-bold text-orange-600">{totals.lowStockItems}</p>
          </div>
        </motion.div>

        <div className="mb-6 flex flex-col gap-3 rounded-2xl border border-slate-100 bg-white p-4 shadow-sm md:flex-row md:items-center">
          <input
            type="search"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Buscar item..."
            className="min-w-0 flex-1 rounded-lg border border-slate-300 px-4 py-2 outline-none transition focus:border-pink-500 focus:ring-2 focus:ring-pink-100"
          />
          <button
            onClick={() => void loadBalances()}
            disabled={loading}
            className="rounded-lg border border-slate-300 px-4 py-2 text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:bg-slate-100"
          >
            Buscar
          </button>
        </div>

        <section className="mb-8 rounded-2xl border border-slate-100 bg-white p-6 shadow-sm">
          <div className="mb-5">
            <h2 className="text-lg font-semibold text-slate-900">Ajustar estoque</h2>
            <p className="mt-1 text-sm text-slate-600">
              Posta um documento V2 real: entrada cria lote; saida consome lotes por FEFO.
            </p>
          </div>

          <div className="grid gap-4 lg:grid-cols-4">
            <label className="block text-sm font-semibold text-slate-700">
              Item
              <select
                value={adjustmentForm.itemId}
                onChange={(event) =>
                  setAdjustmentForm({ ...adjustmentForm, itemId: event.target.value })
                }
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
              >
                <option value="">Selecione</option>
                {balances.map((item) => (
                  <option key={item.itemId} value={item.itemId}>
                    {item.itemName}
                  </option>
                ))}
              </select>
            </label>

            <label className="block text-sm font-semibold text-slate-700">
              Direcao
              <select
                value={adjustmentForm.direction}
                onChange={(event) => updateAdjustmentDirection(event.target.value)}
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
              >
                <option value="OUT">Saida</option>
                <option value="IN">Entrada</option>
              </select>
            </label>

            <label className="block text-sm font-semibold text-slate-700">
              Razao
              <select
                value={adjustmentForm.reasonCode}
                onChange={(event) =>
                  setAdjustmentForm({ ...adjustmentForm, reasonCode: event.target.value })
                }
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
              >
                {adjustmentReasons[adjustmentForm.direction].map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>

            <label className="block text-sm font-semibold text-slate-700">
              Quantidade atomica
              <input
                value={adjustmentForm.quantityAtomic}
                onChange={(event) =>
                  setAdjustmentForm({ ...adjustmentForm, quantityAtomic: event.target.value })
                }
                className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                placeholder="250"
              />
            </label>
          </div>

          {adjustmentForm.direction === "IN" && (
            <div className="mt-4 grid gap-4 lg:grid-cols-3">
              <label className="block text-sm font-semibold text-slate-700">
                Valor de estoque
                <input
                  value={adjustmentForm.inventoryValue}
                  onChange={(event) =>
                    setAdjustmentForm({ ...adjustmentForm, inventoryValue: event.target.value })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="25,00"
                />
              </label>
              <label className="block text-sm font-semibold text-slate-700">
                Lote
                <input
                  value={adjustmentForm.lotCode}
                  onChange={(event) =>
                    setAdjustmentForm({ ...adjustmentForm, lotCode: event.target.value })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="AJUSTE-001"
                />
              </label>
              <label className="block text-sm font-semibold text-slate-700">
                Validade
                <input
                  type="date"
                  value={adjustmentForm.expiresOn}
                  onChange={(event) =>
                    setAdjustmentForm({ ...adjustmentForm, expiresOn: event.target.value })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                />
              </label>
            </div>
          )}

          <div className="mt-5 flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
            <p className="text-sm text-slate-600">
              {selectedAdjustmentBalance
                ? `Saldo atual: ${formatQuantity(
                    selectedAdjustmentBalance.quantityAtomic,
                    selectedAdjustmentBalance.baseUnitCode,
                  )}`
                : "Selecione um item."}
            </p>
            <button
              type="button"
              onClick={() => void postAdjustment()}
              disabled={postingAdjustment || loading || !adjustmentForm.itemId}
              className="rounded-xl bg-pink-600 px-5 py-3 font-semibold text-white transition hover:bg-pink-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              {postingAdjustment ? "Postando..." : "Postar ajuste"}
            </button>
          </div>
        </section>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="overflow-hidden rounded-2xl border border-slate-100 bg-white shadow-sm"
        >
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="border-b border-slate-200 bg-slate-50">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">Item</th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Saldo atual
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Valor
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Reposição
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Capacidades
                  </th>
                  <th className="px-6 py-4 text-left text-sm font-semibold text-slate-900">
                    Último documento
                  </th>
                </tr>
              </thead>
              <tbody>
                {loading ? (
                  <tr>
                    <td colSpan="6" className="px-6 py-12 text-center text-slate-500">
                      Carregando saldos...
                    </td>
                  </tr>
                ) : balances.length === 0 ? (
                  <tr>
                    <td colSpan="6" className="px-6 py-12 text-center text-slate-500">
                      <Package size={32} className="mx-auto mb-3 text-slate-400" />
                      Nenhum saldo de estoque encontrado.
                    </td>
                  </tr>
                ) : (
                  balances.map((item) => {
                    const lowStock =
                      item.reorderQuantityAtomic != null &&
                      item.quantityAtomic <= item.reorderQuantityAtomic;
                    const capabilities = activeCapabilities(item.capabilities);

                    return (
                      <tr
                        key={item.itemId}
                        className={`border-b border-slate-100 hover:bg-slate-50 ${
                          lowStock ? "bg-orange-50" : ""
                        }`}
                      >
                        <td className="px-6 py-4">
                          <div className="font-medium text-slate-900">{item.itemName}</div>
                          <div className="text-xs text-slate-500">#{item.itemId}</div>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-700">
                          {formatQuantity(item.quantityAtomic, item.baseUnitCode)}
                        </td>
                        <td className="px-6 py-4 text-sm font-semibold text-slate-900">
                          {formatInventoryValue(item.inventoryValueMicro)}
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-700">
                          {item.reorderQuantityAtomic == null ? (
                            <span className="text-slate-400">Não definido</span>
                          ) : (
                            formatQuantity(item.reorderQuantityAtomic, item.baseUnitCode)
                          )}
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex flex-wrap gap-2">
                            {capabilities.length === 0 ? (
                              <span className="text-sm text-slate-400">Nenhuma</span>
                            ) : (
                              capabilities.map(([key, label]) => (
                                <span
                                  key={key}
                                  className="rounded-full bg-pink-50 px-2 py-1 text-xs font-medium text-pink-700"
                                >
                                  {label}
                                </span>
                              ))
                            )}
                          </div>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-700">
                          {item.lastDocumentId == null ? (
                            <span className="text-slate-400">Nenhum</span>
                          ) : (
                            `#${item.lastDocumentId}`
                          )}
                        </td>
                      </tr>
                    );
                  })
                )}
              </tbody>
            </table>
          </div>
        </motion.div>
      </main>
    </>
  );
};

export default InventoryPage;
