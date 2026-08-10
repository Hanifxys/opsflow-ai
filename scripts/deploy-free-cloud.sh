#!/usr/bin/env bash
# =============================================================================
# OpsFlow — 100% Free Cloud & VPS Automated Deployment Script
# =============================================================================
# Supports: Oracle Cloud Always Free VPS (ARM 24GB RAM), Render, Fly.io, Ubuntu/Debian

set -e

echo "====================================================================="
echo "🚀 OpsFlow 100% Free Cloud Deployment Initializing..."
echo "====================================================================="

# 1. Install Docker & Docker Compose if missing
if ! command -v docker &> /dev/null; then
    echo "[INFO] Installing Docker Engine..."
    curl -fsSL https://get.docker.com | sh
    sudo usermod -aG docker $USER
fi

if ! docker compose version &> /dev/null; then
    echo "[INFO] Installing Docker Compose plugin..."
    sudo apt-get update && sudo apt-get install -y docker-compose-plugin
fi

# 2. Spin up Docker Compose Stack
echo "[INFO] Launching OpsFlow Self-Hosted Container Stack..."
docker compose up -d --remove-orphans

echo ""
echo "====================================================================="
echo "✨ OpsFlow Free Cloud Deployment Successful!"
echo "====================================================================="
echo "🐳 Portainer Container Manager : http://$(curl -s ifconfig.me):9005"
echo "📊 Grafana Observability UI   : http://$(curl -s ifconfig.me):3000 (admin/admin)"
echo "🪣 MinIO S3 Object Console    : http://$(curl -s ifconfig.me):9001 (opsflow/opsflow_dev_secret)"
echo "🐰 RabbitMQ Broker Console    : http://$(curl -s ifconfig.me):15672 (opsflow/opsflow_dev)"
echo "📈 Prometheus Metrics Engine  : http://$(curl -s ifconfig.me):9090"
echo "🔍 Elasticsearch Search API   : http://$(curl -s ifconfig.me):9200"
echo "====================================================================="
