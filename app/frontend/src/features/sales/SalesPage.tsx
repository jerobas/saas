import { ArrowClockwise, ShoppingCart, Tag } from "@phosphor-icons/react";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
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
  line: SaleLineFormState;
}

const todayISO = () => new Date().toISOString().slice(0, 10);

const emptyLine: SaleLineFormState = {
  itemId: "",
  quantityAtomic: "",
  enteredUnitCode: "",
  enteredPackagingName: "",
  conversionNumeratorAtomic: "",
  conversionDenominator: "1",
  commercialTotal: "",
  lotId: "",
};

const emptyForm = (): SaleFormState => ({
  counterpartyId: "",
  occurredOn: todayISO(),
  reasonCode: "",
  notes: "",
  line: emptyLine,
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
  const [selectedItem, setSelectedItem] = useState<ItemResponse | null>(null);
  const [customers, setCustomers] = useState<CounterpartyResponse[]>([]);
  const [eligibleLots, setEligibleLots] = useState<LotResponse[]>([]);
  const [balance, setBalance] = useState<InventoryBalanceResponse | null>(null);
  const [sales, setSales] = useState<SaleDocumentResponse[]>([]);
  const [form, setForm] = useState<SaleFormState>(() => emptyForm());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const activePackagings = useMemo(
    () =>
      selectedItem?.packagings.filter(
        (packaging) => packaging.archivedAtMs === null || packaging.archivedAtMs === undefined,
      ) ?? [],
    [selectedItem],
  );

  const itemNames = useMemo(
    () => new Map(items.map((item) => [item.id, item.name] as const)),
    [items],
  );

  const customerNames = useMemo(
    () => new Map(customers.map((customer) => [customer.id, customer.name] as const)),
    [customers],
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

    setSelectedItem(detail);
    setEligibleLots(lots);
    setBalance(currentBalance);
    setForm((current) => ({
      ...current,
      line: lineForItem(current.line, detail),
    }));
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
      const firstItem = itemPage.items[0];
      if (firstItem) {
        await loadItemContext(firstItem.id, todayISO());
      } else {
        setSelectedItem(null);
        setEligibleLots([]);
        setBalance(null);
      }
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar vendas."),
      });
    } finally {
      setLoading(false);
    }
  }, [loadItemContext]);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  const selectItem = async (itemId: string) => {
    setForm((current) => ({ ...current, line: { ...current.line, itemId } }));
    const id = parseInteger(itemId);
    if (!id) {
      setSelectedItem(null);
      setEligibleLots([]);
      setBalance(null);
      return;
    }
    setLoading(true);
    setMessage(null);
    try {
      await loadItemContext(id, form.occurredOn);
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar o item vendavel."),
      });
    } finally {
      setLoading(false);
    }
  };

  const changeOccurredOn = async (occurredOn: string) => {
    setForm((current) => ({ ...current, occurredOn }));
    if (!selectedItem) return;
    try {
      setEligibleLots(await inventoryGateway.listEligibleFefoLots(selectedItem.id, occurredOn));
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar lotes elegiveis."),
      });
    }
  };

  const selectPackaging = (packagingID: string) => {
    if (!selectedItem) return;
    if (packagingID === "base") {
      setForm((current) => ({
        ...current,
        line: lineForItem(current.line, selectedItem),
      }));
      return;
    }
    const packaging = activePackagings.find((current) => String(current.id) === packagingID);
    if (!packaging) return;
    setForm((current) => ({
      ...current,
      line: lineForPackaging(current.line, selectedItem, packaging),
    }));
  };

  const postSale = async () => {
    if (saving) return;
    const itemId = parseInteger(form.line.itemId);
    const counterpartyId = parseInteger(form.counterpartyId);
    const quantityAtomic = parseInteger(form.line.quantityAtomic);
    const conversionNumeratorAtomic = parseInteger(form.line.conversionNumeratorAtomic);
    const conversionDenominator = parseInteger(form.line.conversionDenominator);
    const commercialTotalMinor = parseMoneyMinor(form.line.commercialTotal);
    const lotId = parseInteger(form.line.lotId);

    if (!itemId || !quantityAtomic || !conversionNumeratorAtomic || !conversionDenominator) {
      setMessage({
        type: "error",
        text: "Informe item, quantidade atomica e conversao para postar a venda.",
      });
      return;
    }
    if (commercialTotalMinor === undefined) {
      setMessage({ type: "error", text: "Informe o total comercial da venda." });
      return;
    }
    if (commercialTotalMinor === 0 && form.reasonCode.length === 0) {
      setMessage({
        type: "error",
        text: "Venda com total zero precisa da razao PROMOTION ou SAMPLE.",
      });
      return;
    }
    const reasonCode: SaleReason | undefined =
      commercialTotalMinor === 0 ? (form.reasonCode as SaleReason) : undefined;

    setSaving(true);
    setMessage(null);
    try {
      const posted = await saleGateway.postSale({
        idempotencyKey: buildIdempotencyKey(),
        counterpartyId,
        occurredOn: form.occurredOn,
        reasonCode,
        notes: optionalText(form.notes),
        lines: [
          {
            itemId,
            quantityAtomic,
            enteredUnitCode: form.line.enteredUnitCode,
            enteredPackagingName: optionalText(form.line.enteredPackagingName),
            conversionNumeratorAtomic,
            conversionDenominator,
            commercialTotalMinor,
            lotId,
          },
        ],
      });
      setMessage({ type: "success", text: `Venda ${saleLabel(posted)} postada.` });
      setForm((current) => ({
        ...current,
        reasonCode: "",
        notes: "",
        line: {
          ...current.line,
          quantityAtomic: "",
          commercialTotal: "",
          lotId: "",
        },
      }));
      await Promise.all([refreshSelectedInventory(itemId, form.occurredOn), loadSales()]);
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel postar a venda."),
      });
    } finally {
      setSaving(false);
    }
  };

  const refreshSelectedInventory = async (itemId: number, occurredOn: string) => {
    const [lots, currentBalance] = await Promise.all([
      inventoryGateway.listEligibleFefoLots(itemId, occurredOn),
      inventoryGateway.getInventoryBalance(itemId).catch(() => null),
    ]);
    setEligibleLots(lots);
    setBalance(currentBalance);
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

      <section className="mx-auto grid max-w-7xl gap-6 px-6 py-8 xl:grid-cols-[420px_1fr]">
        <aside className="space-y-6">
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
                <p className="text-sm text-slate-600">Cliente opcional, uma linha por vez.</p>
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

              <label className="block text-sm font-semibold text-slate-700">
                Item vendavel
                <select
                  value={form.line.itemId}
                  onChange={(event) => void selectItem(event.target.value)}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  <option value="">Selecione</option>
                  {items.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Unidade / embalagem
                <select
                  value={selectedPackagingValue(form.line, activePackagings)}
                  onChange={(event) => selectPackaging(event.target.value)}
                  disabled={!selectedItem}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500 disabled:bg-slate-100"
                >
                  <option value="base">
                    Unidade base {selectedItem ? `(${selectedItem.baseUnit.symbol})` : ""}
                  </option>
                  {activePackagings.map((packaging) => (
                    <option key={packaging.id} value={packaging.id}>
                      {packaging.name}
                    </option>
                  ))}
                </select>
              </label>

              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Quantidade atomica
                  <input
                    value={form.line.quantityAtomic}
                    onChange={(event) =>
                      setForm({
                        ...form,
                        line: { ...form.line, quantityAtomic: event.target.value },
                      })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="100"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Total comercial
                  <input
                    value={form.line.commercialTotal}
                    onChange={(event) =>
                      setForm({
                        ...form,
                        line: { ...form.line, commercialTotal: event.target.value },
                      })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="25,00"
                  />
                </label>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Conversao numerador
                  <input
                    value={form.line.conversionNumeratorAtomic}
                    onChange={(event) =>
                      setForm({
                        ...form,
                        line: { ...form.line, conversionNumeratorAtomic: event.target.value },
                      })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Denominador
                  <input
                    value={form.line.conversionDenominator}
                    onChange={(event) =>
                      setForm({
                        ...form,
                        line: { ...form.line, conversionDenominator: event.target.value },
                      })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
              </div>

              <label className="block text-sm font-semibold text-slate-700">
                Lote de saida de estoque
                <select
                  value={form.line.lotId}
                  onChange={(event) =>
                    setForm({ ...form, line: { ...form.line, lotId: event.target.value } })
                  }
                  disabled={!selectedItem}
                  className="mt-2 w-full rounded-xl border border-slate-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500 disabled:bg-slate-100"
                >
                  <option value="">FEFO automatico</option>
                  {eligibleLots.map((lot) => (
                    <option key={lot.id} value={lot.id}>
                      {lotLabel(lot)}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Razao para venda sem valor
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
                disabled={saving || loading || form.line.itemId.length === 0}
                className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
              >
                <Tag size={18} />
                {saving ? "Postando..." : "Postar venda"}
              </button>
            </div>
          </div>
        </aside>

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

          <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-xl font-bold text-slate-950">Item selecionado</h2>
            {!selectedItem ? (
              <p className="mt-4 text-sm text-slate-500">
                {loading ? "Carregando..." : "Nenhum item vendavel ativo encontrado."}
              </p>
            ) : (
              <div className="mt-4 grid gap-4 md:grid-cols-3">
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-xs font-semibold uppercase text-slate-500">Produto</p>
                  <p className="mt-1 font-semibold text-slate-950">{selectedItem.name}</p>
                  <p className="text-sm text-slate-500">base {selectedItem.baseUnit.code}</p>
                </div>
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-xs font-semibold uppercase text-slate-500">Saldo</p>
                  <p className="mt-1 font-semibold text-slate-950">
                    {balance?.quantityAtomic ?? 0} atomicos
                  </p>
                  <p className="text-sm text-slate-500">
                    {formatInventoryMicro(balance?.inventoryValueMicro ?? 0)}
                  </p>
                </div>
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-xs font-semibold uppercase text-slate-500">Preco padrao</p>
                  <p className="mt-1 font-semibold text-slate-950">
                    {selectedItem.defaultSalePrice == null
                      ? "-"
                      : formatMoneyMinor(selectedItem.defaultSalePrice)}
                  </p>
                  <p className="text-sm text-slate-500">apenas referencia</p>
                </div>
              </div>
            )}
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-200 p-6">
              <h2 className="text-xl font-bold text-slate-950">Lotes disponiveis</h2>
              <p className="mt-1 text-sm text-slate-500">
                Se nenhum lote for escolhido, o backend aloca automaticamente por FEFO.
              </p>
            </div>
            {eligibleLots.length === 0 ? (
              <p className="p-6 text-sm text-slate-500">
                Nenhum lote elegivel para o item/data selecionados.
              </p>
            ) : (
              <div className="divide-y divide-slate-100">
                {eligibleLots.map((lot) => (
                  <div key={lot.id} className="grid gap-2 p-6 md:grid-cols-[1fr_160px_160px]">
                    <div>
                      <p className="font-semibold text-slate-950">
                        {lot.lotCode ?? `Lote #${lot.id}`}
                      </p>
                      <p className="text-sm text-slate-500">
                        origem {lot.sourceKind} #{lot.sourceDocumentId} · {lot.originatedOn}
                      </p>
                    </div>
                    <p className="text-sm text-slate-700">
                      disponivel: <strong>{lot.availableQuantityAtomic}</strong>
                    </p>
                    <p className="text-sm text-slate-700">
                      vence: <strong>{lot.expiresOn ?? "-"}</strong>
                    </p>
                  </div>
                ))}
              </div>
            )}
          </section>

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
