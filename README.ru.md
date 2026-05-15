# Lenker

[English](README.md) | Русский

Lenker - early-stage open-source VPN ecosystem для провайдеров и пользователей.

Это ещё не готовый VPN-сервис. Сейчас репозиторий сфокусирован на backend foundation для панели провайдера, managed nodes, подписок и одного MVP protocol path: `VLESS + Reality + XTLS Vision`.

## Что Такое Lenker

Lenker должен стать self-hosted стеком для VPN-операций:

```text
provider panel -> node agent -> subscriptions -> client app -> future marketplace
```

Первый продуктовый milestone, `MVP v0.1`, намеренно узкий. Его цель - доказать, что провайдер может управлять пользователями, тарифами, подписками и нодами до того, как проект начнёт расти в billing, marketplace, multi-protocol support или production client distribution.

## Текущий Статус

Lenker находится в активной foundation-разработке.

Сейчас в репозитории есть:

- foundation для `panel-api`;
- bcrypt password verification для admin auth;
- admin session middleware через `Authorization: Bearer <session_token>`;
- admin CRUD slice для users, plans и subscriptions;
- local dev bootstrap для создания первого admin;
- PostgreSQL migrations для identity, subscriptions и node foundation tables;
- OpenAPI draft и lightweight validation;
- GitHub Actions для backend и OpenAPI checks;
- foundation для `node-agent`;
- первый contract между `panel-api` и `node-agent` для registration и heartbeat.

Пока не готово:

- production VPN runtime;
- реальный Xray process control;
- signed config deployment;
- реальный rollback executor;
- полный mTLS/certificate lifecycle;
- production client app.

## MVP v0.1 Scope

Входит в `MVP v0.1`:

- provider panel backend;
- node agent;
- users, plans и subscriptions;
- node registration и heartbeat;
- PostgreSQL-backed state;
- REST API и OpenAPI draft;
- manual renewal, API и webhook foundation;
- целевые client app платформы: Android, Windows, macOS;
- один production protocol path: `VLESS + Reality + XTLS Vision`.

Явно не входит в `MVP v0.1`:

- marketplace implementation;
- built-in billing или payment processing;
- provider ranking, reviews или commission flow;
- Telegram bot как core module;
- iOS или Linux client;
- production multi-protocol support;
- white-label builds;
- enterprise SSO;
- migration tools from other panels;
- full analytics или support ticketing.

## Что Работает Сейчас

Backend foundation:

- `GET /healthz`;
- `POST /api/v1/auth/admin/login`;
- admin-protected users API;
- admin-protected plans API;
- admin-protected subscriptions API;
- `POST /api/v1/nodes/register`;
- `POST /api/v1/nodes/{id}/heartbeat`.

Node-agent foundation:

- `GET /healthz`;
- `GET /status`;
- env-based config loading;
- registration payload builder;
- heartbeat payload builder;
- config revision и rollback placeholder models.

Local tooling:

- PostgreSQL migrations через `golang-migrate/migrate`;
- first-admin bootstrap CLI;
- OpenAPI validation;
- unit и contract tests;
- GitHub Actions CI.

## Структура Репозитория

```text
.
├── apps/
│   ├── client-app/
│   └── panel-web/
├── docs/
│   ├── adr/
│   ├── openapi/
│   ├── MVP_SPEC.md
│   ├── api.md
│   ├── architecture.md
│   ├── business-model.md
│   ├── database.md
│   └── roadmap.md
├── migrations/
├── scripts/
├── services/
│   ├── node-agent/
│   └── panel-api/
├── Makefile
├── README.md
├── README.ru.md
├── go.work
└── package.json
```

## Быстрый Старт

Для текущей backend-разработки нужны:

- Go 1.22+;
- Ruby для lightweight OpenAPI validator;
- PostgreSQL;
- `golang-migrate/migrate`.

Задайте local database URL:

```sh
export LENKER_DATABASE_URL='postgres://lenker:lenker@localhost:5432/lenker?sslmode=disable'
export LENKER_DATABASE_PING=true
```

Примените migrations:

```sh
make migrate-up
```

Создайте первого local admin:

```sh
ADMIN_EMAIL=owner@example.com ADMIN_PASSWORD='change-me-now' make bootstrap-admin
```

Запустите panel API:

```sh
make run-panel-api
```

Запустите node-agent foundation:

```sh
make run-node-agent
```

Более полный local flow с curl-примерами есть в [services/panel-api/README.md](services/panel-api/README.md).

## Проверки И CI

Запустить все текущие проверки:

```sh
make test
```

Эта команда запускает:

- `go test ./...` в `services/panel-api`;
- `go test ./...` в `services/node-agent`;
- OpenAPI validation для `docs/openapi/panel-api.v1.yaml`.

Отдельные команды:

```sh
make test-panel-api
make test-node-agent
make openapi-lint
```

GitHub Actions запускает `make test` на push и pull requests.

## Документация

- [MVP spec](docs/MVP_SPEC.md)
- [Architecture](docs/architecture.md)
- [Database model](docs/database.md)
- [REST API plan](docs/api.md)
- [OpenAPI draft](docs/openapi/panel-api.v1.yaml)
- [OpenAPI notes](docs/openapi/README.md)
- [Roadmap](docs/roadmap.md)
- [Business model boundary](docs/business-model.md)
- [Architecture decision records](docs/adr/README.md)
- [panel-api README](services/panel-api/README.md)
- [node-agent README](services/node-agent/README.md)

## Business Model Boundary

Lenker планируется как open-source core с коммерческими сервисами вокруг него, а не как искусственно урезанная self-host demo.

Self-hosted core должен оставаться полезным для маленьких провайдеров. Будущие коммерческие направления могут включать Lenker Cloud, managed nodes, paid support, enterprise governance, billing plugins, migration services и marketplace trust services.

Проект не должен монетизировать user data, DNS history, browsing history, connection logs, hidden telemetry, provider logs или pay-to-win marketplace ranking.

Marketplace и billing не входят в `MVP v0.1`.

## Security And Privacy Stance

Lenker должен быть privacy-first по умолчанию:

- minimal logging by default;
- без продажи user data или traffic history;
- без hidden telemetry;
- без billing или marketplace tables в MVP schema;
- session и node tokens хранятся как hashes там, где это уже реализовано;
- full mTLS и certificate rotation запланированы, но ещё не завершены.

Этот репозиторий пока не production-hardened. Не воспринимайте его как готовую secure VPN platform.

## Roadmap

Текущее направление:

1. Завершить backend foundation для provider operations.
2. Довести foundations для node registration, heartbeat, config revision, apply и rollback.
3. Добавить panel web flows для admins.
4. Добавить client app flow для Android, Windows и macOS.
5. Усилить release packaging, deployment docs, security policy, backup и recovery.

Post-MVP темы вроде marketplace, provider verification, billing adapters, Lenker Cloud, paid support и enterprise features остаются более поздней работой.

## Contributing Status

Проект пока не готов к широким внешним contributions. Полезны early issues, architecture feedback, security concerns и focused backend review.

Перед большими PR нужно сверяться с фиксированным `MVP v0.1` scope и не добавлять marketplace, billing, multi-protocol runtime или production VPN logic раньше времени.

## License Note

Финальная license policy ещё не зафиксирована полностью.

Текущая рекомендация:

- AGPL-3.0 для `panel-api` и `node-agent`;
- совместимая open-source license для `panel-web` и `client-app`;
- permissive licensing для будущих SDK/specs, где это полезно;
- отдельная trademark policy для имени и логотипа Lenker.

Не считайте licensing финальным, пока не добавлены `LICENSE` и license ADR.
