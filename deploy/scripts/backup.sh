#!/usr/bin/env bash
set -euo pipefail

# Script de Backup para PostgreSQL de Atalaya
# Requiere DATABASE_URL configurado en el entorno

if [ -z "${DATABASE_URL:-}" ]; then
  echo "Error: DATABASE_URL no está configurado."
  exit 1
fi

TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_DIR="${BACKUP_DIR:-./backups}"
BACKUP_FILE="${BACKUP_DIR}/atalaya_db_${TIMESTAMP}.sql.gz"

mkdir -p "${BACKUP_DIR}"

echo "Iniciando respaldo de PostgreSQL Atalaya en ${BACKUP_FILE}..."
pg_dump "${DATABASE_URL}" --clean --if-exists --no-owner --no-privileges | gzip > "${BACKUP_FILE}"

echo "Respaldo completado exitosamente: ${BACKUP_FILE}"
ls -lh "${BACKUP_FILE}"
