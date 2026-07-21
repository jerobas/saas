import { ArrowCounterClockwise, ClockCounterClockwise } from "@phosphor-icons/react";
import { useCallback, useEffect, useMemo, useState } from "react";
import {
  type AdjustmentDocumentResponse,
  type AdjustmentReason,
  adjustmentGateway,
} from "../../gateways/desktopBridge";

interface InventoryItemLabel {
  itemId: number;
  itemName: string;
}

interface AdjustmentHistoryProps {
  items: InventoryItemLabel[];
  refreshKey: number;
  onReverse: (documentId: number) => void;
}

const reasonLabels: Record<AdjustmentReason, string> = {
  OPENING_BALANCE: "Saldo inicial",
  FREE_STOCK: "Entrada gratuita",
  WASTE: "Perda",
  EXPIRY: "Validade",
  DAMAGE: "Dano",
  SAMPLE: "Amostra",
  PHYSICAL_COUNT: "Contagem física",
  DOCUMENTED_CORRECTION: "Correção documentada",
};

const integerFormatter = new Intl.NumberFormat("pt-BR");

const formatBusinessDate = (value: string) => {
  const [year, month, day] = value.split("-");
  return year && month && day ? `${day}/${month}/${year}` : value;
};

function AdjustmentHistory({ items, refreshKey, onReverse }: AdjustmentHistoryProps) {
  const [documents, setDocuments] = useState<AdjustmentDocumentResponse[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const itemNames = useMemo(
    () => new Map(items.map((item) => [item.itemId, item.itemName])),
    [items],
  );

  const loadHistory = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const page = await adjustmentGateway.listAdjustments({ pageSize: 8 });
      setDocuments(page.items ?? []);
    } catch (loadError) {
      setError(
        loadError instanceof Error
          ? loadError.message
          : "Não foi possível carregar os ajustes recentes.",
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadHistory();
  }, [loadHistory, refreshKey]);

  return (
    <section className="mb-8 overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
      <div className="flex flex-col gap-3 border-b border-slate-200 px-6 py-5 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-3">
          <div className="rounded-xl bg-slate-100 p-2 text-slate-700">
            <ClockCounterClockwise size={22} />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-slate-900">Ajustes recentes</h2>
            <p className="mt-1 text-sm text-slate-600">
              Últimos documentos imutáveis, do mais novo para o mais antigo.
            </p>
          </div>
        </div>
        <button
          type="button"
          onClick={() => void loadHistory()}
          disabled={loading}
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-50 disabled:cursor-not-allowed disabled:text-slate-400"
        >
          {loading ? "Atualizando..." : "Atualizar histórico"}
        </button>
      </div>

      {loading && documents.length === 0 ? (
        <p className="px-6 py-8 text-center text-sm text-slate-500">Carregando ajustes...</p>
      ) : error ? (
        <div
          role="alert"
          className="m-5 rounded-xl border border-red-200 bg-red-50 p-4 text-sm text-red-800"
        >
          {error}
        </div>
      ) : documents.length === 0 ? (
        <p className="px-6 py-8 text-center text-sm text-slate-500">
          Nenhum ajuste foi postado ainda.
        </p>
      ) : (
        <div className="divide-y divide-slate-100">
          {documents.map((document) => (
            <article
              key={document.id}
              className="flex flex-col gap-4 px-6 py-5 lg:flex-row lg:items-center lg:justify-between"
            >
              <div className="min-w-0 flex-1">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-semibold text-slate-900">Documento #{document.id}</span>
                  <span className="rounded-full bg-pink-50 px-2.5 py-1 text-xs font-semibold text-pink-700">
                    {reasonLabels[document.reasonCode]}
                  </span>
                  <span className="text-xs text-slate-500">
                    {formatBusinessDate(document.occurredOn)} · sequência {document.postingSequence}
                  </span>
                </div>
                <div className="mt-2 flex flex-wrap gap-x-5 gap-y-1 text-sm text-slate-600">
                  {document.lines.slice(0, 3).map((line) => (
                    <span key={line.id}>
                      {itemNames.get(line.itemId) ?? `Item #${line.itemId}`} ·{" "}
                      {line.direction === "IN" ? "entrada" : "saída"} de{" "}
                      {integerFormatter.format(line.quantityAtomic)} unidades atômicas
                    </span>
                  ))}
                  {document.lines.length > 3 && <span>+{document.lines.length - 3} linhas</span>}
                </div>
                {document.notes && (
                  <p className="mt-2 truncate text-sm text-slate-500">{document.notes}</p>
                )}
              </div>
              <button
                type="button"
                onClick={() => onReverse(document.id)}
                aria-label={`Reverter ajuste #${document.id}`}
                className="inline-flex shrink-0 items-center justify-center gap-2 rounded-xl border border-amber-300 bg-amber-50 px-4 py-2 text-sm font-semibold text-amber-800 transition hover:bg-amber-100"
              >
                <ArrowCounterClockwise size={17} />
                Reverter
              </button>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

export default AdjustmentHistory;
