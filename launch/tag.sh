#!/usr/bin/env bash

# Usage: ./tag.sh 
# Automatically create a tag on dev and push to origin

# -------------------------------
# 全局配置
# -------------------------------
SCRIPT_DIR=$(dirname "$0")
PROJECT_DIR=$(dirname "${SCRIPT_DIR}")
ENV_FILE="${PROJECT_DIR}/docker/.env.dev"

# -------------------------------
# 颜色配置
# -------------------------------
source "${SCRIPT_DIR}/colors.sh"

# -------------------------------
# 品牌显示
# -------------------------------
print_icon() {
  source "${SCRIPT_DIR}/branch.sh" && print_icon
}

print_icon

echo -e "${BRIGHT_RED}==================================${NC}"
echo -e "${BRIGHT_RED}⚠️ 警告: 此标签操作将会把 dev 分支合并到 main 并触发一次正式发布, 请确认后执行！${NC}"
echo -e "${BRIGHT_RED}⚠️ 警告: 请确保本地 dev 分支是最新的！${NC}"
echo -e "${BRIGHT_RED}==================================${NC}"

# 获取所有标签信息
echo -e "${BLUE}获取标签信息...${NC}"
git fetch --tags --quiet

# 获取本地标签
LOCAL_TAGS=$(git tag -l "v*" | sort -V)

# 获取远程标签（去除refs/tags/前缀）
REMOTE_TAGS=$(git ls-remote --tags origin | grep -v '\^{}' | awk '{print $2}' | sed 's|refs/tags/||' | grep "^v" | sort -V)

# 显示所有标签
echo -e "${CYAN}==== 标签列表 ====${NC}"

# 如果没有找到任何标签
if [ -z "$LOCAL_TAGS" ] && [ -z "$REMOTE_TAGS" ]; then
    echo "  没有找到标签"
    NEXT_TAG="v0.1.0"
else
    # 使用更兼容的方式处理标签分类，避免使用关联数组
    
    # 显示同步的标签
    echo -e "${GREEN}本地和远程同步的标签:${NC}"
    BOTH_TAGS=""
    for tag in $LOCAL_TAGS; do
        if echo "$REMOTE_TAGS" | grep -q "^$tag$"; then
            echo -e "  ${GREEN}${tag}${NC}"
            BOTH_TAGS="$BOTH_TAGS $tag"
        fi
    done
    [ -z "$BOTH_TAGS" ] && echo "  (无)"
    
    # 显示仅本地的标签
    echo -e "${YELLOW}仅本地存在的标签:${NC}"
    LOCAL_ONLY_TAGS=""
    for tag in $LOCAL_TAGS; do
        if ! echo "$REMOTE_TAGS" | grep -q "^$tag$"; then
            echo -e "  ${YELLOW}${tag}${NC}"
            LOCAL_ONLY_TAGS="$LOCAL_ONLY_TAGS $tag"
        fi
    done
    [ -z "$LOCAL_ONLY_TAGS" ] && echo "  (无)"
    
    # 显示仅远程的标签
    echo -e "${BLUE}仅远程存在的标签:${NC}"
    REMOTE_ONLY_TAGS=""
    for tag in $REMOTE_TAGS; do
        if ! echo "$LOCAL_TAGS" | grep -q "^$tag$"; then
            echo -e "  ${BLUE}${tag}${NC}"
            REMOTE_ONLY_TAGS="$REMOTE_ONLY_TAGS $tag"
        fi
    done
    [ -z "$REMOTE_ONLY_TAGS" ] && echo "  (无)"
    
    # 获取最新标签用于建议下一个版本号
    # 从所有标签中找出最新的一个
    ALL_TAGS=$(echo -e "$LOCAL_TAGS\n$REMOTE_TAGS" | sort -V | uniq)
    LATEST_TAG=$(echo "$ALL_TAGS" | tail -n1)
    
    # 计算下一个推荐tag
    if [ -n "$LATEST_TAG" ] && [[ "$LATEST_TAG" =~ v([0-9]+)\.([0-9]+)\.([0-9]+) ]]; then
        MAJOR="${BASH_REMATCH[1]}"
        MINOR="${BASH_REMATCH[2]}"
        PATCH="${BASH_REMATCH[3]}"
        
        # 默认增加补丁版本
        NEXT_PATCH=$((PATCH + 1))
        NEXT_TAG="v${MAJOR}.${MINOR}.${NEXT_PATCH}"
    else
        # 如果没有符合格式的tag，设置默认值
        NEXT_TAG="v0.1.0"
    fi
fi

echo -e "${CYAN}================${NC}"

# 用户输入tag
echo -e "${YELLOW}请输入标签 (格式: v0.0.0) [默认: ${NEXT_TAG}]:${NC}"
read -p "> " tag_name

# 使用默认值如果没有输入
if [ -z "$tag_name" ]; then
    tag_name=$NEXT_TAG
fi

# 验证tag格式
if ! [[ $tag_name =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}错误: 标签格式必须是 v0.0.0${NC}"
    exit 1
fi

# 检查标签是否已存在
if echo "$LOCAL_TAGS $REMOTE_TAGS" | grep -q "^$tag_name$"; then
    echo -e "${RED}错误: 标签 ${tag_name} 已存在!${NC}"
    exit 1
fi

# 询问tag描述
echo -e "${YELLOW}请输入标签描述 (可选):${NC}"
read -p "> " description

# 确认操作
echo -e "${BLUE}准备创建标签: ${GREEN}${tag_name}${NC}"
if [ ! -z "$description" ]; then
    echo -e "${BLUE}描述: ${GREEN}${description}${NC}"
fi
echo -e "${YELLOW}确认? (y/n)${NC}"
read -p "> " confirm

if [ "$confirm" != "y" ] && [ "$confirm" != "Y" ]; then
    echo -e "${RED}已取消${NC}"
    exit 0
fi

# 切换到dev分支
echo -e "${BLUE}切换到dev分支...${NC}"
git checkout dev
git pull origin dev

# 创建tag
if [ -z "$description" ]; then
    git tag "$tag_name"
else
    git tag -a "$tag_name" -m "$description"
fi

# 推送tag
echo -e "${BLUE}推送标签到远程...${NC}"
git push origin "$tag_name"

echo -e "${GREEN}✅ 标签 ${tag_name} 已创建并推送到远程!${NC}"