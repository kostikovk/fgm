# AAA with JavaScript/TypeScript, Vitest, and Bun

Use the runner, file suffix, module style, and package scripts already established in the project. Do not mix imports from `vitest` and `bun:test` in one test file or migrate runners as part of an AAA change. The examples below are self-contained teaching fixtures; in a project, import the real subject instead of redeclaring its implementation in the test.

## Contents

1. [Vitest: TypeScript and table-driven scenarios](#vitest-typescript-and-table-driven-scenarios)
2. [Bun: JavaScript and an asynchronous outcome](#bun-javascript-and-an-asynchronous-outcome)
3. [Framework decisions](#framework-decisions)
4. [Validation and official references](#validation-and-official-references)

## Vitest: TypeScript and table-driven scenarios

Save this complete example as `label.vitest.test.ts` in an isolated example project. Every expanded title identifies the expected outcome and the condition. The subject is named by `describe`.

```ts
import { describe, expect, test } from "vitest";

function normalizeLabel(input: string): string {
  const label = input.trim();
  if (label.length === 0) throw new Error("Label is required");
  return label;
}

describe("normalizeLabel", () => {
  test.each([
    { condition: "surrounded by spaces", input: " invoice ", expected: "invoice" },
    { condition: "surrounded by tabs", input: "\tinvoice\t", expected: "invoice" },
  ])("should return the trimmed label when input is $condition", ({ input, expected }) => {
    // Arrange
    const rawLabel = input;

    // Act
    const actual = normalizeLabel(rawLabel);

    // Assert
    expect(actual).toBe(expected);
  });

  test("should throw a required-label error when input contains only whitespace", () => {
    // Arrange
    const input = " \t ";

    // Act: define the operation; the matcher invokes it below.
    const act = () => normalizeLabel(input);

    // Assert
    expect(act).toThrow(new Error("Label is required"));
  });
});
```

For JavaScript, use a `.test.js` file and remove the parameter/return annotations. Keep the same AAA structure and titles. The exact error message is asserted here because it is part of this example's contract; use the project's actual error contract.

## Bun: JavaScript and an asynchronous outcome

Save this complete example separately as `label.bun.test.js`. It uses the same behavior with an asynchronous API to show how the action and assertions remain distinct.

```js
import { describe, expect, test } from "bun:test";

async function normalizeLabel(input) {
  const label = input.trim();
  if (label.length === 0) throw new Error("Label is required");
  return label;
}

describe("normalizeLabel", () => {
  test("should resolve to the trimmed label when input has surrounding spaces", async () => {
    // Arrange
    const input = " invoice ";
    const expected = "invoice";

    // Act
    const actual = await normalizeLabel(input);

    // Assert
    expect(actual).toBe(expected);
  });

  test("should reject with a required-label error when input contains only whitespace", async () => {
    // Arrange
    const input = " \t ";

    // Act
    const operation = normalizeLabel(input);

    // Assert
    await expect(operation).rejects.toThrow(/^Label is required$/);
  });
});
```

Bun also runs `.test.ts` files. To adapt this example to TypeScript, annotate the subject as `async function normalizeLabel(input: string): Promise<string>`; the test bodies need no other changes.

## Framework decisions

- Capture successful async results with `await` in Act. For expected rejection, create the promise in Act and immediately await its rejection assertion in Assert. Always await or return async matcher assertions so the runner observes failures.
- For synchronous exceptions, prefer the named deferred action shown above. The matcher invokes it and verifies the exception; this is a supported combined Act/Assert boundary. Inline `expect(() => subject()).toThrow(...)` is also idiomatic when it matches project conventions. Do not eagerly call the subject before passing it to `toThrow`.
- Use explicit capture when several properties of the same exception must be inspected or the user requires the subject to execute literally in Act. Keep assertions outside the `try` and make the no-exception case fail. Capture once; do not invoke a stateful subject repeatedly just to run different exception matchers.
- Match only the promised error contract. In Vitest and Bun, a string supplied to `toThrow` matches a substring; use an anchored regular expression for an exact message. Check the runner's semantics before assuming an expected Error instance also proves the thrown value's class. When both class and properties matter and one matcher cannot express them, capture the error once and verify them directly.
- Arrange mocks with the chosen runner's APIs and restore them using its lifecycle hooks. Keep fixture mutation isolated per test; use existing fake-time seams rather than adding sleeps.
- Assert a dependency call only when it is part of the behavior named in `should ... when ...`. Prefer observable return values and effects to implementation-specific call counts.

## Validation and official references

Prefer project scripts. In an isolated project with Vitest installed, run `npx vitest run label.vitest.test.ts`; with Bun, run `bun test ./label.bun.test.js`. Keep the runners' examples separate so one runner does not discover the other runner's test files. Run the project's TypeScript check separately when applicable; a passing test run does not prove type correctness.

- [Vitest test API](https://vitest.dev/api/test.html): test names and parameterized scenarios.
- [Vitest expect API](https://vitest.dev/api/expect.html): error assertions and async matchers.
- [Fluent Assertions exception guidance](https://fluentassertions.com/exceptions/): an explicit AAA example defines a deferred action in Act and invokes it through an exception assertion. This supports the structural convention, not a requirement to use its .NET APIs in JavaScript.
- [Bun test runner](https://bun.sh/docs/test): JavaScript/TypeScript tests and `bun test`.
- [Bun writing tests](https://bun.sh/docs/test/writing-tests): `bun:test`, assertions, and async tests.
- [Bun Expect API](https://bun.com/reference/bun/test/Expect): constructor and message matching for `toThrow`.
