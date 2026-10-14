# Next.js decisions

Use alongside [React decisions](react.md) and the [frontend contract and owner gates](../../SKILL.md). Load the affected router section, then the cross-router security, consistency or hydration section relevant to the change. Router coexistence does not require a migration.

## Ground the version and router

Inspect the installed Next.js/React versions, lockfile, `next.config.*`, build scripts, affected route and imported data/client modules. Distinguish `app/` or `src/app/` from `pages/` or `src/pages/` at the affected path. In a mixed project, select APIs for that path rather than inferring one router from package presence.

Check rendering/deployment mode and configuration that changes caching behavior, including Cache Components when supported/enabled. Match official documentation to the installed version and configuration before proposing routing, caching or server-operation APIs. Unversioned docs and current examples cannot establish older defaults. If version-specific evidence is unavailable, preserve a supported precedent and report the unresolved choice; do not silently upgrade.

## App Router boundaries

- Trace the module graph at the interactive boundary. `'use client'` marks a client module boundary, not browser-only execution. Keep it near the interactive subtree when that preserves the contract; marking a broad layout client can draw imported code into the client graph.
- Keep secrets and privileged dependencies in the server-owned path. Pass only the data the user may receive, using the supported React serialization contract. A server-rendered child passed through composition is different from importing its server implementation into a client module; inspect the actual imports.
- Place reads where the contract requires their execution, freshness and authorization, reusing the established server data layer or client cache. Server-rendered data does not imply every later interactive update must become an effect fetch.
- Select App Router navigation, route loading/error boundaries and server operations only after version/configuration checks. Route boundaries handle their supported execution scope; they do not automatically cover every client request or recoverable form error. Keep local pending/error/empty/recovery states aligned with [interaction guidance](../accessibility-and-interaction.md).

## Pages Router

Use the page's established data-loading path. `getServerSideProps` supports request-time server rendering; `getStaticProps` supports static generation under the installed version's regeneration policy. Browser-owned refreshes can use the existing client data mechanism. Choose among them from the actual personalization/freshness contract rather than moving the page to App Router.

Page props are exposed to the client even when produced on the server; omit secrets and unauthorized fields. Keep privileged imports within the server data path, and inspect any imports used by the rendered page. Use the supported Pages Router navigation and API endpoint conventions; App Router `next/navigation`, special loading files or Server Action examples are not drop-in replacements for a Pages path.

## Authorization and cache scope

Treat these as separate decisions:

1. **Authorization:** derive identity from the server's trusted session/context and verify access to the requested resource near its data operation. A client-supplied user ID or a hidden button is not authority. Check mutations independently of the page that exposes them; `'use server'` alone does not authorize callers.
2. **Cache isolation:** trace every cache holding personalized data, including custom module maps and the existing client cache. Identify lifetime, key, tenant/user/resource scope and logout/account-switch behavior. Authorized fetching can still leak data if a later cache hit belongs to another session. Avoid shared caching of private data unless its permitted scope is established.
3. **Freshness:** determine what changes after an authorized write. Invalidation cannot repair a cache whose identity boundary was wrong.

Distinguish render/request memoization, persistent server data/output caches and client/router caches. They have different lifetimes and invalidation behavior. Do not infer one layer's policy from another or encode unconditional defaults; verify configured behavior using matching-version documentation and the actual read path.

## Mutation freshness

Trace write → subsequent read/cache → visible result. State whether the contract allows stale data, requires immediate read-your-own-writes, or uses an optimistic view. Preserve user input and handle uncertain completion through [async/API guidance](../async-and-api.md).

Choose an existing supported update mechanism that reaches the affected layer: client-cache update/invalidation, a fresh server read, or version-appropriate path/tag invalidation. Refreshing a route is not evidence that every persistent cache was invalidated. Stale-while-revalidate may serve an old result and therefore needs a different check from immediate freshness. Verify API availability, invocation context and actual semantics before selecting a mechanism; do not paste a current cache/action example into an older project.

Check the submitting view and any other affected consumer specified by the contract. Optimistic state needs a defined reconciliation/error outcome; a successful local update alone does not prove another consumer receives the new server value.

## Hydration and verification

Client Components can participate in server prerendering. Keep browser globals out of paths executed on the server and make initial server/client output agree. If browser-only state is required, define its initial presentation and update through the supported lifecycle or existing browser-only boundary. Disabling SSR or suppressing hydration warnings needs a concrete contract-based reason, not a blanket response to a mismatch.

Use the installed project scripts to check the relevant build/render/navigation behavior. [React DOM tests](../../../aaa-testing/references/react-tests.md) can establish local UI outcomes; they do not execute Next server boundaries, cache isolation or hydration. For private caches, use controlled two-session identities; for mutation freshness, observe the actual subsequent read and displayed result. Report decision review, DOM tests, Next execution and browser checks separately, with missing environments explicit.

## Official sources

Use these topic routes with matching-version evidence; their live examples and defaults may differ from the project.

- [Server and Client Components](https://nextjs.org/docs/app/getting-started/server-and-client-components): client module boundaries, prerendering, composition and permitted props.
- [App Router migration guide](https://nextjs.org/docs/app/guides/migrating/app-router-migration): App/Pages API distinctions; reading it does not authorize migration.
- [getServerSideProps](https://nextjs.org/docs/pages/api-reference/functions/get-server-side-props) and [getStaticProps](https://nextjs.org/docs/pages/api-reference/functions/get-static-props): Pages rendering/data paths and client-visible props.
- [Data Security](https://nextjs.org/docs/app/guides/data-security) and [use server](https://nextjs.org/docs/app/api-reference/directives/use-server): server-owned data and operation authorization.
- [Server Actions guide](https://nextjs.org/docs/app/guides/server-actions) and [revalidatePath](https://nextjs.org/docs/app/api-reference/functions/revalidatePath): version/configuration-dependent mutation consistency mechanisms.
