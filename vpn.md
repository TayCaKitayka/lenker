# Open-source VPN ecosystem

## Краткая суть

Цель - сделать open-source экосистему для VPN-провайдеров и пользователей, где есть:

- панель управления для провайдера;
- серверные ноды, которые можно добавлять, обновлять, мониторить и объединять в группы;
- выбор протоколов на уровне основного сервера, ноды, группы нод или конкретного тарифного плана;
- приложение для пользователя, которое провайдер может выдавать как свое, но при этом оно остается частью открытой экосистемы;
- самоуправление подпиской: обновить ключ, перевыпустить конфиг, посмотреть трафик, устройства, срок, статус серверов;
- каталог провайдеров по цене, отзывам, качеству, регионам, доступным протоколам и публичной статистике.

Главная идея: не очередная панель для выдачи ссылок, а стандарт для всей цепочки "провайдер -> инфраструктура -> подписка -> приложение -> отзывы -> выбор провайдера".

## Почему это нужно

Сейчас рынок VPN-панелей и клиентов раздроблен:

- провайдеры используют разные панели, ботов, Excel-таблицы, ручные ключи и самописные скрипты;
- пользователь часто получает просто ссылку, QR-код или инструкцию, а когда ключ не работает - идет в Telegram к поддержке;
- клиенты вроде Hiddify, Happ, v2rayN, NekoRay/NekoBox решают подключение, но не решают нормальную экосистему провайдеров;
- панели вроде Marzban, 3x-ui и Hiddify Manager решают управление пользователями, но не дают единого открытого marketplace и нормального пользовательского lifecycle;
- нет прозрачного рейтинга провайдеров: цена, uptime, скорость, перегрузка нод, политика логов и качество поддержки почти нигде не проверяются.

Этот проект должен закрыть разрыв между панелью, приложением и выбором провайдера.

## Что взять из существующих проектов

### Marzban

Что полезно взять:

- REST API как первый-class интерфейс, а не только веб-панель;
- управление сотнями пользователей с лимитами по трафику и сроку;
- несколько протоколов на одного пользователя;
- генерация подписок в разных форматах: Web, V2Ray, Sing-box, Clash, Outline;
- QR-коды и share-ссылки;
- Marzban-node как идея распределения нагрузки по разным серверам и локациям;
- несколько администраторов;
- мониторинг ресурсов и потребления трафика;
- CLI для обслуживания;
- Telegram bot как дополнительный канал, но не как основная архитектура.

Что сделать лучше:

- ноды должны быть не "дополнительной фичей", а центральной сущностью;
- провайдеру нужен нормальный lifecycle ноды: bootstrap, проверка окружения, установка агента, healthcheck, drain, обновление, rollback;
- подписка должна быть не просто ссылка, а объект с политиками, устройствами, правами, историей ключей и self-service действиями.

Источники:

- https://gozargah.github.io/marzban/en/docs/introduction
- https://gozargah.github.io/marzban/en/docs/marzban-node
- https://marzban-docs.sm1ky.com/components/subscriptions/

### 3x-ui

Что полезно взять:

- простую и понятную панель управления Xray;
- мониторинг статуса системы;
- поиск по inbound и клиентам;
- поддержку нескольких протоколов и нескольких пользователей;
- лимиты по трафику, сроку и IP;
- настраиваемые Xray-шаблоны;
- HTTPS для панели и автоматизацию SSL;
- импорт/экспорт базы;
- светлую/темную тему;
- понятную структуру inbounds/clients.

Что сделать лучше:

- 3x-ui сам предупреждает, что проект не стоит использовать как production-основу. В нашей системе production-готовность должна быть базовым требованием;
- нужна строгая RBAC-модель, аудит действий, 2FA, безопасное хранение секретов;
- нельзя хранить root-пароли серверов после установки ноды;
- нужна нормальная миграция конфигов, контроль версий и rollback;
- нужен публичный API и SDK, чтобы провайдеры могли интегрировать биллинг, сайты и CRM.

Источники:

- https://github.com/MHSanaei/3x-ui
- https://mhsanaei.github.io/3x-ui/
- https://github.com/MHSanaei/3x-ui/wiki

### Hiddify Manager и Hiddify App

Что полезно взять из Manager:

- multi-core подход: Xray и Sing-box;
- multi-domain подход;
- простую установку;
- автоматические обновления;
- автоматические бэкапы;
- Cloudflare/CDN-интеграции;
- управление активными конфигурациями;
- лимиты по времени и трафику;
- dedicated user page, где пользователь видит потребление и конфиги;
- smart proxy режимы: только заблокированные сайты, все кроме локальных сайтов, весь трафик;
- DoH, WARP и дополнительные обходные режимы как опциональные модули.

Что полезно взять из App:

- кроссплатформенность: iOS, Android, Windows, macOS, Linux;
- Sing-box как универсальный core;
- импорт подписки одним кликом/deeplink;
- поддержку разных subscription formats;
- показ потребления пользователю;
- выбор самой быстрой ноды по ping/latency;
- несколько профилей;
- TUN-режим и split routing.

Что сделать лучше:

