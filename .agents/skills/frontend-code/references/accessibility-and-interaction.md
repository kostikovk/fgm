# Accessibility and Interaction

Apply the grounding and implementation boundaries in [Frontend Code](../SKILL.md). Use the product contract and existing accessible primitives to decide which interactions and states apply; not every component needs every async state.

## Contents

1. [Semantics and names](#semantics-and-names)
2. [Keyboard and focus](#keyboard-and-focus)
3. [Forms and validation](#forms-and-validation)
4. [Async interaction states](#async-interaction-states)
5. [Verification boundaries](#verification-boundaries)

## Semantics and names

Prefer native controls for their intended behavior and established project primitives for composite widgets. Use links for navigation and buttons for actions. Set button type deliberately inside a form so an auxiliary action does not accidentally submit it. A clickable generic element with an ARIA role still needs its interaction behavior implemented. [HTML button semantics](https://html.spec.whatwg.org/multipage/form-elements.html#the-button-element), [WAI guidance on native semantics](https://www.w3.org/WAI/ARIA/apg/practices/read-me-first/).

Give controls an accessible name that identifies their purpose. Prefer a visible associated label for inputs; proximity and placeholder text do not establish label association. Check the rendered association, including unique IDs for repeated fields. Preserve names and semantics when wrapping a native control or an existing primitive. [WAI labeling controls](https://www.w3.org/WAI/tutorials/forms/labels/).

## Keyboard and focus

Choose the interaction pattern before implementing a custom widget: entry point, activation keys, movement within it, dismissal, and focus destination. Preserve visible focus and a meaningful tab order. Do not add focusability without the corresponding keyboard behavior.

For a modal dialog, move focus inside on opening, keep keyboard focus within the modal while open, support Escape dismissal, and normally return focus to the invoking control on close. If that control no longer exists, choose a destination consistent with the workflow. Reuse an established dialog primitive and verify its behavior rather than adding only `role="dialog"`. [WAI modal dialog pattern](https://www.w3.org/WAI/ARIA/apg/patterns/dialog-modal/).

## Forms and validation

Keep entered values recoverable after a rejected submission. Associate field-specific instructions/errors with their controls; identify invalid fields and explain how to correct them. A form-level failure or summary should be discoverable without relying solely on color or an unrelated toast. Decide how submission errors are announced or focused without moving focus for every keystroke. [WAI user notifications](https://www.w3.org/WAI/tutorials/forms/notifications/).

Treat pending submission as an interaction state: communicate progress, prevent duplicates according to the operation contract, and retain a usable recovery path. UI disabling does not establish server authorization or idempotency. Ambiguous mutation outcomes follow [Recovery and mutations](async-and-api.md#recovery-and-mutations).

## Async interaction states

Distinguish initial/pending retrieval, successful populated results, successful empty results, and retrieval failure when applicable. An empty collection is evidence of successful absence only after a successful response. A failure should provide the permitted recovery action; changing filters or creating an item may be appropriate for empty success, not for an unavailable API.

If the contract retains previously valid data during refresh, show that data with its refresh/error status instead of erasing it. Avoid showing an old entity's results as the new entity's current data. Preserve useful input across failure and retry; define when success clears a draft or changes the view.

Expose meaningful pending, result, and failure status to assistive technology using an appropriate status/live-region mechanism when focus remains elsewhere. Choose announcement urgency and frequency to fit the event; announcing every intermediate update can overwhelm users. A live region does not replace required focus handling in a dialog or error workflow. [WCAG status messages guidance](https://www.w3.org/WAI/WCAG22/Understanding/status-messages.html).

## Verification boundaries

Use [AAA Testing](../../aaa-testing/SKILL.md) for test structure and public UI assertions. Verify relevant names, roles, error associations, and transitions with the existing runner. Exercise actual keyboard navigation, focus movement/return, visible focus, and layout behavior in a browser when validating those interactions. Add assistive-technology checks where the requested accessibility scope needs them.

Record DOM assertions, automated accessibility checks, browser interaction, and assistive-technology checks separately. Passing role queries or an automated scan alone does not establish full accessibility or standards conformance. Report any required check that could not run.
