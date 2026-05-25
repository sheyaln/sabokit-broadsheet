# Social-media MVP plan

Coordinated multi-channel posting alongside email broadcasts. BYO credentials model — each operator registers their own platform apps. Targets: Bluesky, Facebook Pages, Instagram Business/Creator, TikTok.

## Why BYO instead of centrally-registered apps

Three options were considered:

| Model | Pro | Con |
|---|---|---|
| **Central app** (we register Broadside with each platform) | One-click connect for end users | We eat Meta/TikTok app review, business verification, and the maintenance tail of every platform breakage |
| **BYO credentials** (each operator registers their own app) | Zero gatekeeping by us, ships immediately, fork-honest | Operators do a 20-minute developer-account setup per platform |
| **Hybrid** (we register, operators can override) | Best of both | Doubles the surface area to maintain |

Picked BYO. Aligns with the self-hosted identity. Operators who can run Postgres can register a Meta app. The setup friction is real but bounded; the central-app maintenance tail is unbounded.

## Domain shape

New domain, parallel to `broadcast`. Reasoning: broadcasts are deeply email-specific (MJML, deliverability, suppression lists, A/B subject tests, open/click tracking via redirects). Retrofitting them to be multi-channel would touch every layer for a feature with no validated demand. Cheaper to ship a thin parallel domain and converge later if usage justifies it.

```
internal/domain/social_account.go
internal/domain/social_post.go
internal/domain/social_provider.go

internal/service/social/
  service.go              # orchestration: post creation, scheduling, dispatch
  bluesky/provider.go
  bluesky/provider_test.go
  facebook/provider.go
  facebook/provider_test.go
  instagram/provider.go
  instagram/provider_test.go
  tiktok/provider.go
  tiktok/provider_test.go

internal/repository/social_account_postgres.go
internal/repository/social_post_postgres.go
internal/repository/social_post_dispatch_postgres.go

internal/http/social_handler.go
internal/http/social_account_handler.go
```

## Storage

Workspace-scoped tables (per the existing workspace-DB pattern, not the system DB).

```sql
-- One row per platform account an end user has connected.
CREATE TABLE social_account (
  id              VARCHAR(32) PRIMARY KEY,
  user_id         VARCHAR(32) NOT NULL,
  provider_kind   VARCHAR(32) NOT NULL,   -- bluesky | facebook | instagram | tiktok
  provider_user_id VARCHAR(255) NOT NULL, -- the platform's user/page/IG-business-id
  handle          VARCHAR(255) NOT NULL,  -- display handle, for the UI
  access_token    BYTEA NOT NULL,         -- encrypted, same pattern as ESP keys
  refresh_token   BYTEA,                  -- encrypted, nullable (Bluesky app-passwords don't refresh)
  expires_at      TIMESTAMPTZ,
  scopes          JSONB,
  meta            JSONB,                  -- platform-specific extras (e.g. FB Page ID for IG)
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON social_account (user_id, provider_kind);
CREATE UNIQUE INDEX ON social_account (provider_kind, provider_user_id);

-- One row per composed post (across one or more target accounts).
CREATE TABLE social_post (
  id              VARCHAR(32) PRIMARY KEY,
  created_by      VARCHAR(32) NOT NULL,
  status          VARCHAR(32) NOT NULL,   -- draft | scheduled | sending | sent | partial_failure | failed
  scheduled_at    TIMESTAMPTZ,            -- null = draft, set = scheduled
  sent_at         TIMESTAMPTZ,
  content         JSONB NOT NULL,         -- { default: {...}, overrides: { instagram: {...} } }
  media           JSONB,                  -- [{ url, mime, alt_text, ... }]
  target_accounts JSONB NOT NULL,         -- array of social_account.id
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- One row per (post, target) — tracks per-platform outcome.
CREATE TABLE social_post_dispatch (
  id                VARCHAR(32) PRIMARY KEY,
  social_post_id    VARCHAR(32) NOT NULL REFERENCES social_post(id) ON DELETE CASCADE,
  social_account_id VARCHAR(32) NOT NULL REFERENCES social_account(id),
  status            VARCHAR(32) NOT NULL, -- pending | sending | sent | failed
  remote_post_id    VARCHAR(255),         -- the platform's returned post id
  error             TEXT,
  attempts          INT NOT NULL DEFAULT 0,
  last_attempt_at   TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX ON social_post_dispatch (status, social_post_id);
```

Workspace settings get a per-platform credential block:

```jsonb
{
  "social_providers": {
    "facebook": {
      "client_id": "...",
      "client_secret_encrypted": "...",
      "redirect_uri": "https://broadside.example.org/api/social.oauthCallback?provider=facebook"
    },
    "instagram": { ... },
    "tiktok": { ... },
    "bluesky": { "enabled": true }
  }
}
```

`bluesky` has no operator-side credentials because app-passwords are per-user. Treat as a feature flag.

