#!/bin/bash
# Generate ezBookkeeping configuration from install/config wizard values.
set -u

CONF_DIR="${TRIM_PKGETC}"
CONF_FILE="${CONF_DIR}/ezbookkeeping.ini"
TEMPLATE="${TRIM_APPDEST}/conf/ezbookkeeping.ini"
VAR_DIR="${TRIM_PKGVAR}"

mkdir -p "${CONF_DIR}" "${VAR_DIR}/data" "${VAR_DIR}/files" "${VAR_DIR}/log"

# Keep previous values when a wizard field is empty (config wizard re-run).
old_value() {
    if [ -f "${CONF_FILE}" ]; then
        grep -E "^$1 = " "${CONF_FILE}" 2>/dev/null | head -n 1 | cut -d '=' -f 2- | sed 's/^ *//;s/ *$//'
    fi
}

PORT="${wizard_http_port:-}"
[ -z "${PORT}" ] && PORT="$(old_value http_port)"
[ -z "${PORT}" ] && PORT="8080"

DB_TYPE="${wizard_db_type:-}"
[ -z "${DB_TYPE}" ] && DB_TYPE="$(old_value type)"
[ -z "${DB_TYPE}" ] && DB_TYPE="sqlite3"

DB_HOST="${wizard_db_host:-}"
[ -z "${DB_HOST}" ] && DB_HOST="$(old_value host)"
[ -z "${DB_HOST}" ] && DB_HOST="127.0.0.1:3306"

DB_NAME="${wizard_db_name:-}"
[ -z "${DB_NAME}" ] && DB_NAME="$(old_value name)"
[ -z "${DB_NAME}" ] && DB_NAME="ezbookkeeping"

DB_USER="${wizard_db_user:-}"
[ -z "${DB_USER}" ] && DB_USER="$(old_value user)"
[ -z "${DB_USER}" ] && DB_USER="ezbookkeeping"

DB_PASS="${wizard_db_passwd:-}"
if [ -z "${DB_PASS}" ] && [ -f "${CONF_FILE}" ]; then
    DB_PASS="$(old_value passwd)"
fi

# Login is handled by the fnOS unified gateway (X-Trim-* headers on app.sock),
# so the OAuth flow is always disabled and no client credentials are needed.
ENABLE_FNOS="false"

if [ -f "${CONF_FILE}" ]; then
    OLD_SECRET="$(old_value secret_key)"
fi
SECRET="${OLD_SECRET:-}"
if [ -z "${SECRET}" ]; then
    if command -v openssl >/dev/null 2>&1; then
        SECRET="$(openssl rand -hex 32)"
    else
        SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
    fi
fi

if [ ! -f "${TEMPLATE}" ]; then
    echo "找不到配置模板：${TEMPLATE}" > "${LOG_FILE}"
    exit 1
fi

sed \
    -e "s|^http_port = .*|http_port = ${PORT}|" \
    -e "s|^static_root_path = .*|static_root_path = ${TRIM_APPDEST}/public|" \
    -e "s|^local_filesystem_path = .*|local_filesystem_path = ${VAR_DIR}/files|" \
    -e "s|^log_path = .*|log_path = ${VAR_DIR}/log/ezbookkeeping.log|" \
    -e "s|^db_path = .*|db_path = ${VAR_DIR}/data/ezbookkeeping.db|" \
    -e "s|^type = sqlite3.*|type = ${DB_TYPE}|" \
    -e "s|^host = 127.0.0.1:3306.*|host = ${DB_HOST}|" \
    -e "s|^name = ezbookkeeping.*|name = ${DB_NAME}|" \
    -e "s|^user = root.*|user = ${DB_USER}|" \
    -e "s|^passwd =.*|passwd = ${DB_PASS}|" \
    -e "s|^secret_key =.*|secret_key = ${SECRET}|" \
    -e "s|^enable_register = .*|enable_register = true|" \
    -e "s|^enable_fnos_auth = .*|enable_fnos_auth = ${ENABLE_FNOS}|" \
    -e "s|^enable_fnos_gateway_auth = .*|enable_fnos_gateway_auth = true|" \
    "${TEMPLATE}" > "${CONF_FILE}"

if [ ! -s "${CONF_FILE}" ]; then
    echo "生成配置文件失败：${CONF_FILE}" > "${LOG_FILE}"
    exit 1
fi

# Keep the desktop entry port in sync with the configured port.
UI_CONFIG="${TRIM_APPDEST}/ui/config"
if [ -f "${UI_CONFIG}" ]; then
    sed -i "s|\"port\": \"[0-9]*\"|\"port\": \"${PORT}\"|" "${UI_CONFIG}"
fi

exit 0
