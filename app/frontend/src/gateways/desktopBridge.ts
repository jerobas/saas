type BridgeMethod = (...args: unknown[]) => Promise<unknown>;
type ServiceRegistry = Record<string, Record<string, BridgeMethod>>;

interface WailsBridge {
  service?: ServiceRegistry;
  main?: ServiceRegistry;
  wails?: ServiceRegistry;
}

declare global {
  interface Window {
    go?: WailsBridge;
  }
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

export interface PurchaseCursorRequest {
  postingSequence: number;
  id: number;
}

export interface PurchaseCursorResponse {
  postingSequence: number;
  id: number;
}

export interface PurchaseListRequest {
  after?: PurchaseCursorRequest | null;
  pageSize?: number;
}

export interface PurchasePageResponse {
  items: PurchaseDocumentResponse[];
  next?: PurchaseCursorResponse | null;
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

export interface InventoryBalanceCursorRequest {
  itemName: string;
  itemId: number;
}

export interface InventoryBalanceCursorResponse {
  itemName: string;
  itemId: number;
}

export interface InventoryBalanceListRequest {
  includeArchived?: boolean;
  search?: string | null;
  after?: InventoryBalanceCursorRequest | null;
  pageSize?: number;
}

export interface InventoryBalancePageResponse {
  items: InventoryBalanceResponse[];
  next?: InventoryBalanceCursorResponse | null;
}

export interface InventoryBalanceResponse {
  itemId: number;
  itemName: string;
  baseUnitCode: string;
  itemArchivedAtMs?: number | null;
  quantityAtomic: number;
  inventoryValueMicro: number;
  lastDocumentId?: number | null;
  updatedAtMs: number;
  capabilities: CapabilitiesResponse;
  reorderQuantityAtomic?: number | null;
}

export interface LotResponse {
  id: number;
  itemId: number;
  sourceLineId: number;
  sourcePostingSequence: number;
  initialQuantityAtomic: number;
  consumedQuantityAtomic: number;
  restoredQuantityAtomic: number;
  availableQuantityAtomic: number;
  lotCode?: string | null;
  originatedOn: string;
  expiresOn?: string | null;
  createdAtMs: number;
  sourceDocumentId: number;
  sourceKind: string;
  sourceOccurredOn: string;
}

export interface LedgerCursorRequest {
  postingSequence: number;
  lineOrder: number;
  lineId: number;
}

export interface LedgerCursorResponse {
  postingSequence: number;
  lineOrder: number;
  lineId: number;
}

export interface ItemLedgerPageRequest {
  itemId: number;
  after?: LedgerCursorRequest | null;
  pageSize?: number;
}

export interface LedgerEntryPageResponse {
  items: LedgerEntryResponse[];
  next?: LedgerCursorResponse | null;
}

export interface LedgerEntryResponse {
  lineId: number;
  documentId: number;
  postingSequence: number;
  lineOrder: number;
  documentKind: string;
  occurredOn: string;
  postedAtMs: number;
  itemId: number;
  direction: string;
  quantityAtomic: number;
  inventoryValueMicro: number;
  commercialTotalMinor?: number | null;
  currencyCode: string;
  currencyMinorDigits: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  reversesLineId?: number | null;
  idempotencyKey: string;
  counterpartyId?: number | null;
  counterpartyName?: string | null;
  reasonCode?: string | null;
  notes?: string | null;
  reversesDocumentId?: number | null;
}

export interface AllocationResponse {
  id: number;
  lineId: number;
  lotId: number;
  quantityAtomic: number;
  effect: string;
  restoresAllocationId?: number | null;
  createdAtMs: number;
  sourceLineId: number;
  lotInitialQuantityAtomic: number;
  lotCode?: string | null;
  originatedOn: string;
  expiresOn?: string | null;
}

export type AdjustmentReason =
  | "OPENING_BALANCE"
  | "FREE_STOCK"
  | "PHYSICAL_COUNT"
  | "WASTE"
  | "EXPIRY"
  | "DAMAGE"
  | "SAMPLE"
  | "DOCUMENTED_CORRECTION";

export type StockDirection = "IN" | "OUT";

export interface AdjustmentPostRequest {
  idempotencyKey: string;
  occurredOn: string;
  reasonCode: AdjustmentReason;
  notes?: string | null;
  lines: AdjustmentLineRequest[];
}

export interface AdjustmentLineRequest {
  itemId: number;
  direction: StockDirection;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  inventoryValueMicro?: number | null;
  lotCode?: string | null;
  expiresOn?: string | null;
}

export interface AdjustmentDocumentResponse {
  id: number;
  idempotencyKey: string;
  postingSequence: number;
  occurredOn: string;
  postedAtMs: number;
  currencyCode: string;
  currencyMinorDigits: number;
  reasonCode: AdjustmentReason;
  notes?: string | null;
  lines: AdjustmentLineResponse[];
}

export interface AdjustmentLineResponse {
  id: number;
  lineOrder: number;
  itemId: number;
  direction: StockDirection;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  inventoryValueMicro: number;
  lotId?: number | null;
  lotCode?: string | null;
  originatedOn?: string | null;
  expiresOn?: string | null;
  allocations: AdjustmentAllocationResponse[];
}

export interface AdjustmentAllocationResponse {
  id: number;
  lotId: number;
  quantityAtomic: number;
}

export interface ReversalPostRequest {
  idempotencyKey: string;
  targetDocumentId: number;
  occurredOn: string;
  notes?: string | null;
}

export interface ReversalDocumentResponse {
  id: number;
  idempotencyKey: string;
  postingSequence: number;
  targetDocumentId: number;
  occurredOn: string;
  postedAtMs: number;
  currencyCode: string;
  currencyMinorDigits: number;
  reasonCode: "EXACT_REVERSAL";
  notes?: string | null;
  lines: ReversalLineResponse[];
}

export interface ReversalLineResponse {
  id: number;
  lineOrder: number;
  itemId: number;
  direction: StockDirection;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  inventoryValueMicro: number;
  commercialTotalMinor?: number | null;
  reversesLineId: number;
  allocations: ReversalAllocationResponse[];
}

export interface ReversalAllocationResponse {
  id: number;
  lotId: number;
  quantityAtomic: number;
  restoresAllocationId?: number | null;
}

export interface ProductionPostRequest {
  idempotencyKey: string;
  recipeRevisionId: number;
  occurredOn: string;
  directCostMicro: number;
  notes?: string | null;
  output: ProductionOutputRequest;
  inputs: ProductionComponentRequest[];
}

export interface ProductionOutputRequest {
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  lotCode?: string | null;
  expiresOn?: string | null;
}

export interface ProductionComponentRequest {
  itemId: number;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  lotId?: number | null;
}

export interface ProductionDocumentResponse {
  id: number;
  idempotencyKey: string;
  postingSequence: number;
  recipeRevisionId: number;
  outputItemId: number;
  occurredOn: string;
  postedAtMs: number;
  currencyCode: string;
  currencyMinorDigits: number;
  directCostMicro: number;
  notes?: string | null;
  outputLine: ProductionLineResponse;
  inputLines: ProductionLineResponse[];
}

export interface ProductionLineResponse {
  id: number;
  lineOrder: number;
  itemId: number;
  direction: StockDirection;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  inventoryValueMicro: number;
  lotId?: number | null;
  lotCode?: string | null;
  originatedOn?: string | null;
  expiresOn?: string | null;
  allocations: ProductionAllocationResponse[];
}

export interface ProductionAllocationResponse {
  id: number;
  lotId: number;
  quantityAtomic: number;
}

export type SaleReason = "PROMOTION" | "SAMPLE";

export interface SalePostRequest {
  idempotencyKey: string;
  counterpartyId?: number | null;
  occurredOn: string;
  reasonCode?: SaleReason | null;
  notes?: string | null;
  lines: SaleLineRequest[];
}

export interface SaleLineRequest {
  itemId: number;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  commercialTotalMinor: number;
  lotId?: number | null;
}

export interface SaleDocumentResponse {
  id: number;
  idempotencyKey: string;
  postingSequence: number;
  counterpartyId?: number | null;
  occurredOn: string;
  postedAtMs: number;
  currencyCode: string;
  currencyMinorDigits: number;
  reasonCode?: SaleReason | null;
  notes?: string | null;
  lines: SaleLineResponse[];
}

export interface SaleLineResponse {
  id: number;
  lineOrder: number;
  itemId: number;
  direction: StockDirection;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  inventoryValueMicro: number;
  commercialTotalMinor: number;
  allocations: SaleAllocationResponse[];
}

export interface SaleAllocationResponse {
  id: number;
  lotId: number;
  quantityAtomic: number;
}

export interface SaleCursorRequest {
  postingSequence: number;
  id: number;
}

export interface SaleCursorResponse {
  postingSequence: number;
  id: number;
}

export interface SaleListRequest {
  after?: SaleCursorRequest | null;
  pageSize?: number;
}

export interface SalePageResponse {
  items: SaleDocumentResponse[];
  next?: SaleCursorResponse | null;
}

export interface RecipeCursorRequest {
  name: string;
  id: number;
}

export interface RecipeCursorResponse {
  name: string;
  id: number;
}

export interface RecipeListRequest {
  archiveFilter?: ArchiveFilter;
  search?: string | null;
  after?: RecipeCursorRequest | null;
  pageSize?: number;
}

export interface RecipePageResponse {
  items: RecipeSummaryResponse[];
  next?: RecipeCursorResponse | null;
}

export interface CurrentRecipeRevisionSummaryResponse {
  id: number;
  number: number;
  standardYieldQuantityAtomic: number;
}

export interface RecipeSummaryResponse {
  id: number;
  name: string;
  outputItemId: number;
  outputItemName: string;
  createdAtMs: number;
  updatedAtMs: number;
  archivedAtMs?: number | null;
  currentRevision: CurrentRecipeRevisionSummaryResponse;
}

export interface RecipeResponse {
  id: number;
  name: string;
  outputItemId: number;
  createdAtMs: number;
  updatedAtMs: number;
  archivedAtMs?: number | null;
  currentRevision: RecipeRevisionResponse;
}

export interface RecipeCreateRequest {
  name: string;
  outputItemId: number;
  revision: RecipeRevisionWriteRequest;
}

export interface RecipePublishRevisionRequest {
  expectedLatestRevision: number;
  expectedUpdatedAtMs: number;
  revision: RecipeRevisionWriteRequest;
}

export interface RecipeRenameRequest {
  name: string;
  expectedUpdatedAtMs: number;
}

export interface RecipeRevisionWriteRequest {
  standardYieldQuantityAtomic: number;
  instructions: string;
  preparationTimeMinutes: number;
  estimatedDirectCostMicro?: number | null;
  components: RecipeComponentRequest[];
}

export type RecipeComponentSourceType = "UNIT" | "PACKAGING";

export interface RecipeComponentRequest {
  order: number;
  itemId: number;
  quantityAtomic: number;
  sourceType: RecipeComponentSourceType;
  unitCode?: string | null;
  packagingId?: number | null;
}

export interface RecipeRevisionResponse {
  id: number;
  recipeId: number;
  number: number;
  standardYieldQuantityAtomic: number;
  instructions: string;
  preparationTimeMinutes: number;
  estimatedDirectCostMicro?: number | null;
  createdAtMs: number;
  components: RecipeComponentResponse[];
}

export interface RecipeComponentResponse {
  id: number;
  revisionId: number;
  order: number;
  itemId: number;
  quantityAtomic: number;
  enteredUnitCode: string;
  enteredPackagingName?: string | null;
  conversionNumeratorAtomic: number;
  conversionDenominator: number;
  createdAtMs: number;
}

export type ReportingGranularity = "DAY" | "MONTH";

export interface ReportingPeriodRequest {
  fromOccurredOn: string;
  toOccurredOn: string;
  granularity?: ReportingGranularity;
}

export interface ReportingPeriodResponse {
  fromOccurredOn: string;
  toOccurredOn: string;
  granularity: ReportingGranularity;
}

export interface SalesReportResponse {
  period: ReportingPeriodResponse;
  currencyCode: string;
  currencyMinorDigits: number;
  totalSalesCount: number;
  commercialTotalMinor: number;
  cogsInventoryValueMicro: number;
  grossMarginInventoryValueMicro: number;
  grossMarginBasisPoints?: number | null;
  averageCommercialTotalMinor?: number | null;
  growthBasisPoints?: number | null;
  salesRevenueSeries: ReportingSeriesResponse[];
  monthlyRevenueSeries: ReportingSeriesResponse[];
  monthlySalesSeries: ReportingSeriesResponse[];
  topProductsByQuantity: ReportingItemMetricResponse[];
  topProductsByRevenue: ReportingItemMetricResponse[];
  freeSales: ReportingReasonMetricResponse;
  salesByCustomer: ReportingCounterpartyMetricResponse[];
  anonymousSales: ReportingCounterpartyMetricResponse;
}

export interface InventoryReportResponse {
  period: ReportingPeriodResponse;
  currencyCode: string;
  currencyMinorDigits: number;
  totalInventoryValueMicro: number;
  lowStockItemCount: number;
  zeroStockSellableCount: number;
  lowStockItems: ReportingItemMetricResponse[];
  expiringLots7Days: ReportingLotMetricResponse[];
  expiringLots30Days: ReportingLotMetricResponse[];
  expiredLotsWithStock: ReportingLotMetricResponse[];
  inventoryValueByItem: ReportingItemMetricResponse[];
}

export interface PurchaseReportResponse {
  period: ReportingPeriodResponse;
  currencyCode: string;
  currencyMinorDigits: number;
  purchaseSpendSeries: ReportingSeriesResponse[];
  topSuppliersBySpend: ReportingCounterpartyMetricResponse[];
  freeStockEntries: ReportingSeriesResponse[];
}

export interface ProductionReportResponse {
  period: ReportingPeriodResponse;
  currencyCode: string;
  currencyMinorDigits: number;
  productionByRecipeProduct: ReportingItemMetricResponse[];
  directCostSeries: ReportingSeriesResponse[];
  yieldVariance: ReportingItemMetricResponse[];
}

export interface AdjustmentReportResponse {
  period: ReportingPeriodResponse;
  currencyCode: string;
  currencyMinorDigits: number;
  negativeByReason: ReportingReasonMetricResponse[];
  positiveByReason: ReportingReasonMetricResponse[];
  exactReversals: ReportingSeriesResponse[];
}

export interface CategoryMixReportResponse {
  period: ReportingPeriodResponse;
  available: boolean;
  unavailableReason?: string | null;
  rows: CategoryMixRowResponse[];
}

export interface CategoryMixRowResponse {
  categoryName: string;
  quantityAtomic: number;
  commercialTotalMinor: number;
  shareBasisPoints: number;
}

export interface ReportingSeriesResponse {
  bucket: string;
  label: string;
  documentCount: number;
  salesCount: number;
  quantityAtomic: number;
  commercialTotalMinor: number;
  inventoryValueMicro: number;
  directCostInventoryValueMicro: number;
  grossMarginInventoryValueMicro: number;
}

export interface ReportingItemMetricResponse {
  itemId?: number | null;
  itemName: string;
  recipeId?: number | null;
  recipeName?: string | null;
  baseUnitCode?: string | null;
  documentCount: number;
  quantityAtomic: number;
  commercialTotalMinor: number;
  inventoryValueMicro: number;
  directCostInventoryValueMicro: number;
  reorderQuantityAtomic?: number | null;
  standardYieldAtomic?: number | null;
  actualYieldAtomic?: number | null;
  varianceAtomic?: number | null;
}

export interface ReportingCounterpartyMetricResponse {
  counterpartyId?: number | null;
  counterpartyName?: string | null;
  documentCount: number;
  commercialTotalMinor: number;
}

export interface ReportingReasonMetricResponse {
  reasonCode: string;
  documentCount: number;
  quantityAtomic: number;
  commercialTotalMinor: number;
  inventoryValueMicro: number;
}

export interface ReportingLotMetricResponse {
  lotId: number;
  itemId: number;
  itemName: string;
  lotCode?: string | null;
  expiresOn?: string | null;
  availableQuantityAtomic: number;
  inventoryValueMicro: number;
}

async function invoke<T>(service: string, method: string, ...args: unknown[]): Promise<T> {
  const bridgeMethod =
    window.go?.service?.[service]?.[method] ??
    window.go?.main?.[service]?.[method] ??
    window.go?.wails?.[service]?.[method];

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
  getPurchase: (id: number) =>
    invoke<PurchaseDocumentResponse>("PurchaseHandler", "GetPurchase", id),
  listPurchases: (request: PurchaseListRequest) =>
    invoke<PurchasePageResponse>("PurchaseHandler", "ListPurchases", request),
  postPurchase: (request: PurchasePostRequest) =>
    invoke<PurchaseDocumentResponse>("PurchaseHandler", "PostPurchase", request),
};

export const adjustmentGateway = {
  postAdjustment: (request: AdjustmentPostRequest) =>
    invoke<AdjustmentDocumentResponse>("AdjustmentHandler", "PostAdjustment", request),
};

export const reversalGateway = {
  postReversal: (request: ReversalPostRequest) =>
    invoke<ReversalDocumentResponse>("ReversalHandler", "PostReversal", request),
};

export const productionGateway = {
  postProduction: (request: ProductionPostRequest) =>
    invoke<ProductionDocumentResponse>("ProductionHandler", "PostProduction", request),
};

export const saleGateway = {
  getSale: (id: number) => invoke<SaleDocumentResponse>("SaleHandler", "GetSale", id),
  listSales: (request: SaleListRequest) =>
    invoke<SalePageResponse>("SaleHandler", "ListSales", request),
  postSale: (request: SalePostRequest) =>
    invoke<SaleDocumentResponse>("SaleHandler", "PostSale", request),
};

export const recipeGateway = {
  getRecipe: (id: number) => invoke<RecipeResponse>("RecipeHandler", "GetRecipe", id),
  getRecipeRevision: (id: number) =>
    invoke<RecipeRevisionResponse>("RecipeHandler", "GetRecipeRevision", id),
  listRecipeRevisions: (recipeId: number) =>
    invoke<RecipeRevisionResponse[]>("RecipeHandler", "ListRecipeRevisions", recipeId),
  listRecipes: (request: RecipeListRequest) =>
    invoke<RecipePageResponse>("RecipeHandler", "ListRecipes", request),
  createRecipe: (request: RecipeCreateRequest) =>
    invoke<RecipeResponse>("RecipeHandler", "CreateRecipe", request),
  publishRecipeRevision: (id: number, request: RecipePublishRevisionRequest) =>
    invoke<RecipeRevisionResponse>("RecipeHandler", "PublishRecipeRevision", id, request),
  renameRecipe: (id: number, request: RecipeRenameRequest) =>
    invoke<RecipeResponse>("RecipeHandler", "RenameRecipe", id, request),
  archiveRecipe: (id: number, request: VersionedRequest) =>
    invoke<RecipeResponse>("RecipeHandler", "ArchiveRecipe", id, request),
  restoreRecipe: (id: number, request: VersionedRequest) =>
    invoke<RecipeResponse>("RecipeHandler", "RestoreRecipe", id, request),
};

export const inventoryGateway = {
  getInventoryBalance: (itemId: number) =>
    invoke<InventoryBalanceResponse>("InventoryHandler", "GetInventoryBalance", itemId),
  listInventoryBalances: (request: InventoryBalanceListRequest) =>
    invoke<InventoryBalancePageResponse>("InventoryHandler", "ListInventoryBalances", request),
  listItemLotFacts: (itemId: number) =>
    invoke<LotResponse[]>("InventoryHandler", "ListItemLotFacts", itemId),
  listEligibleFefoLots: (itemId: number, businessDate: string) =>
    invoke<LotResponse[]>("InventoryHandler", "ListEligibleFEFOLots", itemId, businessDate),
  listItemLedgerPage: (request: ItemLedgerPageRequest) =>
    invoke<LedgerEntryPageResponse>("InventoryHandler", "ListItemLedgerPage", request),
  listLineAllocations: (lineId: number) =>
    invoke<AllocationResponse[]>("InventoryHandler", "ListLineAllocations", lineId),
};

export const reportingGateway = {
  getSalesReport: (request: ReportingPeriodRequest) =>
    invoke<SalesReportResponse>("ReportingHandler", "GetSalesReport", request),
  getInventoryReport: (request: ReportingPeriodRequest) =>
    invoke<InventoryReportResponse>("ReportingHandler", "GetInventoryReport", request),
  getPurchaseReport: (request: ReportingPeriodRequest) =>
    invoke<PurchaseReportResponse>("ReportingHandler", "GetPurchaseReport", request),
  getProductionReport: (request: ReportingPeriodRequest) =>
    invoke<ProductionReportResponse>("ReportingHandler", "GetProductionReport", request),
  getAdjustmentReport: (request: ReportingPeriodRequest) =>
    invoke<AdjustmentReportResponse>("ReportingHandler", "GetAdjustmentReport", request),
  getCategoryMixReport: (request: ReportingPeriodRequest) =>
    invoke<CategoryMixReportResponse>("ReportingHandler", "GetCategoryMixReport", request),
};

export const ExportDatabase = () => invoke<void>("DatabaseService", "Export");
