/**
 * Environment 模块统一导出
 */

// 页面组件
export { default as EnvironmentDetail } from './EnvironmentDetail';
export { default as EnvironmentPage } from './EnvironmentPage';

// Hooks
export * from '@/hooks/queries/useEnvironmentQueries';
export {
  useEnvironment,
  useLabInfo,
  useLabMembersList,
} from '@/hooks/useEnvironment';

// Core
export { EnvironmentCore } from '@/core/environmentCore';

// Store
export { useEnvironmentStore } from '@/store/environmentStore';

// Types
export type * from '@/types/environment';
