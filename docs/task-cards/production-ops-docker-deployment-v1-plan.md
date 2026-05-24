# `Production Ops Docker Deployment` v1 计划

更新时间：2026-05-24

## 任务目标

本任务卡用于把 RadishMind 后续部署路线固定为 Radish 同类口径：开发期宿主机直跑，本地容器验证使用独立 local compose，测试和生产共用部署态 compose，并通过镜像轨道、固定 tag、公开 URL、密钥和外部反代配置区分环境。

参考项目默认使用在线仓库 `https://github.com/laugh0608/Radish`。本地外部项目路径只作为当次协作输入，不写入长期文档。

## 当前结论

- `host_dev`：继续使用现有 `scripts/run-platform-service.*`、`scripts/run-radishmind-console-dev.*`、`go run` 和 `npm run dev`，不走 Compose。
- `docker_local`：后续使用 `deploy/docker-compose.local.yaml`，允许本地 build，只用于容器构建和启动验证。
- `docker_test`：后续使用 `deploy/docker-compose.yaml`，默认使用 `RADISHMIND_IMAGE_TRACK=test` 或固定 `RADISHMIND_IMAGE_TAG` 拉取预构建镜像。
- `docker_prod`：后续复用 `deploy/docker-compose.yaml`，默认使用 `RADISHMIND_IMAGE_TRACK=release` 或固定 `RADISHMIND_IMAGE_TAG`，外部反代负责 HTTPS，容器内部优先 HTTP。
- 测试和生产的差异应收敛为镜像轨道 / 固定 tag、公开 URL、secret 来源、provider profile、反代证书、数据目录和日志目录。

## v1 范围

1. `docker-deployment-mode-definition`
   - 固定 local / test / prod 模式和停止线。
   - 明确只参考 Radish 的部署形态，不复制 Radish 的业务服务拆分。
   - 当前只形成模式定义，不创建可运行部署声明。
2. `docker-local-compose`
   - 新增 platform 和 console 的本地容器验证编排。
   - 允许本地 build，默认 mock provider。
   - 不接真实 executor、confirmation、writeback 或 replay。
3. `docker-test-prod-compose`
   - 新增测试 / 生产共用部署态 compose。
   - 通过 `RADISHMIND_IMAGE_TRACK=test/release` 或固定 `RADISHMIND_IMAGE_TAG` 区分镜像。
   - 生产必须等待 secret backend、正式 CORS / auth policy、外部反代和 provider health policy。
4. `docker-image-build-publish`
   - 为 platform 和 console 补 Dockerfile 与镜像命名规则。
   - CI 镜像发布规则后续对齐 tag 后缀：`v*-dev`、`v*-test`、`v*-release`。
5. `deployment-readiness-smoke`
   - 先做 `docker compose config` / 静态展开检查。
   - 再逐步引入本地容器 smoke、测试环境 smoke 和生产前复核记录。

## 非目标

- 不直接把当前本地 console 壳声明为 production console。
- 不在没有 runtime config / 静态资产策略前声明 console production packaging ready。
- 不把 Docker Compose 当作 secret backend。
- 不把容器 `restart` 策略当作完整 process supervisor。
- 不把 mock provider、本地 CORS 或 Vite preview 用于生产。
- 不直接修改 Radish、RadishFlow 或 RadishCatalyst 外部工作区。

## 停止线

- `docker_local` 只能证明本地容器验证可用，不代表测试或生产 ready。
- `docker_test` 可以使用测试 provider profile 和测试域名，但不能使用生产 secret。
- `docker_prod` 必须要求真实 secret 来源、外部 HTTPS 入口、正式 provider profile、生产 CORS / auth policy 和部署后 smoke。
- 在 Docker 资产落地前，P3 checklist 中的 production secret backend、process supervisor、deployment environment isolation 和 console production packaging 仍保持 `not_satisfied`。

## 下一步

推进 `docker-local-compose`：先补 platform / console 的 Dockerfile 草案和 `deploy/docker-compose.local.yaml`，并用静态 checker 固定不泄漏 secret、不提交 build 产物、不声明 production ready。
