type BridgeMethod = (...args: unknown[]) => Promise<unknown>;

interface WailsBridge {
  service?: Record<string, Record<string, BridgeMethod>>;
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

async function invoke<T>(service: string, method: string, ...args: unknown[]): Promise<T> {
  const bridgeMethod = window.go?.service?.[service]?.[method];

  if (typeof bridgeMethod !== "function") {
    throw new Error(`Desktop bridge method ${service}.${method} is unavailable.`);
  }

  return (await bridgeMethod(...args)) as T;
}

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

export const ImportDatabase = () => invoke<void>("DatabaseService", "Import");
