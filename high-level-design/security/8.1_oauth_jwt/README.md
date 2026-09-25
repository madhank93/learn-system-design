# OAuth 2.0 and JWT

OAuth 2.0 is the industry-standard authorization framework that lets users grant third-party applications limited access to their resources without handing over credentials. JWT (JSON Web Token) is the stateless token format that makes this work at scale — a resource server can verify a user's identity and permissions by validating the token's cryptographic signature without any database call. Together they underpin every "Login with Google" button, every mobile app that calls an API, and every machine-to-machine service account in a modern backend. Getting the details wrong (wrong grant type, insecure storage, missing claim validation) is a direct path to account takeover.

---

## OAuth 2.0 Roles

```mermaid
flowchart LR
    RO["🧑 Resource Owner\n(User)"]
    Client["📱 Client\n(Your App)"]
    AS["🔐 Authorization Server\n(Google, Auth0, Keycloak)"]
    RS["🗄️ Resource Server\n(Your API)"]

    RO -->|grants permission| Client
    Client -->|requests authorization| AS
    AS -->|issues tokens| Client
    Client -->|presents access token| RS
    RS -->|validates token| AS

    classDef user fill:#F39C12,stroke:#D68910,color:#000
    classDef client fill:#3498DB,stroke:#2471A3,color:#fff
    classDef authserver fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef api fill:#27AE60,stroke:#1E8449,color:#fff

    class RO user
    class Client client
    class AS authserver
    class RS api
```

| Role | Responsibility |
|---|---|
| Resource Owner | The user who owns the data and grants access |
| Client | The application requesting access on the user's behalf |
| Authorization Server | Issues access tokens after authenticating the user |
| Resource Server | Hosts the protected resources; validates tokens |

---

## Grant Types

```mermaid
flowchart TD
    Q1{Client is browser\nor mobile app?}
    Q2{Client can store\na secret securely?}
    Q3{Machine-to-machine?}

    Q1 -->|Yes| PKCE["Authorization Code + PKCE\n(most common for public clients)"]
    Q1 -->|No| Q2
    Q2 -->|Yes - server-side| AuthCode["Authorization Code\n(server-side web app)"]
    Q2 -->|No| Q3
    Q3 -->|Yes| CC["Client Credentials\n(service accounts, M2M)"]
    Q3 -->|No| Device["Device Code\n(TV, CLI tools)"]

    classDef recommended fill:#27AE60,stroke:#1E8449,color:#fff
    classDef standard fill:#3498DB,stroke:#2471A3,color:#fff
    classDef m2m fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef device fill:#F39C12,stroke:#D68910,color:#000

    class PKCE recommended
    class AuthCode standard
    class CC m2m
    class Device device
```

| Grant Type | Use case | Client secret? |
|---|---|---|
| Authorization Code + PKCE | SPA, mobile, public client | No — PKCE replaces secret |
| Authorization Code | Server-side web app | Yes — stored server-side only |
| Client Credentials | Backend → backend, service accounts | Yes — stored server-side only |
| Device Code | TV apps, CLI tools | No |
| Resource Owner Password | Legacy migration only — avoid | Yes — avoid entirely |

---

## Authorization Code + PKCE Flow

```mermaid
sequenceDiagram
    participant User
    participant App as Client App
    participant AS as Authorization Server
    participant API as Resource Server

    App->>App: Generate code_verifier (random 32 bytes)<br/>code_challenge = BASE64URL(SHA256(verifier))
    App->>AS: GET /authorize?response_type=code<br/>&client_id=...&code_challenge=...&scope=read:profile
    AS->>User: Show login page
    User->>AS: Enter credentials + approve scopes
    AS->>App: Redirect to callback?code=AUTH_CODE&state=...
    App->>AS: POST /token<br/>code=AUTH_CODE + code_verifier
    AS->>AS: Verify SHA256(code_verifier) == code_challenge
    AS->>App: { access_token, refresh_token, id_token, expires_in }
    App->>API: GET /profile<br/>Authorization: Bearer ACCESS_TOKEN
    API->>API: Validate token signature + claims
    API->>App: 200 { user profile data }
```

