import { useAuth } from '@/hooks/useAuth';

export function UserInfo() {
  const { user, logout, isLoading } = useAuth();
  
  if (isLoading) {
    return <div className="animate-pulse">加载中...</div>;
  }
  
  if (!user) return null;
  
  return (
    <div className="flex items-center gap-2">
      {user.avatar && (
        <img 
          src={user.avatar} 
          alt="用户头像" 
          className="w-8 h-8 rounded-full"
        />
      )}
      <span>{user.displayName || user.email}</span>
      <button 
        className="px-2 py-1 bg-red-500 text-white rounded hover:bg-red-600" 
        onClick={logout}
      >
        退出
      </button>
    </div>
  );
}

export function LoginButton() {
  const { login, isAuthenticated, isLoading } = useAuth();
  
  if (isLoading) {
    return <div className="animate-pulse">加载中...</div>;
  }
  
  if (isAuthenticated) {
    return <UserInfo />;
  }
  
  return (
    <button 
      className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      onClick={login}
    >
      登录
    </button>
  );
}
