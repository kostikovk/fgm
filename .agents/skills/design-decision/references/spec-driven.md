# Spec-Driven Development (SDD) & API Excellence

This guide describes specification-driven development and API contract maintenance. Identify the project's authoritative contract source and follow its existing conventions, generation workflow, and test tooling before introducing alternatives.

**Core Idea:** Agree the intended contract before implementing behavior, then keep its representation, implementation, tests, and documentation aligned. Contract-first projects author the specification; code-first projects generate it from maintained code, schemas, or annotations. Keep one authoritative source rather than independently maintaining generated output. AI behavior uses prompt specifications and evals as described in §4.

**When to Use This:**
- When designing or changing an HTTP, gRPC, or event-driven API.
- When building a feature that relies on a structured, reliable output from an AI model.
- When development between frontend and backend teams needs to happen in parallel.

---

## Table of Contents
1.  [Spec-as-Code (The "What")](#1-spec-as-code-the-what)
2.  [The SDD Process (The "How")](#2-the-sdd-process-the-how)
3.  [API Design Conventions](#3-api-design-conventions)
4.  [AI-Native Specs](#4-ai-native-specs)
5.  [Cross-References](#5-cross-references)

---

## 1. Spec-as-Code (The "What")
The specification is a machine-readable file that defines the contract.

- **OpenAPI (for HTTP APIs, including REST):** Describes operations, paths, parameters, request bodies, responses, schemas, and security requirements. For the affected operations, specify required/optional fields, status and error responses, and relevant pagination, filtering, authentication, or authorization behavior. Document important semantics that a schema alone cannot express. Keep the OpenAPI version and schema dialect compatible with the project's validators, generators, and framework; do not upgrade merely because a newer specification exists.
  *Illustrative `openapi.yaml` fragment; the complete document also needs its version, metadata, and referenced definitions:*
  ```yaml
  paths:
    /users/{userId}:
      get:
        summary: Get user by ID
        parameters:
          - name: userId
            in: path
            required: true
            schema:
              type: string
        responses:
          '200':
            description: A single user.
            content:
              application/json:
                schema:
                  $ref: '#/components/schemas/User'
  ```
- **Protocol Buffers (for gRPC):** Uses `.proto` files to define services and messages for high-performance RPC.
- **AsyncAPI (for Events):** Defines channels, messages, and publishers/subscribers for event-driven architectures (Kafka, RabbitMQ).

## 2. The SDD Process (The "How")
1.  **Find the contract and its owner:** Inspect existing descriptions, routes, schemas, consumers, tests, and generation/validation configuration. Discover actual paths rather than assuming `openapi.yaml` exists. Establish whether the maintained source is the specification or code that generates it. Read a relevant ADR if one exists; record a new ADR only for a consequential design trade-off, not every endpoint.
2.  **Define and update the affected contract:** Agree operations, inputs, outputs, errors, and relevant security requirements. In contract-first projects, update the maintained description before implementing the behavior. In code-first projects, update the authoritative code/schema/annotations through the applicable test-first workflow and regenerate the description. Preserve unrelated operations and supported consumer behavior; do not edit generated descriptions independently of their source. For a new HTTP API without established conventions, OpenAPI and contract-first design are useful candidates.
3.  **Validate description and implementation:** Run the existing description validator and generation/drift checks. Test actual requests, responses, serialization, status/error mapping, and security behavior relevant to the change. Distinguish schema conformance from consumer contracts. Prism provides OpenAPI-based mocks and a validation proxy that checks exercised requests/responses against the current spec. Pact generates consumer-driven interaction contracts from tests of actual consumer code and verifies them against the provider implementation. They are complementary, not interchangeable; neither proves all behavior correct, and neither is a required new dependency.
4.  **Mocking & Codegen when useful:** Generate mocks or clients when they support a real need, such as parallel frontend development or maintained typed clients. Reuse the project's tools and update affected generated artifacts through their generator. Keep business tests independent of incidental generated-code structure.
5.  **Check compatibility separately:** Conformance to the current schema does not establish compatibility with previous versions. Diff against the previously supported specs and verify actual consumer/provider code for supported version combinations, including behavior tests for semantic changes. A newly required input, removed/renamed field, changed status/error envelope, or new enum value can affect existing consumers. Expand/contract migrations and tolerant readers help staged evolution but do not guarantee semantic compatibility. Report checked versions and any verification gap; do not claim current-schema validation proves old-client compatibility.

> **Sources:** [Pact: How Pact works](https://docs.pact.io/getting_started/how_pact_works) — consumer tests and provider verification; [Prism: Validation Proxy](https://github.com/stoplightio/prism/blob/master/docs/guides/03-validation-proxy.md) — observed traffic validation against OpenAPI.

> **OpenAPI sources:** [OpenAPI Specification](https://spec.openapis.org/oas/v3.2.1.html) defines the HTTP description format; the cited version is a reference, not an upgrade requirement. [OpenAPI Initiative best practices](https://github.com/OAI/learn.openapis.org/blob/main/best-practices.md) recommends design-first while describing both design-first and code-first workflows with one source of truth. [FastAPI's generated OpenAPI](https://fastapi.tiangolo.com/how-to/extending-openapi/) is an example of a framework-owned code-first workflow, not a recommendation to adopt FastAPI.

## 3. API Design Conventions
If the project has an API style guide, follow it. If not, Google's API Improvement Proposals (aip.dev) are a well-documented default:
- **Resource-Oriented Design:** Name resources, not actions (`GET /users/{id}`, not `GET /getUser`).
- **Standard Methods:** Consistent semantics for `GET`, `POST`, `PUT`, `PATCH`, `DELETE`; `PUT` and `DELETE` idempotent.
- **Pagination & Filtering:** Standardized query parameters for collections; cursor-based pagination for large or fast-changing sets.
- **Errors:** A single error envelope with a machine-readable code, translated from domain errors at the transport layer (see `engineering-principles.md` §7).

> **Source:** [Google AIP-180: Backwards compatibility](https://google.aip.dev/180) — distinguishes source, wire, and semantic compatibility; apply alongside the project's compatibility policy.

## 4. AI-Native Specs
The same principles apply to AI, but the "spec" takes a different form.

- **Prompt Specifications:** Version-control the system prompt as a specification of intended persona, capabilities, and constraints, not an enforcement boundary. Pair it with output schemas, evals, and server-side authorization (see `ai-engineering.md` §§1–3).
- **Evals-Driven Development:** Write the evaluation tests (evals) that define correct behavior *before* iterating on the prompt. The model's output must pass them.
- **Output Schema Enforcement:** Constrain model outputs to a schema (JSON Schema, Pydantic, Zod) and validate before use. Schema conformance establishes shape and declared constraints, not factual, semantic, or business correctness, nor reliability equivalent to a traditional API. Apply domain validation and task-specific evals; define refusal, invalid-output, and fail-closed paths (see `ai-engineering.md` §§1–2).

> **Source:** [JSON Schema: What is a schema?](https://json-schema.org/understanding-json-schema/about) — data organization and validation, not a guarantee of correct model behavior.

## 5. Cross-References
- `../SKILL.md` — the decision protocol and the canon.
- `ai-engineering.md` — AI-specific patterns (RAG, agents, evals).
- `engineering-principles.md` §7 — error translation across boundaries; §9 — contract programming.
- `../assets/adr-template.md` — where the design decision behind the spec gets recorded.
