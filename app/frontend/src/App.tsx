import { lazy, Suspense, type ReactNode } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import AppLayout from "./components/AppLayout";
import { AppProvider } from "./context/AppContext";

const DashboardPage = lazy(() => import("./features/dashboard/DashboardPage"));
const DatabasePage = lazy(() => import("./features/database/DatabasePage"));
const EnterprisePage = lazy(() => import("./features/settings/EnterprisePage"));
const InventoryPage = lazy(() => import("./features/inventory/InventoryPage"));
const ProductsPage = lazy(() => import("./features/catalog/ProductsPage"));
const ProductionPage = lazy(() => import("./features/production/ProductionPage"));
const PurchasesPage = lazy(() => import("./features/purchases/PurchasesPage"));
const RecipesPage = lazy(() => import("./features/recipes/RecipesPage"));
const SalesPage = lazy(() => import("./features/sales/SalesPage"));

const desktopPage = (page: ReactNode) => (
  <AppLayout>
    <Suspense fallback={<p className="p-6 text-slate-600">Carregando...</p>}>{page}</Suspense>
  </AppLayout>
);

function App() {
  return (
    <AppProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={desktopPage(<DashboardPage />)} />
          <Route path="/inventory" element={desktopPage(<InventoryPage />)} />
          <Route path="/batches" element={<Navigate to="/inventory?view=lots" replace />} />
          <Route path="/purchases" element={desktopPage(<PurchasesPage />)} />
          <Route path="/production" element={desktopPage(<ProductionPage />)} />
          <Route path="/recipes" element={desktopPage(<RecipesPage />)} />
          <Route path="/products" element={desktopPage(<ProductsPage />)} />
          <Route path="/sales" element={desktopPage(<SalesPage />)} />
          <Route path="/enterprise" element={desktopPage(<EnterprisePage />)} />
          <Route path="/database" element={desktopPage(<DatabasePage />)} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AppProvider>
  );
}

export default App;
