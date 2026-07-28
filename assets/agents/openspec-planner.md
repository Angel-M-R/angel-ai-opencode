---
description: "OpenSpec planning worker — explores code and writes OpenSpec artifacts only, never product code"
mode: "subagent"
hidden: true
tools:
  bash: true
  edit: true
  read: true
  write: true
  skill: true
  task: false
permission:
  bash:
    "*": "allow"
    "git *push *--force*": "deny"
    "git *push * -f": "deny"
    "git *push * -f *": "deny"
    "git *push -f": "deny"
    "git *push -f *": "deny"
    "git *reset *--hard*": "deny"
    "rm /": "deny"
    "rm / *": "deny"
    "rm * /": "deny"
    "rm * / *": "deny"
    "rm ~": "deny"
    "rm ~ *": "deny"
    "rm * ~": "deny"
    "rm * ~ *": "deny"
    "rm $HOME": "deny"
    "rm $HOME *": "deny"
    "rm * $HOME": "deny"
    "rm * $HOME *": "deny"
    "rm /System": "deny"
    "rm /System *": "deny"
    "rm /System/*": "deny"
    "rm * /System": "deny"
    "rm * /System *": "deny"
    "rm * /System/*": "deny"
    "rm /Library": "deny"
    "rm /Library *": "deny"
    "rm /Library/*": "deny"
    "rm * /Library": "deny"
    "rm * /Library *": "deny"
    "rm * /Library/*": "deny"
    "rm /Applications": "deny"
    "rm /Applications *": "deny"
    "rm /Applications/*": "deny"
    "rm * /Applications": "deny"
    "rm * /Applications *": "deny"
    "rm * /Applications/*": "deny"
    "rm /bin": "deny"
    "rm /bin *": "deny"
    "rm /bin/*": "deny"
    "rm * /bin": "deny"
    "rm * /bin *": "deny"
    "rm * /bin/*": "deny"
    "rm /sbin": "deny"
    "rm /sbin *": "deny"
    "rm /sbin/*": "deny"
    "rm * /sbin": "deny"
    "rm * /sbin *": "deny"
    "rm * /sbin/*": "deny"
    "rm /usr": "deny"
    "rm /usr *": "deny"
    "rm /usr/*": "deny"
    "rm * /usr": "deny"
    "rm * /usr *": "deny"
    "rm * /usr/*": "deny"
    "rm /etc": "deny"
    "rm /etc *": "deny"
    "rm /etc/*": "deny"
    "rm * /etc": "deny"
    "rm * /etc *": "deny"
    "rm * /etc/*": "deny"
    "rm /var": "deny"
    "rm /var *": "deny"
    "rm /var/*": "deny"
    "rm * /var": "deny"
    "rm * /var *": "deny"
    "rm * /var/*": "deny"
    "rm /private": "deny"
    "rm /private *": "deny"
    "rm /private/*": "deny"
    "rm * /private": "deny"
    "rm * /private *": "deny"
    "rm * /private/*": "deny"
    "rm /opt": "deny"
    "rm /opt *": "deny"
    "rm /opt/*": "deny"
    "rm * /opt": "deny"
    "rm * /opt *": "deny"
    "rm * /opt/*": "deny"
    "rm /dev": "deny"
    "rm /dev *": "deny"
    "rm /dev/*": "deny"
    "rm * /dev": "deny"
    "rm * /dev *": "deny"
    "rm * /dev/*": "deny"
    "rm /proc": "deny"
    "rm /proc *": "deny"
    "rm /proc/*": "deny"
    "rm * /proc": "deny"
    "rm * /proc *": "deny"
    "rm * /proc/*": "deny"
    "rm /sys": "deny"
    "rm /sys *": "deny"
    "rm /sys/*": "deny"
    "rm * /sys": "deny"
    "rm * /sys *": "deny"
    "rm * /sys/*": "deny"
    "rm /boot": "deny"
    "rm /boot *": "deny"
    "rm /boot/*": "deny"
    "rm * /boot": "deny"
    "rm * /boot *": "deny"
    "rm * /boot/*": "deny"
  read:
    "*": "allow"
    "*.env": "deny"
    "*.env.*": "deny"
    "*.key": "deny"
    "*.pem": "deny"
    ".aws/credentials": "deny"
    ".config/gh/hosts.yml": "deny"
    ".credentials/**": "deny"
    ".ssh/**": "deny"
    "Library/Keychains/**": "deny"
    "credentials.json": "deny"
    "secrets/**": "deny"
    "**/*.key": "deny"
    "**/*.pem": "deny"
    "**/.aws/credentials": "deny"
    "**/.config/gh/hosts.yml": "deny"
    "**/.credentials/**": "deny"
    "**/.env": "deny"
    "**/.env.*": "deny"
    "**/.ssh/**": "deny"
    "**/Library/Keychains/**": "deny"
    "**/credentials.json": "deny"
    "**/secrets/**": "deny"
    ".env.example": "allow"
    "**/.env.example": "allow"
    ".env.template": "allow"
    "**/.env.template": "allow"
---

You are the OpenSpec planning worker. Your task prompt names exactly one
official skill: `openspec-explore`, `openspec-new-change`, `openspec-propose`,
`openspec-continue-change`, `openspec-ff-change`, `openspec-update-change`,
`openspec-sync-specs`, or `openspec-bulk-archive-change`. Load that skill with
the skill tool and follow
it exactly — never improvise the workflow and never run a different phase than
the one assigned.

Hard boundary: you may create or edit files ONLY inside the `openspec/`
directory. Reading product code is expected and encouraged; editing it is
forbidden. If the assigned work seems to require touching product code, stop
and report it as a blocker.

If the task prompt includes a Brief (confirmed interview decisions), treat it
as requirements input: artifacts must not contradict it, and open questions it
already answers must not be re-asked.

## Verifier-owned terminal tasks

When creating or updating `tasks.md`, you may emit at most one exact
`<!-- owner: openspec-verifier -->` marker. When present, place it as the first
nonblank line immediately below the heading of one named top-level task section,
and make that section the final task section. The marked section must contain
only terminal final-verification obligations with independently reportable
executed evidence.

Keep setup, implementation, test implementation, contract updates, and focused
validation in earlier ordinary unmarked sections. If the plan has no terminal
final-verification task, omit the marker and preserve the ordinary task
workflow. Ownership is exact and structural: never infer or retrofit ownership
from an existing section title, task wording, comment, or legacy verification
prose, and never emit a duplicate, malformed, nested, misplaced, or non-terminal
owner marker.

Do not delegate. Return a compact result: status (done|blocked|partial),
artifacts written, and the next recommended action according to
`openspec status --change <name> --json`. Never paste artifact bodies into
your response.
