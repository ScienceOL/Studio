import { AuthUtils } from '@/utils/auth';
import { CheckCircle2, Loader2, XCircle } from 'lucide-react';
import { motion } from 'motion/react';
import { useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate, useSearchParams } from 'react-router-dom';

function getCookie(name: string): string | null {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    const cookieValue = parts.pop()?.split(';').shift() || null;
    return cookieValue ? decodeURIComponent(cookieValue) : null;
  }
  return null;
}

export default function LoginCallback() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [searchParams] = useSearchParams();
  const [status, setStatus] = useState<'processing' | 'success' | 'error'>(
    'processing'
  );
  const [message, setMessage] = useState('');
  const hasProcessed = useRef(false);

  useEffect(() => {
    if (hasProcessed.current) return;
    hasProcessed.current = true;

    const handleCallback = async () => {
      try {
        const error = searchParams.get('error');
        const callbackStatus = searchParams.get('status');

        if (error) {
          const errorMsg = decodeURIComponent(error);
          setStatus('error');
          if (errorMsg.includes('state verification failed')) {
            setMessage(t('login.callback.stateError'));
          } else {
            setMessage(t('login.callback.error', { error: errorMsg }));
          }
          return;
        }

        if (callbackStatus !== 'success') {
          setStatus('error');
          setMessage(t('login.callback.statusError'));
          return;
        }

        const token = getCookie('access_token');
        const refreshToken = getCookie('refresh_token');
        const userInfoEncoded = getCookie('user_info');

        if (!token || !refreshToken) {
          setStatus('error');
          setMessage(t('login.callback.tokenError'));
          return;
        }

        let userInfo = null;
        if (userInfoEncoded) {
          try {
            const userJSON = atob(userInfoEncoded);
            userInfo = JSON.parse(userJSON);
          } catch {
            console.error('Failed to parse user info from cookie');
          }
        }

        AuthUtils.saveAuthInfo(
          {
            accessToken: token,
            refreshToken: refreshToken,
            expiresIn: 3600,
            tokenType: 'Bearer',
          },
          userInfo
        );

        setStatus('success');
        setMessage(t('login.callback.success'));

        const returnUrl = sessionStorage.getItem('login_return_url');
        const from =
          returnUrl || (location.state as { from?: string })?.from || '/';
        sessionStorage.removeItem('login_return_url');

        setTimeout(() => {
          navigate(from, { replace: true });
        }, 1200);
      } catch {
        setStatus('error');
        setMessage(t('login.callback.unknownError'));
      }
    };

    handleCallback();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleRetry = () => {
    AuthUtils.redirectToLogin();
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-b from-black via-[#05060e] to-[#0a0b16] px-8">
      <motion.div
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.4, ease: [0.16, 1, 0.3, 1] }}
        className="w-full max-w-sm text-center"
      >
        {status === 'processing' && (
          <div className="flex flex-col items-center gap-4">
            <Loader2 className="h-8 w-8 animate-spin text-indigo-400" />
            <p className="text-[14px] text-neutral-400">
              {t('login.callback.processing')}
            </p>
          </div>
        )}

        {status === 'success' && (
          <div className="flex flex-col items-center gap-4">
            <CheckCircle2 className="h-8 w-8 text-emerald-400" />
            <p className="text-[14px] font-medium text-emerald-400">
              {message}
            </p>
          </div>
        )}

        {status === 'error' && (
          <div className="flex flex-col items-center gap-4">
            <XCircle className="h-8 w-8 text-red-400" />
            <p className="text-[14px] text-red-400">{message}</p>
            <div className="mt-2 flex gap-3">
              <button
                onClick={handleRetry}
                className="rounded-lg bg-indigo-500 px-5 py-2.5 text-[13px] font-semibold text-white transition-colors hover:bg-indigo-400"
              >
                {t('login.callback.retry')}
              </button>
              <button
                onClick={() => navigate('/')}
                className="rounded-lg bg-white/[0.06] px-5 py-2.5 text-[13px] font-medium text-white/70 transition-colors hover:bg-white/[0.1]"
              >
                {t('login.callback.home')}
              </button>
            </div>
          </div>
        )}
      </motion.div>
    </div>
  );
}
