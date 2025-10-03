import { BrowserRouter, Route, Routes } from 'react-router-dom';
import App from './app/App';
import CallbackPage from './app/login/CallbackPage';
import LoginPage from './app/login/LoginPage';

export default function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<App />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/login/callback" element={<CallbackPage />} />
      </Routes>
    </BrowserRouter>
  );
}
