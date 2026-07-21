# Manual Validation Evidence

Validated against `assets/agents/angel-orchestrator.md` on 2026-07-21. These are instruction walkthroughs; no automated tests were added or modified.

Stable rule references: **Fresh-state invariant**, **Tree rule**, **Cadence rule**, **Batch rule**, **Stop rule**, **Completion rule**, and **Review-fix routing**.

## Scenario Matrix

| Scenario | Explicit input | Rule reference | Confirmed outcome |
|---|---|---|---|
| Mixed task state | Sections `Alpha` (1/2) and `Beta` (1/3), with completed tasks `1.1`, `2.1` and pending tasks `1.2`, `2.2`, `2.3` | Fresh-state invariant; Tree rule | Complete tree reports 2/5 overall, preserves both sections and every task, and uses only `✓` for completed and `☐` for pending. |
| Task cadence | Mixed state; select **After each task** | Cadence rule; Batch rule | Dispatch only `1.2`; after a clean result refresh and render the full tree, return control, and retain the cadence without asking again. |
| Section cadence | Mixed state; select **After each section** | Cadence rule; Batch rule | Dispatch pending `1.2` from `Alpha`; after explicit continuation dispatch only `2.2` and `2.3` from `Beta`, with a refresh and pause at each section boundary. |
| Run-all cadence | Mixed state; select **Run all remaining tasks** | Cadence rule; Batch rule | Dispatch `Alpha` then `Beta` as separate section-bounded batches, refreshing between them; ask no continuation question and never issue an unbounded completion prompt. |
| Unresolvable task state | Status lacks a complete tasks artifact, or resolved `tasks.md` is unreadable | Fresh-state invariant | Stop the planned-task cycle as `blocked`; do not substitute cached state. |
| Non-clean worker result | Planned-task implementer returns `blocked` or `partial` | Stop rule | Surface evidence and dispatch no further batch. |
| Failed check | Worker returns `done`, but a test/build command exits non-zero | Stop rule | Treat the command failure as authoritative, report command and exit code, and stop. |
| Plan deviation | Worker reports work outside or deviation from the requested batch | Stop rule | Surface the deviation; do not broaden or substitute work. |
| State conflict | Fresh state conflicts with the requested batch or completion report during or after dispatch | Fresh-state invariant; Stop rule | Surface the conflict; do not retry around it or continue run-all. |
| Stale pre-dispatch batch | Fresh state shows the intended task or section already complete | Fresh-state invariant; Batch rule | Skip stale dispatch and use the recomputed next bounded batch. |
| Completed planned tasks | Clean batch; fresh state shows every task `✓` and no pending task | Completion rule | Ask no cadence or continuation question; automatically dispatch `openspec-verifier`. |
| Unsuccessful verification | Verifier fails, blocks, or lacks complete executed evidence | Completion rule; Verification policy | Stop before review or archive and report the result. |
| Successful verification | Verifier reports successful executed test/build evidence | Completion rule; Review gate | Enter the existing one-time review selection gate unchanged. |
| Selected review fixes | User selects findings `#1`, `#3`, and `#5` after verification | Review-fix routing | Dispatch one bounded finding-ID batch without requiring `tasks.md` IDs, reopening cadence, or automatically repeating verification; preserve selected-reviewer rerun routing. |

Tasks 2.1 and 2.2 result: pass.
