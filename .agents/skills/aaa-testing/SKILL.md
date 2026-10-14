---
name: aaa-testing
description: Write, refactor, or review automated tests using Arrange–Act–Assert (AAA) and explicit should ... when ... scenario names. Use when adding behavior-focused tests, separating setup from execution and verification, or untangling interleaved actions and assertions. Includes Go, JavaScript/TypeScript with Vitest and Bun, and React/Vue component tests with Vitest.
---

# AAA Testing

Make each test tell one behavioral story: establish the conditions, perform the behavior, then verify its observable outcome.

## Contents

1. [Navigation — choose a language example](#navigation--choose-a-language-example)
2. [Ground the test](#ground-the-test)
3. [State exactly what is tested](#state-exactly-what-is-tested)
4. [Separate the phases](#separate-the-phases)
5. [Choose the test shape](#choose-the-test-shape)
6. [Verify and report](#verify-and-report)
7. [Source](#source)

## Navigation — choose a language example

The shared grounding, scenario naming, AAA phases, test-shape, and verification rules below apply to every test. References illustrate them; they do not replace these requirements. A fragment identifies a heading, not an automatic partial read. Locate it and read through the next heading of the same or higher level with bounded reads. If section reads are unavailable, load only the relevant reference file. Follow prerequisite links; do not load every reference by default.

| Need | Read |
|---|---|
| Go single-behavior example | [A single behavior](references/go-tests.md#a-single-behavior) |
| Go table-driven cases | [Table-driven subtests](references/go-tests.md#table-driven-subtests) |
| Go errors, cleanup and test-specific decisions | [Go-specific decisions](references/go-tests.md#go-specific-decisions) |
| Vitest / TypeScript, including parameterized tests | [Vitest scenarios](references/javascript-typescript-tests.md#vitest-typescript-and-table-driven-scenarios) |
| Bun / JavaScript asynchronous example | [Bun scenario](references/javascript-typescript-tests.md#bun-javascript-and-an-asynchronous-outcome) |
| Framework-specific assertions and mock cleanup | [Framework decisions](references/javascript-typescript-tests.md#framework-decisions) |
| React component test | [React tests](references/react-tests.md#tests-countertesttsx); read its [fixture](references/react-tests.md#component-fixture-countertsx) if needed to understand the example |
| React project setup and component-test decisions | [Apply the pattern](references/react-tests.md#apply-the-pattern-in-a-project) |
| Vue component test | [Vue tests](references/vue-tests.md#tests-countertestts); read its [fixture](references/vue-tests.md#component-fixture-countervue) if needed to understand the example |
| Vue-specific component-test decisions | [Vue-specific decisions](references/vue-tests.md#vue-specific-decisions) |

For another language, apply the shared rules with the project's native runner rather than introducing these example frameworks.

## Ground the test

Inspect the target behavior, nearby tests, and the project's test commands before editing. Derive expected outcomes from the requirement or existing contract, not by reproducing the implementation's algorithm. Reuse the project's framework, fixtures, helpers, and assertion libraries; AAA alone does not justify adding dependencies or changing production interfaces.

For an existing-test refactor, identify the assertions, setup checks, mock expectations, and cleanup that currently protect behavior. Preserve them while moving code between phases. Do not silently weaken a failing assertion to make the test pass. In review-only requests, report findings without changing files.

## State exactly what is tested

Test the observable contract of the named subject: a returned value, a promised error, a state transition, or an externally visible effect under a specific condition. Do not test private implementation details or duplicate the implementation to compute the expected answer. Each scenario must express one behavior; all its assertions must substantiate that behavior.

**Every scenario title must start with lowercase `should` and follow `should <expected observable behavior> when <specific condition>`.** Use this rule for new or edited tests, including every table-driven case. For example:

- `should return the discounted total when a valid coupon is applied`
- `should reject the request when the token is expired`
- `should leave the balance unchanged when the transfer fails`

Name the subject in the enclosing `describe`, Go parent test, or the scenario itself. Avoid names such as `works`, `happy path`, or `should work when valid`: they do not identify the outcome. The asserted outcome and arranged condition must match the title; rename or split a scenario when they do not. A behavior can need several related assertions, and a condition can combine relevant preconditions.

In Vitest and Bun, put the complete title in `test`/`it`; parameterized titles must resolve to this form for every case. In Go, keep the required `Test...` function identifier and put the complete title in `t.Run` or the table row's `name`. The `should ... when ...` rule governs scenario titles, not Go discovery identifiers or suite headings.

## Separate the phases

**Arrange**

- Prepare inputs, expected outcomes, the system under test, and only the dependencies needed for the scenario.
- Establish fixture preconditions and configure test doubles here. A setup failure should stop the scenario with a useful diagnostic.
- Keep mutable state independent across cases and register cleanup when acquiring resources.

**Act**

- Execute the behavior under test and capture its returned values, error, or observable effect.
- Keep one conceptual action per scenario. A workflow may require several calls when their sequence is the behavior being tested; do not force such a scenario into an artificial single call.
- Keep the action recognizable at the call site; avoid helpers that obscure setup, execution, and verification together.
- For an expected synchronous exception, prefer a named deferred action such as `const act = () => subject(input)`, followed by the runner's exception assertion. Creating this function does not execute the subject: the matcher performs the invocation. An inline closure is also acceptable when it follows the project's convention; AAA does not require manual `try/catch`.

**Assert**

- Verify the promised return values, errors, state changes, or externally visible effects. Multiple assertions are appropriate when they describe the same outcome.
- Assert interactions only when the call, arguments, or call count are part of the contract, rather than incidental implementation details.
- Include enough actual/expected context to diagnose a failure. Stop dependent checks when a prerequisite fails; independent checks can continue.
- Avoid starting another scenario or performing additional unrelated actions here. An exception matcher may invoke the deferred action as a combined Act/Assert boundary. Split other interleaved act/assert sequences into separate tests unless they express the workflow under test.

Blank lines are usually enough to separate the phases. Add `// Arrange`, `// Act`, and `// Assert` comments when they help readers or the project requires them; comments alone do not make a test follow AAA.

## Choose the test shape

- Apply the mandatory `should ... when ...` title rule while preserving the repository's file and suite organization.
- Use separate tests for different behaviors. Use a table when cases share the same action and assertion structure; each row or subtest still has its own AAA flow.
- Extract repeated fixture work or assertions only when the helper makes the behavior easier to read. Keep the action and important expectations visible at the call site.
- Control nondeterministic inputs through existing seams. Do not introduce sleeps, live services, or shared mutable fixtures into unit tests to make AAA appear complete. Preserve intentional integration-test boundaries.
- Cover the meaningful success, boundary, and failure cases in the requested scope. AAA organizes tests; it does not by itself prove independence or sufficient coverage.

## Verify and report

Run the changed tests using the project's runner, then applicable formatting and static checks. For a refactor, run the relevant tests before and after when possible. Report the files changed, behavior covered or preserved, commands run, and any validation gap. If execution is unavailable, distinguish suggested commands from commands actually run.

For a review, identify the concrete phase or behavioral ambiguity and propose the smallest useful correction; do not flag a test solely because it lacks AAA comments.

Before finishing, verify that each new or edited scenario has the required title and that its Arrange and Assert phases demonstrate the named condition and outcome. In a review-only request, report violations rather than editing them.

## Source

The AAA framing is adapted from the supplied [Dwarves Memo article](https://memo.d.foundation/golang/testing-made-simple-best-practices-for-golang-test#1-writing-effective-test-cases-with-aaa-pattern). The scenario naming rule is the user's convention. Language references link primary documentation for API details. The examples are original, not copies of the article.
