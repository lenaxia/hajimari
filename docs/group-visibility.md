# Group-Based Visibility — Feature Specification

Status: implemented on branch `feat/group-visibility`
Fork base: `toboshii/hajimari@b07024d`

## 1. Overview

Adds per-request, group-based filtering of apps and bookmarks. Identity is
delegated entirely to a forward-auth reverse proxy (e.g. Authelia via Traefik
`forwardAuth` middleware) which injects a trusted group-membership header.
Hajimari never authenticates; it only authorizes *visibility*.

Ingress example:

```yaml
metadata:
  annotations:
    hajimari.io/enable: "true"
    hajimari.io/group: Media
    hajimari.io/visible-groups: "family,friends"
```

Custom apps and bookmark groups (config):

```yaml
customApps:
  - group: Media
    apps:
      - name: Test
        url: https://example.com
        visibleGroups: [family, friends]
globalBookmarks:
  - group: Admin Links
    visibleGroups: [admins]
    bookmarks: [...]
```

CRD (`applications.hajimari.io`):

```yaml
spec:
  name: Proxmox
  group: Infrastructure
  url: https://proxmox.example.com
  visibleGroups: [admins]
```

## 2. Requirements

### Functional

| ID | Requirement |
|----|-------------|
| FR1 | An app whose source carries `visible-groups` metadata renders only for requests whose user is a member of at least one listed group. |
| FR2 | An app/bookmark with **no** visibility metadata is visible to every requester (identical to upstream behavior). |
| FR3 | The literal `*` in a visibility list means "any authenticated user". |
| FR4 | Group names match case-insensitively; surrounding whitespace is ignored. |
| FR5 | Group membership is read from a configurable HTTP header (config key `groupsHeader`, default `Remote-Groups`), comma-separated. |
| FR6 | Filtering applies uniformly to: discovered ingress apps, custom apps (per app), CRD apps (per app), and bookmark groups (per group). |
| FR7 | The API response shape is unchanged; `visibleGroups` is never serialized to clients (no metadata leak). |
| FR8 | No frontend changes required for basic filtering. |
| FR9 | Members of a configured admin group (config `adminGroups`, default `admins`) may pass `?group=<groups>` on the page or API to preview the dashboard exactly as that membership would see it. |
| FR10 | Impersonation accepts a comma separated list or repeated params (union view); an empty value (`?group=`) previews the groupless (unauthenticated-equivalent) view. |
| FR11 | Non-admins passing `?group=` receive `403 Forbidden`; the parameter never grants visibility on its own. |

### Non-functional

| ID | Requirement |
|----|-------------|
| NFR1 | Filtering is O(apps) per request, single pass, no locks. |
| NFR2 | Group memberships are not logged. |
| NFR3 | Fully backwards compatible: stock configs and annotations behave exactly as upstream. |
| NFR4 | Filtering never mutates the shared kube-app cache (concurrent requests with different groups must not influence each other). |

## 3. Invariants

- **INV1 (default-open):** absence of visibility metadata ⇒ visible to all.
  Rationale: preserves upstream behavior; restricted tiles are the explicit opt-in.
- **INV2 (fail-closed for restricted content):** if metadata *is* present, a
  request without a matching group — including a missing/empty header (direct
  pod access, misconfigured proxy chain) — never sees the tile.
- **INV3 (per-request isolation):** group resolution happens per request in the
  handlers; no shared state may capture one user's groups.
