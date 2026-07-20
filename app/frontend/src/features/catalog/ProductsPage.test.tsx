import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
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

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

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
    expect(screen.queryByText("Unidades")).not.toBeInTheDocument();
    expect(gatewayMocks.catalogGateway.listItems).toHaveBeenCalledWith({
      archiveFilter: "ALL",
      requireCapabilities: { purchasable: false, producible: false, sellable: false },
      pageSize: 100,
    });
  });

  it("only allows default sale price when the item is sellable", async () => {
    const user = userEvent.setup();
    gatewayMocks.catalogGateway.createItem.mockResolvedValue({
      id: 2,
      name: "Flour",
      sku: null,
      description: null,
      baseUnitCode: "g",
      capabilities: { purchasable: true, producible: false, sellable: true },
      defaultSalePrice: 1250,
      reorderQuantityAtomic: null,
      createdAtMs: 1_700_000_000_002,
      updatedAtMs: 1_700_000_000_002,
      archivedAtMs: null,
      baseUnit: gram,
      packagings: [],
    });

    render(<ProductsPage />);

    const salePrice = await screen.findByPlaceholderText("12,50");
    expect(salePrice).toBeDisabled();
    expect(screen.getByText("Marque Venda para informar preco.")).toBeInTheDocument();

    await user.click(screen.getByLabelText("Venda"));
    expect(salePrice).toBeEnabled();

    await user.type(screen.getByPlaceholderText("Farinha de trigo"), "Flour");
    await user.type(salePrice, "12,50");
    await user.click(screen.getByRole("button", { name: "Criar" }));

    expect(gatewayMocks.catalogGateway.createItem).toHaveBeenCalledWith(
      expect.objectContaining({
        name: "Flour",
        baseUnitCode: "g",
        capabilities: { purchasable: true, producible: false, sellable: true },
        defaultSalePrice: 1250,
      }),
    );
  });
});
