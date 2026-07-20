---
name: technical-grilling
description: Relentlessly grill the user about the technical decisions of a plan until shared understanding is confirmed. Use before planning a non-trivial change, after product questions if any.
---

# Technical Grilling

Stress-test the technical shape of a change by walking its decision tree with
the user. The plan is guilty until proven understood: every major decision gets
questioned once, every answer can open new branches, and nothing is acted on
until the user confirms shared understanding.

## Rules

- Ask ONE question at a time with the `question` tool, then STOP and wait.
  Never batch, never answer your own question and move on.
- Facts vs decisions: how the code currently works, what dependencies exist,
  what patterns the repo uses — those are facts; investigate them with tools
  and never ask the user. Tradeoffs, priorities, and interfaces are decisions;
  those always go to the user.
- For every question, present the realistic options with your recommended one
  first, marked as recommended, with a one-line reason. Recommending is
  mandatory; deciding is forbidden.
- Follow the tree: an answer that creates a new dependency, risk, or option
  spawns a follow-up branch. Track open branches; do not declare done while
  any branch is unexplored.
- Match the user's conversation language.

## The decision tree

For each major technical decision in the plan, probe:

1. **Approach** — is there a meaningfully simpler alternative? Why not that one?
2. **Dependencies** — what does this decision now depend on, and what starts
   depending on it?
3. **Failure modes** — what breaks when this fails half-way? What does recovery
   look like?
4. **Data & interfaces** — what shape crosses the boundary, and who owns it?
5. **Testing** — how will we know it works? What is hard to test in this design?
6. **Migration & rollout** — does anything existing need to move? Can this ship
   incrementally?
7. **Reversibility** — if this turns out wrong in a month, how expensive is
   undoing it?

## Ending condition

The grilling ends only when BOTH hold: no unexplored branches remain, and the
user explicitly confirms shared understanding of the resulting plan.

## Output

End with a `## Technical decisions` bullet list: one line per confirmed
decision with its chosen option. Combined with `## Product decisions` (if a
product interview ran), this is the Brief the orchestrator passes to the
planner. Do not act on the plan until the user confirms the Brief.
