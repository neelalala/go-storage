#!/usr/bin/env bash
set -euo pipefail

# Всегда выполняем команды из корня проекта
cd "$(dirname "$0")/.."

RELEASES_DIR=".releases"
CURRENT_FILE="${RELEASES_DIR}/current"
LOG_FILE="${RELEASES_DIR}/history.log"
COMPOSE="docker compose -f compose.yaml -f compose.prod.yaml"

SERVICES=(
    "go-storage/users"
    "go-storage/metadata"
    "go-storage/migrate"
    "go-storage/web"
    "go-storage/gateway"
    "go-storage/storage"
)

# Получение версии: SHA коммита git или дата и время, если репозитория нет
get_version() {
    if git rev-parse --git-dir >/dev/null 2>&1; then
        echo $(git rev-parse --short HEAD)
    else
        echo $(date -u +"%Y-%m-%dT%H-%M-%SZ")
    fi
}

# 1. BUILD: Сборка всех Docker-образов с тегом версии и тегом latest, загрузка в minikube
cmd_build() {
    local version=$(get_version)
    echo "==> Building images for version: ${version}"
    RELEASE_VERSION="${version}" $COMPOSE build

    echo "==> Tagging images as :latest..."
    for svc in "${SERVICES[@]}"; do
        docker tag "${svc}:${version}" "${svc}:latest"
    done

    echo "==> Ensuring minikube is running..."
    minikube status >/dev/null 2>&1 || minikube start

    echo "==> Loading images into minikube..."
    local images_to_load=()
    for svc in "${SERVICES[@]}"; do
        images_to_load+=("${svc}:${version}" "${svc}:latest")
    done
    minikube image load "${images_to_load[@]}"
}

# 2. RELEASE: Создание неизменяемого релиза
cmd_release() {
    local version=$(get_version)
    local desc="${1:-Production release}"

    mkdir -p "${RELEASES_DIR}"

    # Release number
    local count=1
    if [ -f "${LOG_FILE}" ]; then
        local last_line
        last_line=$(grep " | CREATE | " "${LOG_FILE}" 2>/dev/null | tail -n 1 || true)
        if [ -z "${last_line}" ]; then
            last_line=$(tail -n 1 "${LOG_FILE}" 2>/dev/null || true)
        fi
        if [ -n "${last_line}" ]; then
            local last_id
            last_id=$(echo "${last_line}" | awk '{print $1}')
            local last_num
            last_num=$(echo "${last_id}" | sed -E 's/^v([0-9]+).*/\1/')
            if [ -n "${last_num}" ]; then
                count=$((10#${last_num} + 1))
            fi
        fi
    fi
    while ls "${RELEASES_DIR}/"$(printf "v%03d-" "${count}")*.env >/dev/null 2>&1; do
        count=$((count + 1))
    done

    local release_id
    release_id=$(printf "v%03d-%s" "${count}" "${version}")

    # Saving release environment 
    local release_env="${RELEASES_DIR}/${release_id}.env"
    if [ -f "${release_env}" ]; then
        echo "Error: release ${release_id} already exists!" >&2
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

# 3. UP: Запуск активного релиза в Kubernetes
cmd_up() {
    [ -f "${CURRENT_FILE}" ] || cmd_release
    local id
    id=$(cat "${CURRENT_FILE}")
    local env_file="${RELEASES_DIR}/${id}.env"

    if [ ! -f "${env_file}" ]; then
        echo "Error: Release environment file not found: ${env_file}" >&2
        exit 1
    fi

    local version
    version=$(grep '^RELEASE_VERSION=' "${env_file}" | cut -d'=' -f2-)
    echo "==> Running release: ${id}) in Kubernetes"

    # Ensure minikube is running
    minikube status >/dev/null 2>&1 || minikube start

    if ! minikube image ls 2>/dev/null | grep -q "go-storage/gateway:${version}"; then
        echo "==> Release images for ${version} not found in minikube. Building and loading..."
        cmd_build "${version}"
    else
        # Tag as :latest
        for svc in "${SERVICES[@]}"; do
            docker tag "${svc}:${version}" "${svc}:latest"
        done
        local latest_images=()
        for svc in "${SERVICES[@]}"; do
            latest_images+=("${svc}:latest")
        done
        minikube image load "${latest_images[@]}"
    fi

    # Delete previous migration job if it exists so it can run again for the release
    kubectl delete job migrate --ignore-not-found >/dev/null 2>&1 || true

    kubectl apply -f k8s/

    kubectl rollout restart deployment web gateway metadata users 2>/dev/null || true
    kubectl rollout restart statefulset storage 2>/dev/null || true

    echo "==> Release ${id} applied successfully"
}

# 4. DOWN: Остановка запущенных ресурсов в Kubernetes
cmd_down() {
    echo "==> Stopping release in Kubernetes..."
    kubectl delete -f k8s/ --ignore-not-found
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

    local version
    version=$(grep '^RELEASE_VERSION=' "${RELEASES_DIR}/${target}.env" | cut -d'=' -f2-)
    echo "==> Rolling back to release: ${target})"

    # Ensure minikube is running
    minikube status >/dev/null 2>&1 || minikube start

    # Build images
    if ! minikube image ls 2>/dev/null | grep -q "go-storage/gateway:${version}"; then
        cmd_build "${version}"
    else
        # Tag as :latest
        for svc in "${SERVICES[@]}"; do
            docker tag "${svc}:${version}" "${svc}:latest"
        done
        local latest_images=()
        for svc in "${SERVICES[@]}"; do
            latest_images+=("${svc}:latest")
        done
        minikube image load "${latest_images[@]}"
    fi

    echo "${target} | $(date -u +"%Y-%m-%dT%H:%M:%SZ") | ROLLBACK" >> "${LOG_FILE}"
    echo "${target}" > "${CURRENT_FILE}"

    kubectl rollout restart deployment web gateway metadata users 2>/dev/null || true
    kubectl rollout restart statefulset storage 2>/dev/null || true

    echo "==> Rollback to release: ${target} complete"
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
    echo "==> Running migration: $*"
    minikube status >/dev/null 2>&1 || minikube start

    local pod_name="migrate-cli-$RANDOM"
    local overrides=$(cat <<EOF
{
  "spec": {
    "containers": [
      {
        "name": "${pod_name}",
        "image": "go-storage/migrate:latest",
        "imagePullPolicy": "Never",
        "envFrom": [
          {
            "secretRef": {
              "name": "postgres-secret"
            }
          }
        ]
      }
    ]
  }
}
EOF
)
    kubectl run "${pod_name}" --rm -i --restart=Never \
        --image="go-storage/migrate:latest" \
        --image-pull-policy=Never \
        --overrides="${overrides}" \
        -- ./migrate "$@"
}

case "${1:-}" in
    build)    cmd_build ;;
    release)  cmd_release "${2:-}" ;;
    up)       cmd_up ;;
    down)     cmd_down ;;
    rollback) cmd_rollback "${2:-}" ;;
    list)     cmd_list ;;
    current)  cmd_current ;;
    migrate)  shift; cmd_migrate "$@" ;;
    web)      minikube service web ;;
    status)   kubectl get all ;;
    stop)     minikube stop ;;
    clean)    minikube delete ;;
    *)        echo "Usage: $0 {build|release|up|run|down|rollback|list|current|migrate|web|status|stop|clean}"; exit 1 ;;
esac
