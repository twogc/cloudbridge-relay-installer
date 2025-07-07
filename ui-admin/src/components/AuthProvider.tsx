import React, { createContext, useContext, useEffect, useState } from 'react';
import { userManager, getUser, hasAdminRole } from '../config/keycloak';
import type { User } from '../config/keycloak';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isLoading: boolean;
  login: () => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: React.ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const isAuthenticated = user !== null;
  const isAdmin = hasAdminRole(user);

  const login = () => {
    userManager.signinRedirect();
  };

  const logout = () => {
    userManager.signoutRedirect();
  };

  useEffect(() => {
    const handleUserLoaded = async () => {
      try {
        const currentUser = await getUser();
        setUser(currentUser);
      } catch (error) {
        console.error('Error loading user:', error);
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    const handleUserUnloaded = () => {
      setUser(null);
    };

    const handleSilentRenewError = (error: Error) => {
      console.error('Silent renew error:', error);
      setUser(null);
    };

    const handleAccessTokenExpired = () => {
      console.log('Access token expired');
      setUser(null);
    };

    const handleUserSignedOut = () => {
      console.log('User signed out');
      setUser(null);
    };

    // Подписываемся на события
    userManager.events.addUserLoaded(handleUserLoaded);
    userManager.events.addUserUnloaded(handleUserUnloaded);
    userManager.events.addSilentRenewError(handleSilentRenewError);
    userManager.events.addAccessTokenExpired(handleAccessTokenExpired);
    userManager.events.addUserSignedOut(handleUserSignedOut);

    // Проверяем, есть ли уже пользователь
    userManager.getUser().then((user) => {
      if (user && !user.expired) {
        getUser().then(setUser);
      }
      setIsLoading(false);
    });

    return () => {
      userManager.events.removeUserLoaded(handleUserLoaded);
      userManager.events.removeUserUnloaded(handleUserUnloaded);
      userManager.events.removeSilentRenewError(handleSilentRenewError);
      userManager.events.removeAccessTokenExpired(handleAccessTokenExpired);
      userManager.events.removeUserSignedOut(handleUserSignedOut);
    };
  }, []);

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isAdmin,
    isLoading,
    login,
    logout,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}; 