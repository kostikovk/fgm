# A small TDD cycle

Use the existing runner. This original TypeScript example uses Vitest; with Bun, change the test import to `bun:test`. Go, React, and Vue use the same development sequence and the relevant examples in `$aaa-testing`.

Requirement: `clampToZero` returns zero for a negative input and leaves a nonnegative input unchanged.

Write the next test first, in `clampToZero.test.ts`:

```ts
import { expect, test } from "vitest";
import { clampToZero } from "./clampToZero";

test("should return zero when the input is negative", () => {
  const input = -3;
  const expected = 0;

  const actual = clampToZero(input);

  expect(actual).toBe(expected);
});
```

If the module does not exist, the first run may report an import error. Record it as a scaffold failure, not behavioral RED. Add only a callable placeholder in `clampToZero.ts`:

```ts
export function clampToZero(input: number): number {
  throw new Error("Not implemented");
}
```

Run the test again. The discovered scenario now fails because the function cannot deliver zero. This is RED for the requested behavior. Implement only enough for this scenario:

```ts
export function clampToZero(input: number): number {
  return 0;
}
```

Run the test and relevant suite: GREEN. Assess refactoring; no useful structural change is needed at this size. Next, add a separate test before expanding the implementation:

```ts
test("should preserve the input when the input is positive", () => {
  const input = 5;
  const expected = 5;

  const actual = clampToZero(input);

  expect(actual).toBe(expected);
});
```

Run both tests. The new test fails with expected `5`, actual `0`, while the earlier case remains green. Update the implementation:

```ts
export function clampToZero(input: number): number {
  return input < 0 ? 0 : input;
}
```

Rerun both cases and relevant regressions. Add a zero-boundary test next; it may already pass because the current implementation satisfies that behavior. Verify that the assertion is meaningful rather than forcing a failure. Only refactor after GREEN, for example by replacing a conditional with a clearer existing utility if appropriate to the project; rerun the tests afterward. No utility or abstraction is required merely to demonstrate REFACTOR.

## What counts as evidence

| Phase | Useful evidence | Does not establish the phase |
| --- | --- | --- |
| RED | Discovered behavioral test fails against pre-change behavior; inspected failure explains the missing outcome | Test authored but never run; syntax/import/setup error; unrelated failure |
| GREEN | Same test and relevant regression tests pass after the scoped implementation | Skipped test; weakened assertion; mocked subject; only typecheck/build passed |
| REFACTOR | Small structural change preserves the contract and green tests, or an explicit assessment finds no useful cleanup | New behavior implemented without another test; cleanup while tests are failing |

Issue reproduction and fix verification require executed behavioral tests. Logs, manual checks, and code inspection help select the scenario but do not establish reproduction or GREEN. If no suitable test can execute, report the issue as unverified and describe the blocker.

An integration or UI test can supply RED when behavior crosses boundaries. Control dependencies through existing seams and observe the public result; a test that only checks a mock's return value cannot validate the implementation. Existing failing tests can supply RED for a bug if they demonstrate precisely the requested defect; do not duplicate a meaningful regression test just for ceremony.

## Sources and adaptation

- [Kent Beck — Canon TDD](https://newsletter.kentbeck.com/p/canon-tdd): list behaviors, turn one into a runnable test, make that test and previous tests pass, optionally refactor, then repeat.
- [Martin Fowler — Test Driven Development](https://martinfowler.com/bliki/TestDrivenDevelopment.html): test-first design and the repeated RED/GREEN/REFACTOR cycle.
- [Martin Fowler — Refactoring](https://refactoring.com/): small structural transformations preserve observable behavior and use frequent tests.
- [Superpowers — test-driven-development](https://github.com/obra/superpowers/blob/8ca22dba9a94f28898bbce59f2537ff4d87c747d/skills/test-driven-development/SKILL.md): an upstream reusable skill with explicit failing/passing verification. This skill adopts the observed-cycle discipline with original instructions and examples; it preserves user work rather than copying blanket deletion or permission rules.

Scaffolding, blocked execution, test-only work, and pure-refactor routing are operational policies chosen for this skill. AAA structure and `should ... when ...` names are the user's convention; the cited TDD sources do not mandate that spelling.