- приложение должно уметь авторизоваться у провайдера, а не только импортировать ссылку;
- пользователь должен иметь self-service: обновить ключ, отвязать устройство, сменить регион, открыть тикет, посмотреть инциденты;
- провайдер должен иметь white-label режим без форка приложения;
- marketplace провайдеров должен быть встроен как отдельный режим, а не смешан с личными конфигами.

Источники:

- https://github.com/hiddify/Hiddify-Manager
- https://hiddify.com/
- https://hiddify.com/app/How-to-use-Hiddify-app/
- https://apps.apple.com/us/app/hiddify-proxy-vpn/id6596777532

### Happ

Что полезно взять:

- максимально простой пользовательский интерфейс;
- rule-based proxy;
- hidden subscriptions;
- encrypted subscriptions;
- локальное хранение конфигов и акцент на приватность;
- поддержку VLESS Reality, VMess, Trojan, Shadowsocks, Socks, Hysteria2;
- понятное предупреждение, что приложение не продает VPN само по себе.

Что сделать лучше:

- open-source статус должен быть ясным и проверяемым;
- скрытые/зашифрованные подписки нужно сделать стандартом экосистемы, а не отдельной фичей клиента;
- добавить связку с аккаунтом провайдера, устройствами, подпиской и marketplace.

Источники:

- https://www.happ.su/
- https://play.google.com/store/apps/details?id=com.happproxy
- https://apps.apple.com/ca/app/happ-proxy-utility/id6504287215

### v2rayN

Что полезно взять:

- сильный desktop-клиент для power users;
- поддержку Xray, Sing-box и других cores;
- subscription groups;
- импорт из URL, QR, clipboard и config-файлов;
- routing modes: system proxy, PAC, TUN, rule-based;
- DNS-настройки;
- speed/latency testing;
- поддержку разных протоколов: VMess, VLESS, Shadowsocks, Trojan, Hysteria2, TUIC, WireGuard, AnyTLS.

Что сделать лучше:

- для обычного пользователя не показывать всю сложность сразу;
- оставить "Advanced mode" для тех, кому нужны ручные routing/DNS/core настройки;
- в основном режиме приложение должно работать как нормальный VPN-продукт: login -> выбрать тариф/провайдера -> connect.

Источники:

- https://github.com/2dust/v2rayN/wiki/Description-of-subscription
- https://github.com/2dust/v2rayN/wiki/Description-of-system-proxy-routing

### NekoRay / NekoBox

Что полезно взять:

- sing-box first подход;
- TUN-режим;
- правила маршрутизации;
- тест задержки;
- поддержку SS, VMess, VLESS, Trojan, Hysteria2, TUIC, WireGuard;
- plugin-подход для отдельных протоколов;
- хороший пример того, что advanced-клиенты удобны для настройки и диагностики.

Что сделать лучше:

- NekoRay больше подходит техническим пользователям, а наш клиент должен быть понятен обычному человеку;
- нельзя ломать пользовательские маршруты при обновлениях: нужен config migration, backup перед обновлением и rollback;
- kill switch и защита от утечек должны быть не опцией для экспертов, а понятным режимом безопасности.

Источники:

- https://github.com/MatsuriDayo/nekoray/releases
- https://nekobox.pro/
- https://nekobox.pro/faq/what-is-nekobox/

## Главные принципы продукта

1. Open-source не только кодом, но и правилами.
   - Открытая лицензия.
   - Публичная модель угроз.
   - Публичная формула рейтинга провайдеров.
   - Публичный формат подписки.
   - Публичный API.

2. Провайдер владеет своим бизнесом.
   - Он может self-host панель.
   - Может брендировать приложение.
   - Может подключить свой биллинг.
   - Может не участвовать в marketplace.

3. Пользователь не должен страдать от технической сложности.
   - В приложении должен быть обычный сценарий: войти, выбрать тариф, нажать connect.
   - Ручные ссылки, QR и advanced configs остаются, но не являются главным UX.

4. Ноды - центральная часть системы.
   - Ноды можно добавлять, группировать, выводить из эксплуатации, обновлять, проверять.
   - Протоколы выбираются по capabilities ноды.
   - Нода не должна быть "черным ящиком".

5. Приватность по умолчанию.
   - Минимум логов.
   - Понятная retention policy.
   - Логи доступа пользователя выключены по умолчанию.
   - Метрики marketplace агрегированные и без раскрытия трафика пользователя.

6. Безопасность панели важнее количества галочек.
   - 2FA.
   - RBAC.
   - Audit log.
   - mTLS между панелью и нодами.
   - Подписанные конфиги.
   - Бэкапы с шифрованием.
   - Секреты не лежат в открытом виде в базе.

## Компоненты экосистемы

### 1. Provider Panel

Панель для владельца VPN-сервиса.

Основные разделы:

- Главная;
- Пользователи;
- Подписки;
- Устройства;
- Ноды;
- Протоколы;
- Тарифы;
- Биллинг;
- Инциденты;
- Логи;
- Аналитика;
- Marketplace;
- Настройки;
- Администраторы и роли;
- API tokens/webhooks;
- Бэкапы и обновления.

