#!/bin/bash

NACOS_URL="http://localhost:8848"
USERNAME="nacos"
PASSWORD="nacos"

echo "Starting Nacos initialization..."

# 等待 Nacos 完全启动
sleep 10

# 创建 YAML 配置内容
CONFIG_YAML=$(cat <<'EOF'
# Protium Go Configuration
server:
  port: 8080

database:
  host: localhost
  port: 5432
  name: protium
  user: postgres
  password: please_change_me

redis:
  host: localhost
  port: 6379

minio:
  endpoint: localhost:9000
  access_key: minioadmin
  secret_key: minioadmin
  bucket: studio

elasticsearch:
  host: localhost
  port: 9200
  username: elastic
  password: please_change_me

mqtt:
  host: localhost
  port: 1883
EOF
)

# 创建配置项 protium-go (YAML 格式)
echo "Creating configuration 'protium-go'..."
curl -X POST "${NACOS_URL}/nacos/v1/cs/configs" \
    -d "dataId=protium-go&group=DEFAULT_GROUP&type=yaml" \
    --data-urlencode "content=${CONFIG_YAML}" \
    -H "Content-Type: application/x-www-form-urlencoded"

echo "Nacos initialization completed!"