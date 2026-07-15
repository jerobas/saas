type BridgeMethod = (...args: unknown[]) => Promise<unknown>;
type ServiceRegistry = Record<string, Record<string, BridgeMethod>>;

interface WailsBridge {
  service?: ServiceRegistry;
  main?: ServiceRegistry;
}

declare global {
  interface Window {
    go?: WailsBridge;
  }
}

export interface LegacyItem {
  id: string;
  name: string;
  unit: string;
  min_stock_alert: number;
}

export interface LegacyBatch {
  id: string;
  item_id: string;
  quantity_remaining: number;
  unit_price: number;
}

export interface LegacyRecipeIngredientInput {
  item_id: string;
  quantity: number;
}

export interface LegacyRecipe {
  id: string;
  name: string;
  ingredients?: LegacyRecipeIngredientInput[];
}

export type ArchiveFilter = "ACTIVE" | "ARCHIVED" | "ALL";
export type CounterpartyRole = "SUPPLIER" | "CUSTOMER";
export type MeasurementDimension = "MASS" | "VOLUME" | "COUNT";

export interface SettingsResponse {
  businessName: string;
  locale: string;
  timezone: string;
  currencyCode: string;
  currencyMinorDigits: number;
  hourlyLaborCost?: number | null;
  defaultGrossMargin?: number | null;
  createdAtMs: number;
  updatedAtMs: number;
}

export interface SettingsUpdateRequest {
  businessName: string;
  locale: string;
  timezone: string;
  currencyCode: string;
  currencyMinorDigits: number;
  hourlyLaborCost?: number | null;
  defaultGrossMargin?: number | null;
  expectedUpdatedAtMs: number;
}

export interface MeasurementUnitResponse {
  code: string;
  name: string;
  symbol: string;
  dimension: MeasurementDimension;
  numeratorAtomic: number;
  denominator: number;
  isItemBase: boolean;
  isSeeded: boolean;
}

export interface CapabilitiesRequest {
  purchasable: boolean;
  producible: boolean;
  sellable: boolean;
}

export interface CapabilitiesResponse {
  purchasable: boolean;
  producible: boolean;
  sellable: boolean;
}

export interface ItemCursorRequest {
  name: string;
  id: number;
}

export interface ItemCursorResponse {
  name: string;
  id: number;
}

export interface ItemListRequest {
  archiveFilter?: ArchiveFilter;
  requireCapabilities: CapabilitiesRequest;
  search?: string | null;
  after?: ItemCursorRequest | null;
  pageSize?: number;
}

export interface ItemPageResponse {
  items: ItemSummaryResponse[];
  next?: ItemCursorResponse | null;
}

export interface ItemSummaryResponse {
  id: number;
  name: string;
  sku?: string | null;
  description?: string | null;
  baseUnitCode: string;
  capabilities: CapabilitiesResponse;
  defaultSalePrice?: number | null;
  reorderQuantityAtomic?: number | null;
  createdAtMs: number;
  updatedAtMs: number;
  archivedAtMs?: number | null;
}

export interface ItemResponse extends ItemSummaryResponse {
  baseUnit: MeasurementUnitResponse;
  packagings: PackagingResponse[];
}

export interface ItemWriteRequest {
  name: string;
  sku?: string | null;
  description?: string | null;
  baseUnitCode: string;
  capabilities: CapabilitiesRequest;
  defaultSalePrice?: number | null;
  reorderQuantityAtomic?: number | null;
}

export interface ItemUpdateRequest extends ItemWriteRequest {
  expectedUpdatedAtMs: number;
}

export interface VersionedRequest {
  expectedUpdatedAtMs: number;
}

export interface PackagingResponse {
  id: number;
  itemId: number;
  name: string;
  enteredUnitCode: string;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  baseUnit: MeasurementUnitResponse;
  enteredUnit: MeasurementUnitResponse;
  createdAtMs: number;
  updatedAtMs: number;
  archivedAtMs?: number | null;
}

export interface PackagingWriteRequest {
  name: string;
  enteredUnitCode: string;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
}

