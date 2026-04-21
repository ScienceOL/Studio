import { useTranslation } from 'react-i18next';

const links = [
  { nameKey: null, name: 'OpenSDL', href: 'https://github.com/ScienceOL/OpenSDL' },
  { nameKey: null, name: 'Xyzen', href: 'https://xyzen.cc' },
  { nameKey: null, name: 'GitHub', href: 'https://github.com/ScienceOL' },
  { nameKey: 'landing.footer.about', name: 'About', href: '/about' },
];

export default function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="bg-black border-t border-white/[0.06]">
      <div className="mx-auto max-w-screen-2xl px-6 py-10 lg:px-8">
        <div className="flex flex-col items-center gap-6 sm:flex-row sm:justify-between">
          <div className="flex items-center gap-6">
            <img
              src="https://storage.sciol.ac.cn/library/BLogo-dark.svg"
              alt="ScienceOL"
              className="h-5 w-5 opacity-40"
            />
            <div className="flex flex-wrap items-center gap-4">
              {links.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  target={link.href.startsWith('http') ? '_blank' : undefined}
                  className="text-xs text-white/30 transition-colors hover:text-white/60 font-mono"
                >
                  {link.nameKey ? t(link.nameKey) : link.name}
                </a>
              ))}
            </div>
          </div>

          <p className="text-[11px] text-white/20 font-mono text-center sm:text-right">
            &copy; {new Date().getFullYear()} 奇迹物语（上海）智能科技有限公司
            <a href="https://beian.miit.gov.cn" className="ml-2 hover:text-white/40">
              沪ICP备2026012508号
            </a>
          </p>
        </div>
      </div>
    </footer>
  );
}
