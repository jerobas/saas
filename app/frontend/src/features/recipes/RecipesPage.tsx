import {
  Archive,
  ArrowClockwise,
  FileText,
  PencilSimple,
  Plus,
  UploadSimple,
} from "@phosphor-icons/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  catalogGateway,
  type ItemSummaryResponse,
  recipeGateway,
  type RecipeResponse,
  type RecipeRevisionResponse,
  type RecipeSummaryResponse,
} from "../../gateways/desktopBridge";

interface RecipeFormState {
  name: string;
  outputItemId: string;
  standardYieldQuantityAtomic: string;
  instructions: string;
  preparationTimeMinutes: string;
  componentItemId: string;
  componentQuantityAtomic: string;
}

const emptyForm = (): RecipeFormState => ({
  name: "",
  outputItemId: "",
  standardYieldQuantityAtomic: "1000",
  instructions: "",
  preparationTimeMinutes: "0",
  componentItemId: "",
  componentQuantityAtomic: "",
});

const parseInteger = (value: string) => {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isFinite(parsed) ? parsed : undefined;
};

const messageFromError = (error: unknown, fallback: string) =>
  error instanceof Error ? error.message : fallback;

const activeOnly = (item: { archivedAtMs?: number | null }) =>
  item.archivedAtMs === undefined || item.archivedAtMs === null;

