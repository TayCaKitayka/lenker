# panel-web

`panel-web` is the React + TypeScript web application for Lenker provider operations.

Current implemented foundation:

- Vite + React entrypoint;
- TypeScript source compilation;
- base responsive layout;
- admin login against `panel-api`;
- admin session stored in `sessionStorage`;
- expired or malformed admin sessions are cleared on load;
- unauthorized API responses clear the session and return the admin to login;
- dashboard shell;
- users management page with list, create, update, suspend, and activate flows;
- plans management page with list, create, update, and archive flows;
- navigation placeholders for subscriptions and nodes.

Run from the repository root:

```sh
npm install
npm run panel-web:dev
```

Build:

```sh
npm run panel-web:build
```

Type-check:

```sh
npm run panel-web:lint
```

Focused session utility tests:

```sh
npm --workspace @lenker/panel-web run test:session
```

Focused users form tests:

```sh
npm --workspace @lenker/panel-web run test:users
```

Focused plans form tests:

```sh
npm --workspace @lenker/panel-web run test:plans
```

Planned `MVP v0.1` provider UI scope:

- real `panel-api` admin login;
- protected dashboard shell;
- users list/create/update/suspend/activate;
- plans list/create/update/archive;
- subscriptions list/create/renew;
- nodes list/detail/drain/undrain/disable/enable;
- loading, empty, unauthorized, and API error states.

Not planned for this first UI layer:

- marketplace UI;
- billing UI;
- marketing landing;
- advanced analytics;
- client app UI.