export interface PackagingCreateRequest extends PackagingWriteRequest {
  itemId: number;
}

export interface PackagingUpdateRequest extends PackagingWriteRequest {
  expectedUpdatedAtMs: number;
}

export interface CounterpartyCursorRequest {
  name: string;
  id: number;
}

export interface CounterpartyCursorResponse {
  name: string;
  id: number;
}

export interface CounterpartyListRequest {
  archiveFilter?: ArchiveFilter;
  role?: CounterpartyRole | null;
  search?: string | null;
  after?: CounterpartyCursorRequest | null;
  pageSize?: number;
}

export interface CounterpartyPageResponse {
  items: CounterpartyResponse[];
  next?: CounterpartyCursorResponse | null;
}

export interface CounterpartyResponse {
  id: number;
  name: string;
  phone?: string | null;
  email?: string | null;
  notes?: string | null;
  roles: CounterpartyRole[];
  createdAtMs: number;
  updatedAtMs: number;
  archivedAtMs?: number | null;
}

export interface CounterpartyWriteRequest {
  name: string;
  phone?: string | null;
  email?: string | null;
  notes?: string | null;
  roles: CounterpartyRole[];
}

export interface CounterpartyUpdateRequest extends CounterpartyWriteRequest {
  expectedUpdatedAtMs: number;
}

export interface PurchasePostRequest {
  idempotencyKey: string;
  counterpartyId?: number | null;
  occurredOn: string;
  reasonCode?: "FREE_STOCK" | null;
  notes?: string | null;
  lines: PurchaseLineRequest[];
}

export interface PurchaseLineRequest {
  itemId: number;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  commercialTotalMinor: number;
  lotCode?: string | null;
  expiresOn?: string | null;
}

export interface PurchaseDocumentResponse {
  id: number;
  idempotencyKey: string;
  postingSequence: number;
  counterpartyId?: number | null;
  occurredOn: string;
  postedAtMs: number;
  currencyCode: string;
  currencyMinorDigits: number;
  reasonCode?: "FREE_STOCK" | null;
  notes?: string | null;
  lines: PurchaseLineResponse[];
}

export interface PurchaseLineResponse {
  id: number;
  lineOrder: number;
  itemId: number;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  inventoryValueMicro: number;
  commercialTotalMinor: number;
  lotId: number;
  lotCode?: string | null;
  originatedOn: string;
  expiresOn?: string | null;
}

async function invoke<T>(service: string, method: string, ...args: unknown[]): Promise<T> {
  const bridgeMethod =
    window.go?.service?.[service]?.[method] ?? window.go?.main?.[service]?.[method];

  if (typeof bridgeMethod !== "function") {
    throw new Error(`Desktop bridge method ${service}.${method} is unavailable.`);
  }

  return (await bridgeMethod(...args)) as T;
}

export const settingsGateway = {
  getSettings: () => invoke<SettingsResponse>("SettingsHandler", "GetSettings"),
  updateSettings: (request: SettingsUpdateRequest) =>
    invoke<SettingsResponse>("SettingsHandler", "UpdateSettings", request),
};

export const referenceDataGateway = {
  getMeasurementUnit: (code: string) =>
    invoke<MeasurementUnitResponse>("ReferenceDataHandler", "GetMeasurementUnit", code),
  listMeasurementUnits: () =>
    invoke<MeasurementUnitResponse[]>("ReferenceDataHandler", "ListMeasurementUnits"),
};

