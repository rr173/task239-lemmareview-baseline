#!/usr/bin/env bash
# 评测镜像构建脚本：用法 bash build_benzhi_docker.sh <镜像名> <平台>
# 平台示例：linux/amd64 或 linux/arm64
set -euo pipefail

IMAGE_NAME="${1:-my-project}"
PLATFORM="${2:-linux/amd64}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "building benzhi image ${IMAGE_NAME} for platform ${PLATFORM}"
docker buildx build --platform "${PLATFORM}" -f benzhi.Dockerfile -t "${IMAGE_NAME}" .

echo "done: ${IMAGE_NAME} (${PLATFORM})"
