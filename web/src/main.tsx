import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter, Routes, Route } from 'react-router-dom'
import App from './app/App'
import LoginPage from './app/login/LoginPage'
import CallbackPage from './app/login/CallbackPage'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/login/callback" element={<CallbackPage />} />
      </Routes>
    </BrowserRouter>
  </StrictMode>,
)
