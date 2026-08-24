# 数学证明引理依赖复核台 (task239-lemmareview)

形式化数学研究者导入定理、推导步骤与引理前提，服务构建依赖图、传播可用前提、检测
循环引用与未覆盖假设，研究者可替换引理、记录假设豁免并发布冻结证明版本。

## 业务闭环

1. 创建证明草稿（编辑中）。
2. 导入推导步骤文本（解析为有序步骤）。
3. 登记引理并标记为可用。
4. 声明前提依赖边（步骤依赖引理或前置步骤结论）。
5. 执行覆盖分析：传播前提、标记覆盖/缺失/循环。
6. 记录假设豁免、替换引理。
7. 冻结版本（绑定引理指纹与图快照）。

## 实体与状态

- 证明草稿：`editing`→`reviewing`→`gap`/`publishable`→`frozen`
- 推导步骤：`parsed`→`covered`/`missing`/`cyclic`
- 引理：`candidate`→`available`/`replaced`/`invalid`
- 证明版本：`draft`→`shared`/`frozen`→`superseded`

## 标准命令

```bash
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...
./lemmareview --addr :8080 --db lemmareview.db   # 启动服务
./lemmareview --smoke-test                       # 自检（不驻留）
```

## API 入口（前缀 /api）

| 能力 | API 入口 | 生产实现 |
| --- | --- | --- |
| 健康检查 | `GET /api/health` | httpapi/server.go |
| 草稿列表/新建 | `GET/POST /api/drafts` | service.CreateDraft → store |
| 草稿查询/改状态 | `GET/PUT /api/drafts/{id}` | service.GetDraft → store |
| 步骤列表/导入 | `GET/POST /api/drafts/{id}/steps` | service.ImportSteps → parse+store |
| 步骤查询 | `GET /api/drafts/{id}/steps/{sid}` | store.GetStep |
| 引理列表/新建 | `GET/POST /api/drafts/{id}/lemmas` | service.CreateLemma → store |
| 引理替换 | `POST /api/lemmas/{lid}/replace` | service.ReplaceLemma → store |
| 前提边列表/新增 | `GET/POST /api/drafts/{id}/premises` | service.AddPremise → store |
| 覆盖分析 | `POST /api/drafts/{id}/analyze` | service.Analyze → graph+cover |
| 豁免列表/新增 | `GET/POST /api/drafts/{id}/exemptions` | service.AddExemption → store |
| 版本列表/冻结 | `GET/POST /api/drafts/{id}/versions` | service.FreezeVersion → version+store |
| 版本查询 | `GET /api/versions/{vid}` | store.GetVersion |
| 版本共享 | `POST /api/versions/{vid}/share` | service.PublishShared |
| 版本替代 | `POST /api/versions/{vid}/supersede` | service.SupersedeVersion |
| 统计 | `GET /api/drafts/{id}/stats` | service.List* |
| 自检 | `GET /api/selfcheck` | httpapi.handleSelfCheck |

## 关键不变量

- 前提边禁止自指（步骤不能依赖自身结论）。
- 步骤顺序约束：依赖方 seq 必须严格大于被依赖方 seq。
- 冻结草稿/版本不可写入（拒绝 ImportSteps、AddPremise、FreezeVersion）。
- 版本冻结时绑定引理集合 SHA-256 指纹与图快照，不可变可追溯。

## 持久化与重启恢复

使用 SQLite（modernc.org/sqlite，纯 Go、CGO 无关），WAL 模式。所有实体落库；
`--smoke-test` 会关闭并重新打开同一数据库验证持久化与重启恢复。