### 2. Node Agent

Легкий агент на каждой ноде.

Задачи:

- принимать подписанные задания от панели;
- управлять Xray/Sing-box/WireGuard/OpenVPN-core;
- отдавать healthcheck и метрики;
- применять конфиги атомарно;
- делать rollback, если новый конфиг не поднялся;
- собирать системные метрики: CPU, RAM, disk, network, uptime;
- отдавать логи по запросу с учетом RBAC;
- поддерживать drain mode, чтобы не пускать новых пользователей перед обслуживанием.

Важно: после bootstrap панель не должна хранить SSH-пароль или root-доступ. Для связи panel <-> node используется mTLS, токены регистрации и rotation.

### 3. Protocol Engine

Слой, который абстрагирует протоколы.

Поддерживаемые cores:

- Xray-core;
- Sing-box;
- WireGuard;
- OpenVPN как legacy/corporate mode;
- возможные плагины в будущем.

Протоколы первого уровня:

- VLESS + Reality + XTLS Vision;
- Trojan;
- Shadowsocks / Shadowsocks 2022;
- Hysteria2;
- TUIC;
- WireGuard;
- VMess как legacy;
- SSH/SOCKS/HTTP только как advanced/diagnostic режимы;
- OpenVPN только если нужен compatibility mode.

Каждый protocol profile должен иметь:

- название;
- core;
- transport;
- порт;
- TLS/Reality/QUIC настройки;
- совместимость с клиентами;
- требования к UDP/TCP;
- требования к домену/сертификату;
- поддерживает ли mobile;
- поддерживает ли fallback;
- риск блокировок/стабильность;
- preset: recommended, mobile, legacy, experimental.

### 4. Subscription Service

Подписка - не просто ссылка, а управляемый объект.

Подписка содержит:

- пользователя;
- тариф;
- срок действия;
- лимит трафика;
- лимит устройств;
- разрешенные регионы;
- разрешенные протоколы;
- список активных ключей;
- историю ротации ключей;
- subscription URL;
- encrypted subscription payload;
- QR/deeplink;
- compatibility exports: native app, Sing-box, Clash, V2Ray, Outline, WireGuard config.

Self-service действия:

- обновить подписку;
- перевыпустить ключ;
- отозвать старый ключ;
- сбросить список устройств;
- сменить preferred region;
- скачать конфиг для стороннего клиента;
- открыть тикет, если не работает.

### 5. User App

Приложение для конечного пользователя.

Режимы:

- Provider mode: пользователь получил приложение или ссылку от своего провайдера.
- Marketplace mode: пользователь выбирает провайдера внутри приложения.
- Manual mode: пользователь импортирует subscription URL/QR/config как в Hiddify/v2rayN.
- Advanced mode: ручные правила, DNS, routing, core logs, TUN, PAC.

Основные экраны:

- Login/Join provider;
- Главный экран подключения;
- Выбор локации;
- Автовыбор лучшей ноды;
- Статус подписки;
- Трафик и лимиты;
- Устройства;
- Обновить/перевыпустить ключ;
- История проблем и инцидентов;
- Поддержка;
- Каталог провайдеров;
- Настройки безопасности;
- Advanced routing.

Ключевые функции:

- один большой connect/disconnect;
- автоматический выбор лучшей ноды по latency, packet loss, load и доступности;
- fallback между протоколами, если провайдер разрешил;
- kill switch;
- split tunneling;
- per-app routing на Android/desktop, где возможно;
- DNS protection;
- IPv6 support;
- leak checks;
- encrypted local storage;
- hidden/encrypted subscriptions;
- deeplink import;
- QR import;
- push-уведомления об истечении подписки и инцидентах;
- offline mode с последним рабочим конфигом;
- понятная диагностика: "сервер перегружен", "подписка истекла", "ключ отозван", "нет UDP", "домен недоступен".

### 6. Provider Marketplace

Каталог провайдеров внутри приложения. Отдельный центральный сайт с рейтингом провайдеров не нужен; если позже понадобится публичная страница, она должна быть вторичной к in-app marketplace.

Что показывать:

- название провайдера;
- verified/unverified статус;
- страны и регионы;
- цена за месяц;
- лимит трафика;
- лимит устройств;
- доступные протоколы;
- поддержка iOS/Android/desktop;
- способы оплаты;
- trial/refund policy;
- средний uptime;
- средняя задержка по регионам;
- средняя скорость по opt-in тестам;
- перегруженность нод;
- история инцидентов;
- среднее время ответа поддержки;
- отзывы пользователей;
- прозрачность privacy policy;
- наличие no-log policy и срок хранения технических логов.

Как не дать провайдерам накручивать рейтинг:

- отзывы только от пользователей с подтвержденной подпиской;
- отзыв подписывается анонимным purchase/subscription proof, без раскрытия личности;
- метрики берутся из нескольких источников: публичный status, клиентские opt-in замеры, независимые probes;
- формула рейтинга открытая;
- провайдер не может удалить негативный отзыв, только ответить или оспорить;
- подозрительные отзывы помечаются;
- провайдерские метрики подписываются ключом панели;
- приложение показывает, какие данные проверены, а какие заявлены самим провайдером.

