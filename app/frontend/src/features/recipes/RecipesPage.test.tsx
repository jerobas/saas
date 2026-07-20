import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import RecipesPage from "./RecipesPage";

const gatewayMocks = vi.hoisted(() => ({
  catalogGateway: {
    getItem: vi.fn(),
    listItems: vi.fn(),
  },
  recipeGateway: {
    getRecipe: vi.fn(),
    listRecipeRevisions: vi.fn(),
    listRecipes: vi.fn(),
    createRecipe: vi.fn(),
    publishRecipeRevision: vi.fn(),
    renameRecipe: vi.fn(),
    archiveRecipe: vi.fn(),
    restoreRecipe: vi.fn(),
  },
}));

vi.mock("../../gateways/desktopBridge", () => gatewayMocks);

const outputItem = {
  id: 10,
  name: "Bolo",
  sku: null,
  description: null,
  baseUnitCode: "g",
  capabilities: { purchasable: false, producible: true, sellable: true },
  defaultSalePrice: null,
  reorderQuantityAtomic: null,
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
};

const componentItem = {
  id: 11,
  name: "Farinha",
  sku: null,
  description: null,
  baseUnitCode: "g",
  capabilities: { purchasable: true, producible: false, sellable: false },
  defaultSalePrice: null,
  reorderQuantityAtomic: null,
  createdAtMs: 1_700_000_000_000,
  updatedAtMs: 1_700_000_000_000,
  archivedAtMs: null,
};

const secondComponentItem = {
  ...componentItem,
  id: 12,
  name: "Acucar",
};

const gramUnit = {
  code: "g",
  name: "gram",
  symbol: "g",
  dimension: "MASS",
  numeratorAtomic: 1000,
  denominator: 1,
  isItemBase: true,
  isSeeded: true,
};

const kilogramUnit = {
  ...gramUnit,
  code: "kg",
  name: "kilogram",
  symbol: "kg",
  numeratorAtomic: 1_000_000,
  isItemBase: false,
};

const componentItemDetail = {
  ...componentItem,
  baseUnit: gramUnit,
  packagings: [],
};

const secondComponentItemDetail = {
  ...secondComponentItem,
  baseUnit: gramUnit,
  packagings: [
    {
      id: 102,
      itemId: 12,
      name: "Pacote 1 kg",
      enteredUnitCode: "kg",
      conversionNumeratorAtomic: 1_000_000,
      conversionDenominator: 1,
      baseUnit: gramUnit,
      enteredUnit: kilogramUnit,
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_000,
      archivedAtMs: null,
    },
  ],
};

const revision = {
  id: 30,
  recipeId: 20,
  number: 1,
  standardYieldQuantityAtomic: 1000,
  instructions: "Misture e asse.",
  preparationTimeMinutes: 45,
  createdAtMs: 1_700_000_000_100,
  components: [
    {
      id: 31,
      revisionId: 30,
      order: 1,
      itemId: 11,
      quantityAtomic: 500,
      enteredUnitCode: "g",
      conversionNumeratorAtomic: 1000,
      conversionDenominator: 1,
      createdAtMs: 1_700_000_000_100,
    },
    {
      id: 32,
      revisionId: 30,
      order: 2,
      itemId: 12,
      quantityAtomic: 200,
      enteredUnitCode: "kg",
      enteredPackagingName: "Pacote 1 kg",
      conversionNumeratorAtomic: 1_000_000,
      conversionDenominator: 1,
      createdAtMs: 1_700_000_000_100,
    },
  ],
};

const createdRecipe = {
  id: 20,
  name: "Receita de bolo",
  outputItemId: 10,
  createdAtMs: 1_700_000_000_100,
  updatedAtMs: 1_700_000_000_100,
  archivedAtMs: null,
  currentRevision: revision,
};

const recipeSummary = {
  id: 20,
  name: "Receita de bolo",
  outputItemId: 10,
  outputItemName: "Bolo",
  createdAtMs: 1_700_000_000_100,
  updatedAtMs: 1_700_000_000_100,
  archivedAtMs: null,
  currentRevision: {
    id: 30,
    number: 1,
    standardYieldQuantityAtomic: 1000,
  },
};

