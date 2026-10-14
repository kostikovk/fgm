# API Contract Changes

Apply the shared discovery and testing gates in [Code Change](../SKILL.md) before using this guidance.

## Keep API contracts synchronized

When adding or changing an HTTP API, locate the affected routes, request/response schemas, consumers, and existing OpenAPI description or generation workflow. Identify the authoritative source: a maintained specification in a contract-first project, or code/schema/annotations that generate it in a code-first project. Update that source and regenerate derived artifacts; preserve the supported specification version and established tooling.

Agree the intended API contract before implementing changed behavior. Use [Spec-as-Code](../../design-decision/references/spec-driven.md#1-spec-as-code-the-what) for the contract format and [The SDD Process](../../design-decision/references/spec-driven.md#2-the-sdd-process-the-how) for source ownership, validation, and consumer compatibility. Read [API Design Conventions](../../design-decision/references/spec-driven.md#3-api-design-conventions) when designing operations, payloads, errors, or security requirements. For a new HTTP API without an established contract workflow, consider OpenAPI and contract-first design, especially when frontend or independent clients consume it. Follow the existing protocol's contract format for gRPC, events, or other API styles.

Verify the description, actual implementation behavior, and compatibility with supported consumers separately. A valid OpenAPI document or implementation conforming to a newly changed schema does not establish that an older client still works. Keep affected contracts, implementation, tests, and maintained clients aligned in the same scoped change. Introduce mocks, generated clients, or additional tools only when a concrete need justifies them.
