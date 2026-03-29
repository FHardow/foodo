import { BrowserRouter, Routes, Route } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { Toaster } from 'sonner'
import Nav from './components/Nav'
import Store from './pages/Store'
import Basket from './pages/Basket'
import OrderStatus from './pages/OrderStatus'
import OrderHistory from './pages/OrderHistory'
import AdminProducts from './pages/admin/Products'
import AdminOrders from './pages/admin/Orders'
import Product from './pages/Product'

const queryClient = new QueryClient()

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <div className="min-h-screen bg-[#faf7f2]">
          <Nav />
          <main className="max-w-5xl mx-auto px-4 py-6">
            <Routes>
              <Route path="/" element={<Store />} />
              <Route path="/basket" element={<Basket />} />
              <Route path="/orders/:id" element={<OrderStatus />} />
              <Route path="/orders" element={<OrderHistory />} />
              <Route path="/products/:id" element={<Product />} />
              <Route path="/admin/products" element={<AdminProducts />} />
              <Route path="/admin/orders" element={<AdminOrders />} />
            </Routes>
          </main>
        </div>
        <Toaster richColors />
      </BrowserRouter>
    </QueryClientProvider>
  )
}