**Why PKCE matters:** Without it, an attacker who intercepts the authorization `code` (e.g., via a malicious app registered to the same redirect URI) can exchange it for tokens. PKCE binds the code to the originating request — only the app that generated `code_verifier` can redeem the code.

---

## JWT Structure

A JWT is three base64url-encoded segments joined by dots: `header.payload.signature`

```mermaid
flowchart LR
    subgraph JWT["JWT Token"]
        H["Header\n{\n  alg: RS256\n  typ: JWT\n  kid: key-2024-01\n}"]
        P["Payload\n{\n  sub: user123\n  iss: auth.example.com\n  aud: api.example.com\n  exp: 1720000000\n  iat: 1719996400\n  scope: read:profile write:posts\n}"]
        S["Signature\nRS256(\n  base64url(header)\n  + . +\n  base64url(payload),\n  rsaPrivateKey\n)"]
    end

    classDef hdr fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef payload fill:#3498DB,stroke:#2471A3,color:#fff
    classDef sig fill:#27AE60,stroke:#1E8449,color:#fff

    class H hdr
    class P payload
    class S sig
```

### Standard Claims

| Claim | Meaning | Must validate |
|---|---|---|
| `sub` | Subject — user's unique ID | Present, matches expected user |
| `iss` | Issuer — who issued the token | Matches your auth server URL |
| `aud` | Audience — intended recipient | Includes your service's identifier |
| `exp` | Expiry (Unix epoch seconds) | Must be in the future |
| `iat` | Issued-at timestamp | Sanity check — not suspiciously old |
| `nbf` | Not-before timestamp | Token invalid before this time |
| `jti` | JWT unique ID | Used for revocation blocklist lookup |

---

## Token Validation Flow

```mermaid
flowchart TD
    Req["Request\nAuthorization: Bearer TOKEN"]
    Decode["Decode header\nExtract alg + kid"]
    Fetch["Fetch public key\nfrom JWKS endpoint\n(cached 60 min, bust on unknown kid)"]
    Verify["Verify signature\nRS256 / ES256 with public key"]
    Claims["Validate claims\nexp · iss · aud · nbf"]
    Allow["✅ Allow\nAttach user context to request"]
    Deny["❌ 401 Unauthorized"]

    Req --> Decode
    Decode --> Fetch
    Fetch --> Verify
    Verify -->|signature invalid| Deny
    Verify -->|valid| Claims
    Claims -->|expired, wrong iss/aud| Deny
    Claims -->|all valid| Allow

    classDef ok fill:#27AE60,stroke:#1E8449,color:#fff
    classDef fail fill:#E74C3C,stroke:#C0392B,color:#fff
    classDef step fill:#3498DB,stroke:#2471A3,color:#fff
    classDef start fill:#F39C12,stroke:#D68910,color:#000

    class Allow ok
    class Deny fail
    class Decode,Fetch,Verify,Claims step
    class Req start
```

**JWKS caching:** Fetching the public key on every request adds latency and load. Cache the JWKS response keyed by `kid` for 60 minutes; evict and re-fetch on encountering an unknown `kid`.

---

## Access Token vs Refresh Token

```mermaid
sequenceDiagram
    participant App
    participant AS as Authorization Server
    participant API

    App->>API: Request with access_token
    API->>App: 401 Unauthorized (token expired)
    App->>AS: POST /token<br/>grant_type=refresh_token<br/>refresh_token=LONG_LIVED_TOKEN
    AS->>AS: Validate refresh token (not revoked, not expired)
    AS->>App: New access_token + rotated refresh_token
    App->>API: Retry with new access_token
    API->>App: 200 OK
```

| | Access Token | Refresh Token |
|---|---|---|
| Lifetime | 15 min – 1 hour | Days to months |
| Sent to | Resource Server on every API call | Authorization Server only |
| If compromised | Limited blast radius — expires soon | High risk — rotate immediately |
| Storage | Memory or httpOnly cookie | Secure httpOnly cookie / OS keystore |
| Revocable | Hard (stateless) — rely on short TTL | Yes — stored server-side in DB |

**Refresh token rotation:** Issue a new refresh token with each use and invalidate the old one. If the old token is used again, it signals theft — invalidate the entire token family (all refresh tokens for that user session).

---

## Token Storage Security

