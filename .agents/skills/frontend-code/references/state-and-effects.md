# State, Components, and Effects

Apply the grounding and implementation boundaries in [Frontend Code](../SKILL.md). This reference owns framework-neutral decisions; load the affected framework reference for its lifecycle and identity mechanisms.

## Contents

1. [State ownership](#state-ownership)
2. [Derive data and choose component boundaries](#derive-data-and-choose-component-boundaries)
3. [Reset and preserve intentionally](#reset-and-preserve-intentionally)
4. [Synchronize external systems](#synchronize-external-systems)

## State ownership

Choose an owner from who changes the value, who consumes it, and how long it must survive. Keep transient interaction state near the interaction. Share state when independently rendered consumers must agree on one value; choose their existing common owner rather than introducing a global store by default.

Distinguish editable drafts from authoritative server records. A draft can intentionally diverge while editing; define when a newer record replaces, merges with, or conflicts with it. Request caches own retrieved records; copying them into another store needs a separate purpose and synchronization contract. For request ownership, use [Async and API Interaction](async-and-api.md#request-ownership-and-stale-completion).

Use URL state when navigation, sharing, reload, or browser history must reproduce the selection. Use persistent storage only when the contract requires survival beyond the view/session, with a defined reset and user scope. Neither URL nor browser storage is an appropriate place for secrets. Do not broaden persistence merely to prevent a local rerender from resetting input.

## Derive data and choose component boundaries

Compute values that are fully determined by current inputs instead of maintaining a synchronized copy. Filtered items, totals, and validity commonly belong here. Store a snapshot only when its historical meaning matters, such as the values actually submitted. Memoization addresses a demonstrated computation or identity need; it does not establish state ownership or correctness.

Place rendering, interaction, and resource ownership behind boundaries that match their responsibilities. Extract a component when it has a coherent contract, independent lifetime, or useful reuse; line count alone does not establish a boundary. Keep the owning value and its update policy explicit at the caller. Prefer existing composition patterns over duplicating state in siblings or giving a presentation component unrelated request responsibilities. General reuse decisions remain with [Code Change](../../code-change/SKILL.md#find-existing-behavior-before-adding-code).

## Reset and preserve intentionally

Define what identifies the edited entity or interaction session. Decide which state survives a same-entity rerender, entity switch, route change, dismissal, or successful submission. For an editor that discards drafts on record switch, tie reset to record identity; preserve draft input during rerenders of the same record. Do not use random identities or reset on every render to force updates.

Check dependent state too: selection, validation errors, pending work, and focus may belong to the previous entity. Reset only the state covered by the transition contract. If unsaved work needs confirmation or persistence, that is product behavior to establish, not an automatic consequence of changing a component key.

## Synchronize external systems

Separate pure rendering/derivation, explicit user events, and synchronization with an external system. A save caused by submission belongs to that event; a subscription that must follow the currently selected room belongs to that room's resource lifetime. An effect is not a general mechanism for forwarding values between internal stores.

For each listener, subscription, timer, observer, or connection, identify acquisition, replacement conditions, and release. Cleanup releases the particular resource acquired by that setup; repeated setup after cleanup must leave the intended active resources. Do not suppress lifecycle checks to hide duplicate subscriptions. Framework-specific dependency and captured-value semantics belong in the framework reference.

Async completion has its own ownership policy in [Async and API Interaction](async-and-api.md#request-ownership-and-stale-completion). Releasing or aborting work does not alone prove that every continuation stops publishing. Verify the relevant transition or resource behavior through the existing runner and [AAA Testing](../../aaa-testing/SKILL.md); this reference does not prescribe new testing dependencies.
