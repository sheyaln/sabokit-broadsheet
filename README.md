# Broadsheet

> *A broadsheet (n.): a single sheet of paper, printed on one side, posted in public places to be read by anyone passing. The format of the IWW, the suffragists, the abolitionist press, the samizdat-passers and the leaflet-droppers. The original mass-communication tool, before the term existed.*

A self-hosted email platform. Fork of [Notifuse](https://github.com/Notifuse/notifuse) (AGPLv3), with OIDC SSO and a small set of fork-specific patches on top of upstream releases. Go backend, React console.

<img alt="Email Editor" src="https://github.com/user-attachments/assets/f650ac1b-58fd-44fb-884d-e9811255f1e4" />

## What it does

- MJML email composer, campaign scheduling, A/B tests
- Lists, segments, contact profiles with custom fields and event timelines
- Transactional REST API, webhooks, Liquid templating
- Send via Amazon SES, Mailgun, Postmark, Mailjet, SparkPost, or SMTP
- Open/click tracking, delivery and bounce metrics, per-campaign reports
- S3-compatible file storage
- Embeddable notification-center widget for end-user subscription management
- Multi-workspace tenancy

## What's different from upstream

- OIDC SSO (Authentik, Keycloak, Okta, Google Workspace, anything OIDC) with IdP group → workspace permission mapping.
- Self-hosted only. No managed tier.
- AGPLv3, no contributor CLA.

Everything else — engine, editor, API, migrations — is upstream Notifuse with patches.

## Architecture

Backend (Go): clean architecture — `internal/domain/`, `internal/service/`, `internal/repository/`, `internal/http/`. Standard library `http.ServeMux`, no web framework.

Frontend (React + TypeScript): `console/` (operator admin UI, Ant Design) and `notification_center/` (embeddable end-user widget).

Storage: PostgreSQL 17, Squirrel query builder, custom version-based migrations. See [CLAUDE.md](CLAUDE.md) for the migration model.

## Project layout

```
cmd/                    # Entry points
internal/               # Application code
  domain/               # Entities, interfaces
  service/              # Business logic
  repository/           # Data access
  http/                 # Handlers, middleware
  database/             # Schema
  migrations/           # Versioned migrations
console/                # React admin UI
notification_center/    # Embeddable widget
pkg/                    # Public packages (logger, mailer, tracing)
config/                 # Configuration
```

## Docs

Engine docs are upstream: [docs.notifuse.com](https://docs.notifuse.com) covers the editor, API, providers, and migration model. Fork-specific behavior (OIDC, IdP group mapping) is documented in this repo.

## Relationship to Notifuse

Broadsheet is a hard fork of Notifuse. We track upstream by force-rebasing `main` onto upstream when syncing, then re-applying our additions on top. We do not contribute back upstream for branding and because the upstream contributor agreement requires full IP assignment.

For the original project, see [github.com/Notifuse/notifuse](https://github.com/Notifuse/notifuse).

## License

[AGPLv3](LICENCE.md), same as upstream. Original Notifuse code remains copyright (C) 2025 Notifuse. Fork contributions inherit the same terms.
