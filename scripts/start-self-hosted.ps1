# =============================================================================
# OpsFlow — 100% Free Self-Hosted Stack & Portal Launcher
# =============================================================================

Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "🚀 Starting OpsFlow 100% Free Self-Hosted Stack & Portainer Console..." -ForegroundColor Green
Write-Host "=====================================================================" -ForegroundColor Cyan

docker compose up -d

Write-Host "`n=====================================================================" -ForegroundColor Cyan
Write-Host "✨ Self-Hosted Platform Directory & Web Consoles Ready:" -ForegroundColor Yellow
Write-Host "=====================================================================" -ForegroundColor Cyan

Write-Host "🐳 Portainer Container Manager : http://localhost:9000" -ForegroundColor Green
Write-Host "📊 Grafana Observability UI   : http://localhost:3000 (admin/admin)" -ForegroundColor Green
Write-Host "🪣 MinIO S3 Object Storage    : http://localhost:9001 (opsflow/opsflow_dev_secret)" -ForegroundColor Green
Write-Host "🐰 RabbitMQ Broker Console    : http://localhost:15672 (opsflow/opsflow_dev)" -ForegroundColor Green
Write-Host "📈 Prometheus Metrics Engine  : http://localhost:9090" -ForegroundColor Green
Write-Host "🔍 Elasticsearch Search API   : http://localhost:9200" -ForegroundColor Green
Write-Host "💾 PostgreSQL Database        : localhost:5432 (opsflow/opsflow_dev)" -ForegroundColor Green
Write-Host "⚡ Redis Cache & Rate Limit   : localhost:6379" -ForegroundColor Green

Write-Host "=====================================================================" -ForegroundColor Cyan
Write-Host "All services running self-hosted & 100% free!" -ForegroundColor Cyan
