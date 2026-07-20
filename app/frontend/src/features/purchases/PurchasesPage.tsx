import { ArrowClockwise } from "@phosphor-icons/react";
import { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import {
  catalogGateway,
  counterpartyGateway,
  type CounterpartyResponse,
  type ItemSummaryResponse,
  type PurchaseDocumentResponse,
  purchaseGateway,
} from "../../gateways/desktopBridge";
import PurchaseDocumentForm from "./PurchaseDocumentForm";

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

function PurchasesPage() {
  const [items, setItems] = useState<ItemSummaryResponse[]>([]);
  const [suppliers, setSuppliers] = useState<CounterpartyResponse[]>([]);
  const [purchases, setPurchases] = useState<PurchaseDocumentResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

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
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Não foi possível carregar compras.",
      });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadPage();
  }, [loadPage]);

  const handlePosted = async (purchase: PurchaseDocumentResponse) => {
    setMessage({ type: "success", text: `Compra ${documentLabel(purchase)} postada.` });
    try {
      await loadPurchases();
    } catch (error) {
      setMessage({
        type: "error",
        text:
          error instanceof Error
            ? error.message
            : "A compra foi postada, mas o histórico não pôde ser atualizado.",
      });
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
                Monte um documento com várias linhas e poste toda a entrada de estoque de uma vez.
              </p>
            </div>
            <button
              type="button"
              onClick={() => void loadPage()}
              disabled={loading}
              className="inline-flex items-center justify-center gap-2 rounded-xl border border-slate-300 px-4 py-3 font-semibold text-slate-700 hover:bg-slate-100 disabled:opacity-50"
            >
              <ArrowClockwise size={18} />
              Recarregar
            </button>
          </div>
        </div>
      </header>

      <div className="mx-auto max-w-7xl space-y-6 px-6 py-8">
        {message && (
          <div
            role="status"
            className={`rounded-2xl border p-4 text-sm font-medium ${
              message.type === "success"
                ? "border-green-200 bg-green-50 text-green-800"
                : "border-red-200 bg-red-50 text-red-800"
            }`}
          >
            {message.text}
          </div>
        )}

        <PurchaseDocumentForm
          items={items}
          suppliers={suppliers}
          disabled={loading}
          onError={(text) => setMessage({ type: "error", text })}
          onPosted={handlePosted}
        />

        <div className="grid gap-4 md:grid-cols-3">
          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <p className="text-sm font-semibold text-slate-500">Compras</p>
            <p className="mt-2 text-3xl font-bold text-slate-950">{purchases.length}</p>
          </div>
          <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
            <p className="text-sm font-semibold text-slate-500">Itens compráveis</p>
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
                    <th className="px-6 py-3 font-semibold">Razão</th>
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
                          <td className="px-6 py-4 text-slate-700">{purchase.reasonCode ?? "-"}</td>
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
                                    className="grid gap-3 rounded-xl bg-slate-50 p-3 md:grid-cols-3"
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
          Depois de postar, confira <strong>Estoque</strong> na visão <strong>Lotes</strong>: ela lê
          o saldo real criado pela compra.
        </div>
      </div>
    </main>
  );
}

export default PurchasesPage;