```mermaid
flowchart TD
    Q{Where is the app running?}

    Q -->|Web browser| Web{Storage choice}
    Q -->|Mobile app| Mobile["✅ OS Keystore\niOS Keychain /\nAndroid Keystore"]
    Q -->|Backend service| Backend["✅ Environment variable\nor Secrets Manager\n(Vault, AWS Secrets Manager)"]

    Web --> HC["✅ httpOnly Cookie\n+ SameSite=Strict\n+ CSRF double-submit\nBest for web apps"]
    Web --> Mem["⚠️ In-memory only\nLost on refresh\nSafe from XSS"]
    Web --> LS["❌ localStorage\nReadable by any XSS\nDo not store tokens here"]

    classDef ok fill:#27AE60,stroke:#1E8449,color:#fff
    classDef warn fill:#F39C12,stroke:#D68910,color:#000
    classDef bad fill:#E74C3C,stroke:#C0392B,color:#fff

    class Mobile,Backend,HC ok
    class Mem warn
    class LS bad
```

---

## JWT Revocation Options

JWTs are stateless — the resource server validates them without calling the auth server. True instant revocation requires extra infrastructure:

| Strategy | Mechanism | Tradeoff |
|---|---|---|
| Short TTL | 15-minute access tokens | Permissions stay stale up to 15 min |
| Revocation blocklist | Store revoked `jti` in Redis; check on each validation | Extra Redis lookup per request |
| Token introspection | Resource server calls `/introspect` per request | Extra network hop; auth server becomes hot path |
| Refresh token revocation | Revoke refresh token; wait for access token to expire | User can't get new tokens; current access token lingers |

**Practical recommendation:** Short-lived access tokens (15 min) + refresh token rotation covers 95% of cases without infrastructure complexity.

---

## RS256 vs HS256

```mermaid
flowchart LR
    subgraph HS256["HS256 — Symmetric (Avoid in production)"]
        H_Sign["Auth server signs\nwith shared secret"]
        H_Verify["Every API verifies\nwith same shared secret"]
        H_Problem["⚠️ Secret must be distributed\nto all verifying services\n→ secret sprawl, exposure risk"]
    end

    subgraph RS256["RS256 — Asymmetric (Recommended)"]
        R_Sign["Auth server signs\nwith RSA private key\n(never leaves auth server)"]
        R_Verify["Any API verifies\nwith RSA public key\nfrom JWKS endpoint"]
        R_Benefit["✅ Public key safe to publish\nNo secret sharing required\nKey rotation via kid"]
    end

    classDef warn fill:#E67E22,stroke:#CA6F1E,color:#fff
    classDef ok fill:#27AE60,stroke:#1E8449,color:#fff

    class HS256,H_Sign,H_Verify,H_Problem warn
    class RS256,R_Sign,R_Verify,R_Benefit ok
```

---

## OpenID Connect (OIDC)

OIDC adds an identity layer on top of OAuth 2.0. OAuth answers "what can this app access?" — OIDC adds "who is the user?"

Requesting the `openid` scope causes the token endpoint to return an `id_token` JWT alongside the `access_token`. The `id_token` carries standard identity claims: `sub`, `email`, `name`, `picture`. Validate it like any JWT.

```mermaid
flowchart LR
    OAuth["OAuth 2.0\nAuthorization only\naccess_token"]
    OIDC["+ OpenID Connect\nAuthentication layer\naccess_token + id_token"]
    JWKS["JWKS Endpoint\n/oauth/jwks\nPublic keys for verification"]
    UserInfo["UserInfo Endpoint\nGET /userinfo\n{ sub, email, name, picture }"]

    OAuth -->|scope=openid| OIDC
    OIDC -->|validate id_token| JWKS
    OIDC -->|or fetch claims| UserInfo

    classDef base fill:#3498DB,stroke:#2471A3,color:#fff
    classDef ext fill:#8E44AD,stroke:#6C3483,color:#fff
    classDef endpoint fill:#27AE60,stroke:#1E8449,color:#fff

    class OAuth base
    class OIDC ext
    class JWKS,UserInfo endpoint
```

---

## Common Vulnerabilities

