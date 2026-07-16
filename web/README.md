# HomeFlow 运营后台

这是 HomeFlow 到家云的 Vue 3 运营工作台，负责经营概览、订单调度和服务团队运营。

## 本地运行

```bash
npm install
npm run dev -- --host 0.0.0.0 --port 4310
```

默认不配置 API 时，打开页面会进入离线演示入口；演示数据只保存在浏览器内存中，不会请求后端。

## 接入真实 API

启动 `../deploy/docker-compose.yml` 中的 API 后，以环境变量指定地址：

```bash
VITE_API_BASE_URL=http://localhost:8080 npm run dev -- --host 0.0.0.0 --port 4310
```

配置 API 后，页面会先显示登录页。登录成功返回的 `data.accessToken` 会保存到浏览器 `localStorage.homeflow_access_token`，后续请求自动携带 Bearer Token；点击左下角退出按钮会调用登出接口并清理本地令牌。

本地服务的演示管理员账号（仅用于开发环境）为 `13900000000` / `demo123456`。生产环境请使用自己的账号，不要将账号、密码或 Token 写入仓库。

## 检查

```bash
npm test
npm run build
```