Рейтинг не должен быть одной магической цифрой. Лучше показывать несколько оценок:

- цена;
- стабильность;
- скорость;
- приватность;
- поддержка;
- покрытие регионов;
- удобство для новичков;
- advanced возможности.

## Панель: подробная структура

### Главная

Показывать:

- текущая нагрузка основного сервера;
- суммарная нагрузка всех нод;
- активные пользователи сейчас;
- трафик за сегодня/неделю/месяц;
- ошибки Xray/Sing-box/WireGuard;
- последние инциденты;
- просроченные подписки;
- ноды в warning/critical;
- очередь задач;
- статус бэкапов;
- версия панели, агента и cores.

Действия:

- открыть окно нагрузки всех серверов;
- посмотреть CPU, RAM, disk, network по всем серверам;
- посмотреть/сохранить логи панели;
- посмотреть/сохранить логи Xray/Sing-box/WireGuard;
- экспортировать диагностический bundle;
- перезапустить панель с подтверждением;
- остановить панель с подтверждением;
- перезапустить core на выбранной ноде с подтверждением;
- остановить core на выбранной ноде с подтверждением;
- перевести ноду в maintenance/drain mode.

### Пользователи

Список пользователей:

- email/phone/telegram/ID;
- статус: active, expired, suspended, trial, blocked;
- тариф;
- дата начала и окончания;
- количество устройств;
- лимит и потребление трафика;
- часто используемые 3 ноды;
- последняя активность;
- источник регистрации;
- риск/fraud флаги.

Фильтры:

- все пользователи;
- активные;
- просроченные;
- trial;
- без трафика;
- превышен лимит;
- скоро истекают;
- заблокированные;
- по тарифу;
- по региону;
- по провайдерскому тегу.

Действия:

- добавить пользователя;
- массово добавить пользователей;
- удалить/заблокировать пользователей;
- продлить подписку;
- сменить тариф;
- сбросить трафик;
- перевыпустить ключ;
- отозвать все ключи;
- сбросить устройства;
- отправить уведомление;
- открыть user page;
- экспортировать список.

Карточка пользователя:

- данные профиля;
- подписки;
- устройства;
- ключи;
- активные конфиги;
- трафик по дням;
- используемые ноды;
- события безопасности;
- история оплат;
- тикеты;
- audit trail действий админов.

### Подписки и тарифы

Тариф:

- название;
- цена;
- срок;
- trial;
- лимит трафика;
- лимит устройств;
- доступные регионы;
- доступные протоколы;
- приоритет пользователя;
- speed cap, если нужен;
- auto-renew;
- правила grace period;
- политика перевыпуска ключей;
- доступ к advanced configs.

Подписка:

- пользователь;
- тариф;
- активна/истекла/заморожена;
- даты;
- остаток трафика;
- устройства;
- ключи;
- preferred region;
- subscription URL;
- encrypted subscription URL;
- exports для сторонних клиентов.

### Ноды

Нода - отдельный сервер или edge-точка.

Поля ноды:

- name;
- region/country/city;
- provider/datacenter;
- public IP;
- domains;
- IPv4/IPv6;
- capacity;
- bandwidth limit;
- cost per month;
- tags;
- supported cores;
- supported protocols;
- current version agent/core;
- health status;
- current users;
- current traffic;
- load;
- maintenance state.

Добавление ноды:

1. Ввести IP/domain.
2. Выбрать способ bootstrap: SSH key, one-time command, cloud-init, Docker, manual token.
3. Панель проверяет ОС, архитектуру, firewall, root/ sudo, порты, DNS, kernel capabilities.
4. Устанавливается node agent.
5. Агент регистрируется через одноразовый token.
6. Настраивается mTLS.
7. Панель определяет capabilities.
8. Провайдер выбирает группы и протоколы.
9. Панель выкатывает первый config.
10. Healthcheck подтверждает готовность.

Действия с нодой:

- включить/выключить;
- drain mode;
- maintenance mode;
- обновить agent;
- обновить Xray/Sing-box;
- перезапустить core;
- применить новый config;
- rollback config;
- посмотреть логи;
- скачать diagnostic bundle;
- удалить ноду;
- перенести пользователей на другие ноды;
- ограничить новые подключения;
- назначить cost/capacity для балансировки.

Группы нод:

- по региону;
- по провайдеру хостинга;
- по протоколам;
- по тарифу;
- по latency;
- по нагрузке;
- по experimental/stable.

### Протоколы

В панели нужен не просто список протоколов, а preset-модель.

Примеры presets:

- Recommended TCP 443: VLESS + Reality + XTLS Vision.
- Mobile UDP: Hysteria2 или TUIC.
- Compatibility: Trojan TLS / Shadowsocks.
- Corporate: WireGuard/OpenVPN.
- Advanced: кастомный Xray/Sing-box JSON.

Для каждого preset:

- описание;
- где применять;
- какие клиенты поддерживают;
- какие порты нужны;
- нужен ли домен;
- нужен ли TLS certificate;
- нужен ли UDP;
- риски;
- нагрузка на CPU;
- fallback behavior;
- template variables.

