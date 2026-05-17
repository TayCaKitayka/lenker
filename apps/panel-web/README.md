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
- subscriptions management page with list, create, update, and renew flows;
- subscriptions page can load a compact read-only access export for the single MVP VLESS Reality path;
- subscriptions page can issue, rotate, and revoke the current subscription access token; plaintext tokens are only shown from issue/rotate responses;
- nodes management page with list, detail, bootstrap token creation, drain, undrain, disable, enable, read-only config revision metadata, and rollback revision creation flows.

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

Focused subscriptions form tests:

```sh
npm --workspace @lenker/panel-web run test:subscriptions
```

Focused nodes form/action tests:

```sh
npm --workspace @lenker/panel-web run test:nodes
```

Planned `MVP v0.1` provider UI scope:

- real `panel-api` admin login;
- protected dashboard shell;
- users list/create/update/suspend/activate;
- plans list/create/update/archive;
- subscriptions list/create/renew;
- subscription access export inspection for the current provider-side MVP path;
- subscription access token issue, rotate, and revoke controls for the current provider-side MVP path;
- nodes list/detail/drain/undrain/disable/enable;
- node config revisions list/detail metadata view with applied-revision rollback action;
- loading, empty, unauthorized, and API error states.

Not planned for this first UI layer:

- marketplace UI;
- billing UI;
- config apply, Xray process control, or node-side rollback execution;
- marketing landing;
- advanced analytics;
- client app UI.

Nodes page note:

The nodes page uses the existing `panel-api` admin Bearer session flow. It can
create one-time plaintext bootstrap token responses and show them in memory, but
does not store bootstrap tokens in browser storage. It can show existing config
revision metadata and create the backend's dummy signed revision metadata for a
selected node. It can request a backend rollback revision from an applied
revision and refresh metadata after the action. The page does not execute config
apply, node file switching, rollback execution, or Xray runtime control itself.
