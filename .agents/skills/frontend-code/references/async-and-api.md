# Async and API Interaction

Apply the grounding and implementation boundaries in [Frontend Code](../SKILL.md). This reference owns frontend consumption and publication decisions; backend/schema changes remain with [API Contract Changes](../../code-change/references/api-contracts.md#keep-api-contracts-synchronized).

## Contents

1. [Consume the existing contract](#consume-the-existing-contract)
2. [Request ownership and stale completion](#request-ownership-and-stale-completion)
3. [Recovery and mutations](#recovery-and-mutations)
4. [Cache and displayed consistency](#cache-and-displayed-consistency)

## Consume the existing contract

Inspect the actual client, response/error mapping, authentication mechanism, and affected consumers. Reuse the established transport/query layer where its contract fits. Preserve differences between unavailable data, successful absence, validation rejection, unauthorized access, and other failures that affect the UI. With native fetch, an HTTP error response is not itself a rejected fetch promise; inspect response status before treating the payload as success. [Fetch response processing](https://fetch.spec.whatwg.org/#fetch-method).

Choose request inputs and result ownership from the feature contract, including record/query identity and relevant user scope. Match client cache keys to inputs that change the result; do not place secrets in keys that are exposed to logs or URLs. Server execution and a client authentication check do not replace server authorization. For Next execution/data boundaries, read the applicable [Next.js guidance](frameworks/nextjs.md).

A mock describes its configured behavior, not proof that the real API supports it. If recovery needs an idempotency key, status lookup, or cancellation guarantee that the current contract lacks, identify that gap rather than inventing an endpoint or response field.

## Request ownership and stale completion

Define which operation may publish data, errors, and pending-state completion. For latest-query-wins behavior, an older request finishing after its replacement must not overwrite any of those states. Use the existing query/client mechanism when it supplies that guarantee, or track operation identity/lifetime at the owning boundary.

Cancellation can reduce unnecessary work, but publication protection must survive a transport or continuation that does not honor it. Native abort signaling can stop fetch and body consumption; it does not establish that arbitrary async work or a server mutation has been undone. [DOM abort signaling](https://dom.spec.whatwg.org/#aborting-ongoing-activities), [Fetch abort handling](https://fetch.spec.whatwg.org/#fetch-method).

Decide what happens when inputs change, the view is dismissed, or the owning entity/user changes. Protect both success and failure continuations, including finalizers that might clear another operation's pending indicator. A cancellation initiated by replacement should not become an unrelated error message for the replacement request. For displayed pending/empty/error behavior, use [Async interaction states](accessibility-and-interaction.md#async-interaction-states).

## Recovery and mutations

Separate duplicate submission prevention from retry safety. A timeout or lost response can leave a mutation's completion unknown. Do not claim it failed to persist or automatically repeat a non-idempotent operation without a supported guarantee. Preserve the draft and explain the uncertainty using the operation's existing confirmation/status/retry policy; an absent policy is a contract question.

For an optimistic update, identify the affected record/cache/view, concurrent operations, and reconciliation strategy before changing displayed data. A blanket rollback to an old snapshot can erase newer work. Prefer established client behavior or a bounded update tied to the operation; show recoverable failure and reconcile with authoritative data when required.

Treat successful persistence and updated display as separate observations. Clear input or navigate only when the contract calls for it. Use isolated request seams for verification; a frontend test does not authorize live charges, messages, or other external mutations.

## Cache and displayed consistency

Trace the source of the visible data and the caches between it and the authoritative result. Determine user/tenant scope and freshness requirements separately. Invalidating a cache cannot repair a read that was authorized for the wrong user; authorization and scope must already be correct.

After mutation, choose the existing mechanism that updates the affected view: apply a returned record, update/invalidate a relevant client cache, refetch, or use a framework-specific server refresh mechanism. Verify which layers that mechanism actually affects. Do not assume a view refresh clears every cache or that request deduplication implies persistent caching.

Use [AAA Testing](../../aaa-testing/SKILL.md) to arrange controlled completion order and public outcomes with the existing runner. Check stale success and stale failure where relevant, successful empty versus rejected retrieval, and mutation recovery/consistency. Pure client or DOM tests do not prove server cache isolation, authorization, hydration, or external persistence; report those boundaries explicitly.
