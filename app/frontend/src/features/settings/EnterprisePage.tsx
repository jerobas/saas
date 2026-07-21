import {
  Archive,
  ArrowClockwise,
  Buildings,
  FloppyDisk,
  Handshake,
  Plus,
} from "@phosphor-icons/react";
import { useCallback, useEffect, useState } from "react";
import {
  counterpartyGateway,
  type CounterpartyResponse,
  type CounterpartyRole,
  settingsGateway,
  type SettingsResponse,
} from "../../gateways/desktopBridge";

interface SettingsFormState {
  businessName: string;
  locale: string;
  timezone: string;
  currencyCode: string;
  currencyMinorDigits: string;
  hourlyLaborCost: string;
  defaultGrossMargin: string;
}

interface CounterpartyFormState {
  name: string;
  phone: string;
  email: string;
  notes: string;
  supplier: boolean;
  customer: boolean;
}

const emptySettingsForm: SettingsFormState = {
  businessName: "",
  locale: "pt-BR",
  timezone: "America/Sao_Paulo",
  currencyCode: "BRL",
  currencyMinorDigits: "2",
  hourlyLaborCost: "",
  defaultGrossMargin: "",
};

const emptyCounterpartyForm: CounterpartyFormState = {
  name: "",
  phone: "",
  email: "",
  notes: "",
  supplier: true,
  customer: false,
};

const optionalText = (value: string) => {
  const trimmed = value.trim();
  return trimmed.length > 0 ? trimmed : undefined;
};

const parseMoneyMinor = (value: string) => {
  const normalized = value.trim().replace(",", ".");
  if (normalized.length === 0) return undefined;
  const parsed = Number.parseFloat(normalized);
  if (!Number.isFinite(parsed)) return undefined;
  return Math.round(parsed * 100);
};

const parseBasisPoints = (value: string) => {
  const normalized = value.trim().replace(",", ".");
  if (normalized.length === 0) return undefined;
  const parsed = Number.parseFloat(normalized);
  if (!Number.isFinite(parsed)) return undefined;
  return Math.round(parsed * 100);
};

const basisPointsToPercent = (value?: number | null) =>
  value === null || value === undefined ? "" : (value / 100).toFixed(2);

const minorToDecimal = (value?: number | null, digits = 2) =>
  value === null || value === undefined ? "" : (value / 10 ** digits).toFixed(digits);

