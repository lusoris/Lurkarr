# ADR-004: Queue Cleaner Strike System

**Status:** Accepted  
**Date:** 2026-02-01

## Context

Download queue management in the *Arr ecosystem requires balancing thoroughness (give downloads time to recover) against cleanup speed (remove bad downloads quickly). Existing tools like Cleanuparr use immediate removal, which can be too aggressive for transient issues (temporary tracker downtime, ISP throttling).

## Decision

Implement a strike-based system with configurable thresholds per condition type:

- **Stalled** — No progress for N minutes (varies by privacy: public vs. private trackers).
- **Slow** — Speed below threshold for N minutes (SABnzbd-aware, uses `timeleft` calculation).
- **Failed import** — Pattern matching on import failure reasons.
- **Metadata stuck** — Item stuck in metadata phase beyond threshold.

Strikes accumulate over a configurable time window. A download is only removed when its strike count reaches the threshold. Strikes decay naturally when the window expires.

The cleaner runs in phases: blocklist check → dedup → stalled/slow/failed → seeding rules → orphan detection.

## Consequences

**Positive:**
- Transient issues don't trigger immediate removal — downloads recover naturally.
- Per-condition thresholds allow fine-tuning (e.g. more lenient for private trackers).
- Strike history is persisted in the database — survives restarts and is auditable.
- Phase-based execution ensures blocklist/dedup runs before expensive checks.

**Negative:**
- More complex than simple "remove after N minutes stalled" approaches.
- Strike accumulation requires periodic cleaner runs to be effective (depends on scheduler).
- Private tracker detection relies on indexer flags — may not be accurate for all indexers.
