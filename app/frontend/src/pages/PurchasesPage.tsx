import { ArrowClockwise, FileText, Plus } from "@phosphor-icons/react";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import {
  catalogGateway,
  counterpartyGateway,
  type CounterpartyResponse,
  type ItemResponse,
  type ItemSummaryResponse,
  type PackagingResponse,
  type PurchaseDocumentResponse,
  purchaseGateway,
} from "../gateways/desktopBridge";

interface PurchaseLineFormState {
  itemId: string;
  quantityAtomic: string;
  enteredUnitCode: string;
  enteredPackagingName: string;
  conversionNumeratorAtomic: string;
  conversionDenominator: string;
  commercialTotal: string;
  lotCode: string;
  expiresOn: string;
}

interface PurchaseFormState {
  counterpartyId: string;
  occurredOn: string;
  freeStock: boolean;
  notes: string;
  line: PurchaseLineFormState;
}

const todayISO = () => new Date().toISOString().slice(0, 10);

const emptyLine: PurchaseLineFormState = {
  itemId: "",
  quantityAtomic: "",
  enteredUnitCode: "",
  enteredPackagingName: "",
  conversionNumeratorAtomic: "",
  conversionDenominator: "1",
  commercialTotal: "",
  lotCode: "",
  expiresOn: "",
};

