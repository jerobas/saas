import { BrowserRouter as Router, Navigate, Route, Routes } from "react-router-dom";
import AppLayout from "./components/AppLayout";
import { AppProvider } from "./context/AppContext";
import BatchesPage from "./pages/BatchesPage";
import DashboardPage from "./pages/DashboardPage";
import DatabasePage from "./pages/DatabasePage";
import EnterprisePage from "./pages/EnterprisePage";
import InventoryPage from "./pages/InventoryPage";
import ProductsPage from "./pages/ProductsPage";
import RecipesPage from "./pages/RecipesPage";
import SalesPage from "./pages/SalesPage";

const desktopPage = (page) => <AppLayout>{page}</AppLayout>;

function App() {
  return (
    <AppProvider>
      <Router>
        <Routes>
          <Route path="/" element={desktopPage(<DashboardPage />)} />
          <Route path="/inventory" element={desktopPage(<InventoryPage />)} />
          <Route path="/batches" element={desktopPage(<BatchesPage />)} />
          <Route path="/recipes" element={desktopPage(<RecipesPage />)} />
          <Route path="/products" element={desktopPage(<ProductsPage />)} />
          <Route path="/sales" element={desktopPage(<SalesPage />)} />
          <Route path="/enterprise" element={desktopPage(<EnterprisePage />)} />
          <Route path="/database" element={desktopPage(<DatabasePage />)} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </AppProvider>
  );
}

export default App;