- **INV4 (identity is upstream's job):** Hajimari trusts only the configured
  header and must be deployed so the header can only be set by the proxy
  (network policy: the Service is reachable only via the forward-auth chain).
- **INV5 (normalization):** comparisons are over trimmed, lowercased names.
- **INV6 (impersonation boundary):** the admin check is evaluated exclusively
  against the trusted header groups; query parameters never influence the
  requester's own privileges. Impersonation replaces the group set (it does
  not union with the admin's own groups).

## 4. Acceptance Criteria

| ID | Scenario | Expected |
|----|----------|----------|
| AC1 | Header `Remote-Groups: family`, app annotated `family,friends` | App rendered |
| AC2 | Header `Remote-Groups: friends`, same app | App rendered |
| AC3 | Header `Remote-Groups: admins`, same app | App hidden |
| AC4 | No annotation, any header (incl. none) | App rendered (stock behavior) |
| AC5 | Annotation `*`, any non-empty header | App rendered |
| AC6 | Annotation present, header absent/empty | App hidden (INV2) |
| AC7 | Header `Family, Admins` (case/space variance) vs `family,admins` | Matches (INV5) |
| AC8 | All apps of a display-group filtered out | Display group omitted entirely |
| AC9 | Bookmark group `visibleGroups: [admins]`, user `family` | Bookmark group hidden; other groups intact |
| AC10 | Two concurrent users (family, admin) alternate requests | Each sees exactly their set on every request (no bleed) |
| AC11 | Stock deployment (no annotations, no proxy header) | API responses byte-identical in shape to upstream |
| AC12 | `go build ./...`, `go vet ./...`, `go test ./...` | All pass |
| AC13 | Header `admins`, `?group=family` | Response identical to a header-only `family` user (apps + bookmarks) |
| AC14 | Header `admins`, `?group=` | Groupless view: only unannotated items |
| AC15 | Header `family` or no header, any `?group=` value | 403 Forbidden |
| AC16 | Header `admins`, no param | Admin's own view unchanged |
| AC17 | Header `admins`, `?group=family,admins` | Union of both groups' views |

## 5. Test Plan

### Unit (automated, this branch)

- `internal/visibility`: header parsing (empty, single, comma list, whitespace,
  case), wildcard semantics, CanSee matrix (FR1–FR5), group/bookmark filters
  (FR6, AC8/AC9), input non-mutation (NFR4).
- `internal/kube/wrappers`: annotation → `VisibleGroups` parsing.
- `internal/hajimari/customapps`: config `visibleGroups` propagation.

### Integration (manual, cluster)

1. Deploy behind Traefik + Authelia `forwardAuth` with
   `authResponseHeaders: Remote-User,Remote-Groups,Remote-Email`.
2. Create three LDAP users: `family-user` (group `family`), `friend-user`
   (group `friends`), `admin-user` (groups `admins,family`).
3. Annotate: Jellyfin `visible-groups: "family,friends"`, Immich
   `visible-groups: "family"`, Proxmox ingress `visible-groups: "admins"`,
   Sonarr unannotated.
4. Assert AC1–AC9 via browser for each user; assert `/api/apps` responses.
5. Negative: port-forward straight to the pod (`kubectl port-forward`) with a
  spoofed `Remote-Groups: admins` header — confirm it *does* filter in (trusted
  header by design) and therefore that the NetworkPolicy/ingress-only exposure
  from INV4 is a deployment requirement, not an application one.

### Regression

- `helm template` with stock values diffs only by new (unused) fields.
- Existing annotations (`enable`, `instance`, `group`, …) unaffected; suite of
  manual checks against stock config with no header present.

## 6. User Experience

### Happy paths

- **Family member** opens `home.example.com` → Authelia session (SSO) → sees
  Media tiles (Jellyfin, Sonarr), Photos (Immich), family bookmarks. No
  awareness that other tiles exist; no errors.
- **Friend** opens the same URL → sees shared Media tiles; family-only tiles
  are absent (not greyed). Single URL for everyone; membership decides content.
- **Admin** → union of `admins` + `family` tiles: everything.
- **New app deployed** with correct annotation → appears for the right groups
  within the 60s cache refresh; no restart, no frontend change.
- **Group membership changed** in LDAP → next request's header reflects it
  (forward-auth re-evaluates per request); tile visibility updates on reload.

### Unhappy paths

- **Not authenticated:** proxy redirects to Authelia login before Hajimari is
  reached. Hajimari itself never shows a login.
- **Authenticated but no matching group:** tile simply not rendered. The page
  never explains *why* (restricted items are invisible, not disabled).
- **Authelia down:** forward-auth fails closed at the proxy (503/redirect);
  Hajimari is never reached with a stale/absent header for annotated tiles
  (INV2 covers the gap anyway).
- **Header missing entirely (broken middleware or direct pod access):**
  unannotated tiles still render; every annotated tile is hidden (INV2). The
  dashboard "looks empty-ish" — this is the designed tamper signal.
- **Typo in annotation group (`famly`):** tile hidden for everyone except
  unauthenticated-free (INV2). Detection: per-group tile-count assertion in
  monitoring (e.g. Uptime Kuma JSON query) or a Kyverno policy constraining
  values to a known enum.
- **Empty/garbage annotation value (`", ,"`):** parses to zero groups ⇒ treated
  as absent ⇒ visible to all (INV1). To hide from everyone, use a group no
  user holds.

### Deliberate non-goals

- Startpages (tokenized URLs) are not filtered in this revision.
- Per-user customization (theme/bookmarks) remains browser-local as upstream.
- Hajimari still performs no authentication and must not be exposed except
  through the forward-auth chain.
