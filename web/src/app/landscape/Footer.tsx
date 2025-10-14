import { BiliBiliIcon, IoLogoWechat } from '@/assets/SocialIcons';
import { EnvelopeIcon } from '@heroicons/react/24/solid';

import { RiZhihuLine } from 'react-icons/ri';

import Logo from '@/assets/Logo';
import type React from 'react';

const navigation = {
  about: [
    { name: 'About', href: '/about' },
    { name: 'Manifesto', href: '/manifesto' },
    { name: 'Join us', href: '/about#joinus' },
  ],
  project: [
    // { name: 'Pricing', href: '#' },
    { name: 'Protium', href: '#' },
    { name: 'Xyzen', href: '#' },
    { name: 'Anti', href: '#' },
    { name: 'LabOS', href: '#' },
  ],
  community: [
    { name: 'Articles', href: '/articles' },
    { name: 'Spaces', href: '/spaces' },
  ],
  legal: [
    { name: 'Claim', href: '#' },
    { name: 'Privacy', href: '#' },
    { name: 'Terms', href: '#' },
  ],
  social: [
    {
      name: 'Email',
      href: 'mailto:Protium@Protium.com',
      icon: (props: React.SVGProps<SVGSVGElement>) => (
        <EnvelopeIcon className="h-6 w-6" {...props} />
      ),
    },
    {
      name: 'GitHub',
      href: 'https://github.com/Protium',
      icon: (props: React.SVGProps<SVGSVGElement>) => (
        <svg fill="currentColor" viewBox="0 0 24 24" {...props}>
          <path
            fillRule="evenodd"
            d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
            clipRule="evenodd"
          />
        </svg>
      ),
    },
    {
      name: 'Twitter',
      href: 'https://twitter.com/Protium',
      icon: (props: React.SVGProps<SVGSVGElement>) => (
        <svg fill="currentColor" viewBox="0 0 24 24" {...props}>
          <path d="M8.29 20.251c7.547 0 11.675-6.253 11.675-11.675 0-.178 0-.355-.012-.53A8.348 8.348 0 0022 5.92a8.19 8.19 0 01-2.357.646 4.118 4.118 0 001.804-2.27 8.224 8.224 0 01-2.605.996 4.107 4.107 0 00-6.993 3.743 11.65 11.65 0 01-8.457-4.287 4.106 4.106 0 001.27 5.477A4.072 4.072 0 012.8 9.713v.052a4.105 4.105 0 003.292 4.022 4.095 4.095 0 01-1.853.07 4.108 4.108 0 003.834 2.85A8.233 8.233 0 012 18.407a11.616 11.616 0 006.29 1.84" />
        </svg>
      ),
    },
    {
      name: 'WeChat',
      href: 'https://open.weixin.qq.com/qr/code?username=deepmd',
      icon: (props: React.SVGProps<SVGSVGElement>) => (
        <IoLogoWechat className="h-6 w-6" {...props} />
      ),
    },

    {
      name: 'Zhihu',
      href: 'https://www.zhihu.com/org/shen-du-shi-neng-deep-potential',
      icon: (props: React.SVGProps<SVGSVGElement>) => (
        <RiZhihuLine className="h-6 w-6" {...props} />
      ),
    },
    {
      name: 'Bilibili',
      href: 'https://space.bilibili.com/626179751/',
      icon: (props: React.SVGProps<SVGSVGElement>) => (
        <BiliBiliIcon className="h-6 w-6" {...props} />
      ),
    },
  ],
};

