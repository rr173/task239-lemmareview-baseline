基于 Go 实现的数学证明引理依赖复核 Web 项目，一款后端服务，完成推导步骤前提可达性传播、引理依赖图循环检测与缺失引用标记，以及证明版本冻结快照发布。

# BENZHI 评测说明

## 项目类型
数学证明引理依赖复核台：形式化数学研究者导入定理、推导步骤与引理前提，服务构建依赖图、传播可用前提、检测循环引用与未覆盖假设，研究者可替换引理、记录假设豁免并发布冻结证明版本。

## 标准命令
- 构建：`CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...`
- 静态检查：`CGO_ENABLED=0 GOTOOLCHAIN=local go vet ./...`
- 测试：`CGO_ENABLED=0 GOTOOLCHAIN=local go test ./...`
- 启动服务：`./lemmareview --addr :8080 --db lemmareview.db`
- 自检（不驻留）：`./lemmareview --smoke-test`

## API 前缀
所有 HTTP 路由以 `/api` 开头，详见 README.md「API 入口」章节。

## Docker 双架构
使用仓库内置脚本建立双架构基线证明：
```bash
python3 scripts/docker_baseline_validation.py --project-dir <项目根> --verify-and-record
```
构建镜像需设置 `GOPROXY=https://goproxy.cn,direct`、`GOSUMDB=sum.golang.google.cn`、`CGO_ENABLED=0`、`GOTOOLCHAIN=local`。

## --smoke-test 契约
`--smoke-test` 不启动长驻服务，而是真实创建草稿、写入推导步骤与引理、执行前提传播与循环检测、冻结版本，关闭并重新打开 SQLite 数据库验证持久化与重启恢复，最终以退出码 0 结束。该行为是 Docker `CMD` 与双架构验证的唯一判据。
