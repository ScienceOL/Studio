import { useEffect, useState } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { AuthUtils } from '@/lib/auth';

interface ProtectedRouteProps {
  children: React.ReactNode;
}

export default function ProtectedRoute({ children }: ProtectedRouteProps) {
  const [isChecking, setIsChecking] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const location = useLocation();

  useEffect(() => {
    const checkAuth = async () => {
      const authenticated = AuthUtils.isAuthenticated();
      
      if (!authenticated) {
        // 尝试刷新 token
        const refreshed = await AuthUtils.refreshToken();
        setIsAuthenticated(refreshed);
      } else {
        setIsAuthenticated(true);
      }
      
      setIsChecking(false);
    };

    checkAuth();
  }, []);

  if (isChecking) {
    return (
      <div className="min-h-screen bg-gray-50 dark:bg-gray-900 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600 dark:text-gray-400">正在验证身份...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    // 保存当前路径，登录后可以重定向回来
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
}
