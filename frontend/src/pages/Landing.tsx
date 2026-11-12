import keycloak from '../auth/keycloak'
import { Button } from '../components/ui/button'

export default function Landing() {
  return (
    <div className="min-h-screen bg-background flex flex-col items-center justify-center gap-8">
      <h1 className="text-4xl font-bold text-foreground">Foodo</h1>
      <div className="flex gap-4">
        <Button onClick={() => keycloak.login()}>Login</Button>
        <Button variant="outline" onClick={() => keycloak.register()}>
          Register
        </Button>
      </div>
    </div>
  )
}
