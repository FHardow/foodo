import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App'
import keycloak from './auth/keycloak'

keycloak
  .init({ onLoad: 'login-required', pkceMethod: 'S256', checkLoginIframe: false })
  .then((authenticated) => {
    if (!authenticated) {
      keycloak.login()
      return
    }

    keycloak.onTokenExpired = () => {
      keycloak.updateToken(30).catch(() => keycloak.login())
    }

    createRoot(document.getElementById('root')!).render(
      <StrictMode>
        <App />
      </StrictMode>,
    )
  })
  .catch(() => {
    document.getElementById('root')!.textContent = 'Failed to connect to authentication server.'
  })
