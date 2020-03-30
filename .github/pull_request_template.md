<!--
Is there any breaking changes?  If so this is a major release, make sure '#major' is in at least one
commit message to get CI to bump the major.  This will prevent automatic down stream dependency
bumping / consuming.  For more information about semantic versioning see: https://semver.org/


Suggested PR template: Fill/delete/add sections as needed. Optionally delete any commented block.
-->
Checklist
---
1. Did you add/update any commands that accept secrets as args/flags?
   * yes: did you update `secretCommandFlags` and/or `secretCommandArgs` in [internal/pkg/analytics/analytics.go](https://github.com/confluentinc/cli/pull/325/files#diff-2d0a5a6a592890b6dff2d6f891316b82R28)
   * no: ok

What
----
<!--
Briefly describe **what** you have changed and **why**.
Optionally include implementation strategy.
-->

References
----------
<!--
Copy&paste links: to Jira ticket, other PRs, issues, Slack conversations...
For code bumps: link to PR, tag or GitHub `/compare/master...master`
-->

Test&Review
------------

<!--
Has it been tested? how?
Copy&paste any handy instructions, steps or requirements that can save time to the reviewer or any reader.
-->

<!--
Open questions / Follow ups
--------------------------
<!--
Optional: anything open to discussion for the reviewer, out of scope, or follow ups.
-->

<!--
Review stakeholders
------------------
<!--
Optional: mention stakeholders or if special context that is required to review.
-->
