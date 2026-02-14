import { motion } from "framer-motion";
import {
  House,
  Package,
  FileText,
  CaretRight,
  ShoppingCart,
  Egg,
} from "phosphor-react";
import { useNavigate, useLocation } from "react-router-dom";
import { useState, useEffect } from "react";

const Sidebar = ({ isCollapsed, setIsCollapsed }) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [user, setUser] = useState(null);

  useEffect(() => {
    // Buscar dados do usuário da sessão armazenada
    const storedSession = localStorage.getItem("supabase_session");
    if (storedSession) {
      try {
        const { user: sessionUser } = JSON.parse(storedSession);
        setUser(sessionUser);
      } catch (err) {
        console.error("Erro ao ler sessão:", err);
      }
    }
  }, []);

  const defaultUser = {
    name: user?.email?.split("@")[0] || "Usuário",
    email: user?.email || "usuario@confeitaria.com",
    avatar: `https://api.dicebear.com/7.x/avataaars/svg?seed=${user?.email || "user"}`,
  };

  const menuItems = [
    {
      icon: <House size={24} />,
      label: "Painel",
      id: "dashboard",
      path: "/",
    },
    {
      icon: <Egg size={24} />,
      label: "Estoque",
      id: "inventory",
      path: "/inventory",
    },
    {
      icon: <Package size={24} />,
      label: "Lotes",
      id: "batches",
      path: "/batches",
    },
    {
      icon: <FileText size={24} />,
      label: "Receitas",
      id: "recipes",
      path: "/recipes",
    },
    {
      icon: <Package size={24} />,
      label: "Produtos",
      id: "products",
      path: "/products",
    },
    {
      icon: <ShoppingCart size={24} />,
      label: "Vendas",
      id: "sales",
      path: "/sales",
    },
  ];

  const isActive = (path) => location.pathname === path;

  return (
    <motion.aside
      initial={{ x: -300 }}
      animate={{ x: 0 }}
      className={`bg-linear-to-b from-slate-900 to-slate-800 text-white transition-all duration-300 ${
        isCollapsed ? "w-20" : "w-64"
      } min-h-screen fixed left-0 top-0 overflow-y-auto z-40`}
    >
      {/* Logo/Brand */}
      <div className="p-6 border-b border-slate-700">
        <motion.div
          initial={{ scale: 0 }}
          animate={{ scale: 1 }}
          className={`flex items-center gap-3 ${isCollapsed ? "justify-center" : ""}`}
        >
          <div className="w-10 h-10 bg-pink-600 rounded-lg flex items-center justify-center font-bold text-lg">
            🧁
          </div>
          {!isCollapsed && (
            <div>
              <h1 className="font-bold text-lg">Sweeters</h1>
              <p className="text-xs text-slate-400">Gestão Completa</p>
            </div>
          )}
        </motion.div>
      </div>

      {/* User Info */}
      {!isCollapsed && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.2 }}
          className="p-6 border-b border-slate-700"
        >
          <img
            src={defaultUser.avatar}
            alt={defaultUser.name}
            className="w-12 h-12 rounded-full mb-4"
          />
          <h3 className="font-semibold text-sm">{user?.name}</h3>
          <p className="text-xs text-slate-400 mb-3">{user?.email}</p>
          <div className="bg-slate-700 rounded-lg p-3">
            <p className="text-xs text-slate-300 mb-1">Loja</p>
            <p className="text-sm font-semibold">{user?.businessName}</p>
          </div>
        </motion.div>
      )}

      {/* Menu Items */}
      <nav className="p-4 space-y-2">
        {menuItems.map((item, i) => (
          <motion.button
            key={i}
            whileHover={{ scale: 1.02 }}
            whileTap={{ scale: 0.98 }}
            onClick={() => navigate(item.path)}
            className={`w-full flex items-center gap-4 px-4 py-3 rounded-lg transition-all ${
              isActive(item.path)
                ? "bg-pink-600 text-white"
                : "hover:bg-slate-700 text-white"
            } ${isCollapsed ? "justify-center" : ""}`}
          >
            <span
              className={isActive(item.path) ? "text-white" : "text-pink-500"}
            >
              {item.icon}
            </span>
            {!isCollapsed && (
              <span className="text-sm font-medium">{item.label}</span>
            )}
          </motion.button>
        ))}
      </nav>

      {/* Collapse Button */}
      <button
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="fixed top-6 bg-pink-600 rounded-full p-2 hover:bg-pink-700 transition-all shadow-lg z-9999"
        style={{
          left: isCollapsed ? "80px" : "256px",
          transform: "translateX(-50%)",
        }}
      >
        <CaretRight
          size={16}
          weight="bold"
          className={`transition-transform ${isCollapsed ? "" : "rotate-180"}`}
        />
      </button>
    </motion.aside>
  );
};

export default Sidebar;
