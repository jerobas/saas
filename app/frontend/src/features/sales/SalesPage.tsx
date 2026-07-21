import { ArrowClockwise, Plus, ShoppingCart, Tag, Trash } from "@phosphor-icons/react";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import ConversionPreview from "../../components/ConversionPreview";
import {
  catalogGateway,
  counterpartyGateway,
  inventoryGateway,
  type CounterpartyResponse,
  type InventoryBalanceResponse,
  type ItemResponse,
  type ItemSummaryResponse,
  type LotResponse,
  type PackagingResponse,
  saleGateway,
  type SaleDocumentResponse,
  type SaleReason,
} from "../../gateways/desktopBridge";

interface SaleLineFormState {
  key: string;
  itemId: string;
  quantityAtomic: string;
  enteredUnitCode: string;
  enteredPackagingName: string;
  conversionNumeratorAtomic: string;
  conversionDenominator: string;
  commercialTotal: string;
  lotId: string;
}

interface SaleFormState {
  counterpartyId: string;
  occurredOn: string;
  reasonCode: "" | SaleReason;
  notes: string;
  lines: SaleLineFormState[];
}

interface SaleLineContext {
  item: ItemResponse;
  eligibleLots: LotResponse[];
  balance: InventoryBalanceResponse | null;
}

const todayISO = () => new Date().toISOString().slice(0, 10);

let nextLineKey = 0;

const emptyLine = (): SaleLineFormState => ({
  key: `sale-line-${nextLineKey++}`,
  itemId: "",
  quantityAtomic: "",
  enteredUnitCode: "",
  enteredPackagingName: "",
  conversionNumeratorAtomic: "",
  conversionDenominator: "1",
  commercialTotal: "",
  lotId: "",
});

const emptyForm = (): SaleFormState => ({
  counterpartyId: "",
  occurredOn: todayISO(),
  reasonCode: "",
  notes: "",
  lines: [],
});

const optionalText = (value: string) => {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
};

