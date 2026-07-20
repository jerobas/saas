import { FileText, MagnifyingGlass, Plus, Trash } from "@phosphor-icons/react";
import { useEffect, useMemo, useState } from "react";
import ConversionPreview from "../../components/ConversionPreview";
import {
  catalogGateway,
  type CounterpartyResponse,
  type ItemResponse,
  type ItemSummaryResponse,
  type PackagingResponse,
  type PurchaseDocumentResponse,
  type PurchaseLineRequest,
  purchaseGateway,
} from "../../gateways/desktopBridge";

interface PurchaseLineFormState {
  key: string;
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
  includeLotDetails: boolean;
  notes: string;
  lines: PurchaseLineFormState[];
}

interface PurchaseDocumentFormProps {
  items: ItemSummaryResponse[];
  suppliers: CounterpartyResponse[];
  disabled?: boolean;
  onError: (message: string) => void;
  onPosted: (purchase: PurchaseDocumentResponse) => Promise<void>;
}

let nextLineKey = 0;

const todayISO = () => new Date().toISOString().slice(0, 10);

const newLine = (): PurchaseLineFormState => ({
  key: `purchase-line-${++nextLineKey}`,
  itemId: "",
  quantityAtomic: "",
  enteredUnitCode: "",
  enteredPackagingName: "",
  conversionNumeratorAtomic: "",
  conversionDenominator: "1",
  commercialTotal: "",
  lotCode: "",
  expiresOn: "",
});

