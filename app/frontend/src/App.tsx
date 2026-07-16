import { lazy, Suspense, type ReactNode } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import AppLayout from "./components/AppLayout";
import { AppProvider } from "./context/AppContext";

const BatchesPage = lazy(() => import("./pages/BatchesPage"));
const DashboardPage = lazy(() => import("./pages/DashboardPage"));
const DatabasePage = lazy(() => import("./pages/DatabasePage"));
const EnterprisePage = lazy(() => import("./pages/EnterprisePage"));
const InventoryPage = lazy(() => import("./pages/InventoryPage"));
const ProductsPage = lazy(() => import("./pages/ProductsPage"));
const ProductionPage = lazy(() => import("./pages/ProductionPage"));
const PurchasesPage = lazy(() => import("./pages/PurchasesPage"));
const RecipesPage = lazy(() => import("./pages/RecipesPage"));
const SalesPage = lazy(() => import("./pages/SalesPage"));

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
          <Route path="/batches" element={desktopPage(<BatchesPage />)} />
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