export default function Footer() {
  return (
    <footer
      className="bg-white dark:bg-neutral-900"
      aria-labelledby="footer-heading"
    >
      <h2 id="footer-heading" className="sr-only">
        Footer
      </h2>
      <div className="mx-auto max-w-8xl px-6 pb-8 pt-16 sm:pt-24 lg:px-8 lg:pt-32 2xl:max-w-screen-2xl">
        <div className="xl:grid xl:grid-cols-3 xl:gap-8">
          <Logo className="h-8 w-8 fill-indigo-800 dark:fill-white" />
          <div className="mt-16 grid grid-cols-2 gap-8 xl:col-span-2 xl:mt-0">
            <div className="md:grid md:grid-cols-2 md:gap-8">
              <div>
                <h3 className="text-sm font-semibold leading-6 text-neutral-900 dark:text-neutral-50">
                  About
                </h3>
                <ul role="list" className="mt-6 space-y-4">
                  {navigation.about.map((item) => (
                    <li key={item.name}>
                      <a
                        href={item.href}
                        className="text-sm leading-6 text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-50"
                      >
                        {item.name}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
              <div className="mt-10 md:mt-0">
                <h3 className="text-sm font-semibold leading-6 text-neutral-900 dark:text-neutral-50">
                  Project
                </h3>
                <ul role="list" className="mt-6 space-y-4">
                  {navigation.project.map((item) => (
                    <li key={item.name}>
                      <a
                        href={item.href}
                        className="text-sm leading-6 text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-50"
                      >
                        {item.name}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
            <div className="md:grid md:grid-cols-2 md:gap-8">
              <div>
                <h3 className="text-sm font-semibold leading-6 text-neutral-900 dark:text-neutral-50">
                  Community
                </h3>
                <ul role="list" className="mt-6 space-y-4">
                  {navigation.community.map((item) => (
                    <li key={item.name}>
                      <a
                        href={item.href}
                        className="text-sm leading-6 text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-50"
                      >
                        {item.name}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
              <div className="mt-10 md:mt-0">
                <h3 className="text-sm font-semibold leading-6 text-neutral-900 dark:text-neutral-50">
                  Legal
                </h3>
                <ul role="list" className="mt-6 space-y-4">
                  {navigation.legal.map((item) => (
                    <li key={item.name}>
                      <a
                        href={item.href}
                        className="text-sm leading-6 text-neutral-600 hover:text-neutral-900 dark:text-neutral-400 dark:hover:text-neutral-50"
                      >
                        {item.name}
                      </a>
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </div>
        </div>
        <div className="mt-16 border-t border-neutral-900/10 pt-8 dark:border-neutral-50/10 sm:mt-20 lg:mt-24 lg:flex lg:items-center lg:justify-between">
          <div>
            <h3 className="text-sm font-semibold leading-6 text-neutral-900 dark:text-neutral-50">
              Subscribe to our Newsletter
            </h3>
            <p className="mt-2 text-sm leading-6 text-neutral-600 dark:text-neutral-400">
              The latest news, articles, and resources, sent to your inbox
              weekly.
            </p>
          </div>
          <form className="mt-6 sm:flex sm:max-w-md lg:mt-0">
            <label htmlFor="email-address" className="sr-only">
              Email address
            </label>
            <input
              type="email"
              name="email-address"
              id="email-address"
              autoComplete="email"
              required
              className="w-full min-w-0 appearance-none rounded-md border-0 bg-white px-3 py-1.5 text-base text-neutral-900 shadow-sm ring-1 ring-inset ring-neutral-300 placeholder:text-neutral-400 focus:ring-2 focus:ring-inset focus:ring-indigo-600 dark:bg-neutral-900 dark:text-white dark:ring-neutral-600 sm:w-56 sm:text-sm sm:leading-6"
              placeholder="Enter your email"
            />
            <div className="mt-4 sm:ml-4 sm:mt-0 sm:flex-shrink-0">
              <button
                type="submit"
                className="flex w-full items-center justify-center rounded-md bg-indigo-600 px-3 py-2 text-sm font-semibold
                  text-white shadow-sm hover:bg-indigo-500 focus-visible:outline focus-visible:outline-offset-2 focus-visible:outline-indigo-600"
              >
                Subscribe
              </button>
            </div>
          </form>
        </div>
        <div className="mt-8 border-t border-neutral-900/10 pt-8 dark:border-neutral-50/10 md:flex md:items-center md:justify-between">
          <div className="flex space-x-6 md:order-2">
            {navigation.social.map((item) => (
              <a
                key={item.name}
                target="_blank"
                href={item.href}
                className="fill-neutral-400 text-neutral-400 hover:fill-neutral-500 hover:text-neutral-500
                 dark:hover:fill-neutral-300 dark:hover:text-neutral-300"
              >
                <span className="sr-only">{item.name}</span>
                <item.icon className="h-6 w-6" aria-hidden="true" />
              </a>
            ))}
          </div>
          <p className="mt-8 text-xs leading-5 text-neutral-500 md:order-1 md:mt-0">
            &copy; {new Date().getFullYear()} Protium. All rights reserved.
            <a href="https://beian.miit.gov.cn" className="ml-2  text-[11px]">
              浙ICP备2022009106号-3
            </a>
          </p>
        </div>
      </div>
    </footer>
  );
}
