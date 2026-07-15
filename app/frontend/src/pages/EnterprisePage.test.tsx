import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import EnterprisePage from "./EnterprisePage";

const gatewayMocks = vi.hoisted(() => ({
  settingsGateway: {
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
  },
  counterpartyGateway: {
    listCounterparties: vi.fn(),
    createCounterparty: vi.fn(),
    updateCounterparty: vi.fn(),
    archiveCounterparty: vi.fn(),
    restoreCounterparty: vi.fn(),
  },
}));

vi.mock("../gateways/desktopBridge", () => gatewayMocks);

describe("EnterprisePage", () => {
  beforeEach(() => {
    gatewayMocks.settingsGateway.getSettings.mockResolvedValue({
      businessName: "Sweet Shop",
      locale: "pt-BR",
      timezone: "America/Sao_Paulo",
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      hourlyLaborCost: 5000,
      defaultGrossMargin: 3000,
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_001,
    });
    gatewayMocks.counterpartyGateway.listCounterparties.mockResolvedValue({
      items: [
        {
          id: 10,
          name: "Supplier Co",
          phone: "555-0100",
          email: "supplier@example.com",
          notes: null,
          roles: ["SUPPLIER"],
          createdAtMs: 1_700_000_000_000,
          updatedAtMs: 1_700_000_000_001,
          archivedAtMs: null,
        },
      ],
      next: null,
    });
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("renders settings and counterparties from the V2 desktop gateway", async () => {
    render(<EnterprisePage />);

    expect(await screen.findByDisplayValue("Sweet Shop")).toBeInTheDocument();
    expect(screen.getByText("Supplier Co")).toBeInTheDocument();
    expect(screen.getAllByText("Fornecedor").length).toBeGreaterThan(0);
    expect(gatewayMocks.settingsGateway.getSettings).toHaveBeenCalledOnce();
    expect(gatewayMocks.counterpartyGateway.listCounterparties).toHaveBeenCalledWith({
      archiveFilter: "ALL",
      pageSize: 100,
    });
  });
});
