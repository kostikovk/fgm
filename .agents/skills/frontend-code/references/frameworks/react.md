# React decisions

Use this reference for React work after the [frontend contract and owner gates](../../SKILL.md). For Next.js, also select the affected router in [Next.js decisions](nextjs.md#ground-the-version-and-router). Shared [state and effects](../state-and-effects.md) and [async/API ownership](../async-and-api.md) supply the behavioral decisions; this file maps them to React.

## Ground the React version

Inspect installed React/React DOM versions, JSX conventions, rendering environment, component callers and nearby tests. A package range or a current documentation example does not establish that an API is supported by the installed project. Verify matching official documentation before adopting version-dependent hooks or compiler assumptions; unavailable evidence is uncertainty, not a reason to upgrade.

## State and identity

- Derive values from current props/state during render when they add no independent lifetime. An effect that stores filtered props creates another value to synchronize; remove that duplication when the contract calls for current results. Memoization is a performance choice supported by the actual workload, not a correctness requirement.
- Choose the owner from the consumers and required lifetime. Lift shared state to their useful common owner; use an existing provider/store when it already owns the contract. Context distributes a value; introducing it does not resolve which component or service should own it.
- Treat props/state as immutable render snapshots. When the next state depends on pending previous state, use the supported updater form rather than a captured value. Preserve established immutable-update utilities; do not add a state library for a bounded update.
- React associates state with component type and position, including keys. For an editor that must discard a draft on record change, a stable record key at the editor boundary can reset its subtree. Verify that same-record rerenders preserve the draft. A key reset also resets descendants and can affect focus; choose its boundary deliberately.
- Use stable domain identity for reorderable items. Random keys, keys that change on every render, and index keys where order/identity changes can lose state or associate it with the wrong item. Avoid nested component definitions that accidentally change component identity.

## Components and hooks

Keep rendering pure: compute output from props, state and context; perform user-triggered operations in the event that owns them, and external synchronization in its lifecycle. Network mutations during render can repeat without a matching user action.

Keep hooks within the installed version's rules. Ordinary state/effect hooks need consistent top-level call order; do not turn an API-specific exception into permission for conditional hook calls. A custom hook shares a coherent piece of behavior, not a reason to hide unrelated operations behind one interface.

Prefer composition and explicit props for bounded relationships. Split a component where a responsibility, reuse boundary or independent lifetime becomes clearer; file length alone does not establish that boundary. Preserve accessible primitives and their interaction contract when wrapping them; use [accessibility and interaction](../accessibility-and-interaction.md) for those decisions.

## Effects and closures

Before adding an effect, identify the external system being synchronized. Derived rendering data and a submission caused by one user event generally belong in render/event logic instead. Use the framework/router or existing data client for data ownership when it already provides the required lifecycle.

An effect captures values from its render. Include its reactive dependencies and change the structure when a dependency exposes an unwanted coupling; suppressing dependency checks can preserve a stale closure. Avoid blanket empty arrays, forced reruns or new memoization solely to hide that coupling.

Match setup with cleanup for the resources it acquires: unsubscribe, remove the same listener, clear the timer, or release the owned connection. Development Strict Mode's extra setup/cleanup cycle is evidence to check symmetry; disabling it or guarding the second setup can hide the defect. Cleanup also needs to handle dependency changes, not only unmount.

For asynchronous reads, identify which request may publish results. Cleanup can invalidate an obsolete completion even when cancellation is ignored; aborting transport alone does not prove stale results cannot replace current results. Apply [async/API ownership](../async-and-api.md) for result ordering, observable states and mutation uncertainty.

## Verification boundaries

Use [AAA for React components](../../../aaa-testing/references/react-tests.md) for public UI assertions and the project's established test environment. For an ordering defect, control completion order through an existing/injected seam; test the obsolete response finishing after the current response. For identity changes, check both intentional reset and preservation. For resource ownership, observe acquisition/release at the relevant lifecycle boundary.

Report DOM assertions separately from browser keyboard/focus checks and server rendering. A mocked effect test or render count alone does not establish the user-visible contract or runtime behavior outside that environment. Implementation order remains owned by [code-change](../../../code-change/SKILL.md) and its TDD prerequisite.

## Official sources

These sources explain mechanisms, not the installed application's version or behavior. Select matching-version guidance when choosing APIs.

- [You Might Not Need an Effect](https://react.dev/learn/you-might-not-need-an-effect): derived values, events, resets and fetching races.
- [Preserving and Resetting State](https://react.dev/learn/preserving-and-resetting-state): position, component identity and keys.
- [Rules of React](https://react.dev/reference/rules): rendering purity, immutable snapshots and hook rules.
- [useState](https://react.dev/reference/react/useState): queued updater functions and deliberate keyed resets.
- [Synchronizing with Effects](https://react.dev/learn/synchronizing-with-effects): synchronization, cleanup and development checks.
- [Lifecycle of Reactive Effects](https://react.dev/learn/lifecycle-of-reactive-effects): dependencies and resynchronization.
