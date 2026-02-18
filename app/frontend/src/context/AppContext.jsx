import { createContext, useState, useRef, useEffect } from "react";
import { EventsOn } from "../../wailsjs/runtime/runtime";
import notificationSound from "../assets/notification.mp3";

export const AppContext = createContext();

export const AppProvider = ({ children }) => {
  const [pixData, setPixData] = useState(null);
  const [notifications, setNotifications] = useState([]);

  const audioRef = useRef(null);
  const unlockedRef = useRef(false);

  useEffect(() => {
    audioRef.current = new Audio(notificationSound);
    audioRef.current.volume = 0.7;
  }, []);

  useEffect(() => {
    EventsOn("notification:new", (payload) => {
      addNotification(payload);
    });
  }, []);

  useEffect(() => {
    const unlockAudio = () => {
      if (unlockedRef.current) return;

      audioRef.current
        ?.play()
        .then(() => {
          audioRef.current?.pause();
          audioRef.current.currentTime = 0;
          unlockedRef.current = true;
        })
        .catch(() => {});
    };

    document.addEventListener("click", unlockAudio, { once: true });
    return () => document.removeEventListener("click", unlockAudio);
  }, []);

  const buildId = () => {
    if (typeof crypto !== "undefined" && crypto.randomUUID) {
      return crypto.randomUUID();
    }
    return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  };

  const addNotification = (notification) => {
    const normalized = {
      id: notification.id || buildId(),
      title: notification.title || "Notificacao",
      message: notification.message || "",
      read: notification.read || false,
      createdAt: notification.createdAt || new Date().toISOString(),
    };

    setNotifications((prev) => [normalized, ...prev]);

    // Tocar som se o áudio estiver desbloqueado
    if (unlockedRef.current && audioRef.current) {
      audioRef.current.currentTime = 0;
      audioRef.current.play().catch(() => {});
    }
  };

  const markAllNotificationsRead = () => {
    setNotifications((prev) => prev.map((item) => ({ ...item, read: true })));
  };

  const clearNotifications = () => {
    setNotifications([]);
  };

  const unreadCount = notifications.filter((item) => !item.read).length;

  return (
    <AppContext.Provider
      value={{
        pixData,
        setPixData,
        notifications,
        addNotification,
        markAllNotificationsRead,
        clearNotifications,
        unreadCount,
      }}
    >
      {children}
    </AppContext.Provider>
  );
};
