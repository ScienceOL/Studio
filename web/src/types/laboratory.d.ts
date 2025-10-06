// 实验室设备动作类型
export interface LaboratoryDeviceAction {
  id: number;
  key: string;
  title?: string;
  description?: string;
  goal?: any;
  feedback?: any;
  result?: any;
}

// 实验室设备分类类型
export interface LaboratoryDeviceClass {
  id: number;
  module: string;
  type?: string;
  status_types?: any;
  action_type_mappings?: any;
  actions: LaboratoryDeviceAction[];
}

// 实验室注册表类型
export interface LaboratoryRegistry {
  id: number;
  key: string;
  lab_name: string;
  lab_uuid: string;
  class: LaboratoryDeviceClass;
}
