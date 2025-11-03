// 环境（实验室）相关类型定义

export interface Lab {
  uuid: string;
  name: string;
  user_id: string;
  is_admin: boolean;
  access_key: string;
  access_secret: string;
  status: string;
  created_at: string;
  updated_at: string;
  description?: string;
  owner_uuid?: string;
  [key: string]: unknown;
}

export interface CreateLabRequest {
  name: string;
  description?: string;
  [key: string]: unknown;
}

export interface UpdateLabRequest {
  uuid: string;
  name?: string;
  description?: string;
  [key: string]: unknown;
}

export interface LabMember {
  uuid: string;
  user_id: string;
  lab_id: number;
  role: string;
  is_admin: boolean;
  lab_uuid?: string;
  user_uuid?: string;
  username?: string;
  created_at?: string;
}

export interface CreateInviteRequest {
  expires_at?: string;
  role?: string;
  [key: string]: unknown;
}

export interface InviteInfo {
  uuid: string;
  lab_uuid: string;
  lab_name?: string;
  inviter_uuid: string;
  inviter_name?: string;
  role: string;
  expires_at: string;
  created_at: string;
}

export interface UserInfo {
  uuid: string;
  username: string;
  email: string;
  avatar?: string;
  created_at: string;
  [key: string]: unknown;
}

export interface LabResource {
  uuid: string;
  lab_uuid: string;
  name: string;
  type: string;
  status: string;
  [key: string]: unknown;
}

// 实验室成员响应（对应后端 LabMemberResp）
export interface LabMemberResponse {
  uuid: string;
  user_id: string;
  lab_id: number;
  role: string;
  is_admin: boolean;
  [key: string]: unknown;
}
