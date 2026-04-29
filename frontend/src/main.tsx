import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App'
import Landing from './pages/Landing'
import keycloak from './auth/keycloak'

keycloak
  .init({ onLoad: 'check-sso', pkceMethod: 'S256', checkLoginIframe: false })
  .then((authenticated) => {
    if (authenticated) {
      keycloak.onTokenExpired = () => {
        keycloak.updateToken(30).catch(() => keycloak.login())
      }
    }

    createRoot(document.getElementById('root')!).render(
      <StrictMode>
        {authenticated ? <App /> : <Landing />}
      </StrictMode>,
    )
  })
  .catch(() => {
    document.getElementById('root')!.textContent = 'Failed to connect to authentication server.'
  })