const formatMoney = (value?: number | null, digits = 2, currency = "BRL") => {
  if (value === null || value === undefined) return "-";
  return new Intl.NumberFormat("pt-BR", {
    style: "currency",
    currency,
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value / 10 ** digits);
};

const formatDateTime = (ms?: number | null) => {
  if (ms === null || ms === undefined) return "-";
  return new Intl.DateTimeFormat("pt-BR", {
    dateStyle: "short",
    timeStyle: "short",
  }).format(new Date(ms));
};

const settingsToForm = (settings: SettingsResponse): SettingsFormState => ({
  businessName: settings.businessName,
  locale: settings.locale,
  timezone: settings.timezone,
  currencyCode: settings.currencyCode,
  currencyMinorDigits: String(settings.currencyMinorDigits),
  hourlyLaborCost: minorToDecimal(settings.hourlyLaborCost, settings.currencyMinorDigits),
  defaultGrossMargin: basisPointsToPercent(settings.defaultGrossMargin),
});

const rolesFromForm = (form: CounterpartyFormState): CounterpartyRole[] =>
  [form.supplier ? "SUPPLIER" : null, form.customer ? "CUSTOMER" : null].filter(
    (role): role is CounterpartyRole => role !== null,
  );

function EnterprisePage() {
  const [settings, setSettings] = useState<SettingsResponse | null>(null);
  const [settingsForm, setSettingsForm] = useState<SettingsFormState>(emptySettingsForm);
  const [counterparties, setCounterparties] = useState<CounterpartyResponse[]>([]);
  const [counterpartyForm, setCounterpartyForm] =
    useState<CounterpartyFormState>(emptyCounterpartyForm);
  const [editingCounterparty, setEditingCounterparty] = useState<CounterpartyResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState<{ type: "success" | "error"; text: string } | null>(null);

  const loadEnterprise = useCallback(async () => {
    setLoading(true);
    setMessage(null);
    try {
      const [loadedSettings, counterpartyPage] = await Promise.all([
        settingsGateway.getSettings(),
        counterpartyGateway.listCounterparties({ archiveFilter: "ALL", pageSize: 100 }),
      ]);
      setSettings(loadedSettings);
      setSettingsForm(settingsToForm(loadedSettings));
      setCounterparties(counterpartyPage.items);
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel carregar a empresa.",
      });
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadEnterprise();
  }, [loadEnterprise]);

  const saveSettings = async () => {
    if (!settings || saving) return;
    setSaving(true);
    setMessage(null);
    try {
      const currencyMinorDigits = Number.parseInt(settingsForm.currencyMinorDigits, 10);
      const updated = await settingsGateway.updateSettings({
        businessName: settingsForm.businessName.trim(),
        locale: settingsForm.locale.trim(),
        timezone: settingsForm.timezone.trim(),
        currencyCode: settingsForm.currencyCode.trim().toUpperCase(),
        currencyMinorDigits,
        hourlyLaborCost: parseMoneyMinor(settingsForm.hourlyLaborCost),
        defaultGrossMargin: parseBasisPoints(settingsForm.defaultGrossMargin),
        expectedUpdatedAtMs: settings.updatedAtMs,
      });
      setSettings(updated);
      setSettingsForm(settingsToForm(updated));
      setMessage({ type: "success", text: "Configuracoes salvas." });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel salvar configuracoes.",
      });
    } finally {
      setSaving(false);
    }
  };

  const resetCounterpartyForm = () => {
    setEditingCounterparty(null);
    setCounterpartyForm(emptyCounterpartyForm);
  };

  const startEditingCounterparty = (counterparty: CounterpartyResponse) => {
    setEditingCounterparty(counterparty);
    setCounterpartyForm({
      name: counterparty.name,
      phone: counterparty.phone ?? "",
      email: counterparty.email ?? "",
      notes: counterparty.notes ?? "",
      supplier: counterparty.roles.includes("SUPPLIER"),
      customer: counterparty.roles.includes("CUSTOMER"),
    });
  };

  const saveCounterparty = async () => {
    if (saving) return;
    const roles = rolesFromForm(counterpartyForm);
    if (roles.length === 0) {
      setMessage({ type: "error", text: "Selecione fornecedor, cliente ou ambos." });
      return;
    }

    setSaving(true);
    setMessage(null);
    try {
      const request = {
        name: counterpartyForm.name.trim(),
        phone: optionalText(counterpartyForm.phone),
        email: optionalText(counterpartyForm.email),
        notes: optionalText(counterpartyForm.notes),
        roles,
      };
      const saved = editingCounterparty
        ? await counterpartyGateway.updateCounterparty(editingCounterparty.id, {
            ...request,
            expectedUpdatedAtMs: editingCounterparty.updatedAtMs,
          })
        : await counterpartyGateway.createCounterparty(request);
      setCounterparties((current) => {
        const withoutSaved = current.filter((item) => item.id !== saved.id);
        return [...withoutSaved, saved].sort((left, right) => left.name.localeCompare(right.name));
      });
      resetCounterpartyForm();
      setMessage({
        type: "success",
        text: editingCounterparty ? "Contraparte atualizada." : "Contraparte criada.",
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel salvar a contraparte.",
      });
    } finally {
      setSaving(false);
    }
  };

  const toggleCounterpartyArchive = async (counterparty: CounterpartyResponse) => {
    setSaving(true);
    setMessage(null);
    try {
      const saved =
        counterparty.archivedAtMs === null || counterparty.archivedAtMs === undefined
          ? await counterpartyGateway.archiveCounterparty(counterparty.id, {
              expectedUpdatedAtMs: counterparty.updatedAtMs,
            })
          : await counterpartyGateway.restoreCounterparty(counterparty.id, {
              expectedUpdatedAtMs: counterparty.updatedAtMs,
            });
      setCounterparties((current) =>
        current.map((currentCounterparty) =>
          currentCounterparty.id === saved.id ? saved : currentCounterparty,
        ),
      );
      setMessage({
        type: "success",
        text: saved.archivedAtMs ? "Contraparte arquivada." : "Contraparte restaurada.",
      });
    } catch (error) {
      setMessage({
        type: "error",
        text: error instanceof Error ? error.message : "Nao foi possivel alterar arquivamento.",
      });
    } finally {
      setSaving(false);
    }
  };

  const activeCounterparties = counterparties.filter(
    (counterparty) => counterparty.archivedAtMs === null || counterparty.archivedAtMs === undefined,
  );

  return (
    <main className="min-h-screen bg-slate-50">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto max-w-7xl px-6 py-8">
          <div className="flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
            <div>
              <p className="text-sm font-semibold uppercase tracking-wide text-pink-600">
                Empresa V2
              </p>
              <h1 className="mt-2 text-3xl font-bold text-slate-950">
                Configuracoes e contraparte
              </h1>
              <p className="mt-2 max-w-2xl text-slate-600">
                Dados locais reais para moeda, custos, margens, fornecedores e clientes.
              </p>
            </div>
            <button
              type="button"
              onClick={loadEnterprise}
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
              <Buildings size={28} className="text-pink-600" />
              <div>
                <h2 className="text-lg font-bold text-slate-950">Configuracoes</h2>
                <p className="text-sm text-slate-600">
                  {settings
                    ? `Atualizado em ${formatDateTime(settings.updatedAtMs)}`
                    : "Carregando..."}
                </p>
              </div>
            </div>

            <div className="mt-5 space-y-4">
              <label className="block text-sm font-semibold text-slate-700">
                Nome do negocio
                <input
                  value={settingsForm.businessName}
                  onChange={(event) =>
                    setSettingsForm({ ...settingsForm, businessName: event.target.value })
                  }
                  className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                />
              </label>
              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Locale
                  <input
                    value={settingsForm.locale}
                    onChange={(event) =>
                      setSettingsForm({ ...settingsForm, locale: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Fuso
                  <input
                    value={settingsForm.timezone}
                    onChange={(event) =>
                      setSettingsForm({ ...settingsForm, timezone: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Moeda
                  <input
                    value={settingsForm.currencyCode}
                    onChange={(event) =>
                      setSettingsForm({ ...settingsForm, currencyCode: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 uppercase outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Casas
                  <input
                    value={settingsForm.currencyMinorDigits}
                    onChange={(event) =>
                      setSettingsForm({
                        ...settingsForm,
                        currencyMinorDigits: event.target.value,
                      })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <label className="block text-sm font-semibold text-slate-700">
                  Custo/hora
                  <input
                    value={settingsForm.hourlyLaborCost}
                    onChange={(event) =>
                      setSettingsForm({ ...settingsForm, hourlyLaborCost: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="50,00"
                  />
                </label>
                <label className="block text-sm font-semibold text-slate-700">
                  Margem padrao %
                  <input
                    value={settingsForm.defaultGrossMargin}
                    onChange={(event) =>
                      setSettingsForm({
                        ...settingsForm,
                        defaultGrossMargin: event.target.value,
                      })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="30"
                  />
                </label>
              </div>
              <button
                type="button"
                onClick={saveSettings}
                disabled={saving || loading || !settings}
                className="inline-flex w-full items-center justify-center gap-2 rounded-xl bg-pink-600 px-4 py-3 font-semibold text-white hover:bg-pink-700 disabled:bg-slate-300"
              >
                <FloppyDisk size={18} />
                Salvar configuracoes
              </button>
            </div>
          </div>
        </aside>

        <div className="space-y-6">
          <div className="grid gap-4 md:grid-cols-3">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Contrapartes</p>
              <p className="mt-2 text-3xl font-bold text-slate-950">{counterparties.length}</p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Ativas</p>
              <p className="mt-2 text-3xl font-bold text-green-700">
                {activeCounterparties.length}
              </p>
            </div>
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <p className="text-sm font-semibold text-slate-500">Custo/hora</p>
              <p className="mt-2 text-3xl font-bold text-blue-700">
                {formatMoney(
                  settings?.hourlyLaborCost,
                  settings?.currencyMinorDigits,
                  settings?.currencyCode,
                )}
              </p>
            </div>
          </div>

          <div className="grid gap-6 lg:grid-cols-[360px_1fr]">
            <div className="rounded-3xl border border-slate-200 bg-white p-6 shadow-sm">
              <div className="flex items-center gap-3">
                <Handshake size={28} className="text-pink-600" />
                <h2 className="text-lg font-bold text-slate-950">
                  {editingCounterparty ? "Editar contraparte" : "Nova contraparte"}
                </h2>
              </div>
              <div className="mt-5 space-y-4">
                <label className="block text-sm font-semibold text-slate-700">
                  Nome
                  <input
                    value={counterpartyForm.name}
                    onChange={(event) =>
                      setCounterpartyForm({ ...counterpartyForm, name: event.target.value })
                    }
                    className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    placeholder="Fornecedor ABC"
                  />
                </label>
                <div className="grid grid-cols-2 gap-3">
                  <label className="block text-sm font-semibold text-slate-700">
                    Telefone
                    <input
                      value={counterpartyForm.phone}
                      onChange={(event) =>
                        setCounterpartyForm({ ...counterpartyForm, phone: event.target.value })
                      }
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    />
                  </label>
                  <label className="block text-sm font-semibold text-slate-700">
                    Email
                    <input
                      value={counterpartyForm.email}
                      onChange={(event) =>
                        setCounterpartyForm({ ...counterpartyForm, email: event.target.value })
                      }
                      className="mt-2 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                    />
                  </label>
                </div>
                <label className="block text-sm font-semibold text-slate-700">
                  Observacoes
                  <textarea
                    value={counterpartyForm.notes}
                    onChange={(event) =>
                      setCounterpartyForm({ ...counterpartyForm, notes: event.target.value })
                    }
                    className="mt-2 min-h-20 w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:ring-2 focus:ring-pink-500"
                  />
                </label>
                <div>
                  <p className="text-sm font-semibold text-slate-700">Papeis</p>
                  <div className="mt-2 grid grid-cols-2 gap-2 text-sm">
                    <label className="flex items-center gap-2 rounded-xl bg-slate-100 px-3 py-2">
                      <input
                        type="checkbox"
                        checked={counterpartyForm.supplier}
                        onChange={(event) =>
                          setCounterpartyForm({
                            ...counterpartyForm,
                            supplier: event.target.checked,
                          })
                        }
                      />
                      Fornecedor
                    </label>
                    <label className="flex items-center gap-2 rounded-xl bg-slate-100 px-3 py-2">
                      <input
                        type="checkbox"
                        checked={counterpartyForm.customer}
                        onChange={(event) =>
                          setCounterpartyForm({
                            ...counterpartyForm,
                            customer: event.target.checked,
                          })
                        }
                      />
                      Cliente
                    </label>
                  </div>
                </div>
                <div className="flex gap-3">
                  <button
                    type="button"
                    onClick={saveCounterparty}
                    disabled={saving || loading || counterpartyForm.name.trim().length === 0}
                    className="inline-flex flex-1 items-center justify-center gap-2 rounded-xl bg-slate-900 px-4 py-3 font-semibold text-white hover:bg-slate-800 disabled:bg-slate-300"
                  >
                    <Plus size={18} />
                    {editingCounterparty ? "Salvar" : "Criar"}
                  </button>
                  {editingCounterparty && (
                    <button
                      type="button"
                      onClick={resetCounterpartyForm}
                      className="rounded-xl border border-slate-300 px-4 py-3 font-semibold text-slate-700"
                    >
                      Cancelar
                    </button>
                  )}
                </div>
              </div>
            </div>

            <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white shadow-sm">
              <div className="border-b border-slate-200 px-6 py-4">
                <h2 className="text-lg font-bold text-slate-950">Fornecedores e clientes</h2>
              </div>
              {loading ? (
                <p className="p-6 text-slate-600">Carregando contrapartes...</p>
              ) : counterparties.length === 0 ? (
                <p className="p-6 text-slate-600">Nenhuma contraparte cadastrada ainda.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-sm">
                    <thead className="bg-slate-50 text-slate-600">
                      <tr>
                        <th className="px-6 py-3 font-semibold">Nome</th>
                        <th className="px-6 py-3 font-semibold">Papeis</th>
                        <th className="px-6 py-3 font-semibold">Contato</th>
                        <th className="px-6 py-3 font-semibold">Status</th>
                        <th className="px-6 py-3 font-semibold">Acoes</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {counterparties.map((counterparty) => (
                        <tr key={counterparty.id} className="hover:bg-slate-50">
                          <td className="px-6 py-4">
                            <p className="font-semibold text-slate-950">{counterparty.name}</p>
                            <p className="text-xs text-slate-500">
                              Atualizado em {formatDateTime(counterparty.updatedAtMs)}
                            </p>
                          </td>
                          <td className="px-6 py-4 text-slate-700">
                            {counterparty.roles
                              .map((role) => (role === "SUPPLIER" ? "Fornecedor" : "Cliente"))
                              .join(" / ")}
                          </td>
                          <td className="px-6 py-4 text-slate-700">
                            <p>{counterparty.phone ?? "-"}</p>
                            <p className="text-xs text-slate-500">{counterparty.email ?? ""}</p>
                          </td>
                          <td className="px-6 py-4">
                            <span
                              className={`rounded-full px-3 py-1 text-xs font-bold ${
                                counterparty.archivedAtMs
                                  ? "bg-slate-200 text-slate-700"
                                  : "bg-green-100 text-green-700"
                              }`}
                            >
                              {counterparty.archivedAtMs ? "Arquivada" : "Ativa"}
                            </span>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex flex-wrap gap-2">
                              <button
                                type="button"
                                onClick={() => startEditingCounterparty(counterparty)}
                                className="rounded-lg border border-slate-300 px-3 py-2 font-semibold text-slate-700"
                              >
                                Editar
                              </button>
                              <button
                                type="button"
                                onClick={() => void toggleCounterpartyArchive(counterparty)}
                                disabled={saving}
                                className="inline-flex items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 font-semibold text-slate-700 disabled:opacity-50"
                              >
                                {counterparty.archivedAtMs ? (
                                  <ArrowClockwise size={16} />
                                ) : (
                                  <Archive size={16} />
                                )}
                                {counterparty.archivedAtMs ? "Restaurar" : "Arquivar"}
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
          </div>
        </div>
      </section>
    </main>
  );
}

export default EnterprisePage;
