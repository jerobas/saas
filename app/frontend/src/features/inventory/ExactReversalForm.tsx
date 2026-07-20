import { ArrowCounterClockwise, Warning } from "@phosphor-icons/react";
import { useEffect, useRef, useState } from "react";
import { type ReversalDocumentResponse, reversalGateway } from "../../gateways/desktopBridge";

interface ExactReversalFormProps {
  onPosted?: (reversal: ReversalDocumentResponse) => void;
  prefillDocumentId?: number | null;
  prefillRequestKey?: number;
}

const todayISO = () => new Date().toISOString().slice(0, 10);

const optionalText = (value: string) => {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
};

const buildIdempotencyKey = () =>
  `reversal-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`;

function ExactReversalForm({
  onPosted,
  prefillDocumentId,
  prefillRequestKey,
}: ExactReversalFormProps) {
  const targetInputRef = useRef<HTMLInputElement>(null);
  const [targetDocumentId, setTargetDocumentId] = useState("");
  const [occurredOn, setOccurredOn] = useState(todayISO);
  const [notes, setNotes] = useState("");
  const [confirmed, setConfirmed] = useState(false);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);
  const [latestReversal, setLatestReversal] = useState<ReversalDocumentResponse | null>(null);

  useEffect(() => {
    if (!prefillDocumentId || prefillDocumentId <= 0) return;
    setTargetDocumentId(String(prefillDocumentId));
    setMessage(null);
    setLatestReversal(null);
    targetInputRef.current?.focus();
  }, [prefillDocumentId, prefillRequestKey]);

  const postReversal = async () => {
    if (saving) return;
    const normalizedDocumentId = targetDocumentId.trim();
    const documentId = Number(normalizedDocumentId);
    if (
      !/^\d+$/.test(normalizedDocumentId) ||
      !Number.isSafeInteger(documentId) ||
      documentId <= 0
    ) {
      setMessage({ type: "error", text: "Informe um ID de documento válido." });
      return;
    }
    if (!confirmed) {
      setMessage({ type: "error", text: "Confirme que esta é uma correção de lançamento." });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const posted = await reversalGateway.postReversal({
        idempotencyKey: buildIdempotencyKey(),
        targetDocumentId: documentId,
        occurredOn,
        notes: optionalText(notes),
      });
      setLatestReversal(posted);
      setMessage({
        type: "success",
        text: `Reversão #${posted.id} criada para o documento #${posted.targetDocumentId}.`,
      });
      setTargetDocumentId("");
      setNotes("");
      setConfirmed(false);
      onPosted?.(posted);
    } catch (error) {
      setMessage({
        type: "error",
        text:
          error instanceof Error
            ? error.message
            : "Não foi possível reverter o documento exatamente.",
      });
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="mb-8 rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div className="flex items-start gap-3">
          <div className="rounded-xl bg-amber-100 p-2 text-amber-700">
            <ArrowCounterClockwise size={24} />
          </div>
          <div>
            <h2 className="text-lg font-semibold text-slate-900">Reversão exata</h2>
            <p className="mt-1 max-w-2xl text-sm text-slate-600">
              Corrige integralmente um lançamento elegível criando um novo documento imutável. Não
              use este fluxo para devoluções físicas.
            </p>
          </div>
        </div>
        <div className="flex max-w-md items-start gap-2 rounded-xl bg-amber-50 px-3 py-2 text-xs text-amber-900">
          <Warning size={18} className="mt-0.5 shrink-0" />
          <span>
            O documento precisa ser o último lançamento dos itens afetados, não pode ter sido
            revertido e seus lotes precisam continuar reversíveis.
          </span>
        </div>
      </div>

      {message && (
        <div
          role={message.type === "error" ? "alert" : "status"}
          className={`mt-5 rounded-xl border px-4 py-3 text-sm font-medium ${
            message.type === "success"
              ? "border-green-200 bg-green-50 text-green-800"
              : "border-red-200 bg-red-50 text-red-800"
          }`}
        >
          {message.text}
        </div>
      )}

      <div className="mt-5 grid gap-4 lg:grid-cols-[180px_200px_1fr]">
        <label className="block text-sm font-semibold text-slate-700">
          ID do documento
          <input
            ref={targetInputRef}
            inputMode="numeric"
            value={targetDocumentId}
            onChange={(event) => setTargetDocumentId(event.target.value)}
            placeholder="Ex.: 42"
            className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
          />
        </label>
        <label className="block text-sm font-semibold text-slate-700">
          Data da reversão
          <input
            type="date"
            value={occurredOn}
            onChange={(event) => setOccurredOn(event.target.value)}
            className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
          />
        </label>
        <label className="block text-sm font-semibold text-slate-700">
          Observações
          <input
            value={notes}
            onChange={(event) => setNotes(event.target.value)}
            placeholder="Motivo da correção"
            className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
          />
        </label>
      </div>

      <div className="mt-5 flex flex-col gap-4 border-t border-slate-200 pt-5 lg:flex-row lg:items-center lg:justify-between">
        <label className="flex items-start gap-3 text-sm text-slate-700">
          <input
            type="checkbox"
            checked={confirmed}
            onChange={(event) => setConfirmed(event.target.checked)}
            className="mt-1"
          />
          <span>
            <strong className="block text-slate-900">Confirmo que é uma correção de dados.</strong>A
            reversão será um novo lançamento e não apagará o histórico original.
          </span>
        </label>
        <button
          type="button"
          onClick={() => void postReversal()}
          disabled={saving || !targetDocumentId || !confirmed}
          className="inline-flex min-w-52 items-center justify-center gap-2 rounded-xl bg-amber-600 px-5 py-3 font-semibold text-white transition hover:bg-amber-700 disabled:bg-slate-300"
        >
          <ArrowCounterClockwise size={18} />
          {saving ? "Revertendo..." : "Reverter documento"}
        </button>
      </div>

      {latestReversal && (
        <div className="mt-5 rounded-2xl bg-slate-50 p-4">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            Última reversão criada · sequência {latestReversal.postingSequence}
          </p>
          <div className="mt-3 grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {latestReversal.lines.map((line) => (
              <div
                key={line.id}
                className="rounded-xl border border-slate-200 bg-white p-3 text-sm"
              >
                <p className="font-semibold text-slate-900">
                  Item #{line.itemId} · {line.direction === "IN" ? "Entrada" : "Saída"}
                </p>
                <p className="mt-1 text-slate-600">
                  {line.quantityAtomic} unidades atômicas · unidade registrada{" "}
                  {line.enteredUnitCode}
                  {" · "}reverte linha #{line.reversesLineId}
                </p>
              </div>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}

export default ExactReversalForm;
