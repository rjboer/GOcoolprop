# Validation Correctness Repair Plan

> For agentic workers: REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Remove false-pass paths and make validation acceptance depend on executed valid comparisons, statistical audit evidence, reference-defined domains, and complete deterministic suite accounting.

**Architecture:** Add counters and acceptance at the fluid level, keep consistent invalid/error outcomes separate from numeric passes, and make the executor produce actual audit observations. Use bounded reference batches with transaction-safe workers and propagate every storage error. Add cell models for recursive validity/deviation refinement.

## Tasks

1. Add counters, statistical audit summaries, confidence bounds, and fluid acceptance tests.
2. Replace candidate-derived pressure truncation; add reference fluid metadata/domain handshake.
3. Select property/stage/pair tolerances and expose candidate phase.
4. Add bounded batched reference workers, per-batch deadlines, panic isolation, and progress events.
5. Execute and reconcile every planned suite; propagate writes/moves/index errors and sort reports.
6. Add adaptive cell subdivision for invalid, mixed, failed, and unresolved cells.
7. Run focused red/green tests, full Go verification, and a three-fluid smoke that must not pass with zero valid numerical comparisons.

## Acceptance

A fluid cannot pass with zero valid numeric comparisons, consistent errors do not increment valid passes, every mandatory suite reconciles planned/completed counts, reference pressure coverage reaches reference Pmax, and statistical acceptance uses the configured family-wise upper failure bound.
