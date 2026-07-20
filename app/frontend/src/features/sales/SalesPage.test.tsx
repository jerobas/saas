import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import SalesPage from "./SalesPage";

const gatewayMocks = vi.hoisted(() => ({
  catalogGateway: {
    listItems: vi.fn(),
    getItem: vi.fn(),
  },
  counterpartyGateway: {
    listCounterparties: vi.fn(),
  },
  inventoryGateway: {
    getInventoryBalance: vi.fn(),
    listEligibleFefoLots: vi.fn(),
  },
  saleGateway: {
    listSales: vi.fn(),
    postSale: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

const cakeSummary = {
  id: 10,
  name: "Bolo",
  sku: "BOLO-001",
  description: null,
  baseUnitCode: "g",
  capabilities: { purchasable: false, producible: true, sellable: true },
  defaultSalePrice: 1500,
  reorderQuantityAtomic: null,
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
};

const cakeDetail = {
  ...cakeSummary,
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

const customer = {
  id: 20,
  name: "Cliente",
  phone: null,
  email: null,
  notes: null,
  roles: ["CUSTOMER"],
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
};

const cakeLot = {
  id: 30,
  itemId: 10,
  sourceLineId: 40,
  sourcePostingSequence: 2,
  initialQuantityAtomic: 100,
  consumedQuantityAtomic: 0,
  restoredQuantityAtomic: 0,
  availableQuantityAtomic: 100,
  lotCode: "BOLO-1",
  originatedOn: "2026-07-17",
  expiresOn: "2026-07-20",
  createdAtMs: 1_700_000_000_000,
  sourceDocumentId: 50,
  sourceKind: "PRODUCTION",
  sourceOccurredOn: "2026-07-17",
};

const cakeBalance = {
  itemId: 10,
  itemName: "Bolo",
  baseUnitCode: "g",
  itemArchivedAtMs: null,
  quantityAtomic: 100,
  inventoryValueMicro: 3_000_000,
  lastDocumentId: 50,
  updatedAtMs: 1_700_000_000_000,
  capabilities: { purchasable: false, producible: true, sellable: true },
  reorderQuantityAtomic: null,
};

const postedSale = {
  id: 90,
  idempotencyKey: "sale-test",
  postingSequence: 3,
  counterpartyId: 20,
  occurredOn: "2026-07-18",
  postedAtMs: 1_700_000_000_100,
  currencyCode: "BRL",
  currencyMinorDigits: 2,
  reasonCode: null,
  notes: null,
  lines: [
    {
      id: 91,
      lineOrder: 1,
      itemId: 10,
      direction: "OUT",
      quantityAtomic: 20,
      enteredUnitCode: "g",
      enteredPackagingName: null,
      conversionNumeratorAtomic: 1000,
      conversionDenominator: 1,
      inventoryValueMicro: 600_000,
      commercialTotalMinor: 2500,
      allocations: [{ id: 92, lotId: 30, quantityAtomic: 20 }],
    },
  ],
};

describe("SalesPage", () => {
  beforeEach(() => {
    gatewayMocks.catalogGateway.listItems.mockResolvedValue({ items: [cakeSummary], next: null });
    gatewayMocks.catalogGateway.getItem.mockResolvedValue(cakeDetail);
    gatewayMocks.counterpartyGateway.listCounterparties.mockResolvedValue({
      items: [customer],
      next: null,
    });
    gatewayMocks.inventoryGateway.getInventoryBalance.mockResolvedValue(cakeBalance);
    gatewayMocks.inventoryGateway.listEligibleFefoLots.mockResolvedValue([cakeLot]);
    gatewayMocks.saleGateway.listSales
      .mockResolvedValueOnce({ items: [], next: null })
      .mockResolvedValue({ items: [postedSale], next: null });
    gatewayMocks.saleGateway.postSale.mockResolvedValue(postedSale);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("posts a sale through the V2 gateway and refreshes inventory context", async () => {
    const user = userEvent.setup();

    render(<SalesPage />);

    expect(await screen.findByText("BOLO-1")).toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText("Cliente"), "20");
    await user.type(screen.getByLabelText("Quantidade atomica"), "20");
    await user.type(screen.getByLabelText("Total comercial"), "25,00");
    await user.selectOptions(screen.getByLabelText("Lote de saida de estoque"), "30");
    await user.click(screen.getByRole("button", { name: "Postar venda" }));

    expect(gatewayMocks.saleGateway.postSale).toHaveBeenCalledWith(
      expect.objectContaining({
        counterpartyId: 20,
        reasonCode: undefined,
        lines: [
          expect.objectContaining({
            itemId: 10,
            quantityAtomic: 20,
            enteredUnitCode: "g",
            conversionNumeratorAtomic: 1000,
            conversionDenominator: 1,
            commercialTotalMinor: 2500,
            lotId: 30,
          }),
        ],
      }),
    );
    expect(await screen.findByText("#90 / seq 3")).toBeInTheDocument();
    expect(screen.getByText("2026-07-18")).toBeInTheDocument();
    expect(screen.getByText("Detalhe da venda: alocacoes e custo")).toBeInTheDocument();
    expect(screen.getByText("Lotes consumidos")).toBeInTheDocument();
    expect(screen.getByText(/lote #30:/)).toBeInTheDocument();
    expect(gatewayMocks.saleGateway.listSales).toHaveBeenCalledWith({ pageSize: 25 });
    expect(gatewayMocks.inventoryGateway.listEligibleFefoLots).toHaveBeenCalledTimes(2);
    expect(gatewayMocks.inventoryGateway.getInventoryBalance).toHaveBeenCalledTimes(2);
  });
});