function RecipesPage() {
  const [recipes, setRecipes] = useState<RecipeSummaryResponse[]>([]);
  const [outputItems, setOutputItems] = useState<ItemSummaryResponse[]>([]);
  const [componentItems, setComponentItems] = useState<ItemSummaryResponse[]>([]);
  const [selectedRecipe, setSelectedRecipe] = useState<RecipeResponse | null>(null);
  const [revisions, setRevisions] = useState<RecipeRevisionResponse[]>([]);
  const [form, setForm] = useState<RecipeFormState>(() => emptyForm());
  const [renameName, setRenameName] = useState("");
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const selectedComponentItem = useMemo(
    () => componentItems.find((item) => String(item.id) === form.componentItemId) ?? null,
    [form.componentItemId, componentItems],
  );

  const availableComponentItems = useMemo(
    () =>
      componentItems.filter((item) => String(item.id) !== form.outputItemId && activeOnly(item)),
    [componentItems, form.outputItemId],
  );

  const catalogItemNames = useMemo(() => {
    const names = new Map<number, string>();
    [...outputItems, ...componentItems].forEach((item) => {
      names.set(item.id, item.name);
    });
    return names;
  }, [componentItems, outputItems]);

  const loadSelectedRecipe = useCallback(async (id: number) => {
    const [recipe, revisionList] = await Promise.all([
      recipeGateway.getRecipe(id),
      recipeGateway.listRecipeRevisions(id),
    ]);
    setSelectedRecipe(recipe);
    setRevisions(revisionList);
    setRenameName(recipe.name);
  }, []);

  const loadPage = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const [recipePage, produciblePage, itemPage] = await Promise.all([
        recipeGateway.listRecipes({ archiveFilter: "ALL", pageSize: 100 }),
        catalogGateway.listItems({
          archiveFilter: "ACTIVE",
          requireCapabilities: { purchasable: false, producible: true, sellable: false },
          pageSize: 100,
        }),
        catalogGateway.listItems({
          archiveFilter: "ACTIVE",
          requireCapabilities: { purchasable: false, producible: false, sellable: false },
          pageSize: 100,
        }),
      ]);
      setRecipes(recipePage.items);
      setOutputItems(produciblePage.items);
      setComponentItems(itemPage.items);

      const firstRecipe = recipePage.items[0];
      if (firstRecipe) {
        await loadSelectedRecipe(firstRecipe.id);
      } else {
        setSelectedRecipe(null);
        setRevisions([]);
        setRenameName("");
      }

      setForm((current) => ({
        ...current,
        outputItemId: current.outputItemId || String(produciblePage.items[0]?.id ?? ""),
        componentItemId: current.componentItemId || String(itemPage.items[0]?.id ?? ""),
      }));
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel carregar receitas."),
      });
    } finally {
      setLoading(false);
    }
  }, [loadSelectedRecipe]);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  const buildRevisionRequest = () => {
    const outputItemId = parseInteger(form.outputItemId);
    const componentItemId = parseInteger(form.componentItemId);
    const standardYieldQuantityAtomic = parseInteger(form.standardYieldQuantityAtomic);
    const componentQuantityAtomic = parseInteger(form.componentQuantityAtomic);
    const preparationTimeMinutes = parseInteger(form.preparationTimeMinutes);

    if (
      !outputItemId ||
      !componentItemId ||
      !standardYieldQuantityAtomic ||
      !componentQuantityAtomic
    ) {
      setMessage({
        type: "error",
        text: "Informe item de saida, rendimento, componente e quantidade.",
      });
      return null;
    }
    if (outputItemId === componentItemId) {
      setMessage({ type: "error", text: "A receita nao pode consumir o proprio item de saida." });
      return null;
    }
    if (!selectedComponentItem) {
      setMessage({ type: "error", text: "Selecione um componente valido." });
      return null;
    }

    return {
      outputItemId,
      revision: {
        standardYieldQuantityAtomic,
        instructions: form.instructions,
        preparationTimeMinutes: preparationTimeMinutes ?? 0,
        components: [
          {
            order: 1,
            itemId: componentItemId,
            quantityAtomic: componentQuantityAtomic,
            sourceType: "UNIT" as const,
            unitCode: selectedComponentItem.baseUnitCode,
          },
        ],
      },
    };
  };

  const createRecipe = async () => {
    if (saving) return;
    const revisionRequest = buildRevisionRequest();
    if (!revisionRequest) return;
    if (form.name.trim().length === 0) {
      setMessage({ type: "error", text: "Informe o nome da receita." });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const created = await recipeGateway.createRecipe({
        name: form.name.trim(),
        outputItemId: revisionRequest.outputItemId,
        revision: revisionRequest.revision,
      });
      setForm(emptyForm());
      await loadPage();
      await loadSelectedRecipe(created.id);
      setMessage({ type: "success", text: `Receita "${created.name}" criada.` });
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel criar a receita."),
      });
    } finally {
      setSaving(false);
    }
  };

  const publishRevision = async () => {
    if (saving || !selectedRecipe) return;
    const revisionRequest = buildRevisionRequest();
    if (!revisionRequest) return;

    setSaving(true);
    setMessage(null);
    try {
      const revision = await recipeGateway.publishRecipeRevision(selectedRecipe.id, {
        expectedLatestRevision: selectedRecipe.currentRevision.number,
        expectedUpdatedAtMs: selectedRecipe.updatedAtMs,
        revision: revisionRequest.revision,
      });
      await loadPage();
      await loadSelectedRecipe(selectedRecipe.id);
      setMessage({ type: "success", text: `Revisao ${revision.number} publicada.` });
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel publicar a revisao."),
      });
    } finally {
      setSaving(false);
    }
  };

  const renameSelectedRecipe = async () => {
    if (saving || !selectedRecipe) return;
    const name = renameName.trim();
    if (name.length === 0) {
      setMessage({ type: "error", text: "Informe o novo nome da receita." });
      return;
    }
    if (name === selectedRecipe.name) {
      setMessage({ type: "error", text: "O novo nome precisa ser diferente do atual." });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const updated = await recipeGateway.renameRecipe(selectedRecipe.id, {
        name,
        expectedUpdatedAtMs: selectedRecipe.updatedAtMs,
      });
      await loadPage();
      await loadSelectedRecipe(updated.id);
      setMessage({ type: "success", text: `Receita renomeada para "${updated.name}".` });
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel renomear a receita."),
      });
    } finally {
      setSaving(false);
    }
  };

  const toggleArchive = async (recipe: RecipeSummaryResponse) => {
    if (saving) return;
    setSaving(true);
    setMessage(null);
    try {
      const updated = recipe.archivedAtMs
        ? await recipeGateway.restoreRecipe(recipe.id, { expectedUpdatedAtMs: recipe.updatedAtMs })
        : await recipeGateway.archiveRecipe(recipe.id, { expectedUpdatedAtMs: recipe.updatedAtMs });
      await loadPage();
      await loadSelectedRecipe(updated.id);
      setMessage({
        type: "success",
        text: updated.archivedAtMs ? "Receita arquivada." : "Receita restaurada.",
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: messageFromError(error, "Nao foi possivel alterar o estado da receita."),
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
                Receitas V2
              </p>
              <h1 className="mt-2 text-3xl font-bold text-slate-950">Receitas e revisoes</h1>
              <p className="mt-2 max-w-2xl text-slate-600">
                Cadastro real local-first: receita com output fixo e revisoes imutaveis.
              </p>
            </div>
            <button
              type="button"
              onClick={() => void loadPage()}
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
                <h2 className="text-lg font-bold text-slate-950">Editar receita</h2>
                <p className="text-sm text-slate-600">Um componente por vez nesta primeira UI.</p>
              </div>
            </div>

            <div className="mt-5 space-y-4">
              <label className="block text-sm font-semibold text-slate-700">
                Nome
                <input
                  value={form.name}
                  onChange={(event) => setForm({ ...form, name: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="Massa de bolo"
                />
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Item de saida
                <select
                  value={form.outputItemId}
                  onChange={(event) =>
                    setForm({
                      ...form,
                      outputItemId: event.target.value,
                      componentItemId:
                        event.target.value === form.componentItemId ? "" : form.componentItemId,
                    })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  <option value="">Selecione</option>
                  {outputItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} / base {item.baseUnitCode}
                    </option>
                  ))}
                </select>
              </label>

              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Rendimento atomico
                  <input
                    value={form.standardYieldQuantityAtomic}
                    onChange={(event) =>
                      setForm({ ...form, standardYieldQuantityAtomic: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Preparo min.
                  <input
                    value={form.preparationTimeMinutes}
                    onChange={(event) =>
                      setForm({ ...form, preparationTimeMinutes: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
              </div>

              <label className="block text-sm font-semibold text-slate-700">
                Componente
                <select
                  value={form.componentItemId}
                  onChange={(event) => setForm({ ...form, componentItemId: event.target.value })}
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                >
                  <option value="">Selecione</option>
                  {availableComponentItems.map((item) => (
                    <option key={item.id} value={item.id}>
                      {item.name} / base {item.baseUnitCode}
                    </option>
                  ))}
                </select>
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Quantidade atomica do componente
                <input
                  value={form.componentQuantityAtomic}
                  onChange={(event) =>
                    setForm({ ...form, componentQuantityAtomic: event.target.value })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="500"
                />
              </label>

              <label className="block text-sm font-semibold text-slate-700">
                Instrucoes
                <textarea
                  value={form.instructions}
                  onChange={(event) => setForm({ ...form, instructions: event.target.value })}
                  className="mt-2 min-h-24 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  placeholder="Misture, asse..."
                />
              </label>

              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                <button
                  type="button"
                  onClick={() => void createRecipe()}
                  disabled={saving || loading}
                  className="inline-flex items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
                >
                  <Plus size={18} />
                  Criar
                </button>
                <button
                  type="button"
                  onClick={() => void publishRevision()}
                  disabled={saving || loading || !selectedRecipe}
                  className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-300 px-4 py-3 font-semibold text-slate-700 hover:bg-slate-100 disabled:opacity-50"
                >
                  <UploadSimple size={18} />
                  Publicar revisao
                </button>
              </div>

              <p className="text-xs text-slate-500">
                Por enquanto a UI usa unidade base do componente. Embalagens ja existem no contrato,
                mas podem entrar no proximo polimento.
              </p>
            </div>
          </div>
        </aside>

        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-3">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Receitas</p>
              <p className="mt-2 text-3xl font-bold text-slate-950">{recipes.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Outputs produciveis</p>
              <p className="mt-2 text-3xl font-bold text-green-700">{outputItems.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Revisoes selecionadas</p>
              <p className="mt-2 text-3xl font-bold text-blue-700">{revisions.length}</p>
            </div>
          </div>

          <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
            <div className="border-b border-slate-200 px-6 py-4">
              <h2 className="text-lg font-bold text-slate-950">Receitas cadastradas</h2>
            </div>
            {loading ? (
              <p className="p-6 text-slate-600">Carregando receitas...</p>
            ) : recipes.length === 0 ? (
              <p className="p-6 text-slate-600">
                Nenhuma receita cadastrada. Crie um item producivel no Catalogo e use o formulario.
              </p>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-slate-50 text-left text-sm text-slate-700">
                    <tr>
                      <th className="px-6 py-3">Nome</th>
                      <th className="px-6 py-3">Output</th>
                      <th className="px-6 py-3">Revisao atual</th>
                      <th className="px-6 py-3">Status</th>
                      <th className="px-6 py-3">Acoes</th>
                    </tr>
                  </thead>
                  <tbody>
                    {recipes.map((recipe) => (
                      <tr key={recipe.id} className="border-t border-slate-100">
                        <td className="px-6 py-4">
                          <button
                            type="button"
                            onClick={() => void loadSelectedRecipe(recipe.id)}
                            className="text-left font-semibold text-slate-950 hover:text-pink-700"
                          >
                            {recipe.name}
                          </button>
                          <div className="text-xs text-slate-500">#{recipe.id}</div>
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-700">
                          {recipe.outputItemName}
                        </td>
                        <td className="px-6 py-4 text-sm text-slate-700">
                          v{recipe.currentRevision.number} / rendimento{" "}
                          {recipe.currentRevision.standardYieldQuantityAtomic}
                        </td>
                        <td className="px-6 py-4">
                          <span
                            className={`rounded-full px-2 py-1 text-xs font-semibold ${
                              recipe.archivedAtMs
                                ? "bg-slate-100 text-slate-600"
                                : "bg-green-100 text-green-700"
                            }`}
                          >
                            {recipe.archivedAtMs ? "Arquivada" : "Ativa"}
                          </span>
                        </td>
                        <td className="px-6 py-4">
                          <button
                            type="button"
                            onClick={() => void toggleArchive(recipe)}
                            disabled={saving}
                            className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100 disabled:opacity-50"
                          >
                            <Archive size={16} />
                            {recipe.archivedAtMs ? "Restaurar" : "Arquivar"}
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>

          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <h2 className="text-lg font-bold text-slate-950">Detalhe</h2>
            {!selectedRecipe ? (
              <p className="mt-3 text-slate-600">Selecione uma receita para ver revisoes.</p>
            ) : (
              <div className="mt-4 space-y-4">
                <div>
                  <p className="text-sm font-semibold uppercase text-slate-500">
                    {selectedRecipe.name}
                  </p>
                  <p className="mt-1 text-sm text-slate-700">
                    Output #{selectedRecipe.outputItemId} · revisao atual v
                    {selectedRecipe.currentRevision.number}
                  </p>
                </div>
                <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">
                  <label className="block text-sm font-semibold text-slate-700">
                    Renomear receita
                    <input
                      value={renameName}
                      onChange={(event) => setRenameName(event.target.value)}
                      className="mt-2 w-full rounded-xl border border-slate-300 bg-white px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    />
                  </label>
                  <button
                    type="button"
                    onClick={() => void renameSelectedRecipe()}
                    disabled={
                      saving ||
                      loading ||
                      renameName.trim().length === 0 ||
                      renameName.trim() === selectedRecipe.name
                    }
                    className="mt-3 inline-flex items-center gap-2 rounded-xl border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100 disabled:opacity-50"
                  >
                    <PencilSimple size={16} />
                    Renomear
                  </button>
                </div>
                <div className="space-y-3">
                  {revisions.map((revision) => (
                    <div key={revision.id} className="rounded-2xl border border-slate-200 p-4">
                      <div className="flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
                        <div>
                          <div className="flex flex-wrap items-center gap-2">
                            <p className="font-semibold text-slate-950">
                              Revisao {revision.number}
                            </p>
                            {revision.id === selectedRecipe.currentRevision.id && (
                              <span className="rounded-full bg-blue-100 px-2 py-1 text-xs font-semibold text-blue-700">
                                Atual
                              </span>
                            )}
                          </div>
                          <p className="text-sm text-slate-600">
                            Rendimento {revision.standardYieldQuantityAtomic} · preparo{" "}
                            {revision.preparationTimeMinutes} min
                          </p>
                        </div>
                        <span className="text-xs text-slate-500">#{revision.id}</span>
                      </div>
                      <p className="mt-3 rounded-xl bg-slate-50 px-3 py-2 text-xs font-medium text-slate-600">
                        Revisao publicada e imutavel. Para alterar a ficha tecnica, publique uma
                        nova revisao.
                      </p>
                      <p className="mt-3 whitespace-pre-wrap text-sm text-slate-700">
                        {revision.instructions || "Sem instrucoes."}
                      </p>
                      <div className="mt-3 space-y-2">
                        {revision.components.map((component) => (
                          <div
                            key={component.id}
                            className="rounded-xl border border-slate-100 bg-white px-3 py-2 text-sm text-slate-700"
                          >
                            <div className="font-semibold text-slate-900">
                              {component.order}.{" "}
                              {catalogItemNames.get(component.itemId) ??
                                `Item #${component.itemId}`}
                            </div>
                            <div className="mt-1 text-xs text-slate-500">
                              {component.quantityAtomic} {component.enteredUnitCode}
                              {component.enteredPackagingName
                                ? ` via ${component.enteredPackagingName}`
                                : " via unidade base"}
                              {" · "}conversao {component.conversionNumeratorAtomic}/
                              {component.conversionDenominator}
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </section>
    </main>
  );
}

export default RecipesPage;
