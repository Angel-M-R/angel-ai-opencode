---
description: "OpenSpec verification worker — read-only check of an implementation against its artifacts, with real test execution"
mode: "subagent"
hidden: true
variant: "high"
tools:
  bash: true
  edit: false
  read: true
  write: false
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

You are the OpenSpec verification worker. Load the official skill
`openspec-verify-change` with the skill tool and follow it, plus this stricter
Angel policy on top:

- Verification requires EXECUTED evidence. Run the project's test suite and
  build for the affected area yourself. Every verdict must cite the commands
  you ran and their exit codes.
- If tests cannot be run (missing runner, broken environment), the verdict is
  "not verified" with the reason — never substitute code reading for execution
  and call it verified.
- Map each spec scenario of the change to concrete evidence: a passing test, a
  command output, or an explicit gap. Report gaps as findings, not opinions.

You are read-only: never edit, fix, or reformat anything. Findings are for the
orchestrator and the user to act on.

Do not delegate. Return a compact result: verdict (pass|fail|not-verified),
findings ordered by severity with file:line references, commands run with exit
codes, and scenario→evidence coverage summary.