export const catalogGateway = {
  getItem: (id: number) => invoke<ItemResponse>("CatalogHandler", "GetItem", id),
  listItems: (request: ItemListRequest) =>
    invoke<ItemPageResponse>("CatalogHandler", "ListItems", request),
  createItem: (request: ItemWriteRequest) =>
    invoke<ItemResponse>("CatalogHandler", "CreateItem", request),
  updateItem: (id: number, request: ItemUpdateRequest) =>
    invoke<ItemResponse>("CatalogHandler", "UpdateItem", id, request),
  archiveItem: (id: number, request: VersionedRequest) =>
    invoke<ItemResponse>("CatalogHandler", "ArchiveItem", id, request),
  restoreItem: (id: number, request: VersionedRequest) =>
    invoke<ItemResponse>("CatalogHandler", "RestoreItem", id, request),
  getItemPackaging: (id: number) =>
    invoke<PackagingResponse>("CatalogHandler", "GetItemPackaging", id),
  createItemPackaging: (request: PackagingCreateRequest) =>
    invoke<PackagingResponse>("CatalogHandler", "CreateItemPackaging", request),
  updateItemPackaging: (id: number, request: PackagingUpdateRequest) =>
    invoke<PackagingResponse>("CatalogHandler", "UpdateItemPackaging", id, request),
  archiveItemPackaging: (id: number, request: VersionedRequest) =>
    invoke<PackagingResponse>("CatalogHandler", "ArchiveItemPackaging", id, request),
  reconfigureArchivedItemPackaging: (id: number, request: PackagingUpdateRequest) =>
    invoke<PackagingResponse>("CatalogHandler", "ReconfigureArchivedItemPackaging", id, request),
  restoreItemPackaging: (id: number, request: VersionedRequest) =>
    invoke<PackagingResponse>("CatalogHandler", "RestoreItemPackaging", id, request),
};

export const counterpartyGateway = {
  getCounterparty: (id: number) =>
    invoke<CounterpartyResponse>("CounterpartyHandler", "GetCounterparty", id),
  listCounterparties: (request: CounterpartyListRequest) =>
    invoke<CounterpartyPageResponse>("CounterpartyHandler", "ListCounterparties", request),
  createCounterparty: (request: CounterpartyWriteRequest) =>
    invoke<CounterpartyResponse>("CounterpartyHandler", "CreateCounterparty", request),
  updateCounterparty: (id: number, request: CounterpartyUpdateRequest) =>
    invoke<CounterpartyResponse>("CounterpartyHandler", "UpdateCounterparty", id, request),
  archiveCounterparty: (id: number, request: VersionedRequest) =>
    invoke<CounterpartyResponse>("CounterpartyHandler", "ArchiveCounterparty", id, request),
  restoreCounterparty: (id: number, request: VersionedRequest) =>
    invoke<CounterpartyResponse>("CounterpartyHandler", "RestoreCounterparty", id, request),
};

export const purchaseGateway = {
  postPurchase: (request: PurchasePostRequest) =>
    invoke<PurchaseDocumentResponse>("PurchaseHandler", "PostPurchase", request),
};

export const CreateItem = (name: string, unit: string, minimumStock: number) =>
  invoke<LegacyItem>("ItemService", "CreateItem", name, unit, minimumStock);

export const DeleteItem = (id: string) => invoke<void>("ItemService", "DeleteItem", id);

export const GetAllItems = () => invoke<LegacyItem[]>("ItemService", "GetAllItems");

export const CreateBatch = (itemId: string, quantity: number, totalPrice: number) =>
  invoke<LegacyBatch>("BatchService", "CreateBatch", itemId, quantity, totalPrice);

export const DeleteBatch = (id: string) => invoke<void>("BatchService", "DeleteBatch", id);

export const GetBatchesByItem = (itemId: string) =>
  invoke<LegacyBatch[]>("BatchService", "GetBatchesByItem", itemId);

export const CreateRecipe = (name: string, ingredients: LegacyRecipeIngredientInput[]) =>
  invoke<LegacyRecipe>("RecipeService", "CreateRecipe", name, ingredients);

export const DeleteRecipe = (id: string) => invoke<void>("RecipeService", "DeleteRecipe", id);

export const GetAllRecipes = () => invoke<LegacyRecipe[]>("RecipeService", "GetAllRecipes");

export const UpdateRecipe = (
  id: string,
  name: string,
  ingredients: LegacyRecipeIngredientInput[],
) => invoke<void>("RecipeService", "UpdateRecipe", id, name, ingredients);

export const ExportDatabase = () => invoke<void>("DatabaseService", "Export");
