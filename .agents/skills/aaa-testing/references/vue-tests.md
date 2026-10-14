# AAA for Vue components

Test the component's public contract: rendered output, user interaction, and emitted events with meaningful payloads. Avoid reading private component state or invoking methods through `wrapper.vm` when the behavior can be observed through the DOM or a public event.

This example uses Vue 3, Vue Test Utils v2, and Vitest with a DOM environment. Reuse the project's runner, mounting helpers, plugins, and mocks; the example does not authorize dependency installation or tooling migration. Vue 2 uses different tooling and component APIs, so adapt to the installed version rather than copying this fixture unchanged.

## Contents

1. [Component fixture: `Counter.vue`](#component-fixture-countervue)
2. [Tests: `Counter.test.ts`](#tests-countertestts)
3. [Vue-specific decisions](#vue-specific-decisions)
4. [Validation and official references](#validation-and-official-references)

## Component fixture: `Counter.vue`

The component displays a count and emits its new value after an enabled increment. A disabled button leaves the count unchanged and emits no change event. In a real project, import the existing component instead of recreating it in a test.

```vue
<script setup lang="ts">
import { ref } from "vue";

const props = withDefaults(
  defineProps<{ initialCount?: number; disabled?: boolean }>(),
  { initialCount: 0, disabled: false },
);
const emit = defineEmits<{ (event: "change", count: number): void }>();
const count = ref(props.initialCount);

function increment() {
  if (props.disabled) return;
  count.value += 1;
  emit("change", count.value);
}
</script>

<template>
  <section aria-label="Counter">
    <p role="status">Count: {{ count }}</p>
    <button :disabled="props.disabled" @click="increment">Increment</button>
  </section>
</template>
```

## Tests: `Counter.test.ts`

These tests check the visible result and event contract of the same action. There is one button in this fixture, making `get("button")` unambiguous; use the project's accessible-query tooling when available for more complex interfaces.

```ts
// @vitest-environment jsdom
import { afterEach, describe, expect, test } from "vitest";
import { enableAutoUnmount, mount } from "@vue/test-utils";
import Counter from "./Counter.vue";

enableAutoUnmount(afterEach);

describe("Counter", () => {
  test("should display and emit the incremented count when the enabled button is clicked", async () => {
    // Arrange
    const wrapper = mount(Counter, { props: { initialCount: 2 } });
    const button = wrapper.get("button");

    // Act
    await button.trigger("click");

    // Assert
    expect(wrapper.get('[role="status"]').text()).toBe("Count: 3");
    expect(wrapper.emitted("change")).toEqual([[3]]);
  });

  test("should keep the count unchanged and emit no change when the button is disabled", async () => {
    // Arrange
    const wrapper = mount(Counter, { props: { initialCount: 2, disabled: true } });
    const button = wrapper.get("button");

    // Act
    await button.trigger("click");

    // Assert
    expect(button.attributes()).toHaveProperty("disabled");
    expect(wrapper.get('[role="status"]').text()).toBe("Count: 2");
    expect(wrapper.emitted("change")).toBeUndefined();
  });
});
```

For JavaScript, use a `.test.js` file; the test bodies above need no changes. A JavaScript component uses `<script setup>` and runtime props/emits declarations instead of the TypeScript declarations above.

## Vue-specific decisions

- Arrange props, slots, providers, and controlled dependency responses before exercising the scenario. Reuse router/store mounting helpers when the component actually depends on them. Mount is setup for a later interaction; place mount in Act when initial rendering itself is the behavior being tested.
- Await `trigger`, `setValue`, and `setProps` when their reactive changes are needed for an assertion. These wait for Vue's next update tick. If state changes outside those helpers, use `await nextTick()` when necessary rather than a sleep.
- `nextTick()` waits for Vue's reactive rendering; it does not resolve unrelated request promises. When a scenario depends on mocked asynchronous work, await its known completion or `flushPromises()` from Vue Test Utils, then the relevant rendering update if needed. Put these waits with the action they complete. `flushPromises()` is not a substitute for pending timers or unresolved network requests.
- Check both the event payload and the promised occurrence count. `wrapper.emitted("change")` contains one array of arguments per emission; `[[3]]` asserts one emission with the value `3`. Avoid checking only whether an event exists.
- Unmount between tests and restore mocks through the existing setup. The explicit `enableAutoUnmount(afterEach)` makes this example self-contained; reuse project cleanup when already configured.

## Validation and official references

Reuse the project's Vue/Vitest config. In an isolated example where the dependencies already exist, this `vitest.config.ts` enables single-file component compilation:

```ts
import { defineConfig } from "vitest/config";
import vue from "@vitejs/plugin-vue";

export default defineConfig({ plugins: [vue()] });
```

Run `npx vitest run Counter.test.ts` and the project's Vue-aware typecheck, commonly `vue-tsc --noEmit`. The test directive selects jsdom but does not install it. A passing Vitest run does not establish `.vue` type correctness.

- [Vue component testing](https://vuejs.org/guide/scaling-up/testing.html#component-testing): public interfaces and Vue testing tools.
- [Vue TypeScript props and emits](https://vuejs.org/guide/typescript/composition-api.html): typed component contracts.
- [Vue Test Utils event handling](https://test-utils.vuejs.org/guide/essentials/event-handling.html): event payloads and occurrence arrays.
- [Vue Test Utils async behavior](https://test-utils.vuejs.org/guide/advanced/async-suspense.html): awaited interactions, `nextTick`, and `flushPromises`.
- [Vue Test Utils API](https://test-utils.vuejs.org/api/): mounting and automatic unmounting.
- [Vue Test Utils installation](https://test-utils.vuejs.org/guide/): Vue 3/v2 tooling and Vitest setup.
