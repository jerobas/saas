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

const kilogram = {
  code: "kg",
  name: "kilogram",
  symbol: "kg",
  dimension: "MASS",
  numeratorAtomic: 1_000_000,
  denominator: 1,
  isItemBase: false,
  isSeeded: true,
};

const chocolate = {
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
};

const chocolateDetail = {
  ...chocolate,
  baseUnit: gram,
  packagings: [],
};

describe("ProductsPage", () => {
  beforeEach(() => {
    gatewayMocks.referenceDataGateway.listMeasurementUnits.mockResolvedValue([gram, kilogram]);
    gatewayMocks.catalogGateway.listItems.mockResolvedValue({
      items: [chocolate],
      next: null,
    });
    gatewayMocks.catalogGateway.getItem.mockResolvedValue(chocolateDetail);
    gatewayMocks.catalogGateway.createItemPackaging.mockResolvedValue({
      id: 3,
      itemId: 1,
      name: "Saco 1,5 kg",
      enteredUnitCode: "kg",
      conversionNumeratorAtomic: 1_500_000,
      conversionDenominator: 1,
      baseUnit: gram,
      enteredUnit: kilogram,
      createdAtMs: 1_700_000_000_003,
      updatedAtMs: 1_700_000_000_003,
      archivedAtMs: null,
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

  it("creates packaging from readable content while preserving the exact conversion", async () => {
    const user = userEvent.setup();

    render(<ProductsPage />);

    await user.click(await screen.findByRole("button", { name: "Chocolate" }));
    await user.type(screen.getByLabelText("Nome da embalagem"), "Saco 1,5 kg");
    await user.clear(screen.getByLabelText("Conteúdo"));
    await user.type(screen.getByLabelText("Conteúdo"), "1,5");
    await user.selectOptions(screen.getByLabelText("Unidade"), "kg");

    expect(screen.getByText("1 embalagem =")).toHaveTextContent("1 embalagem = 1.500 g");
    expect(screen.queryByPlaceholderText("Numerador")).not.toBeInTheDocument();
    expect(screen.queryByPlaceholderText("Denominador")).not.toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Criar embalagem" }));

    expect(gatewayMocks.catalogGateway.createItemPackaging).toHaveBeenCalledWith({
      itemId: 1,
      name: "Saco 1,5 kg",
      enteredUnitCode: "kg",
      conversionNumeratorAtomic: 1_500_000,
      conversionDenominator: 1,
    });
  });
});
