import { useEffect, useState } from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import CadastroPage from "./pages/CadastroPage";
import PixPaymentPage from "./pages/PixPaymentPage";
import DashboardPage from "./pages/DashboardPage";
import LoadingScreen from "./components/LoadingScreen";
import PrivateRoute from "./components/PrivateRoute";
import InventoryPage from "./pages/InventoryPage";
import RecipesPage from "./pages/RecipesPage";
import ProductsPage from "./pages/ProductsPage";
import SalesPage from "./pages/SalesPage";
import BatchesPage from "./pages/BatchesPage";
import DatabasePage from "./pages/DatabasePage";
import AppLayout from "./components/AppLayout";
import { GetUserStatus } from "../wailsjs/go/main/UserService";
import { AppProvider } from "./context/AppContext";

function App() {
  const [loading, setLoading] = useState(true);
  const [isActive, setIsActive] = useState(false);

  useEffect(() => {
    GetUserStatus()
      .then(setIsActive)
      .catch((error) => {
        console.error("Erro ao obter status do usuário:", error);
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <LoadingScreen />;

  return (
    <AppProvider>
      <Router>
        <Routes>
          {/* públicas */}
          <Route path="/sign-in" element={<CadastroPage />} />
          <Route path="/pix-payment" element={<PixPaymentPage />} />

          {/* privadas */}
          <Route
            path="/"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <DashboardPage />
                </AppLayout>
              </PrivateRoute>
            }
          />
          <Route
            path="/inventory"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <InventoryPage />
                </AppLayout>
              </PrivateRoute>
            }
          />
          <Route
            path="/batches"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <BatchesPage />
                </AppLayout>
              </PrivateRoute>
            }
          />

          <Route
            path="/recipes"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <RecipesPage />
                </AppLayout>
              </PrivateRoute>
            }
          />

          <Route
            path="/products"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <ProductsPage />
                </AppLayout>
              </PrivateRoute>
            }
          />

          <Route
            path="/sales"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <SalesPage />
                </AppLayout>
              </PrivateRoute>
            }
          />
          <Route
            path="/database"
            element={
              <PrivateRoute isActive={isActive}>
                <AppLayout>
                  <DatabasePage />
                </AppLayout>
              </PrivateRoute>
            }
          />
        </Routes>
      </Router>
    </AppProvider>
  );
}

export default App;
