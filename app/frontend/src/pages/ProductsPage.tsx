import { Archive, ArrowClockwise, Package, Plus, XCircle } from "@phosphor-icons/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  catalogGateway,
  type CapabilitiesRequest,
  type ItemResponse,
  type ItemSummaryResponse,
  type MeasurementUnitResponse,
  type PackagingResponse,
  referenceDataGateway,
} from "../gateways/desktopBridge";

interface ItemFormState {
  name: string;
  sku: string;
  description: string;
  baseUnitCode: string;
  purchasable: boolean;
  producible: boolean;
  sellable: boolean;
  defaultSalePrice: string;
  reorderQuantityAtomic: string;
}

interface PackagingFormState {
  name: string;
  enteredUnitCode: string;
  conversionNumeratorAtomic: string;
  conversionDenominator: string;
}

const emptyItemForm: ItemFormState = {
  name: "",
  sku: "",
  description: "",
  baseUnitCode: "",
  purchasable: true,
  producible: false,
  sellable: false,
  defaultSalePrice: "",
  reorderQuantityAtomic: "",
};

const emptyPackagingForm: PackagingFormState = {
  name: "",
  enteredUnitCode: "",
  conversionNumeratorAtomic: "1000",
  conversionDenominator: "1",
};

const optionalText = (value: string) => {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
};

