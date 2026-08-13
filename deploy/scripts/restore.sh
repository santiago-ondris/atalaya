#!/usr/bin/env bash
set -euo pipefail

# Script de Restauración para PostgreSQL de Atalaya
# Uso: ./restore.sh <path_al_archivo_backup.sql.gz>

if [ -z "${DATABASE_URL:-}" ]; then
  echo "Error: DATABASE_URL no está configurado."
  exit 1
fi

if [ $# -lt 1 ]; then
  echo "Uso: $0 <archivo_backup.sql.gz>"
  exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "${BACKUP_FILE}" ]; then
  echo "Error: El archivo de respaldo '${BACKUP_FILE}' no existe."
  exit 1
fi

echo "ADVERTENCIA: Se restaurará la base de datos Atalaya desde '${BACKUP_FILE}'."
read -p "¿Desea continuar? (s/N): " CONFIRM
if [[ "${CONFIRM}" != "s" && "${CONFIRM}" != "S" ]]; then
  echo "Operación cancelada."
  exit 0
fi

echo "Restaurando base de datos..."
gunzip -c "${BACKUP_FILE}" | psql "${DATABASE_URL}"

echo "Restauración completada exitosamente."