const parseInteger = (value: string) => {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const parseMoneyMinor = (value: string) => {
  const normalized = value.trim().replace(",", ".");
  if (normalized.length === 0) return undefined;
  const parsed = Number.parseFloat(normalized);
  if (!Number.isFinite(parsed)) return undefined;
  return Math.round(parsed * 100);
};

const formatMoneyMinor = (minor: number) =>
  new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(minor / 100);

const formatInventoryMicro = (micro: number) =>
  new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(micro / 1_000_000);

const buildIdempotencyKey = () => `sale-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

const messageFromError = (error: unknown, fallback: string) =>
  error instanceof Error ? error.message : fallback;

const saleLabel = (sale: SaleDocumentResponse) => `#${sale.id} / seq ${sale.postingSequence}`;

const lotLabel = (lot: LotResponse) => {
  const code = lot.lotCode ? `${lot.lotCode} · ` : "";
  const expiry = lot.expiresOn ? ` · vence ${lot.expiresOn}` : "";
  return `${code}${lot.availableQuantityAtomic} atomicos${expiry}`;
};

function SalesPage() {
  const [items, setItems] = useState<ItemSummaryResponse[]>([]);
  const [customers, setCustomers] = useState<CounterpartyResponse[]>([]);
  const [lineContexts, setLineContexts] = useState<Record<number, SaleLineContext>>({});
  const [sales, setSales] = useState<SaleDocumentResponse[]>([]);
  const [form, setForm] = useState<SaleFormState>(() => emptyForm());
  const [itemToAddId, setItemToAddId] = useState("");
  const [loading, setLoading] = useState(true);
  const [addingItem, setAddingItem] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const itemNames = useMemo(
    () => new Map(items.map((item) => [item.id, item.name] as const)),
    [items],
  );

  const customerNames = useMemo(
    () => new Map(customers.map((customer) => [customer.id, customer.name] as const)),
    [customers],
  );

  const documentTotalMinor = useMemo(
    () =>
      form.lines.reduce((total, line) => total + (parseMoneyMinor(line.commercialTotal) ?? 0), 0),
    [form.lines],
  );

  const loadSales = useCallback(async () => {
    const salePage = await saleGateway.listSales({ pageSize: 25 });
    setSales(salePage.items);
  }, []);

  const loadItemContext = useCallback(async (itemId: number, occurredOn: string) => {
    const [detail, lots] = await Promise.all([
      catalogGateway.getItem(itemId),
      inventoryGateway.listEligibleFefoLots(itemId, occurredOn),
    ]);

    let currentBalance: InventoryBalanceResponse | null;
    try {
      currentBalance = await inventoryGateway.getInventoryBalance(itemId);
    } catch {
      currentBalance = null;
    }

    return { item: detail, eligibleLots: lots, balance: currentBalance } satisfies SaleLineContext;
  }, []);

  const loadPage = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const [itemPage, customerPage, salePage] = await Promise.all([
        catalogGateway.listItems({
          archiveFilter: "ACTIVE",
          requireCapabilities: { purchasable: false, producible: false, sellable: true },
          pageSize: 100,
        }),
        counterpartyGateway.listCounterparties({
          archiveFilter: "ACTIVE",
          role: "CUSTOMER",
          pageSize: 100,
        }),
        saleGateway.listSales({ pageSize: 25 }),
      ]);
      setItems(itemPage.items);
      setCustomers(customerPage.items);
      setSales(salePage.items);
      setItemToAddId((current) => current || String(itemPage.items[0]?.id ?? ""));
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar vendas."),
      });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  const addItem = async () => {
    const itemId = parseInteger(itemToAddId);
    if (!itemId || addingItem) return;
    if (form.lines.some((line) => parseInteger(line.itemId) === itemId)) {
      setMessage({ type: "error", text: "Este item ja esta no carrinho." });
      return;
    }
    setAddingItem(true);
    setMessage(null);
    try {
      const context = await loadItemContext(itemId, form.occurredOn);
      setLineContexts((current) => ({ ...current, [itemId]: context }));
      setForm((current) => ({
        ...current,
        lines: [...current.lines, lineForItem(emptyLine(), context.item)],
      }));
      const nextItem = items.find(
        (item) =>
          item.id !== itemId && !form.lines.some((line) => parseInteger(line.itemId) === item.id),
      );
      setItemToAddId(String(nextItem?.id ?? ""));
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar o item vendavel."),
      });
    } finally {
      setAddingItem(false);
    }
  };

  const updateLine = (key: string, patch: Partial<SaleLineFormState>) => {
    setForm((current) => ({
      ...current,
      lines: current.lines.map((line) => (line.key === key ? { ...line, ...patch } : line)),
    }));
  };

  const removeLine = (key: string) => {
    setForm((current) => ({
      ...current,
      lines: current.lines.filter((line) => line.key !== key),
    }));
  };

  const changeOccurredOn = async (occurredOn: string) => {
    setForm((current) => ({ ...current, occurredOn }));
    if (form.lines.length === 0) return;
    try {
      const refreshed = await Promise.all(
        form.lines.map(async (line) => {
          const itemId = parseInteger(line.itemId);
          if (!itemId) return null;
          const eligibleLots = await inventoryGateway.listEligibleFefoLots(itemId, occurredOn);
          return { itemId, eligibleLots };
        }),
      );
      setLineContexts((current) => {
        const next = { ...current };
        refreshed.forEach((entry) => {
          if (entry && next[entry.itemId]) {
            next[entry.itemId] = { ...next[entry.itemId], eligibleLots: entry.eligibleLots };
          }
        });
        return next;
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar lotes elegiveis."),
      });
    }
  };

  const selectPackaging = (line: SaleLineFormState, packagingID: string) => {
    const itemId = parseInteger(line.itemId);
    const context = itemId ? lineContexts[itemId] : undefined;
    if (!context) return;
    if (packagingID === "base") {
      updateLine(line.key, lineForItem(line, context.item));
      return;
    }
    const packaging = activePackagings(context.item).find(
      (current) => String(current.id) === packagingID,
    );
    if (!packaging) return;
    updateLine(line.key, lineForPackaging(line, context.item, packaging));
  };

  const postSale = async () => {
    if (saving) return;
    const counterpartyId = parseInteger(form.counterpartyId);
    if (form.lines.length === 0) {
      setMessage({ type: "error", text: "Adicione ao menos um item ao carrinho." });
      return;
    }

    const lines = form.lines.map((line) => ({
      itemId: parseInteger(line.itemId),
      quantityAtomic: parseInteger(line.quantityAtomic),
      enteredUnitCode: line.enteredUnitCode,
      enteredPackagingName: optionalText(line.enteredPackagingName),
      conversionNumeratorAtomic: parseInteger(line.conversionNumeratorAtomic),
      conversionDenominator: parseInteger(line.conversionDenominator),
      commercialTotalMinor: parseMoneyMinor(line.commercialTotal),
      lotId: parseInteger(line.lotId),
    }));

    const invalidLineIndex = lines.findIndex(
      (line) =>
        !line.itemId ||
        !line.quantityAtomic ||
        !line.conversionNumeratorAtomic ||
        !line.conversionDenominator ||
        line.commercialTotalMinor === undefined,
    );
    if (invalidLineIndex >= 0) {
      setMessage({
        type: "error",
        text: `Preencha quantidade, unidade e total comercial da linha ${invalidLineIndex + 1}.`,
      });
      return;
    }
    const hasZeroValueLine = lines.some((line) => line.commercialTotalMinor === 0);
    if (hasZeroValueLine && form.reasonCode.length === 0) {
      setMessage({
        type: "error",
        text: "Venda com total zero precisa da razao PROMOTION ou SAMPLE.",
      });
      return;
    }
    const reasonCode: SaleReason | undefined = hasZeroValueLine
      ? (form.reasonCode as SaleReason)
      : undefined;

    setSaving(true);
    setMessage(null);
    try {
      const posted = await saleGateway.postSale({
        idempotencyKey: buildIdempotencyKey(),
        counterpartyId,
        occurredOn: form.occurredOn,
        reasonCode,
        notes: optionalText(form.notes),
        lines: lines.map((line) => ({
          itemId: line.itemId!,
          quantityAtomic: line.quantityAtomic!,
          enteredUnitCode: line.enteredUnitCode,
          enteredPackagingName: line.enteredPackagingName,
          conversionNumeratorAtomic: line.conversionNumeratorAtomic!,
          conversionDenominator: line.conversionDenominator!,
          commercialTotalMinor: line.commercialTotalMinor!,
          lotId: line.lotId,
        })),
      });
      setMessage({ type: "success", text: `Venda ${saleLabel(posted)} postada.` });
      setForm((current) => ({
        ...current,
        reasonCode: "",
        notes: "",
        lines: [],
      }));
      setLineContexts({});
      setItemToAddId(String(items[0]?.id ?? ""));
      await loadSales();
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel postar a venda."),
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <main className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-7xl px-6 py-8">
          <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-wide text-pink-600">
                Vendas V2
              </p>
              <h1 className="mt-2 text-3xl font-bold text-slate-950">Postar venda</h1>
              <p className="mt-2 max-w-2xl text-slate-600">
                Fluxo minimo real: vende item vendavel, aloca estoque por FEFO ou lote manual e
                calcula COGS pelo ledger V2.
              </p>
            </div>
            <button
              type="button"
              onClick={() => void loadPage()}
              disabled={loading || saving}
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-3 font-semibold text-slate-700 shadow-sm hover:bg-slate-50 disabled:opacity-50"
            >
              <ArrowClockwise size={18} />
              Recarregar
            </button>
          </div>
        </div>
      </header>

      <section className="mx-auto max-w-7xl space-y-6 px-6 py-8">
        <div className="space-y-6">
          {message && (
            <div
              className={`rounded-2xl border p-4 text-sm font-medium ${
                message.type === "success"
                  ? "border-green-200 bg-green-50 text-green-800"
                  : "border-red-200 bg-red-50 text-red-800"
              }`}
            >
              {message.text}
            </div>
          )}

          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <div className="flex items-center gap-3">
              <div className="rounded-2xl bg-pink-50 p-3 text-pink-600">
                <ShoppingCart size={24} />
              </div>
              <div>
                <h2 className="text-lg font-bold text-slate-950">Nova venda</h2>
                <p className="text-sm text-slate-600">
                  Monte um carrinho simples e poste todas as linhas no mesmo documento.
                </p>
              </div>
            </div>

            <div className="mt-5 space-y-4">
              <label className="block text-sm font-semibold text-slate-700">
                Cliente
                <select
                  value={form.counterpartyId}
                  onChange={(event) => setForm({ ...form, counterpartyId: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  <option value="">Sem cliente</option>
                  {customers.map((customer) => (
                    <option key={customer.id} value={customer.id}>
                      {customer.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Data de ocorrencia
                <input
                  type="date"
                  value={form.occurredOn}
                  onChange={(event) => void changeOccurredOn(event.target.value)}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                />
              </label>

              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                <label className="block text-sm font-semibold text-slate-700">
                  Item para adicionar
                  <div className="mt-2 flex flex-col gap-2 sm:flex-row">
                    <select
                      value={itemToAddId}
                      onChange={(event) => setItemToAddId(event.target.value)}
                      className="min-w-0 flex-1 rounded-xl border border-slate-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    >
                      <option value="">Selecione</option>
                      {items
                        .filter(
                          (item) =>
                            !form.lines.some((line) => parseInteger(line.itemId) === item.id),
                        )
                        .map((item) => (
                          <option key={item.id} value={item.id}>
                            {item.name}
                          </option>
                        ))}
                    </select>
                    <button
                      type="button"
                      onClick={() => void addItem()}
                      disabled={addingItem || itemToAddId.length === 0}
                      className="inline-flex items-center justify-center gap-2 rounded-xl bg-slate-900 px-4 py-2 font-semibold text-white hover:bg-slate-800 disabled:bg-slate-300"
                    >
                      <Plus size={18} />
                      {addingItem ? "Adicionando..." : "Adicionar item"}
                    </button>
                  </div>
                </label>
              </div>

              <div className="space-y-4">
                {form.lines.length === 0 ? (
                  <div className="rounded-2xl border border-dashed border-slate-300 p-8 text-center text-sm text-slate-500">
                    O carrinho esta vazio. Adicione um item vendavel para comecar.
                  </div>
                ) : (
                  form.lines.map((line, index) => {
                    const itemId = parseInteger(line.itemId);
                    const context = itemId ? lineContexts[itemId] : undefined;
                    const packagings = context ? activePackagings(context.item) : [];
                    return (
                      <article key={line.key} className="rounded-2xl border border-slate-200 p-5">
                        <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                          <div>
                            <p className="text-xs font-semibold uppercase tracking-wide text-pink-600">
                              Linha {index + 1}
                            </p>
                            <h3 className="mt-1 text-lg font-bold text-slate-950">
                              {context?.item.name ?? "Carregando item..."}
                            </h3>
                            <p className="mt-1 text-sm text-slate-500">
                              Saldo: <strong>{context?.balance?.quantityAtomic ?? 0}</strong>{" "}
                              atomicos
                              {context?.item.defaultSalePrice == null
                                ? ""
                                : ` · preco padrao ${formatMoneyMinor(context.item.defaultSalePrice)}`}
                            </p>
                          </div>
                          <button
                            type="button"
                            onClick={() => removeLine(line.key)}
                            aria-label={`Remover linha ${index + 1}`}
                            className="inline-flex items-center gap-2 self-start rounded-xl px-3 py-2 text-sm font-semibold text-red-600 hover:bg-red-50"
                          >
                            <Trash size={17} />
                            Remover
                          </button>
                        </div>

                        <div className="mt-4 grid gap-4 md:grid-cols-2 xl:grid-cols-4">
                          <label className="block text-sm font-semibold text-slate-700">
                            Quantidade atomica
                            <input
                              aria-label={`Quantidade atomica da linha ${index + 1}`}
                              value={line.quantityAtomic}
                              onChange={(event) =>
                                updateLine(line.key, { quantityAtomic: event.target.value })
                              }
                              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                              placeholder="100"
                            />
                          </label>

                          <label className="block text-sm font-semibold text-slate-700">
                            Unidade / embalagem
                            <select
                              aria-label={`Unidade da linha ${index + 1}`}
                              value={selectedPackagingValue(line, packagings)}
                              onChange={(event) => selectPackaging(line, event.target.value)}
                              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                            >
                              <option value="base">
                                Unidade base {context ? `(${context.item.baseUnit.symbol})` : ""}
                              </option>
                              {packagings.map((packaging) => (
                                <option key={packaging.id} value={packaging.id}>
                                  {packaging.name}
                                </option>
                              ))}
                            </select>
                          </label>

                          <label className="block text-sm font-semibold text-slate-700">
                            Total comercial
                            <input
                              aria-label={`Total comercial da linha ${index + 1}`}
                              value={line.commercialTotal}
                              onChange={(event) =>
                                updateLine(line.key, { commercialTotal: event.target.value })
                              }
                              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                              placeholder="25,00"
                            />
                          </label>

                          <label className="block text-sm font-semibold text-slate-700">
                            Lote de saida
                            <select
                              aria-label={`Lote de saida da linha ${index + 1}`}
                              value={line.lotId}
                              onChange={(event) =>
                                updateLine(line.key, { lotId: event.target.value })
                              }
                              className="mt-2 w-full rounded-xl border border-slate-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                            >
                              <option value="">FEFO automatico</option>
                              {context?.eligibleLots.map((lot) => (
                                <option key={lot.id} value={lot.id}>
                                  {lotLabel(lot)}
                                </option>
                              ))}
                            </select>
                          </label>
                        </div>

                        {context && line.enteredPackagingName && (
                          <ConversionPreview
                            label={`1 ${line.enteredPackagingName}`}
                            numeratorAtomic={line.conversionNumeratorAtomic}
                            denominator={line.conversionDenominator}
                            baseUnit={context.item.baseUnit}
                            className="mt-3 rounded-xl bg-slate-100 px-3 py-2 text-sm text-slate-600"
                          />
                        )}
                      </article>
                    );
                  })
                )}
              </div>

              <label className="block text-sm font-semibold text-slate-700">
                Razao para linha sem valor
                <select
                  value={form.reasonCode}
                  onChange={(event) =>
                    setForm({ ...form, reasonCode: event.target.value as "" | SaleReason })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  <option value="">Venda normal</option>
                  <option value="PROMOTION">PROMOTION</option>
                  <option value="SAMPLE">SAMPLE</option>
                </select>
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Observacoes
                <textarea
                  value={form.notes}
                  onChange={(event) => setForm({ ...form, notes: event.target.value })}
                  className="mt-2 min-h-20 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                />
              </label>

              <button
                type="button"
                onClick={() => void postSale()}
                disabled={saving || loading || form.lines.length === 0}
                className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
              >
                <Tag size={18} />
                {saving ? "Postando..." : "Postar venda"}
              </button>
              <p className="text-right text-sm font-semibold text-slate-600">
                {form.lines.length} {form.lines.length === 1 ? "linha" : "linhas"} · Total{" "}
                <strong className="text-lg text-slate-950">
                  {formatMoneyMinor(documentTotalMinor)}
                </strong>
              </p>
            </div>
          </div>
        </div>

        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-3">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Itens vendaveis</p>
              <p className="mt-2 text-3xl font-bold text-slate-950">{items.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Clientes ativos</p>
              <p className="mt-2 text-3xl font-bold text-blue-700">{customers.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Vendas recentes</p>
              <p className="mt-2 text-3xl font-bold text-green-700">{sales.length}</p>
            </div>
          </div>

          <section className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-200 p-6">
              <h2 className="text-xl font-bold text-slate-950">Vendas postadas</h2>
              <p className="mt-1 text-sm text-slate-500">
                Historico persistente carregado dos documentos SALE.
              </p>
            </div>
            {loading ? (
              <p className="p-6 text-sm text-slate-500">Carregando vendas...</p>
            ) : sales.length === 0 ? (
              <p className="p-6 text-sm text-slate-500">Nenhuma venda postada ainda.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="bg-slate-50 text-slate-600">
                    <tr>
                      <th className="px-6 py-3 font-semibold">Documento</th>
                      <th className="px-6 py-3 font-semibold">Data</th>
                      <th className="px-6 py-3 font-semibold">Cliente</th>
                      <th className="px-6 py-3 font-semibold">Linhas</th>
                      <th className="px-6 py-3 font-semibold">Receita</th>
                      <th className="px-6 py-3 font-semibold">COGS</th>
                      <th className="px-6 py-3 font-semibold">Razao</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {sales.map((sale) => {
                      const revenue = sale.lines.reduce(
                        (acc, line) => acc + line.commercialTotalMinor,
                        0,
                      );
                      const cogs = sale.lines.reduce(
                        (acc, line) => acc + line.inventoryValueMicro,
                        0,
                      );
                      return (
                        <Fragment key={sale.id}>
                          <tr className="hover:bg-slate-50">
                            <td className="px-6 py-4">
                              <p className="font-semibold text-slate-950">{saleLabel(sale)}</p>
                              <p className="text-xs text-slate-500">{sale.idempotencyKey}</p>
                            </td>
                            <td className="px-6 py-4 text-slate-700">{sale.occurredOn}</td>
                            <td className="px-6 py-4 text-slate-700">
                              {sale.counterpartyId == null
                                ? "Sem cliente"
                                : (customerNames.get(sale.counterpartyId) ??
                                  `cliente #${sale.counterpartyId}`)}
                            </td>
                            <td className="px-6 py-4 text-slate-700">
                              {sale.lines.map((line) => (
                                <div key={line.id}>
                                  {itemNames.get(line.itemId) ?? `item #${line.itemId}`}:{" "}
                                  {line.quantityAtomic} {line.enteredUnitCode}
                                </div>
                              ))}
                            </td>
                            <td className="px-6 py-4 font-semibold text-slate-900">
                              {formatMoneyMinor(revenue)}
                            </td>
                            <td className="px-6 py-4 font-semibold text-slate-900">
                              {formatInventoryMicro(cogs)}
                            </td>
                            <td className="px-6 py-4 text-slate-700">{sale.reasonCode ?? "-"}</td>
                          </tr>
                          <tr className="bg-slate-50/70">
                            <td colSpan={7} className="px-6 pb-5 pt-0">
                              <div className="rounded-2xl border border-slate-200 bg-white p-4">
                                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                                  Detalhe da venda: alocacoes e custo
                                </p>
                                <div className="mt-3 space-y-3">
                                  {sale.lines.map((line) => (
                                    <div
                                      key={line.id}
                                      className="grid gap-3 rounded-xl bg-slate-50 p-3 md:grid-cols-[1fr_1fr_1fr]"
                                    >
                                      <div>
                                        <p className="font-semibold text-slate-900">
                                          {itemNames.get(line.itemId) ?? `item #${line.itemId}`}
                                        </p>
                                        <p className="text-xs text-slate-500">
                                          linha #{line.id} · {line.quantityAtomic}{" "}
                                          {line.enteredUnitCode}
                                        </p>
                                      </div>
                                      <div className="text-sm text-slate-700">
                                        <p>
                                          Receita:{" "}
                                          <strong>
                                            {formatMoneyMinor(line.commercialTotalMinor)}
                                          </strong>
                                        </p>
                                        <p>
                                          COGS:{" "}
                                          <strong>
                                            {formatInventoryMicro(line.inventoryValueMicro)}
                                          </strong>
                                        </p>
                                      </div>
                                      <div className="text-sm text-slate-700">
                                        <p className="font-semibold text-slate-900">
                                          Lotes consumidos
                                        </p>
                                        {line.allocations.length === 0 ? (
                                          <p className="text-slate-500">Sem alocacoes.</p>
                                        ) : (
                                          line.allocations.map((allocation) => (
                                            <p key={allocation.id}>
                                              lote #{allocation.lotId}:{" "}
                                              <strong>{allocation.quantityAtomic}</strong> atomicos
                                            </p>
                                          ))
                                        )}
                                      </div>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            </td>
                          </tr>
                        </Fragment>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </section>
        </div>
      </section>
    </main>
  );
}

function lineForItem(line: SaleLineFormState, item: ItemResponse): SaleLineFormState {
  return {
    ...line,
    itemId: String(item.id),
    enteredUnitCode: item.baseUnit.code,
    enteredPackagingName: "",
    conversionNumeratorAtomic: String(item.baseUnit.numeratorAtomic),
    conversionDenominator: String(item.baseUnit.denominator),
  };
}

function lineForPackaging(
  line: SaleLineFormState,
  item: ItemResponse,
  packaging: PackagingResponse,
): SaleLineFormState {
  return {
    ...line,
    itemId: String(item.id),
    enteredUnitCode: packaging.enteredUnitCode,
    enteredPackagingName: packaging.name,
    conversionNumeratorAtomic: String(packaging.conversionNumeratorAtomic),
    conversionDenominator: String(packaging.conversionDenominator),
  };
}

function activePackagings(item: ItemResponse) {
  return item.packagings.filter(
    (packaging) => packaging.archivedAtMs === null || packaging.archivedAtMs === undefined,
  );
}

function selectedPackagingValue(line: SaleLineFormState, packagings: PackagingResponse[]) {
  const selected = packagings.find(
    (packaging) =>
      packaging.name === line.enteredPackagingName &&
      packaging.enteredUnitCode === line.enteredUnitCode &&
      String(packaging.conversionNumeratorAtomic) === line.conversionNumeratorAtomic &&
      String(packaging.conversionDenominator) === line.conversionDenominator,
  );
  return selected ? String(selected.id) : "base";
}

export default SalesPage;
