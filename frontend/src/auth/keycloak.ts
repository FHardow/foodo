import Keycloak from 'keycloak-js'

declare global {
  interface Window {
    __e2eRoles?: string[]
  }
}

const keycloak: Keycloak =
  import.meta.env.VITE_E2E_TEST === 'true'
    ? ({
        init: () => Promise.resolve(true),
        token: 'e2e-mock-token',
        subject: 'e2e-mock-user',
        tokenParsed: { sub: 'e2e-mock-user', name: 'E2E User' },
        hasRealmRole: (role: string) =>
          (window.__e2eRoles ?? []).includes(role),
        onTokenExpired: undefined,
        updateToken: () => Promise.resolve(true),
        login: () => {},
        logout: () => {},
      } as unknown as Keycloak)
    : new Keycloak({
        url: import.meta.env.VITE_KEYCLOAK_URL,
        realm: import.meta.env.VITE_KEYCLOAK_REALM,
        clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID,
      })

export default keycloak
