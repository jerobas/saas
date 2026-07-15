import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import ProductsPage from "./ProductsPage";

const gatewayMocks = vi.hoisted(() => ({
  catalogGateway: {
    listItems: vi.fn(),
    getItem: vi.fn(),
    createItem: vi.fn(),
    updateItem: vi.fn(),
    archiveItem: vi.fn(),
    restoreItem: vi.fn(),
    createItemPackaging: vi.fn(),
    archiveItemPackaging: vi.fn(),
    restoreItemPackaging: vi.fn(),
  },
  referenceDataGateway: {
    listMeasurementUnits: vi.fn(),
  },
}));

vi.mock("../gateways/desktopBridge", () => gatewayMocks);

const gram = {
  code: "g",
  name: "gram",
  symbol: "g",
  dimension: "MASS",
  numeratorAtomic: 1000,
  denominator: 1,
  isItemBase: true,
  isSeeded: true,
};

describe("ProductsPage", () => {
  beforeEach(() => {
    gatewayMocks.referenceDataGateway.listMeasurementUnits.mockResolvedValue([gram]);
    gatewayMocks.catalogGateway.listItems.mockResolvedValue({
      items: [
        {
          id: 1,
          name: "Chocolate",
          sku: "CHOCO",
          description: null,
          baseUnitCode: "g",
          capabilities: { purchasable: true, producible: false, sellable: true },
          defaultSalePrice: 1250,
          reorderQuantityAtomic: null,
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

  it("renders catalog items from the V2 desktop gateway", async () => {
    render(<ProductsPage />);

    expect(await screen.findByText("Chocolate")).toBeInTheDocument();
    expect(screen.getByText("Compra / Venda")).toBeInTheDocument();
    expect(gatewayMocks.catalogGateway.listItems).toHaveBeenCalledWith({
      archiveFilter: "ALL",
      requireCapabilities: { purchasable: false, producible: false, sellable: false },
      pageSize: 100,
    });
  });
});
