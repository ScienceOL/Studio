import i18n from 'i18next';
import LanguageDetector from 'i18next-browser-languagedetector';
import { initReactI18next } from 'react-i18next';

import enAboutJson from './en-us/about.json';
import enDashboardJson from './en-us/dashboard.json';
import enHomepageJson from './en-us/homepage.json';
import enLoginJson from './en-us/login.json';
import enManifestoJson from './en-us/manifesto.json';

import zhAboutJson from './zh-cn/about.json';
import zhDashboardJson from './zh-cn/dashboard.json';
import zhHomepageJson from './zh-cn/homepage.json';
import zhLoginJson from './zh-cn/login.json';
import zhManifestoJson from './zh-cn/manifesto.json';

import zhUserPanelJson from './zh-cn/userPanel/userPanel.json';

const resources = {
  en: {
    translation: {
      ...enHomepageJson,
      ...enDashboardJson,
    },

    login: {
      ...enLoginJson,
    },
    about: {
      ...enAboutJson,
    },
    manifesto: {
      ...enManifestoJson,
    },
    environment: {
      details: {
        hero: {
          readDoc: 'Read Docs',
          viewGithub: 'View Github',
        },
        navbar: {
          Overview: 'Overview',
          Resources: 'Resources',
          Models: 'Models',
          Solvent: 'Solvent',
          Discussion: 'Discussion',
          Releases: 'Releases',
          Insights: 'Insights',
          Benchmark: 'Benchmark',
          Setting: 'Setting',
          Registry: 'Registry',
        },
        overview: {
          release: 'Release',
          pinnedArticle: 'Pinned Articles',
        },
      },
    },
  },
  zh: {
    translation: {
      ...zhHomepageJson,
      ...zhDashboardJson,
    },
    login: {
      ...zhLoginJson,
    },
    userPanel: {
      ...zhUserPanelJson,
    },
    about: {
      ...zhAboutJson,
    },
    manifesto: {
      ...zhManifestoJson,
    },
    environment: {
      details: {
        hero: {
          readDoc: '阅读文档',
          viewGithub: '查看 Github',
        },
        navbar: {
          Overview: '概览',
          Resources: '资源',
          Models: '模型',
          Solvent: '溶剂',
          Discussion: '讨论',
          Releases: '发布',
          Insights: '数据',
          Benchmark: '基准测试',
          Setting: '设置',
          Registry: '注册表',
        },
        overview: {
          release: '发布',
          pinnedArticle: '置顶文章',
        },
      },
    },
  },
};

i18n
  .use(LanguageDetector)
  .use(initReactI18next) // passes i18n down to react-i18next
  .init({
    resources,
    debug: import.meta.env.NODE_ENV === 'development',
    defaultNS: 'translation',
    fallbackLng: 'en',

    interpolation: {
      escapeValue: false, // react already safes from xss
    },
  });

export default i18n;
