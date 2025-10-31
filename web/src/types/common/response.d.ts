/**
 * Unified API response types matching backend (Go) structures in service/pkg/common/response.go
 */

// Backend ErrCode is an int. Keep as number on the frontend.
type ErrCode = number;

interface RespError {
  msg: string;
  info?: string[];
}

// Generic response wrapper
interface Resp<T = unknown> {
  code: ErrCode;
  error?: RespError;
  data?: T;
  timestamp?: number; // unix seconds
}

// Pagination request
interface PageReq {
  page: number;
  page_size: number;
}

interface PageReqT<T = unknown> extends PageReq {
  data: T;
}

// Pagination responses
interface PageResp<T = unknown> {
  total: number;
  page: number;
  page_size: number;
  data: T;
}

interface PageMoreResp<T = unknown> {
  has_more: boolean;
  page: number;
  page_size: number;
  data: T;
}

// WebSocket payloads
interface WsMsgType {
  action: string;
  msg_uuid: string; // UUID string
}

interface WSData<T = unknown> extends WsMsgType {
  data?: T;
}

// WebSocket response = Resp with WSData inside data
type WSResp<T = unknown> = Resp<WSData<T>>;
