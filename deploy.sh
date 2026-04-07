#!/bin/bash

# ================================================================
# deploy.sh — Edu Management Backend
# Usage: ./deploy.sh <vm_host> <ssh_key_path> [options]
#
# Examples:
#   ./deploy.sh 192.168.1.100 ~/.ssh/id_rsa
#   ./deploy.sh 192.168.1.100 ~/.ssh/id_rsa --skip-build
#   ./deploy.sh 192.168.1.100 ~/.ssh/id_rsa --version 2.1.0
# ================================================================


set -euo pipefail

# ----------------------------------------------------------------
# CONFIGURATION
# ----------------------------------------------------------------
IMAGE_NAME="edu-management"
IMAGE_VERSION="1.0.0"
VM_USER="ubuntu"
REMOTE_PATH="/home/ubuntu/edu-management"
PLATFORM="linux/amd64"
HEALTH_CHECK_URL="http://localhost:5000/health"
HEALTH_CHECK_RETRIES=10
HEALTH_CHECK_INTERVAL=3  # seconds

BUILD_PATH="./bin/app"
BUILD_PATH_MIGRATION="./bin/migrate"
BUILD_PATH_SEED="./bin/seed"

SERVER_BUILD_PATH="/bin/"
SERVER_CMD_BUILD_PATH="/cmd/"

DOCKER_COMPOSE_FILE="./docker/docker-compose.yml"
DOCKER_FILE="./docker/Dockerfile"

SERVER_DOCKER_COMPOSE_FILE="docker-compose.yml"
SERVER_DOCKER_FILE="Dockerfile"

LOCAL_APP_PATH="./cmd/api"
LOCAL_MIGRATION_PATH="./cmd/migrate"
LOCAL_SEED_PATH="./cmd/seeder"

ENV_FILE=".env.prod"
ENV_CMD_FILE=".env.prod.cmd"

# ----------------------------------------------------------------
# COLORS
# ----------------------------------------------------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ----------------------------------------------------------------
# HELPERS
# ----------------------------------------------------------------
log()     { echo -e "${BLUE}[$(date '+%H:%M:%S')]${RESET} $1"; }
success() { echo -e "${GREEN}[$(date '+%H:%M:%S')] ✓ $1${RESET}"; }
warn()    { echo -e "${YELLOW}[$(date '+%H:%M:%S')] ⚠ $1${RESET}"; }
error()   { echo -e "${RED}[$(date '+%H:%M:%S')] ✗ $1${RESET}" >&2; }
step()    { echo -e "\n${BOLD}${CYAN}━━━ $1 ━━━${RESET}\n"; }

die() {
  error "$1"
  exit 1
}

# Run a command over SSH
remote() {
  ssh -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${VM_USER}@${VM_HOST}" "$@"
}

# ----------------------------------------------------------------
# ARGUMENT PARSING
# ----------------------------------------------------------------
if [[ $# -lt 2 ]]; then
  echo -e "${BOLD}Usage:${RESET} $0 <vm_host> <ssh_key_path> [--skip-build] [--version <version>] [--seed]"
  echo ""
  echo -e "  ${BOLD}--skip-build${RESET}         Skip Docker build — redeploy the current image"
  echo -e "  ${BOLD}--version <ver>${RESET}      Override image version tag (default: ${IMAGE_VERSION})"
  echo -e "  ${BOLD}--seed${RESET}               Run database seeder after deployment"
  echo ""
  exit 1
fi

VM_HOST=$1
SSH_KEY=$2
shift 2

SKIP_BUILD=false
RUN_SEED=false
RUN_MIGRATION=false
RUN_RESET=false # upload dockers and configs

while [[ $# -gt 0 ]]; do
  case "$1" in
    --skip-build)   SKIP_BUILD=true;        shift ;;
    --seed)         RUN_SEED=true;          shift ;;
    --migrate)         RUN_MIGRATION=true;          shift ;;
    --full-reset)         RUN_RESET=true;          shift ;;
    *)              die "Unknown option: $1" ;;
  esac
done


# ----------------------------------------------------------------
# PREFLIGHT CHECKS
# ----------------------------------------------------------------
step "Preflight"

log "Checking required tools..."
for tool in docker ssh scp; do
  command -v "$tool" &>/dev/null || die "'${tool}' is not installed or not in PATH"
done
success "All required tools found"

log "Verifying SSH connectivity to ${VM_HOST}..."
if ! ssh -i "${SSH_KEY}" -o StrictHostKeyChecking=no -o ConnectTimeout=10 \
     "${VM_USER}@${VM_HOST}" exit 2>/dev/null; then
  die "Cannot connect to ${VM_HOST} — check your SSH key and host"
fi
success "SSH connection OK"

log "Configuration:"
echo -e "  Image:       ${BOLD}${BUILD_PATH}${RESET}"
echo -e "  Host:        ${BOLD}${VM_HOST}${RESET}"
echo -e "  User:        ${BOLD}${VM_USER}${RESET}"
echo -e "  Deploy path: ${BOLD}${REMOTE_PATH}${RESET}"
echo -e "  Skip build:  ${BOLD}${SKIP_BUILD}${RESET}"
echo -e "  Run seed:    ${BOLD}${RUN_SEED}${RESET}"
echo -e "  Run migration:    ${BOLD}${RUN_MIGRATION}${RESET}"

# ----------------------------------------------------------------
# BUILD
# ----------------------------------------------------------------
if [[ "${SKIP_BUILD}" == false ]]; then
  step "Build"
  log "Building App ..."
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH} ${LOCAL_APP_PATH} || die "Build failed"
  success "Image built: ${BUILD_PATH}"
else
  warn "Skipping build — using existing image ${BUILD_PATH}"
  
    die "Image ${BUILD_PATH} not found locally — cannot skip build"
