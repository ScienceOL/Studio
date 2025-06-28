#!/bin/bash

# =============================================
# 构建开发环境容器服务控制脚本
# =============================================

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

IMAGE_NAME="registry.cn-shanghai.aliyuncs.com/scienceol/protium:dev"
SERVICE_IMAGE="registry.aissquare.com/web/service:dev"

# =============================================
# 主执行流程
# =============================================

# 显示品牌图标
print_icon

# -------------------------------
# 环境预检
# -------------------------------
echo -e "${BRIGHT_MAGENTA}\n⚙  检查 Docker 环境...${RESET}"
if ! command -v docker &> /dev/null; then
  echo -e "${BRIGHT_RED}❌ 错误：未检测到 Docker 安装${RESET}"
  exit 1
fi
echo -e "${BRIGHT_GREEN}✓ Docker 已就绪${RESET}"

# =============================================
# Docker Compose 配置预览
# =============================================
# echo -e "${BRIGHT_CYAN}\n🔧 合并 Docker Compose 配置...${RESET}"
# echo -e "${YELLOW}══════════════════════════════════════════${RESET}"
# if [ ! -d "${PROJECT_DIR}/.temp" ]; then
#   mkdir "${PROJECT_DIR}/.temp"
# fi
# docker compose -f ${PROJECT_DIR}/docker/docker-compose.base.yaml \
#               -f ${PROJECT_DIR}/docker/docker-compose.prod.yaml \
#               --env-file ${ENV_FILE} config > "${PROJECT_DIR}/.temp/docker-compose.buildx.config.yaml"
# echo -e "配置文件保存至临时文件路径：${BRIGHT_BLUE}${PROJECT_DIR}/.temp/docker-compose.buildx.config.yaml${RESET}"
# echo -e "${YELLOW}══════════════════════════════════════════${RESET}"

# =============================================
# 容器构建
# =============================================
echo -e "${BRIGHT_BLUE}\n🚀 构建 prod 容器服务...${RESET}"
docker compose -f ${PROJECT_DIR}/docker/docker-compose.base.yaml \
              -f ${PROJECT_DIR}/docker/docker-compose.prod.yaml \
              --env-file ${ENV_FILE} build service

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ${SERVICE_IMAGE} \
  --push \
  --cache-from type=registry,ref=${SERVICE_IMAGE} \
  ${PROJECT_DIR}/service/

docker compose -f ${PROJECT_DIR}/docker/docker-compose.base.yaml \
                -f ${PROJECT_DIR}/docker/docker-compose.prod.yaml \
                --env-file ${ENV_FILE} \
                up -d --pull always service

docker compose -f ${PROJECT_DIR}/docker/docker-compose.base.yaml \
              -f ${PROJECT_DIR}/docker/docker-compose.prod.yaml \
              --env-file ${ENV_FILE} build protium

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t ${IMAGE_NAME} \
  --push \
  --cache-from type=registry,ref=${IMAGE_NAME} \
  ${PROJECT_DIR}/protium/

docker compose -f ${PROJECT_DIR}/docker/docker-compose.base.yaml \
                -f ${PROJECT_DIR}/docker/docker-compose.prod.yaml \
                --env-file ${ENV_FILE} \
                up -d --pull always protium


echo -e "\n${BRIGHT_GREEN}✅ prod 环境已成功构建！${RESET}"

# =============================================
# 检查 Docker 镜像是否是多平台构建
# =============================================
echo -e "${BRIGHT_CYAN}\n🔍 检查 Docker 镜像是否是多平台构建...${RESET}"


# 检查 manifest 命令可用性
if ! docker manifest inspect --help &> /dev/null; then
  echo -e "${BRIGHT_RED}❌ 错误：请启用 Docker 实验性功能以支持 manifest 检查${RESET}"
  echo -e "${YELLOW}操作步骤："
  echo "1. 编辑或创建 ~/.docker/config.json"
  echo "2. 添加内容：{ \"experimental\": \"enabled\" }"
  echo "3. 重启 Docker 服务${RESET}"
  exit 1
fi

# 获取 manifest 信息
MANIFEST_OUTPUT=$(docker manifest inspect $IMAGE_NAME 2>&1)
if [ $? -ne 0 ]; then
  echo -e "${BRIGHT_RED}❌ 错误：无法获取镜像 manifest${RESET}"
  echo -e "${YELLOW}可能原因："
  echo "1. 镜像未成功推送到仓库"
  echo "2. 网络连接问题"
  echo "3. 镜像标签不存在${RESET}"
  exit 1
fi

# 解析平台信息
PLATFORMS=$(echo "$MANIFEST_OUTPUT" | jq -r '.manifests[].platform | .architecture + "/" + .os' 2>/dev/null)

if [[ "$PLATFORMS" =~ "amd64/linux" && "$PLATFORMS" =~ "arm64/linux" ]]; then
  echo -e "${BRIGHT_GREEN}✅ 多平台构建验证通过 (amd64 + arm64)${RESET}"
  echo -e "${YELLOW}══════════════════════════════════════════"
  echo "$PLATFORMS" | sort | uniq | sed 's/^/• /'
  echo -e "══════════════════════════════════════════${RESET}"
else
  echo -e "${BRIGHT_RED}❌ 多平台构建验证失败${RESET}"
  echo -e "${YELLOW}已检测到的平台："
  echo "$PLATFORMS" | sort | uniq | sed 's/^/• /'
  echo -e "${BRIGHT_RED}请检查以下配置："
  echo "1. 构建命令是否包含 --platform 参数"
  echo "2. 是否使用了正确的 buildx 构建器"
  echo "3. 镜像是否成功推送到仓库${RESET}"
  exit 1
fi

