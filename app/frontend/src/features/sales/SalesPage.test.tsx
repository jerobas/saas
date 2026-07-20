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

const candySummary = {
  ...cakeSummary,
  id: 11,
  name: "Brigadeiro",
  sku: "BRIG-001",
  defaultSalePrice: 300,
};

const candyDetail = {
  ...candySummary,
  baseUnit: {
    code: "each",
    name: "each",
    symbol: "un",
    dimension: "COUNT" as const,
    numeratorAtomic: 1,
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

const candyLot = {
  ...cakeLot,
  id: 31,
  itemId: 11,
  sourceLineId: 41,
  availableQuantityAtomic: 50,
  lotCode: "BRIG-1",
};

const candyBalance = {
  ...cakeBalance,
  itemId: 11,
  itemName: "Brigadeiro",
  baseUnitCode: "each",
  quantityAtomic: 50,
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
    {
      id: 93,
      lineOrder: 2,
      itemId: 11,
      direction: "OUT",
      quantityAtomic: 5,
      enteredUnitCode: "each",
      enteredPackagingName: null,
      conversionNumeratorAtomic: 1,
      conversionDenominator: 1,
      inventoryValueMicro: 150_000,
      commercialTotalMinor: 1500,
      allocations: [{ id: 94, lotId: 31, quantityAtomic: 5 }],
    },
  ],
};

describe("SalesPage", () => {
  beforeEach(() => {
    gatewayMocks.catalogGateway.listItems.mockResolvedValue({
      items: [cakeSummary, candySummary],
      next: null,
    });
    gatewayMocks.catalogGateway.getItem.mockImplementation((itemId: number) =>
      Promise.resolve(itemId === 10 ? cakeDetail : candyDetail),
    );
    gatewayMocks.counterpartyGateway.listCounterparties.mockResolvedValue({
      items: [customer],
      next: null,
    });
    gatewayMocks.inventoryGateway.getInventoryBalance.mockImplementation((itemId: number) =>
      Promise.resolve(itemId === 10 ? cakeBalance : candyBalance),
    );
    gatewayMocks.inventoryGateway.listEligibleFefoLots.mockImplementation((itemId: number) =>
      Promise.resolve(itemId === 10 ? [cakeLot] : [candyLot]),
    );
    gatewayMocks.saleGateway.listSales
      .mockResolvedValueOnce({ items: [], next: null })
      .mockResolvedValue({ items: [postedSale], next: null });
    gatewayMocks.saleGateway.postSale.mockResolvedValue(postedSale);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("posts a multi-line sale through the V2 gateway", async () => {
    const user = userEvent.setup();

    render(<SalesPage />);

    expect(
      await screen.findByText("O carrinho esta vazio. Adicione um item vendavel para comecar."),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Adicionar item" }));
    expect(await screen.findByRole("option", { name: /BOLO-1/ })).toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText("Item para adicionar"), "11");
    await user.click(screen.getByRole("button", { name: "Adicionar item" }));
    expect(await screen.findByRole("option", { name: /BRIG-1/ })).toBeInTheDocument();
    expect(screen.queryByText("Conversao numerador")).not.toBeInTheDocument();
    expect(screen.queryByText("Denominador")).not.toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText("Cliente"), "20");
    await user.type(screen.getByLabelText("Quantidade atomica da linha 1"), "20");
    await user.type(screen.getByLabelText("Total comercial da linha 1"), "25,00");
    await user.selectOptions(screen.getByLabelText("Lote de saida da linha 1"), "30");
    await user.type(screen.getByLabelText("Quantidade atomica da linha 2"), "5");
    await user.type(screen.getByLabelText("Total comercial da linha 2"), "15,00");
    await user.selectOptions(screen.getByLabelText("Lote de saida da linha 2"), "31");
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
          expect.objectContaining({
            itemId: 11,
            quantityAtomic: 5,
            enteredUnitCode: "each",
            conversionNumeratorAtomic: 1,
            conversionDenominator: 1,
            commercialTotalMinor: 1500,
            lotId: 31,
          }),
        ],
      }),
    );
    expect(await screen.findByText("#90 / seq 3")).toBeInTheDocument();
    expect(screen.getByText("2026-07-18")).toBeInTheDocument();
    expect(screen.getByText("Detalhe da venda: alocacoes e custo")).toBeInTheDocument();
    expect(screen.getAllByText("Lotes consumidos")).toHaveLength(2);
    expect(screen.getByText(/lote #30:/)).toBeInTheDocument();
    expect(screen.getByText(/lote #31:/)).toBeInTheDocument();
    expect(gatewayMocks.saleGateway.listSales).toHaveBeenCalledWith({ pageSize: 25 });
    expect(gatewayMocks.inventoryGateway.listEligibleFefoLots).toHaveBeenCalledTimes(2);
    expect(gatewayMocks.inventoryGateway.getInventoryBalance).toHaveBeenCalledTimes(2);
  });
});
