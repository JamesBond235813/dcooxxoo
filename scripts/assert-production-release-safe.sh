#!/usr/bin/env bash
set -euo pipefail

# Fail closed before any production release. This is a guard, not a deployer.
# Required: EXPECTED_LIVE_BINARY_SHA256=<hash read from the production host>
# Optional for a user-authorized UI release: APPROVE_UI_CHANGE=YES

repo_root="$(git rev-parse --show-toplevel 2>/dev/null)" || {
  echo "拒绝发布：当前目录不是 Git 工作树。" >&2
  exit 2
}
cd "$repo_root"

baseline_commit_file="$repo_root/ops/production-ui-baseline-20261012/baseline-git-commit.txt"
if [[ ! -r "$baseline_commit_file" ]]; then
  echo "拒绝发布：当前仓库没有锁定的生产基线 commit。" >&2
  exit 2
fi
baseline_commit="$(tr -d '[:space:]' < "$baseline_commit_file")"
if ! git merge-base --is-ancestor "$baseline_commit" HEAD 2>/dev/null; then
  echo "拒绝发布：当前 HEAD 不是从锁定的生产源码基线派生。" >&2
  exit 2
fi

if [[ -z "${EXPECTED_LIVE_BINARY_SHA256:-}" ]]; then
  echo "拒绝发布：请先读取生产二进制 SHA256，并设置 EXPECTED_LIVE_BINARY_SHA256。" >&2
  exit 2
fi
if [[ ! "$EXPECTED_LIVE_BINARY_SHA256" =~ ^[[:xdigit:]]{64}$ ]]; then
  echo "拒绝发布：EXPECTED_LIVE_BINARY_SHA256 格式无效。" >&2
  exit 2
fi

if [[ -n "$(git status --porcelain --untracked-files=all)" ]]; then
  echo "拒绝发布：工作树有未提交或未跟踪文件。请在从生产基线建立的干净隔离工作树构建。" >&2
  git status --short >&2
  exit 3
fi

if [[ "${APPROVE_UI_CHANGE:-}" != "YES" ]]; then
  if ! python3 - "$repo_root" <<'PY'
import hashlib
import subprocess
import sys
from pathlib import Path

root = Path(sys.argv[1])
checks = [
    ("frontend/user", "user-source.SHA256SUMS"),
    ("frontend/admin", "admin-source.SHA256SUMS"),
    ("internal/web/dist/user", "user-dist.SHA256SUMS"),
    ("internal/web/dist/admin", "admin-dist.SHA256SUMS"),
]
manifest_root = root / "ops/production-ui-baseline-20261012"
for rel_dir, manifest_name in checks:
    base = root / rel_dir
    manifest = manifest_root / manifest_name
    if not base.is_dir() or not manifest.is_file():
        print(f"缺少锁定的 UI 路径或清单：{rel_dir} / {manifest_name}", file=sys.stderr)
        sys.exit(1)
    expected = {}
    for raw in manifest.read_text().splitlines():
        if not raw.strip():
            continue
        digest, filename = raw.split(None, 1)
        expected[filename.lstrip(" *")] = digest
    if rel_dir.startswith("frontend/"):
        tracked = subprocess.check_output(
            ["git", "ls-files", "--", rel_dir], cwd=root, text=True
        ).splitlines()
        actual_names = {
            "./" + Path(name).relative_to(rel_dir).as_posix()
            for name in tracked if (root / name).is_file()
        }
    else:
        actual_names = {
            "./" + p.relative_to(base).as_posix()
            for p in base.rglob("*") if p.is_file()
        }
    if actual_names != set(expected):
        missing = sorted(set(expected) - actual_names)[:5]
        extra = sorted(actual_names - set(expected))[:5]
        print(f"UI 文件清单变化：{rel_dir}; 缺少={missing}; 新增={extra}", file=sys.stderr)
        sys.exit(1)
    for name, digest in expected.items():
        path = base / name.removeprefix("./")
        actual = hashlib.sha256(path.read_bytes()).hexdigest()
        if actual != digest:
            print(f"UI 文件与生产基线不同：{rel_dir}/{name}", file=sys.stderr)
            sys.exit(1)
PY
  then
    echo "拒绝发布：前后台源码或嵌入静态资源与锁定的生产 UI 不一致。只有用户明确授权改 UI 后，才可设置 APPROVE_UI_CHANGE=YES。" >&2
    exit 5
  fi
else
  echo "已显式确认本次包含 UI 改动；仍需先确认页面差异属于用户授权范围。" >&2
fi

ssh_key="${PRODUCTION_SSH_KEY:-$HOME/.ssh/node1_vnc_ed25519}"
live_hash="$(ssh -i "$ssh_key" -o BatchMode=yes -o ConnectTimeout=10 root@66.94.122.149 \
  "sha256sum /opt/dujiao-next/dujiao-next | cut -d' ' -f1")"
expected_lower="$(printf '%s' "$EXPECTED_LIVE_BINARY_SHA256" | tr '[:upper:]' '[:lower:]')"
if [[ "$live_hash" != "$expected_lower" ]]; then
  echo "拒绝发布：生产版本已变化。期望 $EXPECTED_LIVE_BINARY_SHA256，实际 $live_hash。请重新核对生产基线。" >&2
  exit 6
fi

echo "发布前检查通过：工作树干净，用户端资源基线符合策略，生产二进制与预期一致。"