const emptyForm = (): PurchaseFormState => ({
  counterpartyId: "",
  occurredOn: todayISO(),
  freeStock: false,
  includeLotDetails: false,
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

const normalizedSearch = (value: string) =>
  value
    .normalize("NFD")
    .replace(/\p{Diacritic}/gu, "")
    .toLocaleLowerCase("pt-BR")
    .trim();

const buildIdempotencyKey = () =>
  `purchase-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

function PurchaseDocumentForm({
  items,
  suppliers,
  disabled = false,
  onError,
  onPosted,
}: PurchaseDocumentFormProps) {
  const [form, setForm] = useState<PurchaseFormState>(() => emptyForm());
  const [itemDetails, setItemDetails] = useState<Record<number, ItemResponse>>({});
  const [itemSearch, setItemSearch] = useState("");
  const [itemToAddId, setItemToAddId] = useState("");
  const [addingLine, setAddingLine] = useState(false);
  const [saving, setSaving] = useState(false);

  const filteredItems = useMemo(() => {
    const search = normalizedSearch(itemSearch);
    if (!search) return items;
    return items.filter((item) =>
      normalizedSearch(`${item.name} ${item.sku ?? ""}`).includes(search),
    );
  }, [itemSearch, items]);

  useEffect(() => {
    setItemToAddId((current) =>
      filteredItems.some((item) => String(item.id) === current)
        ? current
        : String(filteredItems[0]?.id ?? ""),
    );
  }, [filteredItems]);

  const documentTotalMinor = form.lines.reduce(
    (total, line) => total + (parseMoneyMinor(line.commercialTotal) ?? 0),
    0,
  );

  const updateLine = (key: string, patch: Partial<PurchaseLineFormState>) => {
    setForm((current) => ({
      ...current,
      lines: current.lines.map((line) => (line.key === key ? { ...line, ...patch } : line)),
    }));
  };

  const addItem = async () => {
    const itemId = parseInteger(itemToAddId);
    if (!itemId || addingLine) {
      onError("Selecione um item para adicionar à compra.");
      return;
    }

    setAddingLine(true);
    try {
      const detail = itemDetails[itemId] ?? (await catalogGateway.getItem(itemId));
      setItemDetails((current) => ({ ...current, [itemId]: detail }));
      setForm((current) => ({
        ...current,
        lines: [...current.lines, lineForItem(newLine(), detail)],
      }));
      setItemSearch("");
    } catch (error) {
      onError(error instanceof Error ? error.message : "Não foi possível adicionar o item.");
    } finally {
      setAddingLine(false);
    }
  };

  const removeLine = (key: string) => {
    setForm((current) => ({
      ...current,
      lines: current.lines.filter((line) => line.key !== key),
    }));
  };

  const toggleLotDetails = (includeLotDetails: boolean) => {
    setForm((current) => ({
      ...current,
      includeLotDetails,
      lines: includeLotDetails
        ? current.lines
        : current.lines.map((line) => ({ ...line, lotCode: "", expiresOn: "" })),
    }));
  };

  const selectPackaging = (line: PurchaseLineFormState, packagingId: string) => {
    const itemId = parseInteger(line.itemId);
    const item = itemId ? itemDetails[itemId] : undefined;
    if (!item) return;

    if (packagingId === "base") {
      updateLine(line.key, lineForItem(line, item));
      return;
    }

    const packaging = activePackagings(item).find((current) => String(current.id) === packagingId);
    if (packaging) updateLine(line.key, lineForPackaging(line, item, packaging));
  };

  const postPurchase = async () => {
    if (saving) return;
    if (form.lines.length === 0) {
      onError("Adicione pelo menos um item à compra.");
      return;
    }

    const requestLines: PurchaseLineRequest[] = [];
    for (const [index, line] of form.lines.entries()) {
      const itemId = parseInteger(line.itemId);
      const quantityAtomic = parseInteger(line.quantityAtomic);
      const conversionNumeratorAtomic = parseInteger(line.conversionNumeratorAtomic);
      const conversionDenominator = parseInteger(line.conversionDenominator);
      const commercialTotalMinor = parseMoneyMinor(line.commercialTotal);

      if (
        !itemId ||
        !quantityAtomic ||
        quantityAtomic <= 0 ||
        !conversionNumeratorAtomic ||
        conversionNumeratorAtomic <= 0 ||
        !conversionDenominator ||
        conversionDenominator <= 0 ||
        !line.enteredUnitCode
      ) {
        onError(`Revise item, quantidade atômica e unidade da linha ${index + 1}.`);
        return;
      }
      if (commercialTotalMinor === undefined || commercialTotalMinor < 0) {
        onError(`Informe o total comercial da linha ${index + 1}.`);
        return;
      }

      requestLines.push({
        itemId,
        quantityAtomic,
        enteredUnitCode: line.enteredUnitCode,
        enteredPackagingName: optionalText(line.enteredPackagingName),
        conversionNumeratorAtomic,
        conversionDenominator,
        commercialTotalMinor,
        lotCode: form.includeLotDetails ? optionalText(line.lotCode) : undefined,
        expiresOn: form.includeLotDetails ? optionalText(line.expiresOn) : undefined,
      });
    }

    const hasFreeStockLine = requestLines.some((line) => line.commercialTotalMinor === 0);
    if (hasFreeStockLine && !form.freeStock) {
      onError("Compra com linha sem custo precisa da razão FREE_STOCK.");
      return;
    }

    setSaving(true);
    try {
      const posted = await purchaseGateway.postPurchase({
        idempotencyKey: buildIdempotencyKey(),
        counterpartyId: parseInteger(form.counterpartyId),
        occurredOn: form.occurredOn,
        reasonCode: hasFreeStockLine ? "FREE_STOCK" : undefined,
        notes: optionalText(form.notes),
        lines: requestLines,
      });
      await onPosted(posted);
      setForm((current) => ({
        ...emptyForm(),
        counterpartyId: current.counterpartyId,
      }));
    } catch (error) {
      onError(error instanceof Error ? error.message : "Não foi possível postar a compra.");
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="rounded-3xl border border-slate-200 bg-white shadow-sm">
      <div className="flex items-center gap-3 border-b border-slate-200 px-6 py-5">
        <FileText size={28} className="text-pink-600" />
        <div>
          <h2 className="text-xl font-bold text-slate-950">Nova compra</h2>
          <p className="text-sm text-slate-600">
            Um documento com cabeçalho compartilhado e todas as linhas da entrada.
          </p>
        </div>
      </div>

      <div className="space-y-6 p-6">
        <div className="grid gap-4 lg:grid-cols-[1fr_220px_1.5fr]">
          <label className="block text-sm font-semibold text-slate-700">
            Fornecedor
            <select
              value={form.counterpartyId}
              onChange={(event) =>
                setForm((current) => ({ ...current, counterpartyId: event.target.value }))
              }
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
            Data de ocorrência
            <input
              type="date"
              value={form.occurredOn}
              onChange={(event) =>
                setForm((current) => ({ ...current, occurredOn: event.target.value }))
              }
              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
            />
          </label>

          <label className="block text-sm font-semibold text-slate-700">
            Observações
            <input
              value={form.notes}
              onChange={(event) =>
                setForm((current) => ({ ...current, notes: event.target.value }))
              }
              placeholder="Informações gerais desta compra"
              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
            />
          </label>
        </div>

        <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
          <h3 className="font-bold text-slate-900">Adicionar item</h3>
          <div className="mt-3 grid gap-3 md:grid-cols-[1fr_1fr_auto]">
            <label className="relative block">
              <span className="sr-only">Buscar item</span>
              <MagnifyingGlass
                size={18}
                className="pointer-events-none absolute left-3 top-3 text-slate-400"
              />
              <input
                type="search"
                aria-label="Buscar item"
                value={itemSearch}
                onChange={(event) => setItemSearch(event.target.value)}
                placeholder="Nome ou SKU"
                className="w-full rounded-xl border border-slate-300 py-2 pl-10 pr-3 outline-none focus:ring-2 focus:ring-pink-500"
              />
            </label>
            <label>
              <span className="sr-only">Item para adicionar</span>
              <select
                aria-label="Item para adicionar"
                value={itemToAddId}
                onChange={(event) => setItemToAddId(event.target.value)}
                className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
              >
                {filteredItems.length === 0 ? (
                  <option value="">Nenhum item encontrado</option>
                ) : (
                  filteredItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} {item.sku ? `· ${item.sku}` : ""}
                    </option>
                  ))
                )}
              </select>
            </label>
            <button
              type="button"
              onClick={() => void addItem()}
              disabled={disabled || addingLine || !itemToAddId}
              className="inline-flex items-center justify-center gap-2 rounded-xl bg-slate-900 px-5 py-2 font-semibold text-white transition hover:bg-slate-700 disabled:bg-slate-300"
            >
              <Plus size={18} />
              {addingLine ? "Adicionando..." : "Adicionar item"}
            </button>
          </div>
        </div>

        <div>
          <div className="mb-3 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
            <h3 className="font-bold text-slate-900">Itens da compra</h3>
            <div className="flex flex-wrap items-center gap-4">
              <span className="text-sm text-slate-500">
                {form.lines.length} {form.lines.length === 1 ? "linha" : "linhas"}
              </span>
              <label className="inline-flex cursor-pointer items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm font-semibold text-slate-700">
                <input
                  type="checkbox"
                  role="switch"
                  aria-label="Informar lote e validade"
                  checked={form.includeLotDetails}
                  onChange={(event) => toggleLotDetails(event.target.checked)}
                  className="peer sr-only"
                />
                <span className="relative h-6 w-11 rounded-full bg-slate-300 transition peer-checked:bg-pink-600 after:absolute after:left-1 after:top-1 after:h-4 after:w-4 after:rounded-full after:bg-white after:transition-transform peer-checked:after:translate-x-5" />
                <span>
                  <span className="block">Informar lote e validade</span>
                  <span className="block text-xs font-normal text-slate-500">
                    Opcional; o estoque cria o lote interno mesmo sem estes dados.
                  </span>
                </span>
              </label>
            </div>
          </div>

          {form.lines.length === 0 ? (
            <div className="rounded-2xl border border-dashed border-slate-300 px-6 py-10 text-center text-sm text-slate-500">
              Busque um item e adicione a primeira linha da compra.
            </div>
          ) : (
            <div className="space-y-4">
              {form.lines.map((line, index) => {
                const item = itemDetails[Number(line.itemId)];
                const packagings = item ? activePackagings(item) : [];
                return (
                  <article
                    key={line.key}
                    className="rounded-2xl border border-slate-200 bg-white p-4 shadow-xs"
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <p className="text-xs font-semibold uppercase tracking-wide text-pink-600">
                          Linha {index + 1}
                        </p>
                        <h4 className="font-bold text-slate-950">
                          {item?.name ?? `Item #${line.itemId}`}
                        </h4>
                      </div>
                      <button
                        type="button"
                        onClick={() => removeLine(line.key)}
                        className="inline-flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-semibold text-red-600 hover:bg-red-50"
                        aria-label={`Remover ${item?.name ?? `linha ${index + 1}`}`}
                      >
                        <Trash size={17} />
                        Remover
                      </button>
                    </div>

                    <div
                      className={`mt-4 grid gap-3 md:grid-cols-2 ${
                        form.includeLotDetails ? "xl:grid-cols-5" : "xl:grid-cols-3"
                      }`}
                    >
                      <label className="block text-sm font-semibold text-slate-700">
                        Quantidade atômica
                        <input
                          aria-label={`Quantidade atômica da linha ${index + 1}`}
                          value={line.quantityAtomic}
                          onChange={(event) =>
                            updateLine(line.key, { quantityAtomic: event.target.value })
                          }
                          placeholder="1000"
                          className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                        />
                      </label>

                      <label className="block text-sm font-semibold text-slate-700">
                        Unidade / embalagem
                        <select
                          aria-label={`Unidade ou embalagem da linha ${index + 1}`}
                          value={selectedPackagingValue(line, packagings)}
                          onChange={(event) => selectPackaging(line, event.target.value)}
                          className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                        >
                          <option value="base">
                            {item ? `${item.baseUnit.symbol} · unidade base` : "Unidade base"}
                          </option>
                          {packagings.map((packaging) => (
                            <option key={packaging.id} value={packaging.id}>
                              {packaging.name}
                            </option>
                          ))}
                        </select>
                      </label>

                      <label className="block text-sm font-semibold text-slate-700">
                        Valor da linha
                        <input
                          aria-label={`Valor da linha ${index + 1}`}
                          value={line.commercialTotal}
                          onChange={(event) =>
                            updateLine(line.key, { commercialTotal: event.target.value })
                          }
                          placeholder="50,00"
                          className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                        />
                      </label>

                      {form.includeLotDetails && (
                        <>
                          <label className="block text-sm font-semibold text-slate-700">
                            Lote
                            <input
                              aria-label={`Lote da linha ${index + 1}`}
                              value={line.lotCode}
                              onChange={(event) =>
                                updateLine(line.key, { lotCode: event.target.value })
                              }
                              placeholder="LOTE-001"
                              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                            />
                          </label>

                          <label className="block text-sm font-semibold text-slate-700">
                            Validade
                            <input
                              type="date"
                              aria-label={`Validade da linha ${index + 1}`}
                              value={line.expiresOn}
                              onChange={(event) =>
                                updateLine(line.key, { expiresOn: event.target.value })
                              }
                              className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                            />
                          </label>
                        </>
                      )}
                    </div>

                    {item && line.enteredPackagingName && (
                      <ConversionPreview
                        label={`1 ${line.enteredPackagingName}`}
                        numeratorAtomic={line.conversionNumeratorAtomic}
                        denominator={line.conversionDenominator}
                        baseUnit={item.baseUnit}
                        className="mt-3 rounded-xl bg-slate-50 px-3 py-2 text-sm text-slate-600"
                      />
                    )}
                  </article>
                );
              })}
            </div>
          )}
        </div>

        <div className="flex flex-col gap-4 border-t border-slate-200 pt-5 md:flex-row md:items-center md:justify-between">
          <div className="space-y-2">
            <p className="text-sm text-slate-500">Total do documento</p>
            <p className="text-2xl font-bold text-slate-950">
              {formatMoneyMinor(documentTotalMinor)}
            </p>
            <label className="flex items-center gap-2 text-sm font-semibold text-slate-700">
              <input
                type="checkbox"
                checked={form.freeStock}
                onChange={(event) =>
                  setForm((current) => ({ ...current, freeStock: event.target.checked }))
                }
              />
              FREE_STOCK para linha sem custo
            </label>
          </div>
          <button
            type="button"
            onClick={() => void postPurchase()}
            disabled={disabled || saving || form.lines.length === 0}
            className="inline-flex min-w-48 items-center justify-center gap-2 rounded-xl bg-pink-600 px-6 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
          >
            <Plus size={18} />
            {saving ? "Postando..." : "Postar compra"}
          </button>
        </div>
      </div>
    </section>
  );
}

const activePackagings = (item: ItemResponse) =>
  item.packagings.filter(
    (packaging) => packaging.archivedAtMs === null || packaging.archivedAtMs === undefined,
  );

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

export default PurchaseDocumentForm;
