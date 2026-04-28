import { ArrowLeft, Loader2 } from 'lucide-react';
import { SiGithub, SiApple } from 'react-icons/si';
import { motion } from 'motion/react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router-dom';

import { AuthUtils } from '@/utils/auth';
import { config } from '@/configs';

function getLoginUrl(provider?: string): string {
  const frontendCallbackURL = `${config.frontendBaseUrl}/login/callback`;
  let url = `${config.apiBaseUrl}/api/auth/login?frontend_callback_url=${encodeURIComponent(frontendCallbackURL)}`;
  if (provider) {
    url += `&provider=${encodeURIComponent(provider)}`;
  }
  return url;
}

export default function LoginPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [loading, setLoading] = useState<string | null>(null);

  useEffect(() => {
    if (AuthUtils.isAuthenticated()) {
      navigate('/', { replace: true });
    }
  }, [navigate]);

  const handleLogin = (provider?: string) => {
    setLoading(provider ?? 'default');
    window.location.href = getLoginUrl(provider);
  };

  const providers = [
    {
      id: 'provider_apple',
      icon: <SiApple className="absolute left-5 h-5 w-5" />,
      label: t('login.apple'),
    },
    {
      id: 'provider_github',
      icon: <SiGithub className="absolute left-5 h-5 w-5" />,
      label: t('login.github'),
    },
    {
      id: 'provider_google',
      icon: (
        <svg className="absolute left-5 h-5 w-5" viewBox="0 0 24 24">
          <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" />
          <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" />
          <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" />
          <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" />
        </svg>
      ),
      label: t('login.google'),
    },
    {
      id: 'provider_bohrium',
      icon: (
        <img
          src="/defaults/icons/bohrium.png"
          alt="Bohrium"
          className="absolute left-5 h-5 w-5 rounded"
        />
      ),
      label: t('login.bohrium'),
    },
  ];

  return (
    <div className="relative min-h-screen overflow-hidden bg-gradient-to-b from-black via-[#05060e] to-[#0a0b16]">
      {/* Back button */}
      <button
        type="button"
        onClick={() => navigate('/')}
        className="absolute left-5 top-5 z-10 flex h-9 w-9 items-center justify-center rounded-lg text-neutral-600 transition-colors hover:bg-white/[0.06] hover:text-neutral-400"
      >
        <ArrowLeft className="h-[18px] w-[18px]" />
      </button>

      <div className="flex min-h-screen w-full items-center justify-center px-8 md:px-4">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: [0.16, 1, 0.3, 1] }}
          className="w-full max-w-[420px]"
        >
          {/* Logo */}
          <img
            src="https://storage.sciol.ac.cn/library/BLogo-dark.svg"
            alt="ScienceOL"
            className="mb-6 h-12 w-12"
          />

          {/* Tagline */}
          <h1 className="text-[28px] font-bold leading-[1.2] tracking-tight text-white md:text-[32px]">
            {t('login.tagline')}
          </h1>
          <p className="mt-4 text-[15px] text-neutral-500">
            {t('login.subtitle')}
          </p>

          {/* OAuth Buttons */}
          <div className="mt-8 flex flex-col gap-3.5">
            {providers.map((provider) => (
              <button
                key={provider.id}
                type="button"
                disabled={!!loading}
                onClick={() =>
                  handleLogin(
                    provider.id === 'default' ? undefined : provider.id
                  )
                }
                className="relative flex h-[52px] w-full items-center rounded-lg bg-white/[0.07] text-[14px] font-medium text-white transition-colors hover:bg-white/[0.13] active:bg-white/[0.05] disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {loading === provider.id ? (
                  <Loader2 className="absolute left-5 h-5 w-5 animate-spin" />
                ) : (
                  provider.icon
                )}
                <span className="flex-1 text-center">{provider.label}</span>
              </button>
            ))}
          </div>

          {/* Terms */}
          <p className="mt-8 text-xs leading-relaxed text-neutral-600">
            {t('login.terms')}
          </p>
        </motion.div>
      </div>
    </div>
  );
}
