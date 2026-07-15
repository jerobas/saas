import { afterEach, describe, expect, it, vi } from "vitest";
import {
  catalogGateway,
  counterpartyGateway,
  GetAllItems,
  referenceDataGateway,
  settingsGateway,
} from "./desktopBridge";

const originalBridge = window.go;

afterEach(() => {
  window.go = originalBridge;
});

describe("desktop bridge", () => {
  it("forwards calls to the Wails runtime", async () => {
    const getAllItems = vi.fn().mockResolvedValue([{ id: "item-1" }]);
    window.go = { service: { ItemService: { GetAllItems: getAllItems } } };

    await expect(GetAllItems()).resolves.toEqual([{ id: "item-1" }]);
    expect(getAllItems).toHaveBeenCalledOnce();
  });

  it("fails clearly when the desktop runtime is unavailable", async () => {
    window.go = undefined;

    await expect(GetAllItems()).rejects.toThrow(
      "Desktop bridge method ItemService.GetAllItems is unavailable.",
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
});
