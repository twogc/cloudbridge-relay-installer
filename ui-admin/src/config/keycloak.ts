import { UserManager } from 'oidc-client-ts';

// Конфигурация Keycloak для OIDC авторизации
export const keycloakConfig = {
  authority: 'https://edge.2gc.ru/realms/cloudbridge',
  client_id: 'relay-client',
  redirect_uri: window.location.origin + '/callback',
  response_type: 'code',
  scope: 'openid profile email roles',
  post_logout_redirect_uri: window.location.origin,
  silent_redirect_uri: window.location.origin + '/silent-renew.html',
  automaticSilentRenew: true,
  monitorSession: true,
  loadUserInfo: true,
};

// Создаем экземпляр UserManager
export const userManager = new UserManager(keycloakConfig);

// Типы для пользователя
export interface User {
  id: string;
  name: string;
  email: string;
  roles: string[];
}

// Функция для получения пользователя
export const getUser = async (): Promise<User | null> => {
  try {
    const user = await userManager.getUser();
    if (user && !user.expired) {

      
      // Get roles from realm_access (realm roles)
      const realmRoles = (user.profile as any).realm_access?.roles || [];
      // Get roles from resource_access (client roles)
      const resourceRoles = (user.profile as any).resource_access || {};
      const clientRoles = Object.values(resourceRoles).flatMap((resource: any) => resource.roles || []);
      
      const roles = [...realmRoles, ...clientRoles];
      
      return {
        id: user.profile.sub || '',
        name: user.profile.name || user.profile.preferred_username || '',
        email: user.profile.email || '',
        roles: roles,
      };
    }
    return null;
  } catch (error) {
    console.error('Error getting user:', error);
    return null;
  }
};

// Функция для проверки роли admin-ui
export const hasAdminRole = (user: User | null): boolean => {
  return user?.roles.includes('admin-ui') || false;
};

// Функция для получения access token
export const getAccessToken = async (): Promise<string | null> => {
  try {
    const user = await userManager.getUser();
    if (user && !user.expired) {
      return user.access_token;
    }
    return null;
  } catch (error) {
    console.error('Error getting access token:', error);
    return null;
  }
}; 