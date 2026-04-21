// 设备信息配置
export interface DeviceInfo {
  id: string;
  name: string;
  nameEn: string;
  description: string;
  specs?: string[];
  usage?: string;
}

export const DEVICE_INFO: Record<string, DeviceInfo> = {
  'liquid-handler': {
    id: 'liquid-handler',
    name: '自动移液工作站',
    nameEn: 'Liquid Handler',
    description: '高精度自动化液体处理系统，用于样品分配、稀释和混合操作',
    specs: ['精度：±1.5%', '工作范围：0.5-1000μL', '96/384孔板兼容'],
    usage: '用于高通量样品制备、PCR反应体系配置、细胞培养等实验',
  },
  microscope: {
    id: 'microscope',
    name: '光学显微镜',
    nameEn: 'Microscope',
    description: '高分辨率光学显微镜，用于细胞观察和微观结构分析',
    specs: ['放大倍数：40x-1000x', '分辨率：0.2μm', '数字成像系统'],
    usage: '用于细胞形态观察、组织切片分析、微生物检测等',
  },
  monitor: {
    id: 'monitor',
    name: '工作站电脑',
    nameEn: 'Workstation',
    description: '实验室数据处理和设备控制工作站',
    specs: ['24英寸4K显示屏', '高性能处理器', '专业图形显卡'],
    usage: '用于实验数据分析、设备程序控制、结果可视化',
  },
  'agv-robot': {
    id: 'agv-robot',
    name: '智能AGV运输机器人',
    nameEn: 'AGV Robot',
    description: '配备6轴机械臂的自主移动机器人，实现实验室自动化物流',
    specs: ['载重：50kg', '精度：±2mm', '6自由度机械臂', '自主导航'],
    usage: '用于样品运输、耗材配送、设备间协作等自动化任务',
  },
  centrifuge: {
    id: 'centrifuge',
    name: '台式离心机',
    nameEn: 'Centrifuge',
    description: '高速台式离心机，用于样品分离和沉淀',
    specs: ['最高转速：15000 rpm', '容量：24×1.5ml', '温度控制：-10~40℃'],
    usage: '用于DNA/RNA提取、蛋白质纯化、细胞分离等',
  },
  'pipette-rack': {
    id: 'pipette-rack',
    name: '移液枪架',
    nameEn: 'Pipette Rack',
    description: '多通道移液枪存储架，配备不同量程移液枪',
    specs: ['容纳数量：4-8支', '量程：0.5-1000μL', '不锈钢材质'],
    usage: '用于存放和快速取用各种规格的移液枪',
  },
  beaker: {
    id: 'beaker',
    name: '烧杯',
    nameEn: 'Beaker',
    description: '标准实验室玻璃烧杯，用于溶液配置和反应',
    specs: ['容量：50-1000ml', '材质：硼硅玻璃', '耐温：-70~500℃'],
    usage: '用于溶液混合、加热反应、样品储存等',
  },
  'storage-cabinet': {
    id: 'storage-cabinet',
    name: '实验室储物柜',
    nameEn: 'Storage Cabinet',
    description: '标准实验室储物柜，用于存放试剂和耗材',
    specs: ['尺寸：90×180×60cm', '防腐蚀材质', '多层分隔'],
    usage: '用于存放化学试剂、实验耗材、个人防护用品等',
  },
  'reagent-rack': {
    id: 'reagent-rack',
    name: '试剂瓶架',
    nameEn: 'Reagent Rack',
    description: '多位试剂瓶存储架，整齐存放各类试剂',
    specs: ['容量：9-16瓶', '防腐蚀托盘', '标签系统'],
    usage: '用于分类存放和管理各种化学试剂、缓冲液等',
  },
  'reagent-bottle': {
    id: 'reagent-bottle',
    name: '试剂瓶',
    nameEn: 'Reagent Bottle',
    description: '标准实验室试剂瓶，密封保存各类试剂',
    specs: ['容量：50-1000ml', '材质：玻璃/塑料', '密封瓶盖'],
    usage: '用于存放和使用各种化学试剂、溶液',
  },
  'petri-dish': {
    id: 'petri-dish',
    name: '培养皿',
    nameEn: 'Petri Dish',
    description: '无菌塑料培养皿，用于微生物和细胞培养',
    specs: ['直径：90mm', '材质：聚苯乙烯', '灭菌处理'],
    usage: '用于细菌培养、细胞培养、菌落计数等',
  },
  'sample-rack': {
    id: 'sample-rack',
    name: '微孔板',
    nameEn: 'Microplate',
    description: '96孔标准微孔板，用于高通量实验',
    specs: ['规格：96/384孔', '材质：聚丙烯', '体积：50-300μL/孔'],
    usage: '用于ELISA、PCR、细胞培养等高通量实验',
  },
};

// 获取设备信息
export function getDeviceInfo(deviceId: string): DeviceInfo | null {
  return DEVICE_INFO[deviceId] || null;
}

// 获取所有可交互设备列表
export function getAllDeviceIds(): string[] {
  return Object.keys(DEVICE_INFO);
}
