#!/bin/bash

# 打印脚本作用和警告信息
cat <<EOF
=== 分支清理脚本 ===
作用：清理本地已失效的 Git 分支（远程已删除且超过 7 天未更新）

警告：
1. 该脚本会永久删除分支，操作不可逆
2. 只清理同时满足以下条件的分支：
   - 远程仓库已删除
   - 最后一次提交超过 7 天
   - 排除当前所在分支
3. 建议先执行 git fetch 确保分支状态最新

EOF

# 确认继续执行
read -p "是否继续？[y/N] " confirm
if [[ ! "$confirm" =~ [yY] ]]; then
    echo "操作已取消"
    exit 0
fi

# 获取远程最新状态
git fetch --prune

# 计算 7 天前的时间戳（兼容 macOS/Linux）
seven_days_ago=$(date -d "7 days ago" +%s 2>/dev/null || date -v-7d +%s)

# 收集需要删除的分支
to_delete=()
while read -r branch; do
    # 跳过空行和当前分支
    if [[ -z "$branch" || "$branch" == *"*"* ]]; then continue; fi

    branch_name=$(echo "$branch" | awk '{print $1}')
    
    # 检查是否远程分支已删除 ([gone])
    if ! echo "$branch" | grep -q '\[gone\]'; then
        continue
    fi
    
    # 获取最后提交时间戳
    last_commit=$(git log -1 --format="%at" "$branch_name")
    
    # 判断是否超过 7 天
    if [[ $last_commit -lt $seven_days_ago ]]; then
        to_delete+=("$branch_name")
    fi
done < <(git branch -vv --no-color)

# 执行删除操作
if [[ ${#to_delete[@]} -eq 0 ]]; then
    echo "没有需要清理的分支"
    exit 0
fi

echo "以下分支将被删除："
printf " - %s\n" "${to_delete[@]}"

read -p "确认删除以上分支？[y/N] " confirm
if [[ "$confirm" =~ [yY] ]]; then
    for branch in "${to_delete[@]}"; do
        git branch -D "$branch"
    done
    echo "已删除 ${#to_delete[@]} 个分支"
else
    echo "操作已取消"
fi