## Migration

```go
// internal/migrations/v33.go (or whatever the next major is at the time)
// HasSystemUpdate=false, HasWorkspaceUpdate=true.
// Creates the three tables above with IF NOT EXISTS.
```

## Provider interface

```go
package social

type ProviderKind string

const (
    KindBluesky   ProviderKind = "bluesky"
    KindFacebook  ProviderKind = "facebook"
    KindInstagram ProviderKind = "instagram"
    KindTikTok    ProviderKind = "tiktok"
)

type Token struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Meta         map[string]string // provider-specific (e.g. IG Business ID, FB Page ID)
}

type Content struct {
    Text  string
    Media []Media // already uploaded to our S3, ready to hand off
}

type Media struct {
    URL      string
    MimeType string
    AltText  string
    Width    int
    Height   int
}

type PostResult struct {
    RemoteID  string
    RemoteURL string // permalink, where available
}

type Provider interface {
    Kind() ProviderKind

    // AuthURL builds the redirect target for the OAuth dance. state is
    // the CSRF token we'll verify on callback. Bluesky's app-password
    // path returns ("", ErrNotApplicable) — we collect creds via form.
    AuthURL(ctx context.Context, providerCfg ProviderConfig, state string) (string, error)

    // ExchangeCode turns the callback's code into a Token.
    ExchangeCode(ctx context.Context, providerCfg ProviderConfig, code string) (Token, *AccountIdentity, error)

    // Refresh returns a new Token. Implementations that don't need
    // refresh (Bluesky app-password) return the input unchanged.
    Refresh(ctx context.Context, providerCfg ProviderConfig, t Token) (Token, error)

    // Post publishes content to the account identified by the token.
    Post(ctx context.Context, t Token, c Content) (PostResult, error)

    // CharLimit returns the platform's text limit; 0 means none.
    CharLimit() int

    // RequiresMedia returns true if Post requires at least one media item.
    RequiresMedia() bool
}

type ProviderConfig struct {
    ClientID     string
    ClientSecret string
    RedirectURI  string
}

type AccountIdentity struct {
    ProviderUserID string
    Handle         string
    Meta           map[string]string
}
```

Each `internal/service/social/<kind>/provider.go` is a self-contained file implementing this. Tests use real platform sandbox modes where they exist (Bluesky has a `bsky.social` test approach via fresh accounts; TikTok has a sandbox tier; Meta has Test Users).

## Per-platform notes

### Bluesky

- Auth: app-password (simplest path) or OAuth (newer, more setup). Ship app-password first.
- Endpoint: `com.atproto.repo.createRecord` on `app.bsky.feed.post`.
- Char limit: 300 graphemes. Need to count graphemes, not bytes.
- Media: separate `com.atproto.repo.uploadBlob` call per image, then reference blob CIDs in the post record.
- Token: app-passwords don't expire; no refresh needed.
- No platform review.

### Facebook Pages

- Auth: OAuth via `https://www.facebook.com/v18.0/dialog/oauth`. Need `pages_manage_posts`, `pages_read_engagement`. Page-scoped access token after exchanging the short-lived user token for a long-lived one (60-day), then asking for `me/accounts` to get per-page tokens.
- Endpoint: `POST /{page-id}/feed` for text, `/photos` or `/videos` for media.
- Char limit: effectively none (63k+).
- Operator setup: register Meta app, add Facebook Login, request `pages_manage_posts` + `pages_read_engagement`, submit for review, complete Business Verification. **Real friction.**

### Instagram Business/Creator

- Auth: piggybacks on Facebook OAuth. IG account must be Business/Creator AND linked to a Facebook Page that the user manages. Token comes from FB, scoped with `instagram_basic` + `instagram_content_publish`.
- Endpoint: two-step container-then-publish:
  1. `POST /{ig-user-id}/media` with `image_url` or `video_url` (must be publicly accessible — our S3 URLs work) → returns container ID.
  2. `POST /{ig-user-id}/media_publish` with `creation_id=<container>`.
- Char limit: 2,200 chars caption.
- Media: **required** for feed posts. No text-only.
- Hashtag limit: 30. Worth a soft warning in the UI at 25.

### TikTok

- Auth: OAuth via TikTok for Developers. Scope: `video.publish` (Direct Post) or `video.upload` (drafts only).
- Endpoint: two-step. Init: `POST /v2/post/publish/inbox/video/init/` returns upload URL. Upload: chunked PUTs to that URL. Status: poll `POST /v2/post/publish/status/fetch/`.
- Char limit: 2,200 chars caption.
- Media: **video required**. No image posts via this API. Aspect ratio + length constraints apply.
- Operator setup: register TikTok for Developers app, add Content Posting API, submit for production review (sandbox works without).

## Scheduling

