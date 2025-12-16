import '@/i18n';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import Router from './router';

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Router />
  </StrictMode>
);
