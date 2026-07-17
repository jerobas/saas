import { afterEach, describe, expect, it, vi } from "vitest";
import {
  adjustmentGateway,
  catalogGateway,
  counterpartyGateway,
  inventoryGateway,
  purchaseGateway,
  referenceDataGateway,
  recipeGateway,
  reportingGateway,
  reversalGateway,
  saleGateway,
  settingsGateway,
} from "./desktopBridge";

const originalBridge = window.go;

afterEach(() => {
  window.go = originalBridge;
});

describe("desktop bridge", () => {
  it("fails clearly when the desktop runtime is unavailable", async () => {
    window.go = undefined;

    await expect(settingsGateway.getSettings()).rejects.toThrow(
      "Desktop bridge method SettingsHandler.GetSettings is unavailable.",
    );
  });

  it("forwards settings calls to the V2 settings handler", async () => {
    const settings = {
      businessName: "Sweeters",
      locale: "pt-BR",
      timezone: "America/Sao_Paulo",
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_001,
    };
    const getSettings = vi.fn().mockResolvedValue(settings);
    const updateSettings = vi.fn().mockResolvedValue({ ...settings, businessName: "New Sweeters" });
    window.go = {
      service: {
        SettingsHandler: {
          GetSettings: getSettings,
          UpdateSettings: updateSettings,
        },
      },
    };

    await expect(settingsGateway.getSettings()).resolves.toEqual(settings);

    const request = {
      businessName: "New Sweeters",
      locale: "pt-BR",
      timezone: "America/Sao_Paulo",
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      expectedUpdatedAtMs: settings.updatedAtMs,
    };
    await expect(settingsGateway.updateSettings(request)).resolves.toEqual({
      ...settings,
      businessName: "New Sweeters",
    });

    expect(getSettings).toHaveBeenCalledOnce();
    expect(updateSettings).toHaveBeenCalledWith(request);
  });

  it("supports the Wails package namespace used by V2 handlers", async () => {
    const settings = {
      businessName: "Sweeters",
      locale: "pt-BR",
      timezone: "America/Sao_Paulo",
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_001,
    };
    const getSettings = vi.fn().mockResolvedValue(settings);
    window.go = {
      wails: {
        SettingsHandler: {
          GetSettings: getSettings,
        },
      },
    };

    await expect(settingsGateway.getSettings()).resolves.toEqual(settings);
    expect(getSettings).toHaveBeenCalledOnce();
  });

  it("forwards reference data calls to the V2 reference data handler", async () => {
    const gram = {
      code: "g",
      name: "Gram",
      symbol: "g",
      dimension: "MASS",
      numeratorAtomic: 1,
      denominator: 1,
      isItemBase: true,
      isSeeded: true,
    };
    const getMeasurementUnit = vi.fn().mockResolvedValue(gram);
    const listMeasurementUnits = vi.fn().mockResolvedValue([gram]);
    window.go = {
      service: {
        ReferenceDataHandler: {
          GetMeasurementUnit: getMeasurementUnit,
          ListMeasurementUnits: listMeasurementUnits,
        },
      },
    };

    await expect(referenceDataGateway.getMeasurementUnit("g")).resolves.toEqual(gram);
    await expect(referenceDataGateway.listMeasurementUnits()).resolves.toEqual([gram]);

    expect(getMeasurementUnit).toHaveBeenCalledWith("g");
    expect(listMeasurementUnits).toHaveBeenCalledOnce();
  });

  it("forwards catalog calls to the V2 catalog handler", async () => {
    const item = {
      id: 10,
      name: "Chocolate",
      baseUnitCode: "g",
      capabilities: { purchasable: true, producible: false, sellable: true },
      baseUnit: {
        code: "g",
        name: "Gram",
        symbol: "g",
        dimension: "MASS",
        numeratorAtomic: 1,
        denominator: 1,
        isItemBase: true,
        isSeeded: true,
      },
      packagings: [],
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_001,
    };
    const listItems = vi.fn().mockResolvedValue({ items: [item], next: null });
    const createItem = vi.fn().mockResolvedValue(item);
    const archiveItem = vi.fn().mockResolvedValue({ ...item, archivedAtMs: 1_700_000_000_002 });
    window.go = {
      service: {
        CatalogHandler: {
          ListItems: listItems,
          CreateItem: createItem,
          ArchiveItem: archiveItem,
        },
      },
    };

    const listRequest = {
      archiveFilter: "ACTIVE" as const,
      requireCapabilities: { purchasable: true, producible: false, sellable: false },
      pageSize: 50,
    };
    const createRequest = {
      name: "Chocolate",
      baseUnitCode: "g",
      capabilities: { purchasable: true, producible: false, sellable: true },
    };
    const versionedRequest = { expectedUpdatedAtMs: item.updatedAtMs };

    await expect(catalogGateway.listItems(listRequest)).resolves.toEqual({
      items: [item],
      next: null,
    });
    await expect(catalogGateway.createItem(createRequest)).resolves.toEqual(item);
    await expect(catalogGateway.archiveItem(item.id, versionedRequest)).resolves.toEqual({
      ...item,
      archivedAtMs: 1_700_000_000_002,
    });

    expect(listItems).toHaveBeenCalledWith(listRequest);
    expect(createItem).toHaveBeenCalledWith(createRequest);
    expect(archiveItem).toHaveBeenCalledWith(item.id, versionedRequest);
  });

  it("forwards packaging calls to the V2 catalog handler", async () => {
    const packaging = {
      id: 30,
      itemId: 10,
      name: "Bag",
      enteredUnitCode: "kg",
      conversionNumeratorAtomic: 1_000,
      conversionDenominator: 1,
      baseUnit: {
        code: "g",
        name: "Gram",
        symbol: "g",
        dimension: "MASS",
        numeratorAtomic: 1,
        denominator: 1,
        isItemBase: true,
        isSeeded: true,
      },
      enteredUnit: {
        code: "kg",
        name: "Kilogram",
        symbol: "kg",
        dimension: "MASS",
        numeratorAtomic: 1_000,
        denominator: 1,
        isItemBase: false,
        isSeeded: true,
      },
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_001,
    };
    const createItemPackaging = vi.fn().mockResolvedValue(packaging);
    const reconfigureArchivedItemPackaging = vi.fn().mockResolvedValue({
      ...packaging,
      conversionNumeratorAtomic: 2_000,
    });
    window.go = {
      main: {
        CatalogHandler: {
          CreateItemPackaging: createItemPackaging,
          ReconfigureArchivedItemPackaging: reconfigureArchivedItemPackaging,
        },
      },
    };

    const createRequest = {
      itemId: packaging.itemId,
      name: packaging.name,
      enteredUnitCode: packaging.enteredUnitCode,
      conversionNumeratorAtomic: packaging.conversionNumeratorAtomic,
      conversionDenominator: packaging.conversionDenominator,
    };
    const updateRequest = {
      name: packaging.name,
      enteredUnitCode: packaging.enteredUnitCode,
      conversionNumeratorAtomic: 2_000,
      conversionDenominator: packaging.conversionDenominator,
      expectedUpdatedAtMs: packaging.updatedAtMs,
    };

    await expect(catalogGateway.createItemPackaging(createRequest)).resolves.toEqual(packaging);
    await expect(
      catalogGateway.reconfigureArchivedItemPackaging(packaging.id, updateRequest),
    ).resolves.toEqual({
      ...packaging,
      conversionNumeratorAtomic: 2_000,
    });

    expect(createItemPackaging).toHaveBeenCalledWith(createRequest);
    expect(reconfigureArchivedItemPackaging).toHaveBeenCalledWith(packaging.id, updateRequest);
  });

  it("forwards counterparty calls to the V2 counterparty handler", async () => {
    const counterparty = {
      id: 20,
      name: "Supplier",
      roles: ["SUPPLIER"],
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_001,
    };
    const listCounterparties = vi.fn().mockResolvedValue({ items: [counterparty], next: null });
    const updateCounterparty = vi
      .fn()
      .mockResolvedValue({ ...counterparty, name: "Supplier Ltd." });
    window.go = {
      service: {
        CounterpartyHandler: {
          ListCounterparties: listCounterparties,
          UpdateCounterparty: updateCounterparty,
        },
      },
    };

    const listRequest = {
      archiveFilter: "ALL" as const,
      role: "SUPPLIER" as const,
      pageSize: 25,
    };
    const updateRequest = {
      name: "Supplier Ltd.",
      roles: ["SUPPLIER" as const],
      expectedUpdatedAtMs: counterparty.updatedAtMs,
    };

    await expect(counterpartyGateway.listCounterparties(listRequest)).resolves.toEqual({
      items: [counterparty],
      next: null,
    });
    await expect(
      counterpartyGateway.updateCounterparty(counterparty.id, updateRequest),
    ).resolves.toEqual({
      ...counterparty,
      name: "Supplier Ltd.",
    });

    expect(listCounterparties).toHaveBeenCalledWith(listRequest);
    expect(updateCounterparty).toHaveBeenCalledWith(counterparty.id, updateRequest);
  });

  it("forwards purchase read and posting calls to the V2 purchase handler", async () => {
    const response = {
      id: 40,
      idempotencyKey: "purchase-1",
      postingSequence: 1,
      counterpartyId: 20,
      occurredOn: "2026-07-15",
      postedAtMs: 1_700_000_000_000,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      lines: [
        {
          id: 50,
          lineOrder: 1,
          itemId: 10,
          quantityAtomic: 1_000,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          inventoryValueMicro: 5_000_000,
          commercialTotalMinor: 500,
          lotId: 60,
          originatedOn: "2026-07-15",
        },
      ],
    };
    const page = {
      items: [response],
      next: { postingSequence: 1, id: 40 },
    };
    const getPurchase = vi.fn().mockResolvedValue(response);
    const listPurchases = vi.fn().mockResolvedValue(page);
    const postPurchase = vi.fn().mockResolvedValue(response);
    window.go = {
      service: {
        PurchaseHandler: {
          GetPurchase: getPurchase,
          ListPurchases: listPurchases,
          PostPurchase: postPurchase,
        },
      },
    };

    const request = {
      idempotencyKey: "purchase-1",
      counterpartyId: 20,
      occurredOn: "2026-07-15",
      lines: [
        {
          itemId: 10,
          quantityAtomic: 1_000,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          commercialTotalMinor: 500,
        },
      ],
    };
    const listRequest = { pageSize: 25 };

    await expect(purchaseGateway.getPurchase(response.id)).resolves.toEqual(response);
    await expect(purchaseGateway.listPurchases(listRequest)).resolves.toEqual(page);
    await expect(purchaseGateway.postPurchase(request)).resolves.toEqual(response);
    expect(getPurchase).toHaveBeenCalledWith(response.id);
    expect(listPurchases).toHaveBeenCalledWith(listRequest);
    expect(postPurchase).toHaveBeenCalledWith(request);
  });

  it("forwards adjustment posting calls to the V2 adjustment handler", async () => {
    const response = {
      id: 41,
      idempotencyKey: "adjustment-1",
      postingSequence: 2,
      occurredOn: "2026-07-16",
      postedAtMs: 1_700_000_000_100,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      reasonCode: "WASTE" as const,
      lines: [
        {
          id: 51,
          lineOrder: 1,
          itemId: 10,
          direction: "OUT" as const,
          quantityAtomic: 250,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          inventoryValueMicro: 1_250_000,
          allocations: [{ id: 70, lotId: 60, quantityAtomic: 250 }],
        },
      ],
    };
    const postAdjustment = vi.fn().mockResolvedValue(response);
    window.go = {
      service: {
        AdjustmentHandler: {
          PostAdjustment: postAdjustment,
        },
      },
    };

    const request = {
      idempotencyKey: "adjustment-1",
      occurredOn: "2026-07-16",
      reasonCode: "WASTE" as const,
      lines: [
        {
          itemId: 10,
          direction: "OUT" as const,
          quantityAtomic: 250,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
        },
      ],
    };

    await expect(adjustmentGateway.postAdjustment(request)).resolves.toEqual(response);
    expect(postAdjustment).toHaveBeenCalledWith(request);
  });

  it("forwards reversal posting calls to the V2 reversal handler", async () => {
    const response = {
      id: 42,
      idempotencyKey: "reverse-adjustment-1",
      postingSequence: 3,
      targetDocumentId: 41,
      occurredOn: "2026-07-16",
      postedAtMs: 1_700_000_000_200,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      reasonCode: "EXACT_REVERSAL" as const,
      lines: [
        {
          id: 52,
          lineOrder: 1,
          itemId: 10,
          direction: "IN" as const,
          quantityAtomic: 250,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          inventoryValueMicro: 1_250_000,
          reversesLineId: 51,
          allocations: [
            {
              id: 71,
              lotId: 60,
              quantityAtomic: 250,
              restoresAllocationId: 70,
            },
          ],
        },
      ],
    };
    const postReversal = vi.fn().mockResolvedValue(response);
    window.go = {
      service: {
        ReversalHandler: {
          PostReversal: postReversal,
        },
      },
    };

    const request = {
      idempotencyKey: "reverse-adjustment-1",
      targetDocumentId: 41,
      occurredOn: "2026-07-16",
    };

    await expect(reversalGateway.postReversal(request)).resolves.toEqual(response);
    expect(postReversal).toHaveBeenCalledWith(request);
  });

  it("forwards sale posting calls to the V2 sale handler", async () => {
    const response = {
      id: 43,
      idempotencyKey: "sale-1",
      postingSequence: 4,
      counterpartyId: 20,
      occurredOn: "2026-07-18",
      postedAtMs: 1_700_000_000_300,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      lines: [
        {
          id: 53,
          lineOrder: 1,
          itemId: 10,
          direction: "OUT" as const,
          quantityAtomic: 20,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          inventoryValueMicro: 600_000,
          commercialTotalMinor: 1_000,
          allocations: [{ id: 72, lotId: 60, quantityAtomic: 20 }],
        },
      ],
    };
    const page = {
      items: [response],
      next: { postingSequence: 4, id: 43 },
    };
    const getSale = vi.fn().mockResolvedValue(response);
    const listSales = vi.fn().mockResolvedValue(page);
    const postSale = vi.fn().mockResolvedValue(response);
    window.go = {
      service: {
        SaleHandler: {
          GetSale: getSale,
          ListSales: listSales,
          PostSale: postSale,
        },
      },
    };

    const request = {
      idempotencyKey: "sale-1",
      counterpartyId: 20,
      occurredOn: "2026-07-18",
      lines: [
        {
          itemId: 10,
          quantityAtomic: 20,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          commercialTotalMinor: 1_000,
        },
      ],
    };
    const listRequest = { pageSize: 25 };

    await expect(saleGateway.getSale(response.id)).resolves.toEqual(response);
    await expect(saleGateway.listSales(listRequest)).resolves.toEqual(page);
    await expect(saleGateway.postSale(request)).resolves.toEqual(response);
    expect(getSale).toHaveBeenCalledWith(response.id);
    expect(listSales).toHaveBeenCalledWith(listRequest);
    expect(postSale).toHaveBeenCalledWith(request);
  });

  it("forwards recipe calls to the V2 recipe handler", async () => {
    const revision = {
      id: 81,
      recipeId: 80,
      number: 1,
      standardYieldQuantityAtomic: 1_000,
      instructions: "Mix and bake.",
      preparationTimeMinutes: 45,
      createdAtMs: 1_700_000_000_000,
      components: [
        {
          id: 82,
          revisionId: 81,
          order: 1,
          itemId: 10,
          quantityAtomic: 500,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          createdAtMs: 1_700_000_000_000,
        },
      ],
    };
    const recipe = {
      id: 80,
      name: "Cake recipe",
      outputItemId: 11,
      createdAtMs: 1_700_000_000_000,
      updatedAtMs: 1_700_000_000_000,
      currentRevision: revision,
    };
    const page = {
      items: [
        {
          id: recipe.id,
          name: recipe.name,
          outputItemId: recipe.outputItemId,
          outputItemName: "Cake",
          createdAtMs: recipe.createdAtMs,
          updatedAtMs: recipe.updatedAtMs,
          currentRevision: {
            id: revision.id,
            number: revision.number,
            standardYieldQuantityAtomic: revision.standardYieldQuantityAtomic,
          },
        },
      ],
      next: null,
    };
    const getRecipe = vi.fn().mockResolvedValue(recipe);
    const getRecipeRevision = vi.fn().mockResolvedValue(revision);
    const listRecipeRevisions = vi.fn().mockResolvedValue([revision]);
    const listRecipes = vi.fn().mockResolvedValue(page);
    const createRecipe = vi.fn().mockResolvedValue(recipe);
    const publishRecipeRevision = vi.fn().mockResolvedValue({ ...revision, id: 83, number: 2 });
    const renameRecipe = vi.fn().mockResolvedValue({ ...recipe, name: "Renamed cake recipe" });
    const archiveRecipe = vi.fn().mockResolvedValue({ ...recipe, archivedAtMs: 1_700_000_000_100 });
    const restoreRecipe = vi.fn().mockResolvedValue(recipe);
    window.go = {
      service: {
        RecipeHandler: {
          GetRecipe: getRecipe,
          GetRecipeRevision: getRecipeRevision,
          ListRecipeRevisions: listRecipeRevisions,
          ListRecipes: listRecipes,
          CreateRecipe: createRecipe,
          PublishRecipeRevision: publishRecipeRevision,
          RenameRecipe: renameRecipe,
          ArchiveRecipe: archiveRecipe,
          RestoreRecipe: restoreRecipe,
        },
      },
    };

    const writeRevision = {
      standardYieldQuantityAtomic: 1_000,
      instructions: "Mix and bake.",
      preparationTimeMinutes: 45,
      components: [
        {
          order: 1,
          itemId: 10,
          quantityAtomic: 500,
          sourceType: "UNIT" as const,
          unitCode: "g",
        },
      ],
    };
    const createRequest = {
      name: recipe.name,
      outputItemId: recipe.outputItemId,
      revision: writeRevision,
    };
    const publishRequest = {
      expectedLatestRevision: 1,
      expectedUpdatedAtMs: recipe.updatedAtMs,
      revision: { ...writeRevision, instructions: "Mix, rest, and bake." },
    };
    const versionedRequest = { expectedUpdatedAtMs: recipe.updatedAtMs };

    await expect(recipeGateway.getRecipe(recipe.id)).resolves.toEqual(recipe);
    await expect(recipeGateway.getRecipeRevision(revision.id)).resolves.toEqual(revision);
    await expect(recipeGateway.listRecipeRevisions(recipe.id)).resolves.toEqual([revision]);
    await expect(recipeGateway.listRecipes({ pageSize: 25 })).resolves.toEqual(page);
    await expect(recipeGateway.createRecipe(createRequest)).resolves.toEqual(recipe);
    await expect(recipeGateway.publishRecipeRevision(recipe.id, publishRequest)).resolves.toEqual({
      ...revision,
      id: 83,
      number: 2,
    });
    await expect(
      recipeGateway.renameRecipe(recipe.id, {
        name: "Renamed cake recipe",
        expectedUpdatedAtMs: recipe.updatedAtMs,
      }),
    ).resolves.toEqual({ ...recipe, name: "Renamed cake recipe" });
    await expect(recipeGateway.archiveRecipe(recipe.id, versionedRequest)).resolves.toEqual({
      ...recipe,
      archivedAtMs: 1_700_000_000_100,
    });
    await expect(recipeGateway.restoreRecipe(recipe.id, versionedRequest)).resolves.toEqual(recipe);

    expect(getRecipe).toHaveBeenCalledWith(recipe.id);
    expect(getRecipeRevision).toHaveBeenCalledWith(revision.id);
    expect(listRecipeRevisions).toHaveBeenCalledWith(recipe.id);
    expect(listRecipes).toHaveBeenCalledWith({ pageSize: 25 });
    expect(createRecipe).toHaveBeenCalledWith(createRequest);
    expect(publishRecipeRevision).toHaveBeenCalledWith(recipe.id, publishRequest);
    expect(renameRecipe).toHaveBeenCalledWith(recipe.id, {
      name: "Renamed cake recipe",
      expectedUpdatedAtMs: recipe.updatedAtMs,
    });
    expect(archiveRecipe).toHaveBeenCalledWith(recipe.id, versionedRequest);
    expect(restoreRecipe).toHaveBeenCalledWith(recipe.id, versionedRequest);
  });

  it("forwards inventory read calls to the V2 inventory handler", async () => {
    const balancePage = {
      items: [
        {
          itemId: 10,
          itemName: "Chocolate",
          baseUnitCode: "g",
          quantityAtomic: 1_000,
          inventoryValueMicro: 5_000_000,
          updatedAtMs: 1_700_000_000_000,
          capabilities: { purchasable: true, producible: false, sellable: true },
        },
      ],
      next: null,
    };
    const lots = [
      {
        id: 60,
        itemId: 10,
        sourceLineId: 50,
        sourcePostingSequence: 1,
        initialQuantityAtomic: 1_000,
        consumedQuantityAtomic: 0,
        restoredQuantityAtomic: 0,
        availableQuantityAtomic: 1_000,
        originatedOn: "2026-07-15",
        createdAtMs: 1_700_000_000_000,
        sourceDocumentId: 40,
        sourceKind: "PURCHASE",
        sourceOccurredOn: "2026-07-15",
      },
    ];
    const ledger = {
      items: [
        {
          lineId: 50,
          documentId: 40,
          postingSequence: 1,
          lineOrder: 1,
          documentKind: "PURCHASE",
          occurredOn: "2026-07-15",
          postedAtMs: 1_700_000_000_000,
          itemId: 10,
          direction: "IN",
          quantityAtomic: 1_000,
          inventoryValueMicro: 5_000_000,
          commercialTotalMinor: 500,
          currencyCode: "BRL",
          currencyMinorDigits: 2,
          enteredUnitCode: "g",
          conversionNumeratorAtomic: 1_000,
          conversionDenominator: 1,
          idempotencyKey: "purchase-1",
        },
      ],
      next: null,
    };
    const listInventoryBalances = vi.fn().mockResolvedValue(balancePage);
    const listItemLotFacts = vi.fn().mockResolvedValue(lots);
    const listEligibleFEFOLots = vi.fn().mockResolvedValue(lots);
    const listItemLedgerPage = vi.fn().mockResolvedValue(ledger);
    window.go = {
      service: {
        InventoryHandler: {
          ListInventoryBalances: listInventoryBalances,
          ListItemLotFacts: listItemLotFacts,
          ListEligibleFEFOLots: listEligibleFEFOLots,
          ListItemLedgerPage: listItemLedgerPage,
        },
      },
    };

    const balanceRequest = { includeArchived: true, pageSize: 25 };
    await expect(inventoryGateway.listInventoryBalances(balanceRequest)).resolves.toEqual(
      balancePage,
    );
    await expect(inventoryGateway.listItemLotFacts(10)).resolves.toEqual(lots);
    await expect(inventoryGateway.listEligibleFefoLots(10, "2026-07-15")).resolves.toEqual(lots);
    await expect(
      inventoryGateway.listItemLedgerPage({ itemId: 10, pageSize: 10 }),
    ).resolves.toEqual(ledger);

    expect(listInventoryBalances).toHaveBeenCalledWith(balanceRequest);
    expect(listItemLotFacts).toHaveBeenCalledWith(10);
    expect(listEligibleFEFOLots).toHaveBeenCalledWith(10, "2026-07-15");
    expect(listItemLedgerPage).toHaveBeenCalledWith({ itemId: 10, pageSize: 10 });
  });

  it("forwards reporting calls to the V2 reporting handler", async () => {
    const request = {
      fromOccurredOn: "2026-07-01",
      toOccurredOn: "2026-07-31",
      granularity: "MONTH" as const,
    };
    const categoryMix = {
      period: request,
      available: false,
      unavailableReason: "Catalog categories/tags are not modeled in V2 yet.",
      rows: [],
    };
    const salesReport = {
      period: request,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      totalSalesCount: 1,
      totalRevenueMinor: 1_000,
      cogsMicro: 600_000,
      grossMarginMicro: 9_400_000,
      grossMarginBasisPoints: 9_400,
      averageTicketMinor: 1_000,
      salesRevenueSeries: [],
      monthlyRevenueSeries: [],
      monthlySalesSeries: [],
      topProductsByQuantity: [],
      topProductsByRevenue: [],
      freeSales: {
        reasonCode: "FREE_SALES",
        documentCount: 0,
        quantityAtomic: 0,
        revenueMinor: 0,
        inventoryValueMicro: 0,
      },
      salesByCustomer: [],
      anonymousSales: {
        documentCount: 0,
        revenueMinor: 0,
        spendMinor: 0,
      },
    };
    const inventoryReport = {
      period: request,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      totalInventoryValueMicro: 4_900_000,
      lowStockItemCount: 1,
      zeroStockSellableCount: 0,
      lowStockItems: [],
      expiringLots7Days: [],
      expiringLots30Days: [],
      expiredLotsWithStock: [],
      inventoryValueByItem: [],
    };
    const purchaseReport = {
      period: request,
      currencyCode: "BRL",
      currencyMinorDigits: 2,
      purchaseSpendSeries: [
        {
          bucket: "2026-07",
          label: "2026-07",
          documentCount: 1,
          salesCount: 0,
          quantityAtomic: 1_000,
          revenueMinor: 500,
          spendMinor: 500,
          inventoryValueMicro: 5_000_000,
          directCostMicro: 0,
          grossMarginMicro: 0,
        },
      ],
      topSuppliersBySpend: [],
      freeStockEntries: [],
    };
    const getSalesReport = vi.fn().mockResolvedValue(salesReport);
    const getInventoryReport = vi.fn().mockResolvedValue(inventoryReport);
    const getPurchaseReport = vi.fn().mockResolvedValue(purchaseReport);
    const getCategoryMixReport = vi.fn().mockResolvedValue(categoryMix);
    window.go = {
      service: {
        ReportingHandler: {
          GetSalesReport: getSalesReport,
          GetInventoryReport: getInventoryReport,
          GetPurchaseReport: getPurchaseReport,
          GetCategoryMixReport: getCategoryMixReport,
        },
      },
    };

    await expect(reportingGateway.getSalesReport(request)).resolves.toEqual(salesReport);
    await expect(reportingGateway.getInventoryReport(request)).resolves.toEqual(inventoryReport);
    await expect(reportingGateway.getPurchaseReport(request)).resolves.toEqual(purchaseReport);
    await expect(reportingGateway.getCategoryMixReport(request)).resolves.toEqual(categoryMix);
    expect(getSalesReport).toHaveBeenCalledWith(request);
    expect(getInventoryReport).toHaveBeenCalledWith(request);
    expect(getPurchaseReport).toHaveBeenCalledWith(request);
    expect(getCategoryMixReport).toHaveBeenCalledWith(request);
  });
});
