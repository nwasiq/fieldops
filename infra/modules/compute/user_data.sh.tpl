#!/bin/bash
# Rendered by infra/modules/compute via templatefile(). The three secrets are read from
# SSM Parameter Store at boot; this file carries none of them.
set -euo pipefail

REGION="${region}"
ENV_NAME="${env}"
IMAGE="${backend_image}"
REGISTRY="${ecr_registry}"
ENV_FILE=/etc/fieldops/backend.env

if ! command -v docker >/dev/null 2>&1; then
  dnf install -y docker
fi
systemctl enable --now docker

param() {
  aws ssm get-parameter --region "$REGION" --with-decryption \
    --name "/fieldops/$ENV_NAME/$1" --query Parameter.Value --output text
}

install -d -m 0700 /etc/fieldops
umask 077
{
  echo "DATABASE_URL=$(param DATABASE_URL)"
  echo "JWT_SECRET=$(param JWT_SECRET)"
  echo "ENCRYPTION_KEY=$(param ENCRYPTION_KEY)"
  echo "PORT=8080"
  echo "SEED_ON_BOOT=${seed_on_boot}"
} > "$ENV_FILE"

aws ecr get-login-password --region "$REGION" \
  | docker login --username AWS --password-stdin "$REGISTRY"
docker pull "$IMAGE"
docker rm -f fieldops-backend >/dev/null 2>&1 || true
docker run -d --name fieldops-backend --restart unless-stopped \
  --env-file "$ENV_FILE" -p 8080:8080 "$IMAGE"
