// src/types/global.d.ts
declare global {
  // 通过引用模块类型生成全局类型
  type NodeTypes =
    import('@/app/(dashboard)/workflow/nodes/nodeTypes').NodeTypes;
}

// 确保文件作为模块，添加 export{} 的作用是
// 使 `d.ts` 不把不在 declare global 的代码当作全局代码
// 防止污染全局命名空间
// export {};
