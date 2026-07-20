import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import PurchasesPage from "./PurchasesPage";

const gatewayMocks = vi.hoisted(() => ({
  catalogGateway: {
    listItems: vi.fn(),
    getItem: vi.fn(),
  },
  counterpartyGateway: {
    listCounterparties: vi.fn(),
  },
  purchaseGateway: {
    listPurchases: vi.fn(),
    postPurchase: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

const flourSummary = {
  id: 10,
  name: "Farinha",
  sku: "FAR-001",
  description: null,
  baseUnitCode: "g",
  capabilities: { purchasable: true, producible: false, sellable: false },
  defaultSalePrice: null,
  reorderQuantityAtomic: null,
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
};

const flourDetail = {
  ...flourSummary,
  baseUnit: {
    code: "g",
    name: "gram",
    symbol: "g",
    dimension: "MASS" as const,
    numeratorAtomic: 1000,
    denominator: 1,
    isItemBase: true,
    isSeeded: true,
  },
  packagings: [],
};

const supplier = {
  id: 20,
  name: "Fornecedor",
  phone: null,
  email: null,
  notes: null,
  roles: ["SUPPLIER"],
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
};

const postedPurchase = {
  id: 30,
  idempotencyKey: "purchase-test",
  postingSequence: 1,
  counterpartyId: 20,
  occurredOn: "2026-07-15",
  postedAtMs: 1_700_000_000_001,
  currencyCode: "BRL",
  currencyMinorDigits: 2,
  lines: [
    {
      id: 40,
      lineOrder: 1,
      itemId: 10,
      quantityAtomic: 1000,
      enteredUnitCode: "g",
      conversionNumeratorAtomic: 1000,
      conversionDenominator: 1,
      inventoryValueMicro: 50_000_000,
      commercialTotalMinor: 5000,
      lotId: 50,
      lotCode: "LOTE-1",
      originatedOn: "2026-07-15",
      expiresOn: null,
    },
  ],
};

describe("PurchasesPage", () => {
  beforeEach(() => {
    gatewayMocks.catalogGateway.listItems.mockResolvedValue({ items: [flourSummary], next: null });
    gatewayMocks.catalogGateway.getItem.mockResolvedValue(flourDetail);
    gatewayMocks.counterpartyGateway.listCounterparties.mockResolvedValue({
      items: [supplier],
      next: null,
    });
    gatewayMocks.purchaseGateway.listPurchases
      .mockResolvedValueOnce({ items: [], next: null })
      .mockResolvedValue({ items: [postedPurchase], next: null });
    gatewayMocks.purchaseGateway.postPurchase.mockResolvedValue(postedPurchase);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("posts a purchase through the V2 gateway and refreshes the history", async () => {
    const user = userEvent.setup();

    render(<PurchasesPage />);

    expect(await screen.findByText("Farinha")).toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText("Fornecedor"), "20");
    await user.clear(screen.getByLabelText("Quantidade atomica"));
    await user.type(screen.getByLabelText("Quantidade atomica"), "1000");
    await user.type(screen.getByLabelText("Total comercial"), "50,00");
    await user.type(screen.getByLabelText("Lote"), "LOTE-1");
    await user.click(screen.getByRole("button", { name: "Postar compra" }));

    expect(gatewayMocks.purchaseGateway.postPurchase).toHaveBeenCalledWith(
      expect.objectContaining({
        counterpartyId: 20,
        reasonCode: undefined,
        lines: [
          expect.objectContaining({
            itemId: 10,
            quantityAtomic: 1000,
            enteredUnitCode: "g",
            conversionNumeratorAtomic: 1000,
            conversionDenominator: 1,
            commercialTotalMinor: 5000,
            lotCode: "LOTE-1",
          }),
        ],
      }),
    );
    expect(await screen.findByText("#30 / seq 1")).toBeInTheDocument();
    expect(screen.getByText("Detalhe da compra: linhas, lotes e valores")).toBeInTheDocument();
    expect(screen.getByText("Lote criado")).toBeInTheDocument();
    expect(screen.getByText(/lote #50/)).toBeInTheDocument();
  });
});
