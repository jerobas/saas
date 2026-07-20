import { ArrowClockwise, Factory, PlayCircle } from "@phosphor-icons/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  catalogGateway,
  inventoryGateway,
  type InventoryBalanceResponse,
  type ItemResponse,
  type LotResponse,
  productionGateway,
  type ProductionDocumentResponse,
  type RecipeResponse,
  type RecipeSummaryResponse,
  recipeGateway,
} from "../../gateways/desktopBridge";

interface ProductionComponentFormState {
  itemId: number;
  quantityAtomic: string;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  lotId: string;
}

interface ProductionFormState {
  recipeId: string;
  occurredOn: string;
  outputQuantityAtomic: string;
  outputLotCode: string;
  outputExpiresOn: string;
  directCost: string;
  notes: string;
  components: ProductionComponentFormState[];
}

const todayISO = () => new Date().toISOString().slice(0, 10);

const emptyForm = (): ProductionFormState => ({
  recipeId: "",
  occurredOn: todayISO(),
  outputQuantityAtomic: "",
  outputLotCode: "",
  outputExpiresOn: "",
  directCost: "0",
  notes: "",
  components: [],
});

const parseInteger = (value: string) => {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const parseMoneyMicro = (value: string) => {
  const normalized = value.trim().replace(",", ".");
  if (normalized.length === 0) return 0;
  const parsed = Number.parseFloat(normalized);
  if (!Number.isFinite(parsed)) return undefined;
  return Math.round(parsed * 1_000_000);
};

const optionalText = (value: string) => {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
};

const messageFromError = (error: unknown, fallback: string) =>
  error instanceof Error ? error.message : fallback;

const buildIdempotencyKey = () =>
  `production-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

const formatInventoryMicro = (micro: number) =>
  new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency: "BRL",
  }).format(micro / 1_000_000);

const estimateInventoryValue = (
  quantityAtomic: number,
  balance: InventoryBalanceResponse | null,
) => {
  if (!balance || balance.quantityAtomic <= 0 || quantityAtomic <= 0) return 0;
  return Math.round((balance.inventoryValueMicro * quantityAtomic) / balance.quantityAtomic);
};

const lotLabel = (lot: LotResponse) => {
  const code = lot.lotCode ? `${lot.lotCode} · ` : "";
  const expiry = lot.expiresOn ? ` · vence ${lot.expiresOn}` : "";
  return `${code}${lot.availableQuantityAtomic} atomicos${expiry}`;
};

function ProductionPage() {
  const [recipes, setRecipes] = useState<RecipeSummaryResponse[]>([]);
  const [selectedRecipe, setSelectedRecipe] = useState<RecipeResponse | null>(null);
  const [outputItem, setOutputItem] = useState<ItemResponse | null>(null);
  const [componentItems, setComponentItems] = useState<Record<number, ItemResponse>>({});
  const [eligibleLots, setEligibleLots] = useState<Record<number, LotResponse[]>>({});
  const [balances, setBalances] = useState<Record<number, InventoryBalanceResponse>>({});
  const [latestProduction, setLatestProduction] = useState<ProductionDocumentResponse | null>(null);
  const [form, setForm] = useState<ProductionFormState>(() => emptyForm());
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const currentRevision = selectedRecipe?.currentRevision ?? null;

  const componentRows = useMemo(
    () =>
      form.components.map((component) => ({
        ...component,
        item: componentItems[component.itemId] ?? null,
        lots: eligibleLots[component.itemId] ?? [],
        balance: balances[component.itemId] ?? null,
      })),
    [balances, componentItems, eligibleLots, form.components],
  );

  const outputBalance = outputItem ? (balances[outputItem.id] ?? null) : null;

  const productionPreview = useMemo(() => {
    const directCostMicro = parseMoneyMicro(form.directCost) ?? 0;
    const outputQuantityAtomic = parseInteger(form.outputQuantityAtomic) ?? 0;
    const components = componentRows.map((component) => {
      const quantityAtomic = parseInteger(component.quantityAtomic) ?? 0;
      const selectedLotID = parseInteger(component.lotId);
      const lot =
        component.lots.find((current) => current.id === selectedLotID) ?? component.lots[0] ?? null;
      return {
        ...component,
        quantityAtomic,
        lot,
        estimatedValueMicro: estimateInventoryValue(quantityAtomic, component.balance),
      };
    });
    const componentValueMicro = components.reduce(
      (acc, component) => acc + component.estimatedValueMicro,
      0,
    );
    return {
      outputQuantityAtomic,
      directCostMicro,
      componentValueMicro,
      estimatedOutputValueMicro: componentValueMicro + directCostMicro,
      components,
    };
  }, [componentRows, form.directCost, form.outputQuantityAtomic]);

  const refreshInventory = useCallback(
    async (recipe: RecipeResponse | null, occurredOn: string) => {
      if (!recipe) return;
      const itemIds = [
        recipe.outputItemId,
        ...recipe.currentRevision.components.map((component) => component.itemId),
      ];
      const uniqueItemIds = [...new Set(itemIds)];

      const [balancePairs, lotPairs] = await Promise.all([
        Promise.all(
          uniqueItemIds.map(
            async (itemId) => [itemId, await inventoryGateway.getInventoryBalance(itemId)] as const,
          ),
        ),
        Promise.all(
          recipe.currentRevision.components.map(
            async (component) =>
              [
                component.itemId,
                await inventoryGateway.listEligibleFefoLots(component.itemId, occurredOn),
              ] as const,
          ),
        ),
      ]);

      setBalances(Object.fromEntries(balancePairs));
      setEligibleLots(Object.fromEntries(lotPairs));
    },
    [],
  );

  const loadRecipe = useCallback(
    async (recipeId: number, occurredOn = form.occurredOn) => {
      const recipe = await recipeGateway.getRecipe(recipeId);
      const output = await catalogGateway.getItem(recipe.outputItemId);
      const componentPairs = await Promise.all(
        recipe.currentRevision.components.map(
          async (component) =>
            [component.itemId, await catalogGateway.getItem(component.itemId)] as const,
        ),
      );

      setSelectedRecipe(recipe);
      setOutputItem(output);
      setComponentItems(Object.fromEntries(componentPairs));
      setForm((current) => ({
        ...current,
        recipeId: String(recipe.id),
        outputQuantityAtomic: String(recipe.currentRevision.standardYieldQuantityAtomic),
        components: recipe.currentRevision.components.map((component) => ({
          itemId: component.itemId,
          quantityAtomic: String(component.quantityAtomic),
          enteredUnitCode: component.enteredUnitCode,
          enteredPackagingName: component.enteredPackagingName,
          conversionNumeratorAtomic: component.conversionNumeratorAtomic,
          conversionDenominator: component.conversionDenominator,
          lotId: "",
        })),
      }));
      await refreshInventory(recipe, occurredOn);
    },
    [form.occurredOn, refreshInventory],
  );

  const loadPage = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const recipePage = await recipeGateway.listRecipes({
        archiveFilter: "ACTIVE",
        pageSize: 100,
      });
      setRecipes(recipePage.items);
      const firstRecipe = recipePage.items[0];
      if (firstRecipe) {
        await loadRecipe(firstRecipe.id, form.occurredOn);
      } else {
        setSelectedRecipe(null);
        setOutputItem(null);
        setComponentItems({});
        setEligibleLots({});
        setBalances({});
      }
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar producao."),
      });
    } finally {
      setLoading(false);
    }
  }, [form.occurredOn, loadRecipe]);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  const reloadLotsForDate = async (occurredOn: string) => {
    setForm((current) => ({ ...current, occurredOn }));
    if (selectedRecipe) {
      await refreshInventory(selectedRecipe, occurredOn);
    }
  };

  const selectRecipe = async (recipeId: string) => {
    setForm((current) => ({ ...current, recipeId }));
    const parsed = parseInteger(recipeId);
    if (!parsed) {
      setSelectedRecipe(null);
      setOutputItem(null);
      return;
    }
    setLoading(true);
    setMessage(null);
    try {
      await loadRecipe(parsed, form.occurredOn);
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar a receita."),
      });
    } finally {
      setLoading(false);
    }
  };

  const updateComponent = (itemId: number, patch: Partial<ProductionComponentFormState>) => {
    setForm((current) => ({
      ...current,
      components: current.components.map((component) =>
        component.itemId === itemId ? { ...component, ...patch } : component,
      ),
    }));
  };

  const postProduction = async () => {
    if (saving || !selectedRecipe || !outputItem || !currentRevision) return;

    const outputQuantityAtomic = parseInteger(form.outputQuantityAtomic);
    const directCostMicro = parseMoneyMicro(form.directCost);
    if (!outputQuantityAtomic) {
      setMessage({ type: "error", text: "Informe a quantidade de saida." });
      return;
    }
    if (directCostMicro === undefined) {
      setMessage({ type: "error", text: "Informe um custo direto valido." });
      return;
    }

    const inputs = form.components.map((component) => ({
      component,
      quantityAtomic: parseInteger(component.quantityAtomic),
      lotId: parseInteger(component.lotId),
    }));
    if (inputs.some((input) => !input.quantityAtomic)) {
      setMessage({ type: "error", text: "Informe a quantidade real de todos os insumos." });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const posted = await productionGateway.postProduction({
        idempotencyKey: buildIdempotencyKey(),
        recipeRevisionId: currentRevision.id,
        occurredOn: form.occurredOn,
        directCostMicro,
        notes: optionalText(form.notes),
        output: {
          quantityAtomic: outputQuantityAtomic,
          enteredUnitCode: outputItem.baseUnit.code,
          conversionNumeratorAtomic: outputItem.baseUnit.numeratorAtomic,
          conversionDenominator: outputItem.baseUnit.denominator,
          lotCode: optionalText(form.outputLotCode),
          expiresOn: optionalText(form.outputExpiresOn),
        },
        inputs: inputs.map(({ component, quantityAtomic, lotId }) => ({
          itemId: component.itemId,
          quantityAtomic: quantityAtomic ?? 0,
          enteredUnitCode: component.enteredUnitCode,
          enteredPackagingName: component.enteredPackagingName,
          conversionNumeratorAtomic: component.conversionNumeratorAtomic,
          conversionDenominator: component.conversionDenominator,
          lotId,
        })),
      });
      setLatestProduction(posted);
      setMessage({ type: "success", text: `Producao #${posted.id} postada.` });
      setForm((current) => ({
        ...current,
        outputLotCode: "",
        notes: "",
        components: current.components.map((component) => ({ ...component, lotId: "" })),
      }));
      await refreshInventory(selectedRecipe, form.occurredOn);
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel postar a producao."),
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
                Producao V2
              </p>
              <h1 className="mt-2 text-3xl font-bold text-slate-950">Postar producao</h1>
              <p className="mt-2 max-w-2xl text-slate-600">
                Fluxo minimo real: usa uma receita publicada, consome insumos por FEFO ou lote
                manual e cria o lote de saida com custo transferido.
              </p>
            </div>
            <button
              type="button"
              onClick={() => void loadPage()}
              className="inline-flex items-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-3 text-sm font-semibold text-slate-700 shadow-sm hover:bg-slate-50"
            >
              <ArrowClockwise size={18} />
              Recarregar
            </button>
          </div>
        </div>
      </header>

      <section className="mx-auto grid max-w-7xl gap-6 px-6 py-8 xl:grid-cols-[420px_1fr]">
        <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="mb-5 flex items-center gap-3">
            <div className="rounded-2xl bg-pink-50 p-3 text-pink-600">
              <Factory size={24} />
            </div>
            <div>
              <h2 className="text-xl font-bold text-slate-950">Nova producao</h2>
              <p className="text-sm text-slate-500">Preview simples antes da postagem.</p>
            </div>
          </div>

          {message && (
            <div
              className={`mb-4 rounded-2xl border px-4 py-3 text-sm ${
                message.type === "success"
                  ? "border-emerald-200 bg-emerald-50 text-emerald-800"
                  : "border-red-200 bg-red-50 text-red-700"
              }`}
            >
              {message.text}
            </div>
          )}

          <div className="space-y-4">
            <label className="block">
              <span className="mb-1 block text-sm font-semibold text-slate-700">Receita</span>
              <select
                value={form.recipeId}
                onChange={(event) => void selectRecipe(event.target.value)}
                className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm"
              >
                <option value="">Selecione</option>
                {recipes.map((recipe) => (
                  <option key={recipe.id} value={recipe.id}>
                    {recipe.name} · v{recipe.currentRevision.number}
                  </option>
                ))}
              </select>
            </label>

            <label className="block">
              <span className="mb-1 block text-sm font-semibold text-slate-700">Data</span>
              <input
                type="date"
                value={form.occurredOn}
                onChange={(event) => void reloadLotsForDate(event.target.value)}
                className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
              />
            </label>

            <div className="grid grid-cols-2 gap-3">
              <label className="block">
                <span className="mb-1 block text-sm font-semibold text-slate-700">
                  Quantidade saida
                </span>
                <input
                  value={form.outputQuantityAtomic}
                  onChange={(event) =>
                    setForm((current) => ({
                      ...current,
                      outputQuantityAtomic: event.target.value,
                    }))
                  }
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-semibold text-slate-700">
                  Custo direto
                </span>
                <input
                  value={form.directCost}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, directCost: event.target.value }))
                  }
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
                  placeholder="0,00"
                />
              </label>
            </div>

            <div className="grid grid-cols-2 gap-3">
              <label className="block">
                <span className="mb-1 block text-sm font-semibold text-slate-700">Lote saida</span>
                <input
                  value={form.outputLotCode}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, outputLotCode: event.target.value }))
                  }
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
                  placeholder="PROD-001"
                />
              </label>
              <label className="block">
                <span className="mb-1 block text-sm font-semibold text-slate-700">
                  Validade saida
                </span>
                <input
                  type="date"
                  value={form.outputExpiresOn}
                  onChange={(event) =>
                    setForm((current) => ({ ...current, outputExpiresOn: event.target.value }))
                  }
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
                />
              </label>
            </div>

            <label className="block">
              <span className="mb-1 block text-sm font-semibold text-slate-700">Notas</span>
              <textarea
                value={form.notes}
                onChange={(event) =>
                  setForm((current) => ({ ...current, notes: event.target.value }))
                }
                className="min-h-20 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
              />
            </label>

            <button
              type="button"
              disabled={saving || loading || !selectedRecipe}
              onClick={() => void postProduction()}
              className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 text-sm font-semibold text-white shadow-sm hover:bg-pink-700 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              <PlayCircle size={18} />
              {saving ? "Postando..." : "Postar producao"}
            </button>
          </div>
        </div>

        <div className="space-y-6">
          <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-xl font-bold text-slate-950">Receita atual</h2>
            {!selectedRecipe || !outputItem ? (
              <p className="mt-4 text-sm text-slate-500">
                {loading ? "Carregando..." : "Nenhuma receita ativa encontrada."}
              </p>
            ) : (
              <div className="mt-4 grid gap-4 md:grid-cols-3">
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-xs font-semibold uppercase text-slate-500">Receita</p>
                  <p className="mt-1 font-semibold text-slate-950">{selectedRecipe.name}</p>
                  <p className="text-sm text-slate-500">v{currentRevision?.number}</p>
                </div>
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-xs font-semibold uppercase text-slate-500">Saida</p>
                  <p className="mt-1 font-semibold text-slate-950">{outputItem.name}</p>
                  <p className="text-sm text-slate-500">base {outputItem.baseUnit.code}</p>
                </div>
                <div className="rounded-2xl bg-slate-50 p-4">
                  <p className="text-xs font-semibold uppercase text-slate-500">Saldo saida</p>
                  <p className="mt-1 font-semibold text-slate-950">
                    {outputBalance?.quantityAtomic ?? 0} atomicos
                  </p>
                  <p className="text-sm text-slate-500">
                    {formatInventoryMicro(outputBalance?.inventoryValueMicro ?? 0)}
                  </p>
                </div>
              </div>
            )}
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-xl font-bold text-slate-950">Preview de producao</h2>
            {!selectedRecipe || !outputItem ? (
              <p className="mt-4 text-sm text-slate-500">
                Selecione uma receita para calcular o preview.
              </p>
            ) : (
              <div className="mt-4 space-y-4">
                <div className="grid gap-4 md:grid-cols-3">
                  <div className="rounded-2xl bg-slate-50 p-4">
                    <p className="text-xs font-semibold uppercase text-slate-500">Yield alvo</p>
                    <p className="mt-1 font-semibold text-slate-950">
                      {productionPreview.outputQuantityAtomic} atomicos
                    </p>
                  </div>
                  <div className="rounded-2xl bg-slate-50 p-4">
                    <p className="text-xs font-semibold uppercase text-slate-500">
                      Custo componentes
                    </p>
                    <p className="mt-1 font-semibold text-slate-950">
                      {formatInventoryMicro(productionPreview.componentValueMicro)}
                    </p>
                  </div>
                  <div className="rounded-2xl bg-slate-50 p-4">
                    <p className="text-xs font-semibold uppercase text-slate-500">Custo estimado</p>
                    <p className="mt-1 font-semibold text-slate-950">
                      {formatInventoryMicro(productionPreview.estimatedOutputValueMicro)}
                    </p>
                    <p className="text-sm text-slate-500">
                      inclui direto {formatInventoryMicro(productionPreview.directCostMicro)}
                    </p>
                  </div>
                </div>

                <div className="divide-y divide-slate-100 rounded-2xl border border-slate-200">
                  {productionPreview.components.length === 0 ? (
                    <p className="p-4 text-sm text-slate-500">Nenhum insumo no preview.</p>
                  ) : (
                    productionPreview.components.map((component) => (
                      <div key={component.itemId} className="grid gap-2 p-4 md:grid-cols-3">
                        <div>
                          <p className="font-semibold text-slate-950">
                            {component.item?.name ?? `Item #${component.itemId}`}
                          </p>
                          <p className="text-sm text-slate-500">
                            {component.quantityAtomic} atomicos esperados
                          </p>
                        </div>
                        <p className="text-sm text-slate-700">
                          FEFO sugerido:{" "}
                          <strong>
                            {component.lot?.lotCode ?? `lote #${component.lot?.id ?? "-"}`}
                          </strong>
                        </p>
                        <p className="text-sm text-slate-700">
                          valor estimado:{" "}
                          <strong>{formatInventoryMicro(component.estimatedValueMicro)}</strong>
                        </p>
                      </div>
                    ))
                  )}
                </div>
              </div>
            )}
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-200 p-6">
              <h2 className="text-xl font-bold text-slate-950">Consumo real</h2>
              <p className="mt-1 text-sm text-slate-500">
                Os valores abaixo começam pela revisão da receita, mas podem ser ajustados antes da
                postagem.
              </p>
            </div>
            <div className="divide-y divide-slate-100">
              {componentRows.length === 0 ? (
                <p className="p-6 text-sm text-slate-500">Nenhum componente carregado.</p>
              ) : (
                componentRows.map((component) => (
                  <div
                    key={component.itemId}
                    className="grid gap-4 p-6 lg:grid-cols-[1fr_160px_260px]"
                  >
                    <div>
                      <p className="font-semibold text-slate-950">
                        {component.item?.name ?? `Item #${component.itemId}`}
                      </p>
                      <p className="mt-1 text-sm text-slate-500">
                        Receita: {component.enteredPackagingName ?? component.enteredUnitCode} ·
                        conversao {component.conversionNumeratorAtomic}/
                        {component.conversionDenominator}
                      </p>
                      <p className="mt-1 text-sm text-slate-500">
                        Saldo: {component.balance?.quantityAtomic ?? 0} atomicos ·{" "}
                        {formatInventoryMicro(component.balance?.inventoryValueMicro ?? 0)}
                      </p>
                    </div>
                    <label className="block">
                      <span className="mb-1 block text-sm font-semibold text-slate-700">
                        Quantidade
                      </span>
                      <input
                        value={component.quantityAtomic}
                        onChange={(event) =>
                          updateComponent(component.itemId, { quantityAtomic: event.target.value })
                        }
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm"
                      />
                    </label>
                    <label className="block">
                      <span className="mb-1 block text-sm font-semibold text-slate-700">
                        Lote de entrada
                      </span>
                      <select
                        value={component.lotId}
                        onChange={(event) =>
                          updateComponent(component.itemId, { lotId: event.target.value })
                        }
                        className="w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm"
                      >
                        <option value="">FEFO automatico</option>
                        {component.lots.map((lot) => (
                          <option key={lot.id} value={lot.id}>
                            {lotLabel(lot)}
                          </option>
                        ))}
                      </select>
                    </label>
                  </div>
                ))
              )}
            </div>
          </section>

          <section className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-xl font-bold text-slate-950">Ultima postagem</h2>
            {!latestProduction ? (
              <p className="mt-4 text-sm text-slate-500">Nenhuma producao postada nesta sessao.</p>
            ) : (
              <div className="mt-4 rounded-2xl bg-slate-50 p-4 text-sm text-slate-700">
                <p className="font-semibold text-slate-950">
                  Producao #{latestProduction.id} · seq {latestProduction.postingSequence}
                </p>
                <p className="mt-1">
                  Saida: {latestProduction.outputLine.quantityAtomic} atomicos · valor{" "}
                  {formatInventoryMicro(latestProduction.outputLine.inventoryValueMicro)}
                </p>
                <p className="mt-1">
                  Custo direto: {formatInventoryMicro(latestProduction.directCostMicro)}
                </p>
              </div>
            )}
          </section>
        </div>
      </section>
    </main>
  );
}

export default ProductionPage;
