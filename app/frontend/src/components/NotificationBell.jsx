import { useContext, useState, useEffect, useRef } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { Bell } from "phosphor-react";
import { AppContext } from "../context/AppContext";

const NotificationBell = () => {
  const {
    notifications,
    unreadCount,
    markAllNotificationsRead,
    clearNotifications,
  } = useContext(AppContext);

  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef(null);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (containerRef.current && !containerRef.current.contains(event.target)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  const toggleOpen = () => {
    const next = !isOpen;
    setIsOpen(next);

    if (next) {
      markAllNotificationsRead();
    }
  };

  return (
    <div className="relative" ref={containerRef}>
      <motion.button
        type="button"
        onClick={toggleOpen}
        aria-label="Notificacoes"
        whileHover={{ scale: 1.06, rotate: -4 }}
        transition={{ type: "spring", stiffness: 300, damping: 18 }}
        className="group relative w-11 h-11 rounded-full bg-gradient-to-br from-slate-800 to-slate-900 shadow-lg border border-slate-700 flex items-center justify-center hover:shadow-xl hover:from-slate-700 hover:to-slate-800 transition-all cursor-pointer"
      >
        <Bell
          size={22}
          className="text-slate-300 transition-colors group-hover:text-pink-500"
          weight="bold"
        />

        {unreadCount > 0 && (
          <motion.span
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            className="absolute -top-1 -right-1 min-w-[18px] h-[18px] px-1 bg-gradient-to-br from-pink-500 to-pink-600 text-white text-[10px] font-semibold leading-[18px] rounded-full text-center shadow-lg border border-pink-400"
          >
            {unreadCount}
          </motion.span>
        )}
      </motion.button>

      <AnimatePresence>
        {isOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10, scale: 0.95 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: -10, scale: 0.95 }}
            transition={{ duration: 0.2 }}
            className="absolute right-0 mt-3 w-80 bg-gradient-to-br from-slate-800 to-slate-900 border border-slate-700 rounded-2xl shadow-2xl overflow-hidden"
          >
            <div className="px-5 py-4 text-sm font-semibold text-white border-b border-slate-700 bg-slate-900/50 flex items-center justify-between">
              <span className="flex items-center gap-2">
                <Bell size={18} weight="bold" className="text-pink-500" />
                Notificações
              </span>

              {notifications.length > 0 && (
                <button
                  onClick={clearNotifications}
                  className="text-xs text-slate-400 font-normal hover:text-pink-500 transition-colors cursor-pointer"
                  title="Limpar todas"
                >
                  Limpar tudo
                </button>
              )}
            </div>

            <div className="max-h-80 overflow-auto">
              {notifications.length === 0 ? (
                <div className="px-5 py-8 text-sm text-slate-400 text-center">
                  <Bell
                    size={32}
                    className="mx-auto mb-2 text-slate-600"
                    weight="duotone"
                  />
                  Sem notificações no momento
                </div>
              ) : (
                notifications.map((item) => (
                  <motion.div
                    key={item.id}
                    initial={{ opacity: 0, x: -10 }}
                    animate={{ opacity: 1, x: 0 }}
                    whileHover={{
                      backgroundColor: "rgba(100,116,139,0.1)",
                    }}
                    className="px-5 py-4 text-sm border-b border-slate-700/50 last:border-b-0 transition-colors cursor-pointer"
                  >
                    <div className="flex items-start gap-3">
                      <div
                        className={`w-2 h-2 rounded-full mt-1.5 ${
                          item.read
                            ? "bg-slate-600"
                            : "bg-pink-500 shadow-lg shadow-pink-500/50"
                        }`}
                      />

                      <div className="flex-1 min-w-0">
                        <div className="font-semibold text-white">
                          {item.title}
                        </div>

                        {item.message && (
                          <div className="text-slate-400 mt-1 text-xs leading-relaxed">
                            {item.message}
                          </div>
                        )}
                      </div>
                    </div>
                  </motion.div>
                ))
              )}
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
};

export default NotificationBell;
