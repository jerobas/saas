import { createContext, useState } from 'react';

export const AppContext = createContext();

export const AppProvider = ({ children }) => {
    const [pixData, setPixData] = useState(null);

    return (
        <AppContext.Provider value={{ pixData, setPixData }}>
            {children}
        </AppContext.Provider>
    );
};