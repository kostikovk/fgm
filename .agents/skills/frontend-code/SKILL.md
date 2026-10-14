---
name: frontend-code
description: Implement, refactor, or review frontend UI and client behavior with guidance on state ownership, components, effects, accessibility, async UI states and API consumption. Adds version-aware React and Next.js detail when those frameworks are used; review-only requests remain read-only.
---

# Frontend Code

Produce UI behavior that fits the project's contracts and stack, including interaction, failure recovery and affected server/client boundaries.

## Ground the change — always apply

Inspect project instructions, manifests and lockfile, affected UI/callers, existing primitives and nearby tests. Identify the observable behavior and existing state/data/API mechanisms before choosing a solution. Establish installed versions, rendering environment and, for Next.js, the affected App/Pages route and relevant cache configuration. Mixed projects route by the changed path, not merely the presence of an `app` directory.

## Shared boundaries — always apply

- Use [code-change](../code-change/SKILL.md) for implementation/reuse decisions and task workspace preparation. Load [TDD](../tdd-workflow/SKILL.md) before executable edits and [AAA](../aaa-testing/SKILL.md) before test edits. Review-only work reports findings without editing; skill advice does not authorize a migration, installation or publication.
- Before choosing new or changed HTML/CSS/browser behavior, inspect existing project mechanisms, consult current native web patterns and check the project's browser/runtime support policy through [Web platform guidance](references/web-platform.md). Use Modern Web Guidance when available or official sources directly; retain suitable existing primitives and do not assume Baseline alone covers project targets. State-only or server-only changes do not require an unrelated platform search.
- Choose the owner and lifetime of state from the contract; derive values rather than synchronize duplicate facts. Separate interaction logic, rendering and external-system synchronization. Preserve intentional reset/persistence and release owned resources.
- Make relevant pending, success, empty and failure states distinguishable and recoverable. Establish which request may publish a result; cancellation alone may not stop stale completion. Preserve useful input and avoid automatic retries of mutations with uncertain completion.
- Prefer semantic controls and existing accessible primitives. Custom interactions need explicit keyboard/focus behavior and meaningful names; automation alone does not prove complete accessibility.
- Preserve server-only data and enforce authorization at server operations. Treat authorization, user/cache scope and mutation-to-view freshness as separate contracts. Follow the installed framework's router/configuration instead of importing current defaults.
- Unknown product behavior returns to [requirements-definition](../requirements-definition/SKILL.md); open technical alternatives to [design-decision](../design-decision/SKILL.md); unknown failure cause to [debugging](../debugging/SKILL.md). A clear bounded implementation continues directly.

## Navigation — choose the affected concern

The grounding and shared boundaries above apply to every task. Load only applicable concern sections and framework detail; a fragment locates a section, so read through the next heading of the same or higher level and follow its prerequisites. If bounded reads are unavailable, read only the relevant file. Another framework can use the shared rules without React assumptions.

| Concern | Read / prerequisite |
|---|---|
| Choosing HTML/CSS/browser APIs, custom UI mechanisms or new dependencies | [Web platform guidance](references/web-platform.md); inspect existing mechanisms and browser/runtime targets before adopting a pattern |
| State, component responsibilities, derived values, reset or effects | [State and effects](references/state-and-effects.md) |
| Controls, forms, keyboard/focus, validation or visible UI states | [Accessibility and interaction](references/accessibility-and-interaction.md) |
| API reads, racing requests, recovery, mutations or cache/view consistency | [Async and API](references/async-and-api.md); schema/backend contract changes also follow [API contract ownership](../code-change/references/api-contracts.md#keep-api-contracts-synchronized) |
| React implementation | [React](references/frameworks/react.md), plus shared sections selected by the concern |
| Next.js implementation | [Next.js](references/frameworks/nextjs.md), its version/router section and applicable [React](references/frameworks/react.md) sections; follow the shared concern routes too |
| React component test mechanics | [AAA React decisions](../aaa-testing/references/react-tests.md#apply-the-pattern-in-a-project); AAA entrypoint is required before test edits |

## Verify and report

Run focused public-behavior tests and applicable project typecheck, lint, build and integration checks. Check actual keyboard/focus, navigation or hydration in a browser when those contracts are affected and a suitable environment exists. Reuse established runners and safe API seams; do not probe live mutation endpoints to validate a fixture.

Report the observable result, state/boundary choice, checks actually run and material gaps. Distinguish supplied records from execution, DOM tests from server/browser behavior, and automated accessibility checks from conformance. Missing prerequisites remain explicit; do not mark blocked checks passed. Inspect the diff for lost input, cleanup, authorization, contract compatibility or unrelated changes.
