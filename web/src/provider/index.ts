import {
  QueryClient,
  QueryClientProvider,
} from '@tanstack/react-query'

import AuthProvider from './AuthProvider';


export default function Providers({ children }: { children: React.ReactNode }) {
  const queryClient = new QueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        {children}
      </AuthProvider>
    </QueryClientProvider>
  )
}
