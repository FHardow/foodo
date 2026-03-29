import keycloak from '../auth/keycloak'

export default function Landing() {
  return (
    <div className="min-h-screen bg-[#faf7f2] flex flex-col items-center justify-center gap-8">
      <h1 className="text-4xl font-bold text-[#3b2a1a]">Bread Order</h1>
      <div className="flex gap-4">
        <button
          onClick={() => keycloak.login()}
          className="px-6 py-2 rounded-lg bg-[#3b2a1a] text-white font-medium hover:bg-[#5a3e28] transition-colors"
        >
          Login
        </button>
        <button
          onClick={() => keycloak.register()}
          className="px-6 py-2 rounded-lg border border-[#3b2a1a] text-[#3b2a1a] font-medium hover:bg-[#f0e8dc] transition-colors"
        >
          Register
        </button>
      </div>
    </div>
  )
}
