# HomeFlow Admin

HomeFlow 到家服务预约、派单与运营管理平台的管理后台和 Go API。

## 目录

- `server/`：Go 1.25 + Gin API
- `web/`：Vue 3 + TypeScript 管理后台
- `deploy/`：MySQL 8.4、Redis 8 Docker Compose
- `docs/`：API 契约、部署与截图

项目采用 MIT 许可证。默认仅提供本地演示数据，不包含任何生产凭据。

## 本地运行后台

```bash
cd web
npm install
npm run dev
```

未配置 API 时，后台会自动使用内置演示数据，并在页面标记“演示数据”。接入 Go API 时，在 `web/.env.local` 配置：

```dotenv
VITE_API_BASE_URL=http://localhost:8080
```

前端适配器会请求 `/api/v1/dashboard/summary`、`/api/v1/admin/orders` 和 `/api/v1/admin/dispatch/recommendations`，并从 `localStorage.homeflow_access_token` 读取 Bearer Token。登录接口为 `POST /api/v1/auth/login`，成功后可将返回的 `data.accessToken` 写入该键；不要把账号、密码或 Token 提交到仓库。若前后端端口不同，请在 API 服务端配置允许前端开发地址的 CORS。

```js
localStorage.setItem('homeflow_access_token', data.accessToken)
```