### Логи и диагностика

Логи разделить:

- panel logs;
- node agent logs;
- core logs: Xray/Sing-box/WireGuard/OpenVPN;
- auth logs;
- billing logs;
- audit logs;
- marketplace sync logs.

Функции:

- фильтр по уровню: info, warning, error, critical;
- фильтр по ноде;
- фильтр по пользователю без раскрытия лишних данных;
- экспорт diagnostic bundle;
- отправка bundle в поддержку;
- автоматическое удаление старых логов;
- настройка retention policy;
- маскирование IP/ключей/токенов.

Важно: в open-source проекте сразу описать, какие логи можно включать, какие выключены по умолчанию и какие опасны для приватности.

### Администраторы и роли

Роли:

- Owner;
- Admin;
- Support;
- Billing;
- Node operator;
- Read-only auditor;
- Marketplace manager.

Права:

- управление пользователями;
- управление тарифами;
- управление нодами;
- доступ к логам;
- доступ к биллингу;
- выпуск ключей;
- удаление пользователей;
- просмотр audit logs;
- управление API tokens;
- управление marketplace profile.

Обязательные функции:

- 2FA;
- session management;
- IP allowlist;
- audit log всех опасных действий;
- подтверждение для restart/stop/delete/revoke;
- emergency recovery codes.

## API

Публичный API нужен с первого MVP.

Основные группы:

- Auth API;
- Users API;
- Subscriptions API;
- Devices API;
- Keys API;
- Nodes API;
- Protocol Profiles API;
- Plans API;
- Billing API;
- Logs API;
- Metrics API;
- Marketplace API;
- Webhooks API.

Примеры endpoint-логики:

- `POST /users`
- `POST /subscriptions`
- `POST /subscriptions/{id}/rotate-key`
- `POST /subscriptions/{id}/revoke-key`
- `GET /users/{id}/usage`
- `GET /nodes`
- `POST /nodes/bootstrap-token`
- `POST /nodes/{id}/drain`
- `POST /nodes/{id}/deploy-config`
- `POST /nodes/{id}/rollback`
- `GET /marketplace/provider-profile`
- `POST /webhooks/billing/payment-succeeded`

Для разработчиков:

- OpenAPI spec;
- SDK для TypeScript/Go/Python;
- webhook signatures;
- idempotency keys;
- API tokens с scopes;
- rate limits.

## Архитектура

Возможный стек:

- Backend: Go или Python/FastAPI.
- Frontend panel: React/TypeScript.
- DB: PostgreSQL.
- Cache/queue: Redis.
- Node agent: Go.
- Metrics: Prometheus-compatible endpoint плюс встроенный lightweight dashboard.
- Logs: встроенный просмотр, опционально Loki/OpenSearch.
- Client app: Flutter, потому что нужен iOS/Android/desktop из одной базы.
- VPN core в приложении: Sing-box first, Xray support там, где нужен.
- Communication: REST для панели/API, gRPC или HTTPS для panel-node, mTLS.
- Config delivery: signed JSON bundles.

Почему Flutter:

- Hiddify показывает, что Flutter подходит для такого класса приложений;
- быстрее сделать mobile + desktop;
- можно разделить UI и core bridge;
- проще white-label.

Почему Sing-box first в клиенте:

- широкий набор протоколов;
- TUN;
- rule-based routing;
- хорошая кроссплатформенность.

Почему не только Xray:

- Xray силен для VLESS/Reality и совместимости с экосистемой;
- Sing-box удобен как универсальный клиентский core;
- WireGuard/OpenVPN нужны как понятные и привычные режимы для части пользователей.

## Безопасность

Обязательные решения:

- mTLS между панелью и нодой;
- signed config bundles;
- config versioning;
- rollback after failed deploy;
- secrets encryption at rest;
- no stored SSH password after bootstrap;
- 2FA для админов;
- RBAC;
- audit log;
- backup encryption;
- automatic security updates как опция;
- dependency scanning;
- release signing;
- reproducible builds, если получится;
- security policy и responsible disclosure.

Для приложения:

- encrypted local storage;
- biometric lock как опция;
- kill switch;
- DNS leak protection;
- IPv6 leak protection;
- WebRTC leak instructions/check;
- защита subscription URL от случайного раскрытия;
- device binding без жесткого fingerprinting;
- clear logout/revoke device.

Для marketplace:

- не собирать историю сайтов;
- opt-in метрики скорости/latency;
- агрегировать данные;
- не публиковать малые выборки, по которым можно деанонимизировать пользователя;
- показывать пользователю, какие метрики он отправляет.

## Приватность и доверие

Для провайдера в marketplace показывать:

- privacy policy;
- срок хранения технических логов;
- какие данные нужны для регистрации;
- можно ли оплатить без KYC;
- есть ли anonymous trial;
- юрисдикция;
- кому принадлежит сервис;
- прозрачность инцидентов;
- есть ли warrant canary/transparency report, если провайдер хочет.

В панели сделать privacy presets:

- Minimal logs: только технические агрегаты.
- Balanced: технические логи с коротким retention.
- Debug: временно включается для диагностики, автоматически выключается.

## Marketplace: модель провайдера

Провайдер может быть:

- self-hosted;
- verified;
- sponsored;
- community;
- private/unlisted.

Provider profile:

- slug;
- display name;
- domains;
- public signing key;
- API endpoint;
- regions;
- plans;
- protocols;
- support contacts;
- status page;
- legal/privacy URLs;
- refund policy;
- logo/theme;
- marketplace visibility.

Верификация провайдера:

1. Проверка домена через DNS или `.well-known/vpn-provider.json`.
2. Публичный signing key.
3. Подписанный provider manifest.
4. Проверка API compatibility.
5. Проверка status endpoint.
6. Ручная/автоматическая модерация.

## Приложение: UX сценарии

### Сценарий 1. Пользователь получил VPN от провайдера

1. Пользователь устанавливает приложение.
2. Открывает provider deeplink или вводит код провайдера.
3. Авторизуется: email, phone, magic link, Telegram, OAuth или provider account.
4. Видит свой тариф, срок, трафик и устройства.
5. Нажимает connect.
6. Если не работает - нажимает "Починить подключение".
7. Приложение проверяет статус подписки, доступность нод, протокол, DNS, UDP, время системы.
8. Если ключ сломан или скомпрометирован - пользователь сам перевыпускает ключ.
9. Старый ключ отзывается.

### Сценарий 2. Пользователь выбирает провайдера

1. Открывает Marketplace.
2. Фильтрует по цене, региону, способу оплаты, устройствам, протоколам.
3. Смотрит verified metrics и отзывы.
4. Выбирает тариф.
5. Оплата идет через провайдера или через marketplace, в зависимости от модели.
6. Подписка появляется в приложении.
7. Пользователь подключается.

### Сценарий 3. Пользователь хочет сторонний клиент

1. Открывает подписку.
2. Выбирает "Экспорт".
3. Получает Sing-box, Clash, V2Ray, WireGuard или QR.
4. Панель отмечает, что это external config.
5. При ротации ключа external config нужно обновить.

### Сценарий 4. Провайдер добавляет новую ноду

1. Открывает "Ноды".
2. Нажимает "Добавить ноду".
3. Выбирает bootstrap method.
4. Агент устанавливается.
5. Панель проверяет health.
6. Провайдер выбирает protocol presets.
7. Нода добавляется в группу.
8. Балансировщик начинает направлять новых пользователей.

## Балансировка и выбор нод

Балансировка должна учитывать:

- регион пользователя;
- выбранный тариф;
- доступность протоколов;
- текущую нагрузку CPU/RAM/network;
- packet loss;
- latency;
- bandwidth cost;
- лимиты хостинга;
- maintenance/drain state;
- историю ошибок;
- предпочтения пользователя.

Режимы:

- auto best;
- closest region;
- cheapest capacity;
- premium-only;
- manual region;
- failover-only.

В приложении показывать просто:

- Авто;
- Страна/город;
- Последний рабочий;
- Самый быстрый.

## Биллинг

В MVP можно не строить собственный billing, а сделать интеграции.

Нужно предусмотреть:

- plans;
- invoices;
- payments;
- promo codes;
- trial;
- renewals;
- grace period;
- refunds;
- webhooks;
- manual payment mark;
- external billing provider adapter.

Важно: marketplace и panel billing лучше разделить. Провайдер может принимать оплату сам, а marketplace только показывает и ведет пользователя.

## White-label

Провайдер должен иметь возможность:

- задать логотип;
- задать цвета;
- задать название;
- задать support links;
- скрыть marketplace;
- включить только свои ноды;
- распространять provider deeplink;
- собрать branded build, если это юридически и технически возможно.

Но лучше сначала сделать dynamic branding без форков:

- приложение одно;
- бренд подтягивается после provider login/deeplink;
- провайдер не собирает свой небезопасный fork;
- обновления приходят всем пользователям.

## Что должно отличать проект

1. Единый стандарт подписки.
2. Нормальный provider API.
3. Self-service для пользователя.
4. Ноды как first-class сущность.
5. Marketplace с проверяемыми метриками.
6. Open-source governance.
7. Privacy-first telemetry.
8. Production-ready security.
9. White-label без форков.
10. Advanced режим без усложнения обычного UX.

## MVP

### MVP 1: панель и ноды

Минимально:

- панель;
- авторизация админов;
- RBAC хотя бы базовый;
- пользователи;
- тарифы;
- подписки;
- лимиты по времени и трафику;
- добавление нод;
- node agent;
- healthcheck;
- Xray support;
- VLESS + Reality preset;
- Hysteria2/TUIC как второй этап, если успеваем;
- subscription links;
- QR/deeplink;
- логи;
- audit log;
- backup/restore;
- REST API.

### MVP 2: приложение

Минимально:

- Android + Windows или Android + iOS, выбор зависит от приоритета;
- Flutter;
- login к провайдеру;
- импорт provider deeplink;
- connect/disconnect;
- автообновление подписки;
- перевыпуск ключа;
- просмотр срока/трафика/устройств;
- выбор региона;
- auto best node;
- базовая диагностика;
- encrypted storage;
- kill switch там, где платформа позволяет.