| Vulnerability | Description | Fix |
|---|---|---|
| `alg: none` attack | Attacker removes signature, sets `alg: none`; naive validators accept | Always allowlist algorithms; reject `none` |
| Algorithm confusion | RS256 token verified as HS256 with public key as secret | Lock algorithm in verifier; never derive from token header alone |
| Missing `aud` check | Token for service A accepted by service B | Always validate `aud` against your service identifier |
| Expired token accepted | Server skips `exp` validation | Never skip — check every claim |
| Token in URL | `GET /api?token=...` logged in proxies and browser history | Tokens in `Authorization: Bearer` header only |
| Overly broad scopes | Access token grants `admin:*` when only `read:profile` needed | Request minimum required scopes; enforce in resource server |

---

## Run This Implementation

```sh
docker compose up --build -d && sleep 2 && bash test.sh
```

```sh
go test -v ./auth-server/
```

The Go implementation includes: authorization code issuance, token exchange, JWT signing (HS256 for simplicity in the demo), and resource server validation middleware.

---

## Interview Angles

**Q: What is the difference between OAuth 2.0 and OpenID Connect?**
OAuth 2.0 is an authorization framework — it grants a client permission to access resources; it says nothing about who the user is. OIDC builds on OAuth 2.0 and adds authentication — it issues an `id_token` JWT with standardized identity claims (`sub`, `email`, `name`) so the client knows *who* granted access.

**Q: How do you revoke a JWT after issuance?**
JWTs are stateless, so revocation requires either: a Redis blocklist checked per request using the `jti` claim (adds latency), relying on a short TTL (15 min) so expired tokens are self-revoking, or using opaque access tokens with introspection endpoints. Most systems use short TTL + revocable refresh tokens.

**Q: Why use RS256 over HS256 for token signing?**
RS256 uses asymmetric keys — the auth server signs with a private key that never leaves it, and any service can verify using the public key published at the JWKS endpoint. HS256 requires distributing the shared secret to every service that validates tokens, creating secret sprawl and exposure risk.

**Q: What is PKCE and when is it required?**
PKCE (Proof Key for Code Exchange) prevents authorization code interception in public clients (mobile apps, SPAs) that cannot securely store a client secret. The client generates a `code_verifier`, sends its SHA256 hash as `code_challenge` in the authorization request, and proves ownership of the verifier when exchanging the code for tokens — so an intercepted code is useless without the original verifier.

**Q: Which claims must you always validate on a JWT?**
At minimum: `exp` (token has not expired), `iss` (token was issued by your expected auth server), `aud` (token is intended for your service), and the signature itself. Skipping `aud` enables cross-service token replay; skipping `exp` enables indefinite reuse of stolen tokens.

---

## Architecture Diagrams

> Generated from D2 sources in [`docs/diagrams/src/8.1_oauth_jwt/`](../../../docs/diagrams/src/8.1_oauth_jwt/).
> Regenerate: `just diagrams-topic 8.1_oauth_jwt`.

### Overview

The OAuth 2.0 authorization-code flow — the client app never sees the password, only tokens — plus the JWT's header.payload.signature anatomy.

![Oauth Jwt — overview](../../../docs/diagrams/svg/8.1_oauth_jwt/overview/index.svg)

**`invalid_signature` scenario:** a tampered token fails signature verification immediately — integrity is checked with math, not a database.

![Oauth Jwt — invalid_signature scenario](../../../docs/diagrams/svg/8.1_oauth_jwt/overview/invalid_signature.svg)

**`token_expired` scenario:** an expired token gets 401; the client redeems its refresh token for a new access token without re-login — short-lived access tokens bound the blast radius of theft.

![Oauth Jwt — token_expired scenario](../../../docs/diagrams/svg/8.1_oauth_jwt/overview/token_expired.svg)

### Request Flow

Stateless validation at the gateway: decode header → fetch the matching JWKS public key (cached) → verify signature → check exp/iss/aud → forward with identity headers. No session store, no auth-DB call.

![Oauth Jwt — request flow](../../../docs/diagrams/svg/8.1_oauth_jwt/request-flow/index.svg)

**`key_rotation` scenario:** the JWKS serves old and new keys simultaneously; the token's kid header selects the right one, so rotation is invisible to users.

![Oauth Jwt — key_rotation scenario](../../../docs/diagrams/svg/8.1_oauth_jwt/request-flow/key_rotation.svg)