Reuse the existing scheduling worker pattern from `broadcast`. Add a task type `social_post_dispatch` that the worker picks up at `scheduled_at`. Worker iterates `social_post_dispatch` rows in `pending` state for that post, calls the appropriate `Provider.Post`, writes back `sent`/`failed`. Idempotent on retry via `attempts` counter + `last_attempt_at`.

Token refresh runs on a separate ticker (hourly): finds `social_account` rows where `expires_at` is within 24 hours, calls `Provider.Refresh`, writes back. Failed refreshes mark the account `requires_reauth` and notify the user.

## HTTP API

RPC-style endpoints matching the existing dot-notation convention:

```
POST /api/social.account.list
POST /api/social.account.connectStart      -> { auth_url } for OAuth providers
POST /api/social.account.connectComplete   -> handles the OAuth callback exchange
POST /api/social.account.disconnect

POST /api/social.post.create
POST /api/social.post.update
POST /api/social.post.delete
POST /api/social.post.schedule
POST /api/social.post.sendNow
GET  /api/social.post.list                  -- query: status, scheduled_after, etc.
GET  /api/social.post.get
GET  /api/social.post.dispatchStatus        -- per-platform outcome detail

GET  /api/social.oauthCallback              -- the redirect target; verifies state, exchanges code
```

## Frontend

New page `/social` paralleling `/broadcasts`:
- List view + calendar view (steal layout from broadcasts)
- Composer with one default editor + per-platform tabs for overrides, each showing platform-specific char counter, media uploader, preview
- Settings → Integrations → Social: per-platform "paste your client_id/secret + redirect URL" form, then per-user "Connect account" button

Reuse existing components:
- S3 file manager for media upload
- Date/time picker from the broadcast scheduler
- AntD form components throughout

## Phasing

| Phase | Scope | Time | Gate |
|---|---|---|---|
| 1 | Domain + repos + migration + provider interface + Bluesky (app-password) | ~1 week | Bluesky end-to-end post works in dev |
| 2 | Composer UI + scheduler wiring + per-user account list | ~1 week | Can schedule a Bluesky post from console, worker dispatches |
| 3 | Facebook Pages provider + OAuth flow + connect UX | ~1.5 weeks | Real Page post from real account, token refresh works |
| 4 | Instagram Business provider (container-publish + media gating) | ~2 weeks | Real IG Business post with image |
| 5 | TikTok provider (chunked video upload + status polling) | ~2 weeks | Real TikTok video post |

**Phase 2 is the gate.** If the composer + scheduler integration doesn't feel right with just Bluesky, fix it before adding Meta. Adding more platforms multiplies the cost of every UX rework.

Total: ~7.5–8 weeks of careful work. Add ~30% buffer for platform-doc rot, OAuth dance debugging, and the inevitable "the sandbox behaves differently from production" surprises.

## Tests

Per the project test convention (CLAUDE.md → testing requirements):

- **Domain layer**: validation rules for `social_post` (e.g., TikTok dispatch requires video media), `social_account` (provider_kind allowed values, token encrypted).
- **Service layer**: mock the `Provider` interface, test orchestration (scheduling, dispatch, retry, partial-failure handling).
- **Repository layer**: sqlmock for all three new tables.
- **HTTP layer**: handler tests for each endpoint.
- **Provider layer**: per-platform unit tests with `httptest.Server` simulating each platform's API. Integration test against the real platform sandbox where feasible (Bluesky: spin up a dedicated test account; TikTok: sandbox tier; Meta: Test Users).

Run after each phase: `make test-domain test-service test-repo test-http` and the new `make test-social` (add a target).

## Open questions

1. **Do we surface the operator setup guide in-product or as docs?** A "Setup Instagram" wizard inside the integrations page would reduce support load but is meaningful frontend work. Docs are cheaper but worse UX. Default: docs first, wizard if usage justifies it.
2. **Cross-platform link tracking.** Email broadcasts get `/r/` redirect tracking. Should social posts get the same? Adds complexity (per-platform link-rewriting, awareness of platform-side link rules like X's t.co). Defer to Phase 6.
3. **Threads on Bluesky / IG Carousels / multi-photo Facebook posts.** Each is a real shape we don't model in `social_post.content`. Defer; ship single-post first, model threads/carousels only when demand is concrete.
4. **What happens when an operator changes their client_secret?** Existing tokens still work until they expire, then refresh fails. Worth a UI banner on the integrations page.
5. **Mastodon.** Not in the target list but trivially fits the provider interface (per-instance OAuth, standard REST). Free addition in Phase 3 or as a community PR.

## Out of scope for MVP

- Social listening / inbound mention aggregation. Different problem; Chatwoot territory.
- DM / unified inbox. Same.
- Analytics beyond "did it post." Per-platform reach/engagement APIs are inconsistent and rate-limited.
- AI caption generation. Cheap to add later if Tiptap's existing AI hooks are wired up; out of scope for plumbing.
- Bulk-import of historical posts.