### MVP 3: marketplace beta

Минимально:

- provider registry;
- provider profiles;
- ручная модерация провайдера;
- фильтры по цене/региону/протоколам;
- verified reviews;
- public status metrics;
- open ranking formula;
- жалобы/appeals.

## Roadmap

### Phase 0. Исследование и стандарт

- описать формат provider manifest;
- описать формат subscription;
- описать panel API;
- описать node agent protocol;
- выбрать лицензию;
- выбрать стек;
- написать threat model.

### Phase 1. Core panel

- админка;
- пользователи;
- подписки;
- ноды;
- Xray config deployment;
- VLESS Reality;
- метрики;
- логи;
- API.

### Phase 2. Client app

- Flutter shell;
- Sing-box bridge;
- provider login;
- subscription sync;
- connect;
- diagnostics;
- key rotation.

### Phase 3. Reliability

- rollback config;
- backup/restore;
- node drain;
- multi-node balancing;
- incident system;
- alerting;
- status page.

### Phase 4. Marketplace

- provider registry;
- reviews;
- ratings;
- public metrics;
- provider verification;
- anti-fraud.

### Phase 5. Ecosystem

- SDK;
- plugins;
- billing adapters;
- white-label;
- external clients compatibility;
- public docs;
- contributor program.

## Лицензия и governance

Варианты:

- AGPL-3.0 для панели и node agent, чтобы SaaS-форки тоже публиковали изменения.
- GPL-3.0 для клиента, если цель - защитить open-source природу приложения.
- Apache-2.0/MIT для SDK, чтобы провайдеры легче интегрировались.

Практичный вариант:

- `panel`: AGPL-3.0;
- `node-agent`: AGPL-3.0;
- `client-app`: GPL-3.0 или MPL-2.0, надо отдельно проверить совместимость с app store и используемыми core libraries;
- `sdk`: Apache-2.0;
- `protocol specs`: CC BY 4.0 или Apache-2.0.

Governance:

- public roadmap;
- RFC-процесс для протоколов и API;
- security policy;
- code of conduct;
- DCO вместо тяжелого CLA;
- public release signing;
- отдельный security advisory процесс.

## Риски

- Слишком много протоколов в начале. Решение: presets и plugin architecture.
- Сложность iOS. Решение: заранее проверить Network Extension, App Store rules и core integration.
- Marketplace может стать местом скама. Решение: verification, signed metrics, verified reviews, moderation.
- Провайдеры могут подделывать статистику. Решение: opt-in client telemetry, public probes, подпись метрик, открытая формула.
- Пользователи могут не понимать advanced configs. Решение: обычный режим по умолчанию, advanced отдельно.
- VPN-тематика юридически чувствительная. Решение: open-source tool, законное использование, прозрачная policy, anti-abuse.
- Панель может стать целью атак. Решение: security-first architecture, минимизация секретов, RBAC, audit, mTLS.
- Хранение логов может убить доверие. Решение: minimal logs by default, clear retention, privacy presets.

## Вопросы, которые нужно решить

1. Какое рабочее название проекта? (Lenker)
2. Первый рынок: Россия/СНГ, весь мир или privacy/VPN community глобально? (Без разницы, проект будет open-source на github, но изначально расчитывается на россию)
3. Что важнее первым: панель или приложение? (панель и база для подключений (приложение))
4. Первые платформы приложения: Android, iOS, Windows, macOS, Linux? (андроид, windows, macos)
5. Marketplace будет централизованный, федеративный или гибридный? 
6. Провайдеры будут проходить ручную верификацию? (Да)
7. Нужен ли встроенный биллинг в MVP или только webhooks/integrations?
8. Какие способы входа нужны: email, phone, Telegram, OAuth, magic link? (email; phone можно оставить как опциональную верификацию без денежных бонусов)
9. Какой минимальный набор протоколов для MVP?
10. Нужен ли white-label build или хватит dynamic branding?
11. Какой уровень логов допустим по умолчанию? (error/warning)
12. Нужен ли Telegram bot как обязательная часть или только optional plugin? (опционально)
13. Будет ли центральный сайт проекта с рейтингом провайдеров? (нет, marketplace нужен внутри приложения)
14. Кто владеет marketplace governance?
15. Нужно ли сразу делать миграцию из Marzban/3x-ui/Hiddify Manager? (нет)

## Ответы и текущие решения

Уже понятно:

- Рабочее название: Lenker.
- Проект open-source на GitHub, но первый практический фокус - Россия/СНГ.
- Первым делаем панель и базу для подключений, приложение идет рядом, но не должно тормозить MVP панели.
- Первые платформы приложения: Android, Windows, macOS.
- Провайдеры проходят ручную верификацию.
- Логирование по умолчанию: только `error` и `warning`.
- Telegram bot - опциональный plugin, не обязательная часть ядра.
- Центральный сайт с рейтингом провайдеров не нужен.
- Marketplace нужен внутри приложения.
- Идея бонуса `10 рублей за phone verification` отброшена.
- Миграции из Marzban/3x-ui/Hiddify Manager не нужны в первой версии.

