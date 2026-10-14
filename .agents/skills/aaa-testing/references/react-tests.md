# AAA for React components

Test the component's public UI contract: what users can see or do under the named condition. Use accessible roles and names, realistic user interactions, and assertions about the rendered outcome. Avoid checking React state, hook calls, or CSS selectors when the behavior can be observed through the UI.

This TypeScript example uses Vitest, React Testing Library, `user-event`, and `jest-dom` in a DOM environment. Reuse these when established in the project; this example does not authorize installing them or changing the project's runner. For JavaScript, use `.jsx` and remove the props type and parameter annotation.

## Contents

1. [Component fixture: `Counter.tsx`](#component-fixture-countertsx)
2. [Tests: `Counter.test.tsx`](#tests-countertesttsx)
3. [Apply the pattern in a project](#apply-the-pattern-in-a-project)
4. [Validation and official references](#validation-and-official-references)

## Component fixture: `Counter.tsx`

This small component makes the example runnable. In a real project, import the existing component instead of recreating it in a test.

```tsx
import * as React from "react";

type CounterProps = {
  initialCount?: number;
  disabled?: boolean;
};

export function Counter({ initialCount = 0, disabled = false }: CounterProps) {
  const [count, setCount] = React.useState(initialCount);

  return (
    <section aria-label="Counter">
      <p role="status">Count: {count}</p>
      <button disabled={disabled} onClick={() => setCount((value) => value + 1)}>
        Increment
      </button>
    </section>
  );
}
```

## Tests: `Counter.test.tsx`

The first scenario tests a visible count increase after a click. The second tests that a disabled control leaves the count unchanged. Each title starts with `should` and names both the expected behavior and condition.

```tsx
// @vitest-environment jsdom
import * as React from "react";
import { afterEach, describe, expect, test } from "vitest";
import { cleanup, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "@testing-library/jest-dom/vitest";
import { Counter } from "./Counter";

afterEach(cleanup);

describe("Counter", () => {
  test("should increase the displayed count when the increment button is clicked", async () => {
    // Arrange
    const user = userEvent.setup();
    render(<Counter initialCount={2} />);
    const button = screen.getByRole("button", { name: "Increment" });

    // Act
    await user.click(button);

    // Assert
    expect(screen.getByRole("status")).toHaveTextContent(/^Count: 3$/);
  });

  test("should keep the displayed count unchanged when the increment button is disabled", async () => {
    // Arrange
    const user = userEvent.setup();
    render(<Counter initialCount={2} disabled />);
    const button = screen.getByRole("button", { name: "Increment" });

    // Act
    await user.click(button);

    // Assert
    expect(button).toBeDisabled();
    expect(screen.getByRole("status")).toHaveTextContent(/^Count: 2$/);
  });
});
```

## Apply the pattern in a project

- Arrange props, providers, controlled dependency responses, and a fresh user-event session. Render is setup when the behavior under test is a later user interaction; if the initial render itself is the behavior, place it in Act.
- Await user interactions in Act. A sequence such as typing into a form and submitting it is one conceptual action when form submission is the scenario under test.
- Assert the user-visible result. For asynchronously appearing UI, await `screen.findByRole(...)`; use `waitFor` for a retrying assertion when needed. Keep user interactions and other side effects outside `waitFor` callbacks because callbacks can run repeatedly.
- Reuse existing router, context, network mocks, and cleanup conventions. The explicit `afterEach(cleanup)` makes this standalone example independent of Vitest globals; avoid duplicating it when project setup already handles cleanup and matcher imports.
- Keep DOM fixtures independent and use the project's configured environment. The example uses jsdom; a Bun React suite needs its established DOM setup and Bun-compatible test hooks and matchers.

## Validation and official references

With the example's dependencies already available, run `npx vitest run Counter.test.tsx` and the project's TypeScript check. The file directive selects jsdom for this test; it does not install jsdom. Preserve existing JSX configuration; the explicit React imports also support a classic JSX transform.

- [React Testing Library example](https://testing-library.com/docs/react-testing-library/example-intro/): render, queries, and user-visible assertions.
- [Query priority](https://testing-library.com/docs/queries/about/#priority): prefer accessible roles and names.
- [user-event setup](https://testing-library.com/docs/user-event/setup/): fresh sessions and awaited interactions.
- [React Testing Library cleanup](https://testing-library.com/docs/react-testing-library/api/#cleanup): unmount between tests.
- [jest-dom with Vitest](https://github.com/testing-library/jest-dom#with-vitest): matcher import and TypeScript setup.
- [Vitest environments](https://vitest.dev/guide/environment.html): jsdom and file-level environment selection.
- [Testing Library async methods](https://testing-library.com/docs/dom-testing-library/api-async/): `findBy` and retrying `waitFor`.
