#!/bin/bash

# 配置参数
NACOS_HOST="127.0.0.1:8848"
NACOS_BASE_URL="http://${NACOS_HOST}/nacos"
USERNAME="nacos"
PASSWORD="nacos"

# 配置信息
NAMESPACE_ID="public"
GROUP_NAME="DEFAULT_GROUP"
DATA_ID="studio-config"
CONTENT="test: xxx"
TYPE="yaml"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 打印带颜色的消息
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 登录函数
login() {
    log_info "尝试登录..."
    response=$(curl -s -X POST "${NACOS_BASE_URL}/v3/auth/user/login" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "username=${USERNAME}&password=${PASSWORD}")
    
    # 检查是否包含 accessToken
    if echo "$response" | grep -q "accessToken"; then
        ACCESS_TOKEN=$(echo "$response" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
        log_info "登录成功，获取到 accessToken"
        return 0
    else
        log_warn "登录失败，响应: $response"
        return 1
    fi
}

# 初始化管理员账号
init_admin() {
    log_info "初始化管理员账号..."
    response=$(curl -s -X POST "http://127.0.0.1:8080/v3/auth/user/admin" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "password=${PASSWORD}")
    
    # log_info "初始化管理员账号响应: $response"
    log_info ${RED}"账号：nacos，密码：nacos"${NC}
    
    # 等待一下再尝试登录
    sleep 2
}

# 检查配置是否存在
check_config_exists() {
    log_info "检查配置是否存在..."
    response=$(curl -s -X GET "${NACOS_BASE_URL}/v3/admin/cs/config" \
        -H "accessToken: ${ACCESS_TOKEN}" \
        -G \
        -d "namespaceId=${NAMESPACE_ID}" \
        -d "groupName=${GROUP_NAME}" \
        -d "dataId=${DATA_ID}")
    
    # 检查响应中是否包含配置数据
    if echo "$response" | grep -q '"code":0'; then
        log_info "配置已存在"
        return 0
    else
        log_info "配置不存在，需要创建"
        return 1
    fi
}

# 创建配置
create_config() {
    log_info "创建配置..."
    response=$(curl -s -X POST "${NACOS_BASE_URL}/v3/admin/cs/config" \
        -H "accessToken: ${ACCESS_TOKEN}" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "namespaceId=${NAMESPACE_ID}" \
        -d "groupName=${GROUP_NAME}" \
        -d "dataId=${DATA_ID}" \
        -d "content=${CONTENT}" \
        -d "type=${TYPE}")
    
    if echo "$response" | grep -q '"code":0'; then
        log_info ${RED}"配置创建成功 dataId=${DATA_ID}, groupName=${GROUP_NAME}, namespaceId=${NAMESPACE_ID}"${NC}
        return 0
    else
        log_error "配置创建失败，响应: $response"
        return 1
    fi
}

# 主函数
main() {
    log_info "开始执行 Nacos 配置初始化脚本..."
    
    # 第一步：尝试登录
    if ! login; then
        log_warn "登录失败，尝试初始化管理员账号..."
        
        # 第二步：初始化管理员账号
        init_admin
        
        # 第三步：重新登录
        if ! login; then
            log_error "重新登录失败，脚本退出"
            exit 1
        fi
    fi
    
    # 第四步：检查配置是否存在
    if ! check_config_exists; then
        # 第五步：创建配置
        if create_config; then
            log_info "配置创建成功"
        else
            log_error "配置创建失败"
            exit 1
        fi
    else
        log_info "配置已存在，无需创建"
    fi
    
    log_info "脚本执行完成"
}

# 执行主函数
main "$@"