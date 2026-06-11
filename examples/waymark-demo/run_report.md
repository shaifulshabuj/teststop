# teststop Report — waymark

**Date:** 2026-06-11 21:54:25
**Duration:** 203109ms
**System:** api (TypeScript)
**Mode:** predicted (no `--target` — structural only, not executed)
**Predicted confidence:** 71.4% _(run with `--target` to verify)_
**Exit Code:** 1 (review needed)

## Predicted Risks (15 total)

| Priority | Title | Area | Edge Case |
|----------|-------|------|-----------|
| critical | Double-clicking Approve on a slow connection approves twice | api/actions/:action_id/approve | yes |
| critical | I approve from the dashboard while my teammate rejects from Slack | api/actions/:action_id/reject | yes |
| critical | Corporate email scanner clicks my approve link before I do | api/actions/approve-via-token | yes |
| high | Slack re-sends my Approve tap because the server answered too slowly | api/slack/interact | yes |
| high | Pasting policy rules from a Google Doc breaks on smart quotes | api/config/policies | yes |
| high | Editing config in two tabs silently reverts my earlier change | api/config | no |
| medium | Hand-typing a page URL with page 0 and an enormous page size | api/actions/paginated | yes |
| critical | Picking files to roll back from a list that went stale overnight | api/sessions/:session_id/rollback-partial | yes |
| high | Paused in one window, resumed in another, now both show different states | api/sessions/:session_id/resume | yes |
| medium | Adding a teammate twice with a pasted email that has a trailing newline | api/team/members | no |
| high | Two admins edit the same approval route and one edit vanishes | api/approval-routes/:route_id | yes |
| medium | Checking on a session from a phone after my laptop fell asleep | api/sessions/:session_id/status | no |
| medium | Creating an approval route named with emoji and formatting from Slack | api/approval-routes | yes |
| medium | Stopping a project from an old bookmarked page while also using the new hub | projects/:id/stop | yes |
| medium | Maintenance archive kicks in while I'm paging through old actions | api/maintenance/archive | yes |

## Execution

- **Target:** _none — predicted only, not executed_
- **Results:** 15 scenarios predicted. Run with `--target <url>` to execute and verify.

## Predicted Failure Modes (0)

_(none)_

## Memory State

- **Stable areas:** projects/:id/stop
- **Volatile areas:** api/sessions/:session_id, api/actions/:action_id, sessions/:id, sessions/:id/pause, api/hub/projects/:id/resume, api/sessions/:session_id/status, api/sessions/:session_id/actions, sessions/:id/resume, api/sessions/:session_id/rollback
- **Retired areas:** projects/:id/stop