const emptyForm = (): PurchaseFormState => ({
  counterpartyId: "",
  occurredOn: todayISO(),
  freeStock: false,
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

const formatInventoryMicro = (micro: number) =>
  new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(micro / 1_000_000);

const formatMoneyMinor = (minor: number) =>
  new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(minor / 100);

const documentLabel = (purchase: PurchaseDocumentResponse) =>
  `#${purchase.id} / seq ${purchase.postingSequence}`;

const buildIdempotencyKey = () =>
  `purchase-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

function PurchasesPage() {
  const [items, setItems] = useState<ItemSummaryResponse[]>([]);
  const [selectedItem, setSelectedItem] = useState<ItemResponse | null>(null);
  const [suppliers, setSuppliers] = useState<CounterpartyResponse[]>([]);
  const [purchases, setPurchases] = useState<PurchaseDocumentResponse[]>([]);
  const [form, setForm] = useState<PurchaseFormState>(() => emptyForm());
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

  const loadPurchases = useCallback(async () => {
    const page = await purchaseGateway.listPurchases({ pageSize: 25 });
    setPurchases(page.items);
  }, []);

  const loadPage = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const [itemPage, supplierPage, purchasePage] = await Promise.all([
        catalogGateway.listItems({
          archiveFilter: "ACTIVE",
          requireCapabilities: { purchasable: true, producible: false, sellable: false },
          pageSize: 100,
        }),
        counterpartyGateway.listCounterparties({
          archiveFilter: "ACTIVE",
          role: "SUPPLIER",
          pageSize: 100,
        }),
        purchaseGateway.listPurchases({ pageSize: 25 }),
      ]);
      setItems(itemPage.items);
      setSuppliers(supplierPage.items);
      setPurchases(purchasePage.items);
      const firstItem = itemPage.items[0];
      if (firstItem) {
        const detail = await catalogGateway.getItem(firstItem.id);
        setSelectedItem(detail);
        setForm((current) => ({
          ...current,
          line: lineForItem(current.line, detail),
        }));
      }
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel carregar compras.",
      });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  const selectItem = async (itemId: string) => {
    const id = Number.parseInt(itemId, 10);
    if (!Number.isFinite(id)) {
      setSelectedItem(null);
      setForm((current) => ({ ...current, line: { ...current.line, itemId } }));
      return;
    }
    const detail = await catalogGateway.getItem(id);
    setSelectedItem(detail);
    setForm((current) => ({ ...current, line: lineForItem(current.line, detail) }));
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

  const postPurchase = async () => {
    if (saving) return;
    const quantityAtomic = parseInteger(form.line.quantityAtomic);
    const conversionNumeratorAtomic = parseInteger(form.line.conversionNumeratorAtomic);
    const conversionDenominator = parseInteger(form.line.conversionDenominator);
    const commercialTotalMinor = parseMoneyMinor(form.line.commercialTotal);
    const itemId = parseInteger(form.line.itemId);
    const counterpartyId = parseInteger(form.counterpartyId);

    if (!itemId || !quantityAtomic || !conversionNumeratorAtomic || !conversionDenominator) {
      setMessage({
        type: "error",
        text: "Informe item, quantidade atomica e conversao para postar a compra.",
      });
      return;
    }
    if (commercialTotalMinor === undefined) {
      setMessage({ type: "error", text: "Informe o total comercial da compra." });
      return;
    }
    if (commercialTotalMinor === 0 && !form.freeStock) {
      setMessage({ type: "error", text: "Compra com total zero precisa da razao FREE_STOCK." });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const posted = await purchaseGateway.postPurchase({
        idempotencyKey: buildIdempotencyKey(),
        counterpartyId,
        occurredOn: form.occurredOn,
        reasonCode: commercialTotalMinor === 0 ? "FREE_STOCK" : undefined,
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
            lotCode: optionalText(form.line.lotCode),
            expiresOn: optionalText(form.line.expiresOn),
          },
        ],
      });
      setMessage({ type: "success", text: `Compra ${documentLabel(posted)} postada.` });
      setForm((current) => ({
        ...emptyForm(),
        counterpartyId: current.counterpartyId,
        line: selectedItem ? lineForItem(emptyLine, selectedItem) : emptyLine,
      }));
      await loadPurchases();
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel postar a compra.",
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
                Compras V2
              </p>
              <h1 className="mt-2 text-3xl font-bold text-slate-950">Compras e entrada</h1>
              <p className="mt-2 max-w-2xl text-slate-600">
                Postagem real local-first: cria documento, linha de entrada, lote e saldo de
                estoque.
              </p>
            </div>
            <button
              type="button"
              onClick={loadPage}
              disabled={loading || saving}
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-300 px-4 py-3 font-semibold text-slate-700 hover:bg-slate-100 disabled:opacity-50"
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
              <FileText size={28} className="text-pink-600" />
              <div>
                <h2 className="text-lg font-bold text-slate-950">Nova compra</h2>
                <p className="text-sm text-slate-600">Fornecedor opcional, uma linha por vez.</p>
              </div>
            </div>

            <div className="mt-5 space-y-4">
              <label className="block text-sm font-semibold text-slate-700">
                Fornecedor
                <select
                  value={form.counterpartyId}
                  onChange={(event) => setForm({ ...form, counterpartyId: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  <option value="">Sem fornecedor</option>
                  {suppliers.map((supplier) => (
                    <option key={supplier.id} value={supplier.id}>
                      {supplier.name}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Data de ocorrencia
                <input
                  type="date"
                  value={form.occurredOn}
                  onChange={(event) => setForm({ ...form, occurredOn: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                />
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Item compravel
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
                    placeholder="1000"
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
                    placeholder="50,00"
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

              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Lote
                  <input
                    value={form.line.lotCode}
                    onChange={(event) =>
                      setForm({ ...form, line: { ...form.line, lotCode: event.target.value } })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="LOTE-001"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Validade
                  <input
                    type="date"
                    value={form.line.expiresOn}
                    onChange={(event) =>
                      setForm({ ...form, line: { ...form.line, expiresOn: event.target.value } })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
              </div>

              <label className="flex items-center gap-2 rounded-xl bg-slate-100 px-3 py-2 text-sm font-semibold text-slate-700">
                <input
                  type="checkbox"
                  checked={form.freeStock}
                  onChange={(event) => setForm({ ...form, freeStock: event.target.checked })}
                />
                FREE_STOCK para compra sem custo
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
                onClick={postPurchase}
                disabled={saving || loading || form.line.itemId.length === 0}
                className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
              >
                <Plus size={18} />
                Postar compra
              </button>
            </div>
          </div>
        </aside>

        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-3">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Compras</p>
              <p className="mt-2 text-3xl font-bold text-slate-950">{purchases.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Itens compraveis</p>
              <p className="mt-2 text-3xl font-bold text-green-700">{items.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Fornecedores ativos</p>
              <p className="mt-2 text-3xl font-bold text-blue-700">{suppliers.length}</p>
            </div>
          </div>

          <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-200 px-6 py-4">
              <h2 className="text-lg font-bold text-slate-950">Compras postadas</h2>
            </div>
            {loading ? (
              <p className="p-6 text-slate-600">Carregando compras...</p>
            ) : purchases.length === 0 ? (
              <p className="p-6 text-slate-600">Nenhuma compra postada ainda.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="bg-slate-50 text-slate-600">
                    <tr>
                      <th className="px-6 py-3 font-semibold">Documento</th>
                      <th className="px-6 py-3 font-semibold">Data</th>
                      <th className="px-6 py-3 font-semibold">Fornecedor</th>
                      <th className="px-6 py-3 font-semibold">Linhas</th>
                      <th className="px-6 py-3 font-semibold">Valor estoque</th>
                      <th className="px-6 py-3 font-semibold">Razao</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {purchases.map((purchase) => {
                      const supplier = suppliers.find(
                        (current) => current.id === purchase.counterpartyId,
                      );
                      const inventoryValue = purchase.lines.reduce(
                        (acc, line) => acc + line.inventoryValueMicro,
                        0,
                      );
                      return (
                        <Fragment key={purchase.id}>
                          <tr className="hover:bg-slate-50">
                            <td className="px-6 py-4">
                              <p className="font-semibold text-slate-950">
                                {documentLabel(purchase)}
                              </p>
                              <p className="text-xs text-slate-500">{purchase.idempotencyKey}</p>
                            </td>
                            <td className="px-6 py-4 text-slate-700">{purchase.occurredOn}</td>
                            <td className="px-6 py-4 text-slate-700">
                              {supplier?.name ?? "Sem fornecedor"}
                            </td>
                            <td className="px-6 py-4 text-slate-700">
                              {purchase.lines.map((line) => (
                                <div key={line.id}>
                                  {itemNames.get(line.itemId) ?? `item #${line.itemId}`}:{" "}
                                  {line.quantityAtomic} {line.enteredUnitCode}
                                  {line.lotCode ? ` / lote ${line.lotCode}` : ""}
                                </div>
                              ))}
                            </td>
                            <td className="px-6 py-4 font-semibold text-slate-900">
                              {formatInventoryMicro(inventoryValue)}
                            </td>
                            <td className="px-6 py-4 text-slate-700">
                              {purchase.reasonCode ?? "-"}
                            </td>
                          </tr>
                          <tr className="bg-slate-50/70">
                            <td colSpan={6} className="px-6 pb-5 pt-0">
                              <div className="rounded-2xl border border-slate-200 bg-white p-4">
                                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                                  Detalhe da compra: linhas, lotes e valores
                                </p>
                                <div className="mt-3 space-y-3">
                                  {purchase.lines.map((line) => (
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
                                          Total comercial:{" "}
                                          <strong>
                                            {formatMoneyMinor(line.commercialTotalMinor)}
                                          </strong>
                                        </p>
                                        <p>
                                          Valor estoque:{" "}
                                          <strong>
                                            {formatInventoryMicro(line.inventoryValueMicro)}
                                          </strong>
                                        </p>
                                      </div>
                                      <div className="text-sm text-slate-700">
                                        <p className="font-semibold text-slate-900">Lote criado</p>
                                        <p>
                                          lote #{line.lotId}
                                          {line.lotCode ? ` · ${line.lotCode}` : ""}
                                        </p>
                                        <p className="text-slate-500">
                                          origem {line.originatedOn}
                                          {line.expiresOn ? ` · vence ${line.expiresOn}` : ""}
                                        </p>
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
          </div>

          <div className="rounded-3xl border border-blue-100 bg-blue-50 p-4 text-sm text-blue-900">
            Depois de postar, confira <strong>Estoque</strong> e <strong>Lotes</strong>: eles leem o
            saldo real criado pela compra.
          </div>
        </div>
      </section>
    </main>
  );
}

function lineForItem(line: PurchaseLineFormState, item: ItemResponse): PurchaseLineFormState {
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
  line: PurchaseLineFormState,
  item: ItemResponse,
  packaging: PackagingResponse,
): PurchaseLineFormState {
  return {
    ...line,
    itemId: String(item.id),
    enteredUnitCode: packaging.enteredUnitCode,
    enteredPackagingName: packaging.name,
    conversionNumeratorAtomic: String(packaging.conversionNumeratorAtomic),
    conversionDenominator: String(packaging.conversionDenominator),
  };
}

function selectedPackagingValue(line: PurchaseLineFormState, packagings: PackagingResponse[]) {
  const selected = packagings.find(
    (packaging) =>
      packaging.name === line.enteredPackagingName &&
      packaging.enteredUnitCode === line.enteredUnitCode &&
      String(packaging.conversionNumeratorAtomic) === line.conversionNumeratorAtomic &&
      String(packaging.conversionDenominator) === line.conversionDenominator,
  );
  return selected ? String(selected.id) : "base";
}

export default PurchasesPage;
