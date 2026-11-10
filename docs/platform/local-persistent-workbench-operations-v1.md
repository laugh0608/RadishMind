# 本机持久工作台运行指南 v1

更新时间：2026-10-31

状态：命令、配置、Web 接线和运维实现已完成；2026-10-31 宿主机 Chrome 的首次使用、Prompt / Workflow、显式保存、冷重启与新目录恢复已完成固定响应验收，证据见 [W44 周志](../devlogs/2026-W44.md#2026-10-31-宿主机工作台验收与修复)。另行授权的 `linux/arm64` Docker 构建 / 运行、挂载限制、重启及容器与宿主机备份迁移也已通过技术验收，见 [Docker 验收记录](../devlogs/2026-W44.md#2026-10-31-docker-工作台交付验收)。`linux/amd64`、Windows PowerShell 和独立用户试用尚未验收；真实模型窗口继续暂停。

范围与停止线见[交付专题](../features/user-workspace/local-persistent-workbench-delivery-v1.md)。工作台只面向回环地址上的内部开发者预览，使用既有本地 Session 和 SQLite。Prompt 与基本 Workflow 共用同一配置；旧开发入口、PostgreSQL 和生产门禁保持各自边界。

## 宿主机准备与启动

需要已安装 Go 1.25、Node.js 22、npm，以及 Python 3.12 或更新版本的虚拟环境，Python 环境中必须有 `requirements-dev.txt` 中的依赖。Web 需要已有 `node_modules`，Go 模块需已在缓存。启动脚本不安装依赖；缺失时明确失败。依赖安装或下载仍按协作规则单独授权。以下从仓库根目录执行。

```bash
mkdir -p var/workbench/secrets var/workbench-backups
cp deploy/workbench-native.example.json var/workbench/platform.json
./start.sh workbench build
./start.sh workbench check --config var/workbench/platform.json
./start.sh workbench init --config var/workbench/platform.json
./start.sh workbench serve --config var/workbench/platform.json
```

`build` 记录当前 Git 提交（未提交修改附加 `-dirty`），构建正式 Web 和 `tmp/workbench-bin/radishmind-platform`，Go 使用 `GOPROXY=off`、`GOTOOLCHAIN=local`，不会自动下载依赖。Web 产物含工作台标记；普通开发构建不能直接交给工作台服务。`check` 检查配置、Web 产物、Python 版本 / `jsonschema` 与桥接脚本，不创建数据、不打开数据库、不启动工作进程、不调用模型。检查成功不能证明数据库或提供方可用。更新构建前先停工作台，构建后完整刷新浏览器；运行中替换 `dist` 会使旧页面继续请求已移除的哈希分块。

样例的 Python 路径为相对配置目录的 `../../.venv/bin/python`。Windows 将其改为 `../../.venv/Scripts/python.exe`，入口改用 `pwsh ./start.ps1 workbench <动作> ...`；其余 Go 参数相同。PowerShell 包装脚本与 Windows 交叉编译是不同验证，当前未执行 PowerShell 运行验收。

`init` 只接受不存在的数据库，复用全部现有迁移，不创建用户、权限或演示应用。`serve` 在前台运行；默认打开 `http://127.0.0.1:4100`。Ctrl+C 停止 HTTP 和 Python 工作进程，等待命令退出后再操作数据。再次使用只执行 `serve`，不要重复 `init`。服务拒绝缺库、待迁移、未知 / 被修改的迁移记录及错误 Web 构建。

## 首次账户与工作区

1. 启动后在 Web 注册自己的本地账户。普通注册不会自动获得管理员权限，页面显示账户 ID 和配置的租户 / 工作区。
2. 停止前台服务并等待退出，用该账户 ID 执行以下一次性初始化。
3. 再次启动，登录或刷新原有会话进入工作区。首次目录为空，由用户建立 Prompt / Workflow 应用。

```bash
./start.sh workbench bootstrap-admin --config var/workbench/platform.json \
  --user-id '<页面显示的账户 ID>' --audit-ref 'audit:local-workbench-first-admin'
./start.sh workbench serve --config var/workbench/platform.json
```

初始化复用既有身份负责模块和审计语义。运行中、账户不存在、已有管理员或重复初始化均拒绝；没有 HTTP 自动提权入口。初始化只使用配置中的租户和工作区，不接受另传作用域绕开配置。

首次配置 Prompt 前，先在当前应用的“API 与密钥”签发包含 `models:read` 的应用密钥，显式交接到 Playground 并加载模型目录，再返回配置页选择已校验模型。该读取不会调用模型；Web Session 不代替 Gateway API 密钥。离开 Playground 清除原始密钥，重新交接需使用新签发的一次性凭据。密钥的调用权限仅按实际任务选择。

普通 Prompt 通过既有草案、模板版本、审查和显式运行流程使用；基本 Workflow 从工作区新建草案、校验、保存、Definition 候选审查、显式激活后运行；结构化档案使用活动版本的输入合同。需要会话结果保留时，在应用交互中选择结构化 Workflow 档案及已激活定义 ID，创建绑定会话，并在本轮执行前显式勾选保存。再次打开时先选择应用，从草案库、Definition 或会话列表精确重读；不必新建草案或重新初始化。应用会话、运行元数据与显式 Result Artifact 各司其职；需要保留答案正文时使用既有结果保存动作，不能把“运行成功”理解为所有正文已经保存。重启只重读已有资源，不重跑历史输入。

工作台首页仅读取真实资源。Agent、RAG、HTTP Tool、Schedule、评测批次、管理路由、配额和计费扩展不在此运行档启用；旧开发演示模式继续提供各自已有能力。页面复用既有组件及语义 token，不另建产品布局。

## 配置与模型

优先级为代码默认值、`local-workbench` 默认值、显式 JSON、显式环境变量，最后验证运行档不变量。选择工作台时必须指定配置文件；未知字段、重复键、`null`、身份 / 存储 / 功能组合冲突直接失败。`--config` 优先选择文件，未传时取 `RADISHMIND_PLATFORM_CONFIG`，再默认 `var/workbench/platform.json`。包装脚本从仓库根目录执行。

| 字段 / 覆盖项 | 含义 |
| --- | --- |
| `runtime_profile` / `RADISHMIND_RUNTIME_PROFILE` | 固定 `local-workbench`；文件与环境选择不同档时拒绝 |
| `listen_addr` / `RADISHMIND_PLATFORM_LISTEN_ADDR` | 宿主机回环监听；改端口时同步公开来源 |
| `sqlite_dev_database_path` | 相对配置目录的数据文件；宿主机样例为 `data/radishmind.db` |
| `workbench.public_origin` / `RADISHMIND_WORKBENCH_PUBLIC_ORIGIN` | 浏览器来源，含端口；必须是 HTTP 回环地址，显式端口不能为 `80` 或带前导零 |
| `workbench.tenant_ref` / `RADISHMIND_WORKBENCH_TENANT_REF` | 当前租户，不创建成员权限 |
| `workbench.workspace_id` / `RADISHMIND_WORKBENCH_WORKSPACE_ID` | 当前工作区，不替代服务端授权 |
| `workbench.environment` / `RADISHMIND_WORKBENCH_ENVIRONMENT` | `development` 或 `test` |
| `workbench.web_assets_path` / `RADISHMIND_WORKBENCH_WEB_ASSETS_PATH` | 已构建 Web 目录，相对配置目录解析 |
| `python_binary` | 可执行文件名按 PATH 查找；带目录的相对路径以配置目录解析 |
| `bridge_script` | 沿用平台脚本路径规则，包装脚本以仓库为工作目录；迁移宿主机后核对 |
| `api_key_file` / `RADISHMIND_PLATFORM_API_KEY_FILE` | 相对配置目录的凭据文件，禁止与 `RADISHMIND_PLATFORM_API_KEY` 同时提供 |

其它已有平台环境变量仍经过统一加载器，完整列表见[平台运行手册](platform-service-operations-runbook-v1.md)。工作台不接受逐组件存储选择或把必需功能关闭，也不接受开启范围外能力。Docker 镜像明确设置 `RADISHMIND_WORKBENCH_CONTAINER=1` 才允许内部 `0.0.0.0`，宿主机仍只映射回环端口；这不代表共享网络部署准入。

默认 `mock` 不联网，只提供有限固定响应，用于技术链验证。五字段诊断等严格结构化正向场景需要明确选择隔离固定响应提供方；产品不会隐式启动它。真实任务效果必须在恢复模型窗口后单独验证。

显式远端接入将 `provider` 设为 `openai-compatible`，填写 `model`、`base_url` 和 `api_key_file`。JSON 不允许 `api_key`，Web 公开配置也不包含凭据、文件路径或提供方 URL。凭据文件由用户单独准备，限制读取权限；容器使用 `/run/secrets/radishmind/provider-key`。启动才读取值，轮换后重启生效。工作台 Python 不读取仓库 `.env`，并隔离继承的模型环境配置；不会自动探测、换模型、下载权重或回退到 mock。

## 冷备份、升级和恢复

先停止服务并等待退出。OS 文件锁覆盖服务、初始化、身份初始化、备份、升级和恢复，不依赖容器 PID；锁文件 `.workbench.lock` 保留在数据目录，进程退出后由 OS 释放锁，不要删除锁文件来强行绕过占用。旧开发入口拒绝打开带工作台标记的数据目录；工作台也拒绝接管无标记的旧库目录。只有显式 `init` 新库或 `restore` 新目录可以建立标记，不能给活动旧库手工补标记来绕过检查。

```bash
./start.sh workbench backup --config var/workbench/platform.json \
  --output var/workbench-backups/manual-001
./start.sh workbench upgrade --config var/workbench/platform.json \
  --output var/workbench-backups/before-upgrade-001
./start.sh workbench restore --source var/workbench-backups/manual-001 \
  --output var/workbench-restored-001
```

备份与恢复目标必须不存在，父目录须先准备。备份用 SQLite `VACUUM INTO` 生成一致快照，经完整性、外键、迁移和 SHA-256 校验后最后写入清单。产物为 `radishmind.db`、`platform.json`、`manifest.json`；数据库含用户材料，应私有保存。普通备份不含凭据文件，远端配置改为恢复目录中的 `secrets/provider-key` 引用。来源版本在清单中保留，不替代实际迁移集合兼容检查。

`upgrade` 只允许当前二进制认识的迁移子集，先成功冷备份，再调用原迁移执行器。未知或校验不符的迁移不能升级；迁移失败保留数据库、备份和错误，不自动启动或回退。未完成的备份目录没有有效清单，不能恢复。必要时从备份恢复到另一目录并使用匹配版本调查，不覆盖唯一旧库。

2026-10-31 新增的 `development_attempts` 与 `development_artifacts` 已纳入共享迁移集合；此前工作台库使用新构建时也须遵循上述显式升级流程。已有宿主机 / 容器验收记录针对当时迁移集，本轮受控开发测试只使用隔离数据库，未替日常库升级。新增表不会自动启用开发通道、SDK 或代码容器。

恢复输出包含 `platform.json` 与 `data/radishmind.db`，保留原库和备份。启动前核对 Python、桥接脚本、Web 目录、监听来源和凭据，再执行 `check`。已知但不完整的迁移集需先 `upgrade`，才能 `serve`。恢复不会自动切换日常配置。

```bash
./start.sh workbench check --config var/workbench-restored-001/platform.json
./start.sh workbench serve --config var/workbench-restored-001/platform.json
```

恢复的是备份时点，之后的成员撤销、密钥吊销和应用修改不会保留。恢复后重新登录并核对账户、成员和 API key；本轮不实现跨时间权限合并。数据库中的历史执行不会自动重放。

## Docker 准备与启动

以下需明确 Docker 构建 / 运行授权。构建会拉取基础镜像并下载 npm、Go 与 Python 依赖；不下载模型，不发布镜像。现有 Platform / Console Compose 不变，工作台使用独立项目。

```bash
mkdir -p deploy/workbench/data deploy/workbench/secrets deploy/workbench-backups
cp deploy/workbench-docker.example.json deploy/workbench/platform.json
cp deploy/.env.workbench.example deploy/.env.workbench
```

编辑 `.env.workbench`，将 `RADISHMIND_WORKBENCH_BUILD_VERSION` 设为本次源代码提交标识；有未提交修改时附加 `-dirty`，默认 `local-workbench-unversioned` 只表示未标识构建。隔离验收时使用独立项目名 `-p <本次项目名>`，避免接管已有工作台；挂载路径建议填写绝对路径，相对路径按 Compose 文件所在目录解析。再使 UID / GID 与准备好的目录所有者一致，例如使用 `id -u` / `id -g` 的结果；不要直接照抄 `10001` 后放宽所有目录权限。运行用户需能读取配置 / 凭据并写入数据与备份目录。目录缺失不会被 Compose 自动创建；权限错误停止，由用户修正目录，不自动递归改所有者。

```bash
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml config --quiet
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml build
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml run --rm --no-deps workbench check --config /etc/radishmind/platform.json
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml run --rm --no-deps workbench init --config /etc/radishmind/platform.json
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml up --no-build
```

服务前台运行，内部端口 `7000`，宿主机默认回环 `4100`。改宿主机端口时同时改 JSON 的 `workbench.public_origin`；两者不相互猜测。Web、公开配置和 API 同源，不需要 Vite 或 Node 常驻。首次注册后 Ctrl+C，确认容器停止，再执行一次性初始化与重新启动：

```bash
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml stop
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml run --rm --no-deps workbench bootstrap-admin --config /etc/radishmind/platform.json --user-id '<账户 ID>' --audit-ref 'audit:docker-workbench-first-admin'
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml up --no-build
```

根文件系统只读，只有数据目录和临时目录可写；配置 / 凭据只读挂载。容器非 root、无额外 capability、不自动重启，停止宽限期为 30 秒，日志有界。健康检查不调用模型，只证明服务已通过启动检查并回应 `/healthz`。2026-10-31 已实际验证本机 `linux/arm64` 构建 / 运行、正常停止、挂载权限与只读限制；`linux/amd64` 仍待实际验收。健康检查通过不能代替 Prompt / Workflow 执行、显式结果保存或独立用户验收。

## Docker 备份与载体迁移

停服后使用相同 Compose 配置，给一次性命令额外挂载备份目录。以下使用 POSIX Shell 的绝对路径；PowerShell 使用对应绝对路径参数。

```bash
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml stop
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml run --rm --no-deps -v "$PWD/deploy/workbench-backups:/backups" workbench backup --config /etc/radishmind/platform.json --output /backups/manual-001
docker compose --env-file deploy/.env.workbench -f deploy/docker-compose.workbench.yaml run --rm --no-deps -v "$PWD/deploy/workbench-backups:/backups" workbench restore --source /backups/manual-001 --output /backups/restored-001
```

容器恢复后，在宿主机审阅 `deploy/workbench-backups/restored-001/platform.json`。若继续容器运行，把数据库路径改为 `/var/lib/radishmind/radishmind.db`、Web 目录改为 `/workspace/web`、Python 改为 `python3`、桥接脚本改为 `/workspace/scripts/run-platform-bridge.py`；监听改为 `0.0.0.0:7000`，公开来源与宿主机映射一致。远端凭据引用改为 `/run/secrets/radishmind/provider-key`，并单独准备凭据挂载。

然后把 Compose 的配置挂载切到该配置文件，数据挂载切到其 `data/` 目录，凭据挂载切到审阅后的目录，执行一次性 `check` 再显式 `up`。不要把仅挂载配置文件时的 `data/radishmind.db` 误当成持久卷路径。

转到宿主机时，数据库使用恢复目录中的 `data/radishmind.db`，监听与公开来源使用相同回环地址，Python / 脚本 / Web 改为宿主机真实路径。宿主机与容器通过备份恢复迁移，不同时共享原数据库。`docker compose ... down` 仅清理容器 / 网络；数据、凭据、备份和本地镜像保留，删除这些材料需要另行明确范围。

## 常见失败与验收记录

| 现象 | 处理 |
| --- | --- |
| 公开配置失败或错误来源 | 检查浏览器地址与 `public_origin`，修复后重试；不回退旧端口或样例 |
| 数据目录被占用 | 找到并停止持有数据的进程 / 容器，等待退出；不删锁、不复制活动数据库 |
| 目录不可写 / 凭据不可读 | 核对实际运行 UID / GID、挂载与路径权限，不改用临时库 |
| 缺库 / 待迁移 | 新目录显式 `init`；已有库先备份并显式 `upgrade`，不覆盖 |
| 模型失败 / 会话失效 / CAS 冲突 | 使用既有失败分类，重新登录或重读版本后由用户决定重试，不自动重放 |
| 中断后重启 | 重读已持久化资源；未显式保留的正文和中断执行不能宣称恢复 |

真实验收应记录配置摘要、端口、隔离目录、构建版本、账户与应用操作、Prompt / Workflow 结果、关闭重开、恢复后重读及进程清理。固定响应只验证技术链；真实模型质量和独立用户日常可用性另记，不合并成生产声明。
