# HomeFlow API Contract v1

Base URL: `/api/v1`

## Response envelope

```json
{ "code": 0, "message": "ok", "data": {}, "traceId": "trace-..." }
```

`code: 0` means success. Errors use stable business codes such as `AUTH_REQUIRED`, `FORBIDDEN`, `VALIDATION_FAILED`, `SLOT_UNAVAILABLE`, `ORDER_STATE_INVALID`, and `IDEMPOTENCY_CONFLICT`.

## Auth

- `POST /auth/login` `{ "phone": "13800000000", "password": "demo123456" }`
- `POST /auth/refresh`
- `POST /auth/logout`
- `GET /auth/me`

The demo server accepts seeded accounts and returns a bearer token. Production SMS/login adapters are intentionally not included.

## Catalog and booking

- `GET /service-categories`
- `GET /services?categoryId=&page=&pageSize=`
- `GET /services/:id`
- `GET /addresses`
- `POST /addresses`
- `GET /availability?serviceId=&date=`
- `POST /orders` with `Idempotency-Key`, `{ "serviceId", "addressId", "date", "slotId", "remark" }`
- `GET /orders?page=&pageSize=&status=`
- `GET /orders/:id`
- `POST /orders/:id/cancel`
- `POST /orders/:id/confirm`
- `POST /orders/:id/review` `{ "rating": 1-5, "content": "..." }`

## Operations and dispatch

- `GET /dashboard/summary`
- `GET /admin/orders?status=&date=&keyword=`
- `POST /admin/orders/:id/assign` `{ "technicianId": "..." }`
- `GET /admin/dispatch/recommendations?orderId=`
- `GET /technicians`
- `POST /technicians`
- `PATCH /technicians/:id`
- `GET /technicians/:id/schedule?date=`
- `GET /services` / `POST /services` / `PATCH /services/:id`
- `GET /reviews`
- `GET /audit-logs`

派单推荐按以下可解释优先级排序：技能匹配、服务区域匹配、当前班次可用、当前订单负载（负载越低越优先）。调度员和管理员可派单；师傅只能操作分配给自己的履约订单。

## Technician workbench

- `GET /workbench/tasks?status=&date=`
- `POST /workbench/orders/:id/accept`
- `POST /workbench/orders/:id/arrive`
- `POST /workbench/orders/:id/start`
- `POST /workbench/orders/:id/proofs` multipart `before[]`, `after[]`, `note`
- `POST /workbench/orders/:id/complete`

## Order states

`pending_confirmation -> pending_dispatch -> assigned -> en_route -> serving -> pending_customer_confirmation -> completed`

Cancellation is allowed from `pending_confirmation`, `pending_dispatch`, and `assigned` with an audit event. Every state-changing request is authenticated, authorized, idempotent where retried, and recorded in `order_events` and `audit_logs`.

预约请求必须带 `Idempotency-Key`。服务端在事务边界内使用唯一业务键保存订单，并以 Redis 8 `SET NX EX` 对时段加短锁；MySQL 8.4 是订单、事件和审计日志的事实来源。开发环境无法连接数据库时 API 仍可使用内存演示模式，生产部署必须配置 `MYSQL_DSN`、`REDIS_ADDR` 和 `JWT_SECRET`。
