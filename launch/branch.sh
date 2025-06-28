#!/bin/bash

# 生成斜体Science OL大标题
function print_icon() {
  # 颜色配置
  local COLOR_HEAD="\033[38;5;214m"    # 开头部分颜色（橙色）
  local COLOR_DECO="\033[38;5;242m"    # 装饰符号颜色（深色）
  local COLOR_TAIL="\033[38;5;220m"    # OL部分颜色（黄色）
  local ITALIC="\033[3m"           # 斜体控制
  local RESET="\033[0m"            # 重置样式
  
  echo -e "${ITALIC}${COLOR_HEAD}"
  echo -e " ██████${COLOR_DECO}╗ ${COLOR_HEAD} ██████╗ ██╗ ███████╗ ███╗   ██╗  ██████╗ ███████╗ ${COLOR_TAIL} ██████╗  ██╗     ${RESET}${ITALIC}${COLOR_HEAD}"
  echo -e "██${COLOR_DECO}╔════╝ ${COLOR_HEAD}██╔════╝ ██║ ██╔════╝ ██╔██╗ ██║ ██╔════╝ ██╔════╝ ${COLOR_TAIL}██╔═══██╗ ██║     ${RESET}${ITALIC}${COLOR_HEAD}"
  echo -e "${COLOR_DECO}╚${COLOR_HEAD}█████${COLOR_DECO}╗  ${COLOR_HEAD}██║      ██║ █████╗   ██║╚██╗██║ ██║      █████╗   ${COLOR_TAIL}██║   ██║ ██║     ${RESET}${ITALIC}${COLOR_HEAD}"
  echo -e " ${COLOR_DECO}╚═══${COLOR_HEAD}██${COLOR_DECO}╗ ${COLOR_HEAD}██║      ██║ ██╔══╝   ██║ ╚████║ ██║      ██╔══╝   ${COLOR_TAIL}██║   ██║ ██║     ${RESET}${ITALIC}${COLOR_HEAD}"
  echo -e "██████${COLOR_DECO}╔╝ ${COLOR_HEAD}╚██████╗ ██║ ███████╗ ██║  ╚███║ ╚██████╗ ███████╗ ${COLOR_TAIL}╚██████╔╝ ███████╗${RESET}${ITALIC}${COLOR_HEAD}"
  echo -e "${COLOR_DECO}╚═════╝  ${COLOR_HEAD} ╚═════╝ ╚═╝ ╚══════╝ ╚═╝   ╚══╝  ╚═════╝ ╚══════╝ ${COLOR_TAIL} ╚═════╝  ╚══════╝${RESET}"
  echo -e ""
}

# 直接执行时显示图标
if [[ "$0" == *"icon.sh" ]]; then
  print_icon
fi