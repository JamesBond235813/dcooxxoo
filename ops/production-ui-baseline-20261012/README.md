# 生产前端锁定基线（2026-10-19 刷新）

## 基准版本

当前线上运行版本是唯一 UI 基准。本次对 `https://ai.ooxxoo.cn` 和本机线上服务 `127.0.0.1:8080` 读取实际页面与静态资源，并与本地生产源码树、云服务器 `/opt/dujiao-next-src` 逐项核对。

- 线上运行二进制 SHA-256：`416a705e9cd5261d2bf96914bbb6a820be90035b889f38d06d92b115585aaa8f`
- 当前页面入口 `/` SHA-256：`82b312f9adf36f4135c53c6b68e4b03df724e262df9a188bb65acff629cceca8`，引用 `/assets/index-DfzD-M5Q.js`、`/assets/index-OIFmFRTC.css`
- 管理端 `/admin/` 会在运行时插入 `<base href="/admin/">`；源码模板使用 `__DJ_ADMIN_BASE__`。
- 4,221 个用户端和管理端静态资源均通过线上 HTTP SHA-256 校验。
- 本地与云端 `frontend/user`、`frontend/admin` 的 691 个源码文件逐项一致；该线上快照亦为本地仓库中的基准。
- `baseline-git-commit.txt` 指向共同祖先 `3147f8b`，本地与 GitHub `main` 均从该快照派生。以此避免旧工作树的提交祖先关系造成误拒；文件清单仍严格锁定源码及 dist 内容。

四份 SHA256 清单是精确校验依据。今后若线上 UI 被正式更新，应从已核实线上运行版本重新生成清单并更新本目录，不能让旧基线阻挡与当前生产版本相同的代码，也不能把未核实的候选 UI 当成新基准。

## 发布门禁

1. 只从干净、由当前仓库主线派生的隔离工作树构建。
2. 发布前运行 `scripts/assert-production-release-safe.sh`，并设置从线上刚读取的 `EXPECTED_LIVE_BINARY_SHA256`。
3. 门禁比较完整源码与嵌入资源清单，同时校验工作树干净、HEAD 含有共同基线祖先、线上二进制哈希未漂移。
4. 用户授权 UI 改动时，仍需审阅改动范围；不得用授权标志掩盖未审阅 UI 差异。
5. 发布前备份；完成后核对健康检查、页面与相关资源。

清单：`user-source.SHA256SUMS`、`admin-source.SHA256SUMS`、`user-dist.SHA256SUMS`、`admin-dist.SHA256SUMS`。`runtime-binary.sha256` 记录采样时运行版本，门禁使用操作者实时读取的值。
