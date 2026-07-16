import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import ProductionPage from "./ProductionPage";

const gatewayMocks = vi.hoisted(() => ({
  catalogGateway: {
    getItem: vi.fn(),
  },
  inventoryGateway: {
    getInventoryBalance: vi.fn(),
    listEligibleFefoLots: vi.fn(),
  },
  productionGateway: {
    postProduction: vi.fn(),
  },
  recipeGateway: {
    listRecipes: vi.fn(),
    getRecipe: vi.fn(),
  },
}));

vi.mock("../gateways/desktopBridge", () => gatewayMocks);

const recipeSummary = {
  id: 1,
  name: "Bolo simples",
  outputItemId: 10,
  outputItemName: "Bolo",
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
  currentRevision: {
    id: 2,
    number: 1,
    standardYieldQuantityAtomic: 100,
  },
};

const recipeDetail = {
  id: 1,
  name: "Bolo simples",
  outputItemId: 10,
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
  currentRevision: {
    id: 2,
    recipeId: 1,
    number: 1,
    standardYieldQuantityAtomic: 100,
    instructions: "Misture.",
    preparationTimeMinutes: 30,
    estimatedDirectCostMicro: null,
    createdAtMs: 1_700_000_000_000,
    components: [
      {
        id: 3,
        revisionId: 2,
        order: 1,
        itemId: 20,
        quantityAtomic: 500,
        enteredUnitCode: "g",
        enteredPackagingName: null,
        conversionNumeratorAtomic: 1000,
        conversionDenominator: 1,
        createdAtMs: 1_700_000_000_000,
      },
    ],
  },
};

const outputItem = {
  id: 10,
  name: "Bolo",
  sku: null,
  description: null,
  baseUnitCode: "g",
  capabilities: { purchasable: false, producible: true, sellable: true },
  defaultSalePrice: 1500,
  reorderQuantityAtomic: null,
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
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

const componentItem = {
  ...outputItem,
  id: 20,
  name: "Farinha",
  capabilities: { purchasable: true, producible: false, sellable: false },
};

const componentLot = {
  id: 30,
  itemId: 20,
  sourceLineId: 40,
  sourcePostingSequence: 1,
  initialQuantityAtomic: 1000,
  consumedQuantityAtomic: 0,
  restoredQuantityAtomic: 0,
  availableQuantityAtomic: 1000,
  lotCode: "FAR-1",
  originatedOn: "2026-07-15",
  expiresOn: "2026-12-31",
  createdAtMs: 1_700_000_000_000,
  sourceDocumentId: 50,
  sourceKind: "PURCHASE",
  sourceOccurredOn: "2026-07-15",
};

const postedProduction = {
  id: 90,
  idempotencyKey: "production-test",
  postingSequence: 2,
  recipeRevisionId: 2,
  outputItemId: 10,
  occurredOn: "2026-07-16",
  postedAtMs: 1_700_000_000_100,
  currencyCode: "BRL",
  currencyMinorDigits: 2,
  directCostMicro: 500_000,
  notes: null,
  outputLine: {
    id: 91,
    lineOrder: 2,
    itemId: 10,
    direction: "IN",
    quantityAtomic: 100,
    enteredUnitCode: "g",
    conversionNumeratorAtomic: 1000,
    conversionDenominator: 1,
    inventoryValueMicro: 3_000_000,
    lotId: 92,
    lotCode: "BOLO-1",
    originatedOn: "2026-07-16",
    expiresOn: null,
    allocations: [],
  },
  inputLines: [
    {
      id: 93,
      lineOrder: 1,
      itemId: 20,
      direction: "OUT",
      quantityAtomic: 500,
      enteredUnitCode: "g",
      conversionNumeratorAtomic: 1000,
      conversionDenominator: 1,
      inventoryValueMicro: 2_500_000,
      lotId: null,
      lotCode: null,
      originatedOn: null,
      expiresOn: null,
      allocations: [{ id: 94, lotId: 30, quantityAtomic: 500 }],
    },
  ],
};

describe("ProductionPage", () => {
  beforeEach(() => {
    gatewayMocks.recipeGateway.listRecipes.mockResolvedValue({
      items: [recipeSummary],
      next: null,
    });
    gatewayMocks.recipeGateway.getRecipe.mockResolvedValue(recipeDetail);
    gatewayMocks.catalogGateway.getItem.mockImplementation((id: number) =>
      Promise.resolve(id === 10 ? outputItem : componentItem),
    );
    gatewayMocks.inventoryGateway.getInventoryBalance.mockImplementation((itemId: number) =>
      Promise.resolve({
        itemId,
        itemName: itemId === 10 ? "Bolo" : "Farinha",
        baseUnitCode: "g",
        itemArchivedAtMs: null,
        quantityAtomic: itemId === 10 ? 0 : 1000,
        inventoryValueMicro: itemId === 10 ? 0 : 5_000_000,
        lastDocumentId: null,
        updatedAtMs: 1_700_000_000_000,
        capabilities: itemId === 10 ? outputItem.capabilities : componentItem.capabilities,
        reorderQuantityAtomic: null,
      }),
    );
    gatewayMocks.inventoryGateway.listEligibleFefoLots.mockResolvedValue([componentLot]);
    gatewayMocks.productionGateway.postProduction.mockResolvedValue(postedProduction);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("posts production through the V2 gateway", async () => {
    const user = userEvent.setup();

    render(<ProductionPage />);

    expect(await screen.findByText("Bolo simples")).toBeInTheDocument();
    expect(screen.getByText("Preview de producao")).toBeInTheDocument();
    expect(screen.getByText("Custo estimado")).toBeInTheDocument();
    expect(screen.getByText("500 atomicos esperados")).toBeInTheDocument();
    expect(screen.getByText("FAR-1")).toBeInTheDocument();
    await user.clear(screen.getByLabelText("Custo direto"));
    await user.type(screen.getByLabelText("Custo direto"), "0,50");
    await user.type(screen.getByLabelText("Lote saida"), "BOLO-1");
    await user.selectOptions(screen.getByLabelText("Lote de entrada"), "30");
    await user.click(screen.getByRole("button", { name: "Postar producao" }));

    expect(gatewayMocks.productionGateway.postProduction).toHaveBeenCalledWith(
      expect.objectContaining({
        recipeRevisionId: 2,
        directCostMicro: 500_000,
        output: expect.objectContaining({
          quantityAtomic: 100,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1000,
          conversionDenominator: 1,
          lotCode: "BOLO-1",
        }),
        inputs: [
          expect.objectContaining({
            itemId: 20,
            quantityAtomic: 500,
            enteredUnitCode: "g",
            conversionNumeratorAtomic: 1000,
            conversionDenominator: 1,
            lotId: 30,
          }),
        ],
      }),
    );
    expect(await screen.findByText("Producao #90 · seq 2")).toBeInTheDocument();
  });
});
