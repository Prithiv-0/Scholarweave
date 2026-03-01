import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User, apiService, authStorage } from '@/api/client';

interface AuthContextType {
    user: User | null;
    token: string | null;
    isLoading: boolean;
    login: (token: string, user: User) => void;
    logout: () => void;
    refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        // Initialize from local storage
        const storedToken = authStorage.getToken();
        if (storedToken) {
            setToken(storedToken);
            // Fetch fresh user profile
            refreshUser().finally(() => setIsLoading(false));
        } else {
            setIsLoading(false);
        }
    }, []);

    const login = (newToken: string, newUser: User) => {
        authStorage.setToken(newToken);
        setToken(newToken);
        setUser(newUser);
    };

    const logout = () => {
        authStorage.clearToken();
        setToken(null);
        setUser(null);
    };

    const refreshUser = async () => {
        try {
            const freshUser = await apiService.me();
            setUser(freshUser);
        } catch (err) {
            console.error('Failed to fetch profile', err);
            logout(); // clear invalid token
        }
    };

    return (
        <AuthContext.Provider value={{ user, token, isLoading, login, logout, refreshUser }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = (): AuthContextType => {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
};
