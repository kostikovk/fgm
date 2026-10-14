# Sources and Adaptation

This skill implements a development procedure inspired by research; it is not an official Karpathy skill and does not promise measured improvement merely because the procedure is present.

## Karpathy's proposal

In his [13 July 2025 post](https://x.com/karpathy/status/1944435412489171119), Andrej Karpathy proposes reviewing attempts and their outcomes, extracting a textual lesson, and modifying a system prompt or lessons database. He separately discusses possible later distillation into model weights and the problem of context growth. The post's text was read through its [Thread Reader copy](https://threadreaderapp.com/thread/1944435412489171119.html), because direct X retrieval was unavailable.

Adaptation: use scoped development lessons and retrieve them before relevant work. Keep persistence separate from weight updates, and revise individual lessons rather than accumulating an unrestricted prompt. The post is a research proposal, not an evaluated guarantee for this repository.

## Research

- [Reflexion: Language Agents with Verbal Reinforcement Learning](https://arxiv.org/abs/2303.11366) — reflection on task feedback stored in episodic memory improves subsequent attempts in the studied environments, including coding benchmarks. It supports the feedback/reflection loop; it does not establish effectiveness on every real development task.
- [ExpeL: LLM Agents Are Experiential Learners](https://arxiv.org/abs/2308.10144) — extracts natural-language insights from experience and recalls them for future tasks without weight updates. It supports learning across tasks rather than retaining only a single conversation's retries.
- [Large Language Models Cannot Self-Correct Reasoning Yet](https://arxiv.org/abs/2310.01798) — finds limitations, sometimes degradation, for intrinsic self-correction without external feedback on the studied reasoning tasks and models. This motivates checking lessons against evidence; it is not proof that every modern model or reflection method fails.

The lesson statuses, project-local Markdown fallback, deduplication procedure, and evaluation scenarios are engineering choices for this portable skill. They were not prescribed or validated by Karpathy. Format checks and synthetic decision evaluations must be reported separately from demonstrated engineering outcomes or a measured reduction in mistakes.