const parseInteger = (value: string) => {
  const trimmed = value.trim();
  if (trimmed.length === 0) return undefined;
  const parsed = Number.parseInt(trimmed, 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const parseMoneyMinor = (value: string) => {
  const normalized = value.trim().replace(",", ".");
  if (normalized.length === 0) return undefined;
  const parsed = Number.parseFloat(normalized);
  if (!Number.isFinite(parsed)) return undefined;
  return Math.round(parsed * 100);
};

const formatMoney = (minor?: number | null) => {
  if (minor === null || minor === undefined) return "-";
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(minor / 100);
};

const formatDateTime = (ms?: number | null) => {
  if (ms === null || ms === undefined) return "-";
  return new Intl.DateTimeFormat("pt-BR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(ms));
};

const capabilityLabels = (capabilities: CapabilitiesRequest) =>
  [
    capabilities.purchasable ? "Compra" : null,
    capabilities.producible ? "Producao" : null,
    capabilities.sellable ? "Venda" : null,
  ]
    .filter(Boolean)
    .join(" / ");

function ProductsPage() {
  const [items, setItems] = useState<ItemSummaryResponse[]>([]);
  const [units, setUnits] = useState<MeasurementUnitResponse[]>([]);
  const [selectedItem, setSelectedItem] = useState<ItemResponse | null>(null);
  const [editingItem, setEditingItem] = useState<ItemSummaryResponse | null>(null);
  const [itemForm, setItemForm] = useState<ItemFormState>(emptyItemForm);
  const [packagingForm, setPackagingForm] = useState<PackagingFormState>(emptyPackagingForm);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const itemBaseUnits = useMemo(() => units.filter((unit) => unit.isItemBase), [units]);

  const selectedBaseUnit = useMemo(
    () => units.find((unit) => unit.code === selectedItem?.baseUnitCode),
    [selectedItem?.baseUnitCode, units],
  );

  const compatiblePackagingUnits = useMemo(() => {
    if (!selectedBaseUnit) return units;
    return units.filter((unit) => unit.dimension === selectedBaseUnit.dimension);
  }, [selectedBaseUnit, units]);

  const refreshSelectedItem = useCallback(async (id: number) => {
    const item = await catalogGateway.getItem(id);
    setSelectedItem(item);
    return item;
  }, []);

  const loadCatalog = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const [loadedUnits, page] = await Promise.all([
        referenceDataGateway.listMeasurementUnits(),
        catalogGateway.listItems({
          archiveFilter: "ALL",
          requireCapabilities: { purchasable: false, producible: false, sellable: false },
          pageSize: 100,
        }),
      ]);
      setUnits(loadedUnits);
      setItems(page.items);
      const defaultBaseUnit = loadedUnits.find((unit) => unit.isItemBase)?.code ?? "";
      const defaultEnteredUnit =
        loadedUnits.find((unit) => !unit.isItemBase)?.code ?? defaultBaseUnit;
      setItemForm((current) => ({
        ...current,
        baseUnitCode: current.baseUnitCode || defaultBaseUnit,
      }));
      setPackagingForm((current) => ({
        ...current,
        enteredUnitCode: current.enteredUnitCode || defaultEnteredUnit,
      }));
      setSelectedItem(null);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel carregar o catalogo.",
      });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadCatalog();
  }, [loadCatalog]);

  const resetItemForm = () => {
    setEditingItem(null);
    setItemForm({
      ...emptyItemForm,
      baseUnitCode: itemBaseUnits[0]?.code ?? "",
    });
  };

  const startEditingItem = (item: ItemSummaryResponse) => {
    setEditingItem(item);
    setItemForm({
      name: item.name,
      sku: item.sku ?? "",
      description: item.description ?? "",
      baseUnitCode: item.baseUnitCode,
      purchasable: item.capabilities.purchasable,
      producible: item.capabilities.producible,
      sellable: item.capabilities.sellable,
      defaultSalePrice:
        item.defaultSalePrice === null || item.defaultSalePrice === undefined
          ? ""
          : String(item.defaultSalePrice / 100),
      reorderQuantityAtomic:
        item.reorderQuantityAtomic === null || item.reorderQuantityAtomic === undefined
          ? ""
          : String(item.reorderQuantityAtomic),
    });
  };

  const persistItem = async () => {
    if (saving) return;
    const capabilities = {
      purchasable: itemForm.purchasable,
      producible: itemForm.producible,
      sellable: itemForm.sellable,
    };
    if (!capabilities.purchasable && !capabilities.producible && !capabilities.sellable) {
      setMessage({ type: "error", text: "Selecione pelo menos uma capacidade para o item." });
      return;
    }
    const defaultSalePrice = parseMoneyMinor(itemForm.defaultSalePrice);
    if (!capabilities.sellable && defaultSalePrice !== undefined) {
      setMessage({
        type: "error",
        text: "Preco de venda so pode ser informado quando a capacidade Venda esta marcada.",
      });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const request = {
        name: itemForm.name.trim(),
        sku: optionalText(itemForm.sku),
        description: optionalText(itemForm.description),
        baseUnitCode: itemForm.baseUnitCode,
        capabilities,
        defaultSalePrice,
        reorderQuantityAtomic: parseInteger(itemForm.reorderQuantityAtomic),
      };
      const saved = editingItem
        ? await catalogGateway.updateItem(editingItem.id, {
            ...request,
            expectedUpdatedAtMs: editingItem.updatedAtMs,
          })
        : await catalogGateway.createItem(request);
      setItems((current) => {
        const withoutSaved = current.filter((item) => item.id !== saved.id);
        return [...withoutSaved, saved].sort((left, right) => left.name.localeCompare(right.name));
      });
      setSelectedItem(saved);
      resetItemForm();
      setMessage({ type: "success", text: editingItem ? "Item atualizado." : "Item criado." });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel salvar o item.",
      });
    } finally {
      setSaving(false);
    }
  };

  const toggleItemArchive = async (item: ItemSummaryResponse) => {
    setSaving(true);
    setMessage(null);
    try {
      const saved =
        item.archivedAtMs === null || item.archivedAtMs === undefined
          ? await catalogGateway.archiveItem(item.id, { expectedUpdatedAtMs: item.updatedAtMs })
          : await catalogGateway.restoreItem(item.id, { expectedUpdatedAtMs: item.updatedAtMs });
      setItems((current) =>
        current.map((currentItem) => (currentItem.id === saved.id ? saved : currentItem)),
      );
      if (selectedItem?.id === saved.id) setSelectedItem(saved);
      setMessage({
        type: "success",
        text: saved.archivedAtMs ? "Item arquivado." : "Item restaurado.",
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel alterar o arquivamento.",
      });
    } finally {
      setSaving(false);
    }
  };

  const createPackaging = async () => {
    if (!selectedItem || saving) return;
    setSaving(true);
    setMessage(null);
    try {
      await catalogGateway.createItemPackaging({
        itemId: selectedItem.id,
        name: packagingForm.name.trim(),
        enteredUnitCode: packagingForm.enteredUnitCode,
        conversionNumeratorAtomic: Number.parseInt(packagingForm.conversionNumeratorAtomic, 10),
        conversionDenominator: Number.parseInt(packagingForm.conversionDenominator, 10),
      });
      await refreshSelectedItem(selectedItem.id);
      setPackagingForm({
        ...emptyPackagingForm,
        enteredUnitCode: compatiblePackagingUnits[0]?.code ?? "",
      });
      setMessage({ type: "success", text: "Embalagem criada." });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel salvar a embalagem.",
      });
    } finally {
      setSaving(false);
    }
  };

  const togglePackagingArchive = async (packaging: PackagingResponse) => {
    if (!selectedItem || saving) return;
    setSaving(true);
    setMessage(null);
    try {
      if (packaging.archivedAtMs === null || packaging.archivedAtMs === undefined) {
        await catalogGateway.archiveItemPackaging(packaging.id, {
          expectedUpdatedAtMs: packaging.updatedAtMs,
        });
      } else {
        await catalogGateway.restoreItemPackaging(packaging.id, {
          expectedUpdatedAtMs: packaging.updatedAtMs,
        });
      }
      await refreshSelectedItem(selectedItem.id);
      setMessage({ type: "success", text: "Embalagem atualizada." });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel atualizar a embalagem.",
      });
    } finally {
      setSaving(false);
    }
  };

  const activeItems = items.filter(
    (item) => item.archivedAtMs === null || item.archivedAtMs === undefined,
  );

  return (
    <main className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-7xl px-6 py-8">
          <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-wide text-pink-600">
                Catalogo V2
              </p>
              <h1 className="mt-2 text-3xl font-bold text-slate-950">Itens e embalagens</h1>
              <p className="mt-2 max-w-2xl text-slate-600">
                Cadastro real local-first. Itens podem ser compraveis, produziveis e/ou vendaveis.
              </p>
            </div>
            <button
              type="button"
              onClick={loadCatalog}
              disabled={loading || saving}
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-300 px-4 py-3 font-semibold text-slate-700 hover:bg-slate-100 disabled:opacity-50"
            >
              <ArrowClockwise size={18} />
              Recarregar
            </button>
          </div>
        </div>
      </header>

      <section className="mx-auto grid max-w-7xl gap-6 px-6 py-8 lg:grid-cols-[360px_1fr]">
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
            <h2 className="text-lg font-bold text-slate-950">
              {editingItem ? "Editar item" : "Novo item"}
            </h2>
            <div className="mt-5 space-y-4">
              <label className="block text-sm font-semibold text-slate-700">
                Nome
                <input
                  value={itemForm.name}
                  onChange={(event) => setItemForm({ ...itemForm, name: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="Farinha de trigo"
                />
              </label>
              <label className="block text-sm font-semibold text-slate-700">
                SKU
                <input
                  value={itemForm.sku}
                  onChange={(event) => setItemForm({ ...itemForm, sku: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="FARINHA-1KG"
                />
              </label>
              <label className="block text-sm font-semibold text-slate-700">
                Unidade base
                <select
                  value={itemForm.baseUnitCode}
                  onChange={(event) =>
                    setItemForm({ ...itemForm, baseUnitCode: event.target.value })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  {itemBaseUnits.map((unit) => (
                    <option key={unit.code} value={unit.code}>
                      {unit.symbol} - {unit.name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="block text-sm font-semibold text-slate-700">
                Descricao
                <textarea
                  value={itemForm.description}
                  onChange={(event) =>
                    setItemForm({ ...itemForm, description: event.target.value })
                  }
                  className="mt-2 min-h-20 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                />
              </label>
              <div>
                <p className="text-sm font-semibold text-slate-700">Capacidades</p>
                <div className="mt-2 grid grid-cols-3 gap-2 text-sm">
                  {[
                    ["purchasable", "Compra"],
                    ["producible", "Producao"],
                    ["sellable", "Venda"],
                  ].map(([key, label]) => (
                    <label
                      key={key}
                      className="flex items-center gap-2 rounded-xl bg-slate-100 px-3 py-2"
                    >
                      <input
                        type="checkbox"
                        checked={
                          itemForm[
                            key as keyof Pick<
                              ItemFormState,
                              "purchasable" | "producible" | "sellable"
                            >
                          ] as boolean
                        }
                        onChange={(event) => {
                          const checked = event.target.checked;
                          setItemForm({
                            ...itemForm,
                            [key]: checked,
                            ...(key === "sellable" && !checked ? { defaultSalePrice: "" } : {}),
                          });
                        }}
                      />
                      {label}
                    </label>
                  ))}
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Preco venda
                  <input
                    value={itemForm.defaultSalePrice}
                    onChange={(event) =>
                      setItemForm({ ...itemForm, defaultSalePrice: event.target.value })
                    }
                    disabled={!itemForm.sellable}
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="12,50"
                  />
                  {!itemForm.sellable && (
                    <span className="mt-1 block text-xs font-normal text-slate-500">
                      Marque Venda para informar preco.
                    </span>
                  )}
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Reposicao atomica
                  <input
                    value={itemForm.reorderQuantityAtomic}
                    onChange={(event) =>
                      setItemForm({ ...itemForm, reorderQuantityAtomic: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="1000"
                  />
                </label>
              </div>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={persistItem}
                  disabled={saving || loading || itemForm.name.trim().length === 0}
                  className="inline-flex flex-1 items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
                >
                  <Plus size={18} />
                  {editingItem ? "Salvar" : "Criar"}
                </button>
                {editingItem && (
                  <button
                    type="button"
                    onClick={resetItemForm}
                    className="rounded-xl border border-slate-300 px-4 py-3 font-semibold text-slate-700"
                  >
                    Cancelar
                  </button>
                )}
              </div>
            </div>
          </div>
        </aside>

        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-2">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Total</p>
              <p className="mt-2 text-3xl font-bold text-slate-950">{items.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Ativos</p>
              <p className="mt-2 text-3xl font-bold text-green-700">{activeItems.length}</p>
            </div>
          </div>

          <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-200 px-6 py-4">
              <h2 className="text-lg font-bold text-slate-950">Itens cadastrados</h2>
            </div>
            {loading ? (
              <p className="p-6 text-slate-600">Carregando catalogo...</p>
            ) : items.length === 0 ? (
              <p className="p-6 text-slate-600">Nenhum item cadastrado ainda.</p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-sm">
                  <thead className="bg-slate-50 text-slate-600">
                    <tr>
                      <th className="px-6 py-3 font-semibold">Nome</th>
                      <th className="px-6 py-3 font-semibold">Capacidades</th>
                      <th className="px-6 py-3 font-semibold">Preco</th>
                      <th className="px-6 py-3 font-semibold">Status</th>
                      <th className="px-6 py-3 font-semibold">Acoes</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {items.map((item) => (
                      <tr key={item.id} className="hover:bg-slate-50">
                        <td className="px-6 py-4">
                          <button
                            type="button"
                            onClick={() => void refreshSelectedItem(item.id)}
                            className="text-left font-semibold text-slate-950 hover:text-pink-700"
                          >
                            {item.name}
                          </button>
                          <p className="text-xs text-slate-500">
                            {item.sku ?? "Sem SKU"} - base {item.baseUnitCode}
                          </p>
                        </td>
                        <td className="px-6 py-4 text-slate-700">
                          {capabilityLabels(item.capabilities)}
                        </td>
                        <td className="px-6 py-4 text-slate-700">
                          {formatMoney(item.defaultSalePrice)}
                        </td>
                        <td className="px-6 py-4">
                          <span
                            className={`rounded-full px-3 py-1 text-xs font-bold ${
                              item.archivedAtMs
                                ? "bg-slate-200 text-slate-700"
                                : "bg-green-100 text-green-700"
                            }`}
                          >
                            {item.archivedAtMs ? "Arquivado" : "Ativo"}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex flex-wrap gap-2">
                            <button
                              type="button"
                              onClick={() => startEditingItem(item)}
                              className="rounded-lg border border-slate-300 px-3 py-2 font-semibold text-slate-700"
                            >
                              Editar
                            </button>
                            <button
                              type="button"
                              onClick={() => void toggleItemArchive(item)}
                              disabled={saving}
                              className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 font-semibold text-slate-700 disabled:opacity-50"
                            >
                              {item.archivedAtMs ? (
                                <ArrowClockwise size={16} />
                              ) : (
                                <Archive size={16} />
                              )}
                              {item.archivedAtMs ? "Restaurar" : "Arquivar"}
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-bold text-slate-950">Embalagens</h2>
                <p className="mt-1 text-sm text-slate-600">
                  {selectedItem
                    ? `Item selecionado: ${selectedItem.name}`
                    : "Selecione um item para ver e cadastrar embalagens."}
                </p>
              </div>
              <Package size={28} className="text-pink-600" />
            </div>

            {selectedItem && (
              <div className="mt-6 grid gap-6 lg:grid-cols-[320px_1fr]">
                <div className="space-y-3">
                  <input
                    value={packagingForm.name}
                    onChange={(event) =>
                      setPackagingForm({ ...packagingForm, name: event.target.value })
                    }
                    className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="Saco 1 kg"
                  />
                  <select
                    value={packagingForm.enteredUnitCode}
                    onChange={(event) =>
                      setPackagingForm({ ...packagingForm, enteredUnitCode: event.target.value })
                    }
                    className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  >
                    {compatiblePackagingUnits.map((unit) => (
                      <option key={unit.code} value={unit.code}>
                        {unit.symbol} - {unit.name}
                      </option>
                    ))}
                  </select>
                  <div className="grid grid-cols-2 gap-3">
                    <input
                      value={packagingForm.conversionNumeratorAtomic}
                      onChange={(event) =>
                        setPackagingForm({
                          ...packagingForm,
                          conversionNumeratorAtomic: event.target.value,
                        })
                      }
                      className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                      placeholder="Numerador"
                    />
                    <input
                      value={packagingForm.conversionDenominator}
                      onChange={(event) =>
                        setPackagingForm({
                          ...packagingForm,
                          conversionDenominator: event.target.value,
                        })
                      }
                      className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                      placeholder="Denominador"
                    />
                  </div>
                  <button
                    type="button"
                    onClick={createPackaging}
                    disabled={saving || packagingForm.name.trim().length === 0}
                    className="w-full rounded-xl bg-slate-900 px-4 py-3 font-semibold text-white hover:bg-slate-800 disabled:bg-slate-300"
                  >
                    Criar embalagem
                  </button>
                </div>

                <div className="space-y-3">
                  {selectedItem.packagings.length === 0 ? (
                    <p className="rounded-2xl bg-slate-50 p-4 text-sm text-slate-600">
                      Nenhuma embalagem cadastrada para este item.
                    </p>
                  ) : (
                    selectedItem.packagings.map((packaging) => (
                      <div
                        key={packaging.id}
                        className="flex items-center justify-between gap-4 rounded-2xl border border-slate-200 p-4"
                      >
                        <div>
                          <p className="font-semibold text-slate-950">{packaging.name}</p>
                          <p className="text-sm text-slate-600">
                            {packaging.enteredUnit.symbol} &rarr;{" "}
                            {packaging.conversionNumeratorAtomic}/{packaging.conversionDenominator}{" "}
                            atomicos de {packaging.baseUnit.symbol}
                          </p>
                          <p className="text-xs text-slate-500">
                            Atualizado em {formatDateTime(packaging.updatedAtMs)}
                          </p>
                        </div>
                        <button
                          type="button"
                          onClick={() => void togglePackagingArchive(packaging)}
                          className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700"
                        >
                          {packaging.archivedAtMs ? (
                            <ArrowClockwise size={16} />
                          ) : (
                            <XCircle size={16} />
                          )}
                          {packaging.archivedAtMs ? "Restaurar" : "Arquivar"}
                        </button>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </section>
    </main>
  );
}

export default ProductsPage;
