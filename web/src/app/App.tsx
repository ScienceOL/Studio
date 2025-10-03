import { useXyzen, Xyzen } from '@sciol/xyzen';
import { useNavigate } from 'react-router-dom';

export default function App() {
  const { panelWidth } = useXyzen();
  const navigate = useNavigate();

  const handleGoLogin = () => {
    navigate('/login');
  };

  return (
    <main className="flex h-full">
      <div
        style={{ width: `calc(100% - ${panelWidth}px)` }}
        className="h-full mt-20 flex flex-col items-center justify-center gap-6"
      >
        <div className="font-bold bg-black text-white p-4 rounded">
          Studio 简易主页
        </div>
        <button
          className="px-4 py-2 bg-blue-600 text-white rounded shadow"
          onClick={handleGoLogin}
        >
          跳转到登录页
        </button>
      </div>
      <Xyzen />
    </main>
  );
}
