import { cleanup, render, screen } from "@testing-library/react";
import type { PropsWithChildren } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

vi.mock("./context/AppContext", () => ({
  AppProvider: ({ children }: PropsWithChildren) => children,
}));

vi.mock("./components/AppLayout", () => ({
  default: ({ children }: PropsWithChildren) => <main>{children}</main>,
}));

vi.mock("./features/dashboard/DashboardPage", () => ({ default: () => <h1>Dashboard</h1> }));
vi.mock("./features/inventory/InventoryPage", () => ({ default: () => <h1>Inventory</h1> }));
vi.mock("./features/inventory/BatchesPage", () => ({ default: () => <h1>Batches</h1> }));
vi.mock("./features/recipes/RecipesPage", () => ({ default: () => <h1>Recipes</h1> }));
vi.mock("./features/catalog/ProductsPage", () => ({ default: () => <h1>Products</h1> }));
vi.mock("./features/production/ProductionPage", () => ({ default: () => <h1>Production</h1> }));
vi.mock("./features/sales/SalesPage", () => ({ default: () => <h1>Sales</h1> }));
vi.mock("./features/settings/EnterprisePage", () => ({ default: () => <h1>Enterprise</h1> }));
vi.mock("./features/database/DatabasePage", () => ({ default: () => <h1>Database</h1> }));

describe("desktop routes", () => {
  beforeEach(() => {
    window.history.replaceState({}, "", "/");
  });

  afterEach(cleanup);

  it("renders the dashboard at the root route", async () => {
    render(<App />);

    expect(await screen.findByRole("heading", { name: "Dashboard" })).toBeInTheDocument();
  });

  it("redirects an unknown route to the dashboard", async () => {
    window.history.replaceState({}, "", "/does-not-exist");

    render(<App />);

    expect(await screen.findByRole("heading", { name: "Dashboard" })).toBeInTheDocument();
    expect(window.location.pathname).toBe("/");
  });
});
