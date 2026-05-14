# OpenAPI

This directory contains draft OpenAPI contracts for Lenker APIs.

## panel-api v1

`panel-api.v1.yaml` documents only the currently implemented `services/panel-api` HTTP surface:

- `GET /healthz`
- admin login
- admin-protected users endpoints
- admin-protected plans endpoints
- admin-protected subscriptions endpoints

This is a draft contract for the implemented backend slice. It intentionally does not include marketplace, billing, node-agent APIs, client app APIs, devices, key rotation, export flows, or VPN/Xray config delivery.

Conservative note:

The OpenAPI file is handwritten for now. The project does not generate server code from OpenAPI and does not add a heavy validation framework at this stage.

## Validation

Validate the draft contract from the repository root:

```sh
make openapi-lint
```

The target uses the local Ruby standard library YAML parser through `scripts/validate-openapi.rb`. It checks that the file parses, has the expected OpenAPI sections, resolves local `$ref` values, and documents only the currently implemented panel-api endpoints.

No code is generated from the spec.
