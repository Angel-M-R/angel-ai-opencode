---
name: investigate
description: Root-cause investigation loop for bugs and unexpected behavior. Use when something fails, regresses, or behaves oddly and the cause is unknown.
---

# Investigate

A disciplined loop for finding the real cause of a failure before touching any
fix. Symptom-patching without a confirmed cause is forbidden.

## The loop

1. **Reproduce first.** No fix before a reliable reproduction. If you cannot
   reproduce it, THAT is the investigation: find the conditions that make it
   appear. Capture the exact command/input and the exact observed output.
2. **Gather evidence.** Exact error text, relevant logs, `git log` for recent
   changes to the affected area, and the smallest input that still fails.
   Shrink the reproduction as far as it will go.
3. **Hypothesize.** Write down at most 2–3 candidate causes, ranked by
   likelihood given the evidence. A hypothesis must predict something
   observable.
4. **Test the cheapest discriminating experiment.** Change ONE variable at a
   time. An experiment that cannot rule out a hypothesis is not worth running.
5. **Find the root, not the symptom.** Once a cause is confirmed, ask once:
   why did the system allow this? If the answer reveals a deeper cause (missing
   validation, wrong assumption, dead invariant), that is the real target.
6. **Fix and prove.** Apply the fix at the root, then rerun the original
   reproduction to show it now passes, and add a regression test that would
   have caught it.

## Report

- Reproduction: command/input and failing output.
- Cause: what was wrong, with the evidence that confirmed it.
- Fix: what changed and why at that level.
- Proof: the rerun result and the regression test.