fi


# ----------------------------------------------------------------
# SEED (optional)
# ----------------------------------------------------------------
if [[ "${RUN_SEED}" == true ]]; then
  step "Seed"

  log "Build seed..."

  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH_SEED} ${LOCAL_SEED_PATH} || die "Build failed"
  success "Image built: ${BUILD_PATH_SEED}"

fi


# ----------------------------------------------------------------
# MIGRATE (optional)
# ----------------------------------------------------------------
if [[ "${RUN_MIGRATION}" == true ]]; then
  step "Migration"
  log "Build migration..."

  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH_MIGRATION} ${LOCAL_MIGRATION_PATH} || die "Build failed"
  success "Image built: ${BUILD_PATH_MIGRATION}"

fi

# ----------------------------------------------------------------
# TRANSFER
# ----------------------------------------------------------------
step "Transfer"
log "Copying builds and related to ${VM_HOST}:${REMOTE_PATH}/ ..."

scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${BUILD_PATH}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}${SERVER_BUILD_PATH}" || die "SCP transfer failed"

if [[ "${RUN_MIGRATION}" == true ]]; then
  scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${BUILD_PATH_MIGRATION}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}${SERVER_CMD_BUILD_PATH}" || die "SCP transfer failed"
fi

if [[ "${RUN_SEED}" == true ]]; then
  scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${BUILD_PATH_SEED}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}${SERVER_CMD_BUILD_PATH}" || die "SCP transfer failed"
fi

if [[ "${RUN_RESET}" == true ]]; then
scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${ENV_FILE}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}/.env" || die "SCP transfer failed"
scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${ENV_CMD_FILE}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}/cmd/.env" || die "SCP transfer failed"
scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${DOCKER_COMPOSE_FILE}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}${SERVER_DOCKER_COMPOSE_FILE}" || die "SCP transfer failed"
scp -i "${SSH_KEY}" -o StrictHostKeyChecking=no "${DOCKER_FILE}" "${VM_USER}@${VM_HOST}:${REMOTE_PATH}${SERVER_DOCKER_FILE}" || die "SCP transfer failed"
fi

success "Transfer complete"

# ----------------------------------------------------------------
# REMOTE DEPLOYMENT
# ----------------------------------------------------------------
step "Deploy"

remote bash -euo pipefail << REMOTE
  set -euo pipefail

  echo "Navigating to ${REMOTE_PATH}..."
  cd "${REMOTE_PATH}" || { echo "Deploy path not found: ${REMOTE_PATH}"; exit 1; }

  echo "Stopping old containers gracefully..."
  docker compose down --timeout 30

  echo "Starting new containers..."
  docker compose up -d --remove-orphans

  echo "Removing dangling images..."
  docker image prune -f
REMOTE

success "Containers started on ${VM_HOST}"

# ----------------------------------------------------------------
# SEED (optional)
# ----------------------------------------------------------------
if [[ "${RUN_SEED}" == true ]]; then
  step "Seed"

  log "Running database seeder..."
  remote bash -euo pipefail << REMOTE
    cd "${REMOTE_PATH}/cmd"
    ./seed || \
      { echo "Seeder failed"; exit 1; }
REMOTE
  success "Seeder complete"
fi


# ----------------------------------------------------------------
# MIGRATE (optional)
# ----------------------------------------------------------------
if [[ "${RUN_MIGRATION}" == true ]]; then
  step "Migration"

  log "Running database migration..."
  remote bash -euo pipefail << REMOTE
    cd "${REMOTE_PATH}/cmd"
    ./migrate || \
      { echo "migration failed"; exit 1; }
REMOTE
  success "migration complete"
fi

# ----------------------------------------------------------------
# HEALTH CHECK
# ----------------------------------------------------------------
step "Health Check"
log "Waiting for application to become healthy..."

attempt=1
while [[ $attempt -le $HEALTH_CHECK_RETRIES ]]; do
  log "Attempt ${attempt}/${HEALTH_CHECK_RETRIES}..."

  HTTP_STATUS=$(remote curl -s -o /dev/null -w "%{http_code}" \
    --max-time 5 "${HEALTH_CHECK_URL}" 2>/dev/null || echo "000")

  if [[ "${HTTP_STATUS}" == "200" ]]; then
    success "Health check passed (HTTP ${HTTP_STATUS})"
    break
  fi

  if [[ $attempt -eq $HEALTH_CHECK_RETRIES ]]; then
    error "Health check failed after ${HEALTH_CHECK_RETRIES} attempts (last status: ${HTTP_STATUS})"
    warn "Fetching recent logs for diagnosis..."
    remote bash -c "cd ${REMOTE_PATH} && docker compose logs --tail=50 app" || true
    die "Deployment completed but application is not healthy — manual inspection required"
  fi

  warn "Not healthy yet (HTTP ${HTTP_STATUS}) — retrying in ${HEALTH_CHECK_INTERVAL}s..."
  sleep "${HEALTH_CHECK_INTERVAL}"
  (( attempt++ ))
done

# ----------------------------------------------------------------
# LOCAL CLEANUP
# ----------------------------------------------------------------
step "Cleanup"
log "Removing local tar file..."
success "Cleaned up"

# ----------------------------------------------------------------
# DONE
# ----------------------------------------------------------------
echo ""
echo -e "${GREEN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${GREEN}${BOLD}  ✓ Deployment successful${RESET}"
echo -e "${GREEN}${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "  Image:   ${BOLD}${BUILD_PATH}${RESET}"
echo -e "  Host:    ${BOLD}${VM_HOST}${RESET}"
echo -e "  Time:    ${BOLD}$(date '+%Y-%m-%d %H:%M:%S')${RESET}"
echo ""