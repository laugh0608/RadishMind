# RadishMind Platform Service Layer

本目录承载 `Go` 平台服务层的最小骨架。

当前职责：

- 启动最小本地 `HTTP` 服务
- 承载 northbound `API` / `gateway` 入口
- 提供观测、部署壳和后续鉴权 / 流式转发落点

当前明确不做：

- 不在这里复制第二套业务真相源
- 不在这里重写模型推理、训练、评测或 `builder`
- 不绕过 `contracts/` 自定义另一套 canonical protocol

当前最小路由：

- `GET /healthz`
- `GET /v1/models`
- `POST /v1/chat/completions`

其中 `/v1/chat/completions` 目前只提供占位响应，用于固定服务层边界；真正的 canonical request / response 翻译与 Python runtime bridge 仍是下一步实现项。