describe("RecipesPage", () => {
  beforeEach(() => {
    gatewayMocks.catalogGateway.getItem.mockImplementation(async (itemId: number) =>
      itemId === 12 ? secondComponentItemDetail : componentItemDetail,
    );
    gatewayMocks.catalogGateway.listItems
      .mockResolvedValueOnce({ items: [outputItem], next: null })
      .mockResolvedValueOnce({
        items: [outputItem, componentItem, secondComponentItem],
        next: null,
      })
      .mockResolvedValue({
        items: [outputItem, componentItem, secondComponentItem],
        next: null,
      });
    gatewayMocks.recipeGateway.listRecipes
      .mockResolvedValueOnce({ items: [], next: null })
      .mockResolvedValue({ items: [recipeSummary], next: null });
    gatewayMocks.recipeGateway.createRecipe.mockResolvedValue(createdRecipe);
    gatewayMocks.recipeGateway.getRecipe.mockResolvedValue(createdRecipe);
    gatewayMocks.recipeGateway.listRecipeRevisions.mockResolvedValue([revision]);
    gatewayMocks.recipeGateway.publishRecipeRevision.mockResolvedValue({
      ...revision,
      id: 32,
      number: 2,
    });
    gatewayMocks.recipeGateway.renameRecipe.mockResolvedValue({
      ...createdRecipe,
      name: "Receita de bolo festa",
      updatedAtMs: 1_700_000_000_200,
    });
    gatewayMocks.recipeGateway.archiveRecipe.mockResolvedValue({
      ...createdRecipe,
      archivedAtMs: 1_700_000_000_200,
    });
    gatewayMocks.recipeGateway.restoreRecipe.mockResolvedValue(createdRecipe);
  });

  afterEach(() => {
    cleanup();
    vi.clearAllMocks();
  });

  it("creates a multi-component recipe with a packaging source", async () => {
    const user = userEvent.setup();

    render(<RecipesPage />);

    await screen.findByText(
      "Nenhuma receita cadastrada. Crie um item producivel no Catalogo e use o formulario.",
    );
    await user.type(screen.getByLabelText("Nome"), "Receita de bolo");
    await user.clear(screen.getByLabelText("Rendimento atomico"));
    await user.type(screen.getByLabelText("Rendimento atomico"), "1000");
    await user.clear(screen.getByLabelText("Preparo min."));
    await user.type(screen.getByLabelText("Preparo min."), "45");
    await user.selectOptions(screen.getByLabelText("Item do componente 1"), "11");
    await user.type(screen.getByLabelText("Quantidade atomica 1"), "500");
    await user.click(screen.getByRole("button", { name: "Adicionar componente" }));
    await user.selectOptions(screen.getByLabelText("Item do componente 2"), "12");
    await user.type(screen.getByLabelText("Quantidade atomica 2"), "200");
    await user.selectOptions(
      screen.getByLabelText("Unidade ou embalagem do componente 2"),
      await screen.findByRole("option", { name: "Pacote 1 kg" }),
    );
    expect(await screen.findByText("1.000 g")).toBeInTheDocument();
    await user.type(screen.getByLabelText("Instrucoes"), "Misture e asse.");
    await user.click(screen.getByRole("button", { name: "Criar" }));

    expect(gatewayMocks.recipeGateway.createRecipe).toHaveBeenCalledWith({
      name: "Receita de bolo",
      outputItemId: 10,
      revision: {
        standardYieldQuantityAtomic: 1000,
        instructions: "Misture e asse.",
        preparationTimeMinutes: 45,
        components: [
          {
            order: 1,
            itemId: 11,
            quantityAtomic: 500,
            sourceType: "UNIT",
            unitCode: "g",
          },
          {
            order: 2,
            itemId: 12,
            quantityAtomic: 200,
            sourceType: "PACKAGING",
            packagingId: 102,
          },
        ],
      },
    });
    expect(await screen.findByText('Receita "Receita de bolo" criada.')).toBeInTheDocument();
    expect(await screen.findByText("Revisao 1")).toBeInTheDocument();
    expect(await screen.findByText("1. Farinha")).toBeInTheDocument();
    expect(await screen.findByText("2. Acucar")).toBeInTheDocument();
    expect(screen.getByText("Atual")).toBeInTheDocument();
    expect(screen.getByText(/Revisao publicada e imutavel/)).toBeInTheDocument();
  });

  it("renames the selected recipe through the V2 gateway", async () => {
    const user = userEvent.setup();
    gatewayMocks.recipeGateway.listRecipes.mockReset();
    gatewayMocks.recipeGateway.listRecipes.mockResolvedValue({
      items: [recipeSummary],
      next: null,
    });

    render(<RecipesPage />);

    expect(await screen.findByText("Revisao 1")).toBeInTheDocument();
    const renameInput = screen.getByLabelText("Renomear receita");
    await user.clear(renameInput);
    await user.type(renameInput, "Receita de bolo festa");
    await user.click(screen.getByRole("button", { name: "Renomear" }));

    expect(gatewayMocks.recipeGateway.renameRecipe).toHaveBeenCalledWith(20, {
      name: "Receita de bolo festa",
      expectedUpdatedAtMs: 1_700_000_000_100,
    });
    expect(
      await screen.findByText('Receita renomeada para "Receita de bolo festa".'),
    ).toBeInTheDocument();
  });
});
