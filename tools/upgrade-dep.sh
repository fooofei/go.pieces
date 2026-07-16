#!/usr/bin/env bash
#
# 把仓库里所有 Go 模块（含 go.mod 的目录）的依赖升级到最新版本。
#
# 默认升级所有模块；也可以只升级指定目录：
#   tools/upgrade-dep.sh                      # 升级所有模块
#   tools/upgrade-dep.sh tools/cltmsg pkg     # 只升级指定模块
#
# 说明：
#   - go get -u ./... 会升级到同一 major 版本内的最新 minor/patch；
#     跨 major 的升级需要手动处理（import path 通常会变），脚本不会自动做。
#   - 每个模块依次执行 tidy -> get -u -> tidy，单个模块失败不会中断其它模块。

set -uo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

upgrade_one() {
  local dir=$1
  echo "==> ${dir#"$repo_root"/}"
  (
    cd "$dir" || exit 1
    go mod tidy || return 1
    go get -u all || return 1
    go mod tidy || return 1
  )
}

main() {
  local -a dirs=()
  local -a failed=()
  local dir

  if (($#)); then
    for dir in "$@"; do
      dirs+=("$(cd "$dir" && pwd)")
    done
  else
    mapfile -t dirs < <(
      find "$repo_root" -name go.mod \
        -not -path '*/.git/*' -not -path '*/vendor/*' \
        -printf '%h\n' | sort
    )
  fi

  for dir in "${dirs[@]}"; do
    if ! upgrade_one "$dir"; then
      echo "!!! failed: $dir" >&2
      failed+=("$dir")
    fi
  done

  if ((${#failed[@]})); then
    echo
    echo "以下模块升级失败，请手动检查：" >&2
    printf '  %s\n' "${failed[@]}" >&2
    return 1
  fi
  echo
  echo "全部模块依赖已升级完成。"
}

main "$@"