Нужно уточнить:

- Phone verification нужна только как защита от абуза/мультиаккаунтов или ее вообще не делать в MVP?
- Marketplace делать в первом MVP приложения или вынести в отдельную phase после панели, нод и подключения?

## Простыми словами про непонятные вопросы

### 5. Marketplace: централизованный, федеративный или гибридный?

Это вопрос о том, где хранится список провайдеров.

- Централизованный: есть один главный каталог Lenker. Все провайдеры попадают туда через модерацию. Проще для пользователей, но больше ответственности на владельцах Lenker.
- Федеративный: каждый провайдер публикует свой файл/manifest у себя на домене, а приложение может подключать разные каталоги. Меньше контроля, сложнее UX.
- Гибридный: провайдер хранит данные у себя, но Lenker ведет проверенный индекс провайдеров. Это самый здравый вариант.

Рекомендация для Lenker: гибридный marketplace внутри приложения. Без отдельного рейтингового сайта. Сначала можно хранить verified-index в GitHub repository проекта, а приложение будет читать этот индекс и показывать каталог провайдеров.

### 7. Встроенный биллинг или webhooks/integrations?

Встроенный биллинг - это когда Lenker сам умеет принимать оплату, создавать счета, продлевать подписки, делать возвраты и работать с платежками.

Webhooks/integrations - это когда оплату принимает внешний сервис, сайт провайдера, Telegram-бот или ручная админка, а Lenker только получает событие: "пользователь оплатил, продлить подписку".

Рекомендация для MVP: не делать полноценный биллинг. Сделать тарифы, подписки, ручное продление, API и webhooks. Платежные интеграции добавить позже как plugins.

### 9. Минимальный набор протоколов для MVP

Это вопрос: какие протоколы реально поддерживать первыми, чтобы не утонуть в сложности.

Рекомендация для MVP:

- обязательно: `VLESS + Reality + XTLS Vision` через Xray;
- желательно вторым: `Hysteria2` через Sing-box, потому что он полезен на нестабильных сетях и mobile;
- позже: `TUIC`, `WireGuard`, `Trojan`, `Shadowsocks 2022`;
- не делать первым приоритетом: `VMess`, потому что это legacy.

Для панели важно сразу проектировать protocol presets, но физически реализовать сначала 1-2 стабильных preset.

### 10. White-label build или dynamic branding?

White-label build - это когда каждый провайдер собирает свое отдельное приложение со своим названием, логотипом и публикацией.

Dynamic branding - это когда приложение Lenker одно, но после входа к провайдеру подтягивает его логотип, цвета, название и ссылки поддержки.

Рекомендация для MVP: dynamic branding. Отдельные white-label builds сильно усложнят обновления, App Store/Google Play, безопасность и поддержку.

### 14. Кто владеет marketplace governance?

Это вопрос: кто решает, какого провайдера пустить в каталог, кого заблокировать, как считать рейтинг и как разбирать жалобы.

Рекомендация для старта:

- до появления большой community этим владеет core team проекта Lenker;
- правила верификации и рейтинга публичные;
- спорные решения фиксируются в GitHub issues/discussions;
- позже можно сделать community moderation/RFC-процесс.

## Обновленное решение по MVP после ответов

MVP Lenker должен быть таким:

1. Панель провайдера.
2. Добавление и управление нодами.
3. Пользователи, тарифы, подписки, лимиты по времени/трафику/устройствам.
4. Protocol preset: `VLESS + Reality + XTLS Vision`.
5. Базовый node agent с healthcheck, deploy config и rollback.
6. Subscription URL, QR и deeplink.
7. Приложение для Android, Windows, macOS.
8. Вход по email.
9. Phone verification не давать денежных бонусов; если понадобится, оставить только как антиабуз-опцию.
10. Логи только `warning/error` по умолчанию.
11. Без встроенного биллинга: ручное продление, API и webhooks.
12. Без центрального сайта рейтинга.
13. Marketplace нужен внутри приложения; для первого MVP можно заложить архитектуру, а полноценный каталог провайдеров вынести в следующую phase.
14. Telegram bot только как optional plugin.
15. Без миграции из других панелей в первой версии.

## Черновое позиционирование

Open-source VPN ecosystem for providers and users.

Для провайдера:

- развернул панель;
- добавил ноды;
- выбрал protocol presets;
- создал тарифы;
- выдал приложение пользователям;
- пользователи сами обновляют ключи и управляют подпиской.

Для пользователя:

- установил приложение;
- вошел к провайдеру или выбрал его в каталоге;
- видит срок, трафик, устройства и статус;
- подключается одной кнопкой;
- если что-то сломалось, приложение диагностирует проблему и может перевыпустить ключ без поддержки.

Для open-source community:

- открытый код;
- открытый API;
- открытый формат подписки;
- открытая формула рейтинга;
- возможность self-host;
- возможность создавать клиентов, панели, плагины и биллинг-интеграции вокруг одного стандарта.
