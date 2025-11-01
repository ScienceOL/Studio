/**
 * 统一分页配置
 */

export const PAGINATION_CONFIG = {
  // 默认每页数量
  DEFAULT_PAGE_SIZE: 20,

  // 默认页码
  DEFAULT_PAGE: 1,

  // 最大每页数量
  MAX_PAGE_SIZE: 2000,

  // 页码选项
  PAGE_SIZE_OPTIONS: [10, 20, 50, 100, 200],
} as const;

/**
 * 分页参数类型
 */
export interface PaginationParams {
  page?: number;
  page_size?: number;
}

/**
 * 规范化分页参数
 * 确保前端发送的参数始终有默认值
 */
export function normalizePaginationParams(
  params?: PaginationParams
): Required<PaginationParams> {
  return {
    page: params?.page || PAGINATION_CONFIG.DEFAULT_PAGE,
    page_size: params?.page_size || PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
  };
}

/**
 * 分页响应数据类型
 */
export interface PaginationResponse<T> {
  data: T[];
  page: number;
  page_size: number;
  has_more?: boolean;
  total?: number;
}
