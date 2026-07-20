---
name: product-grilling
description: Interview the user about the product side of a change before planning — problem, users, business rules, scope, non-goals. Use before creating an OpenSpec change when the user chose product questions.
---

# Product Grilling

Interrogate the product assumptions behind a change BEFORE any planning
artifact exists. The goal is a short list of confirmed product decisions, not
a document.

## Rules

- Ask ONE question at a time with the `question` tool, then STOP and wait.
  Never batch questions, never continue past an unanswered question.
- Offer concrete answer options and put your recommended one first, marked as
  recommended. The user decides; you never assume.
- Facts are yours, decisions are theirs: anything discoverable from the
  codebase, docs, or existing OpenSpec artifacts you investigate with tools —
  never ask the user something you can look up.
- 3–5 questions per round. After a round, summarize the assumptions collected
  so far and ask: correct something / another round / continue to planning.
- Match the user's conversation language.

## Question territory

Pick the questions that expose the biggest unknowns for THIS change; skip
territory that is already obvious or answered:

1. **Problem & trigger** — what hurts today, and what triggered doing this now?
2. **Target user & situation** — who exactly is this for, in what moment?
3. **Success outcome** — what observable change means it worked?
4. **Business rules & constraints** — invariants the solution must respect.
5. **Edge cases** — the weird inputs/situations that usually get discovered late.
6. **Impact of NOT doing it** — what happens if this ships never or later?
7. **First-slice scope** — the smallest version worth shipping.
8. **Non-goals** — what this change deliberately does NOT cover.
9. **Tradeoffs accepted** — what the user is willing to sacrifice (speed,
   polish, generality) and what they are not.

## Output

End with a `## Product decisions` bullet list: one line per confirmed decision,
including scope boundary and non-goals. Ask the user to confirm it. This brief
feeds the technical interview (if any) and then the planner — do not proceed
until confirmed.
