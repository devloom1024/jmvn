## 变更类型（PR 标题前缀）

> 标题请使用 conventional commit 格式：`type: 描述`。合并后会自动触发 release-please 发版。
> `feat:` 和 `fix:` 会触发自动发版，其他类型仅在 CHANGELOG 中记录。

- [ ] `feat:` 新功能（触发发版，minor bump）
- [ ] `fix:` Bug 修复（触发发版，patch bump）
- [ ] `feat!:` / `fix!:` 破坏性变更（触发发版，major bump）
- [ ] `docs:` 文档
- [ ] `refactor:` 重构
- [ ] `test:` 测试
- [ ] `chore:` 工程配置

## 变更说明

<!-- 简要描述本次变更做了什么、为什么这样做 -->

## 测试计划

<!-- 描述如何验证本次变更：手动测试步骤 / 自动化测试覆盖情况 -->

- [ ] `go test ./...` 通过
- [ ] 手动验证：
