Checklist
---------
1. [CRUCIAL] Is the change for CP or CCloud functionalities that are already live in prod?
   * yes: ok
   * no: DO NOT MERGE until the required functionalites are live in prod  
   
2. Did you add/update any commands that accept secrets as args/flags?
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
Copy & paste links to Jira tickets, other PRs, issues, Slack conversations, etc.
For code bumps: link to PR, tag or GitHub `/compare/master...master`
-->

Test & Review
-------------
<!--
Has it been tested? how?
Copy & paste any handy instructions, steps or requirements that can save time to the reviewer or any reader.
-->

<!--
Open questions / Follow ups
---------------------------
Optional: anything open to discussion for the reviewer, out of scope, or follow ups.
-->

<!--
Review stakeholders
-------------------
Optional: mention stakeholders or special context that is required to review.
-->
