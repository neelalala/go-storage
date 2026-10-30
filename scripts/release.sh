#!/usr/bin/env bash
set -euo pipefail

# Всегда выполняем команды из корня проекта
cd "$(dirname "$0")/.."

RELEASES_DIR=".releases"
CURRENT_FILE="${RELEASES_DIR}/current"
LOG_FILE="${RELEASES_DIR}/history.log"
COMPOSE="docker compose -f compose.yaml -f compose.prod.yaml"

# Получение версии: SHA коммита git или дата и время, если репозитория нет
get_version() {
    if git rev-parse --git-dir >/dev/null 2>&1; then
        echo $(git rev-parse --short HEAD)
    else
        echo $(date -u +"%Y-%m-%d-%H-%M-%S")
    fi
}

# 1. BUILD: Сборка всех Docker-образов с тегом версии
cmd_build() {
    local version
    version=$(get_version)
    echo "==> Building image: ${version}"
    RELEASE_VERSION="${version}" $COMPOSE build
}

# 2. RELEASE: Создание неизменяемого релиза
cmd_release() {
    local version
    version=$(get_version)
    local desc="${1:-Production release}"

    mkdir -p "${RELEASES_DIR}"

    # Relese number
    local count=1
    if [ -f "${LOG_FILE}" ]; then
        count=$(($(grep -c " | CREATE | " "${LOG_FILE}" 2>/dev/null || true) + 1))
    fi
    local release_id
    release_id=$(printf "v%03d-%s" "${count}" "${version}")

    # Saving release environment 
    local release_env="${RELEASES_DIR}/${release_id}.env"
    if [ -f "${release_env}" ]; then
        echo "Error: relese ${release_id} already exists!" >&2
        exit 1
    fi

    local src=".env"
    [ -f "$src" ] || src=".env.example"

    echo "RELEASE_VERSION=${version}" > "${release_env}"
    grep -v '^RELEASE_VERSION=' "$src" >> "${release_env}" 2>/dev/null || true
    chmod 444 "${release_env}" # Making read-only

    # Logging and updating current
    echo "${release_id} | $(date -u +"%Y-%m-%dT%H:%M:%SZ") | CREATE | ${desc}" >> "${LOG_FILE}"
    echo "${release_id}" > "${CURRENT_FILE}"

    echo "==> Release created: ${release_id}"
}

# 3. RUN: Запуск активного релиза
cmd_run() {
    [ -f "${CURRENT_FILE}" ] || cmd_release
    local id
    id=$(cat "${CURRENT_FILE}")
    echo "==> Running release: ${id}"
    $COMPOSE --env-file "${RELEASES_DIR}/${id}.env" up -d --no-build
}

# 4. DOWN: Остановка запущенных контейнеров
cmd_down() {
    local env_arg=""
    [ -f "${CURRENT_FILE}" ] && env_arg="--env-file ${RELEASES_DIR}/$(cat "${CURRENT_FILE}").env"
    $COMPOSE ${env_arg} down
}

# 5. ROLLBACK: Откат на предыдущий (или указанный) релиз
cmd_rollback() {
    local target="${1:-}"

    # Previous release as target
    if [ -z "$target" ]; then
        local cur=""
        [ -f "${CURRENT_FILE}" ] && cur=$(cat "${CURRENT_FILE}")
        local prev=""
        for rel in $(grep " | CREATE | " "${LOG_FILE}" 2>/dev/null | awk '{print $1}'); do
            if [ "$rel" = "$cur" ]; then
                break
            fi
            prev="$rel"
        done
        target="$prev"
    fi

    if [ -z "$target" ] || [ ! -f "${RELEASES_DIR}/${target}.env" ]; then
        echo "Error: release for rollback not found" >&2
        exit 1
    fi

    echo "${target} | $(date -u +"%Y-%m-%dT%H:%M:%SZ") | ROLLBACK" >> "${LOG_FILE}"
    echo "${target}" > "${CURRENT_FILE}"
    echo "==> Rollback on release: ${target}"
    cmd_run
}

# 6. LIST: История релизов и откатов
cmd_list() {
    [ -f "${LOG_FILE}" ] || { echo "No releases yet"; return; }
    local cur=""
    [ -f "${CURRENT_FILE}" ] && cur=$(cat "${CURRENT_FILE}")
    while IFS= read -r line; do
        if echo "$line" | grep -q "^${cur} "; then
            echo "=> ${line}  <-- CURRENT"
        else
            echo "   ${line}"
        fi
    done < "${LOG_FILE}"
}

# 7. CURRENT: Показать активный релиз
cmd_current() {
    if [ -f "${CURRENT_FILE}" ]; then
        echo "Current release: $(cat "${CURRENT_FILE}")"
    else
        echo "No current release"
    fi
}

# 8. MIGRATE: Запуск миграций в окружении активного релиза
cmd_migrate() {
    local env_arg=""
    [ -f "${CURRENT_FILE}" ] && env_arg="--env-file ${RELEASES_DIR}/$(cat "${CURRENT_FILE}").env"
    $COMPOSE ${env_arg} run --rm --no-build migrate "$@"
}

case "${1:-}" in
    build)    cmd_build ;;
    release)  cmd_release "${2:-}" ;;
    run)      cmd_run ;;
    down)     cmd_down ;;
    rollback) cmd_rollback "${2:-}" ;;
    list)     cmd_list ;;
    current)  cmd_current ;;
    migrate)  shift; cmd_migrate "$@" ;;
    *)        echo "Usage: $0 {build|release|run|down|rollback|list|current|migrate}"; exit 1 ;;
esac
