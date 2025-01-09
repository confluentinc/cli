Release Notes
-------------
<!--
If this PR introduces any user-facing changes, please document them below. Please delete any unused section titles and placeholders.
Please match the style of previous release notes: https://docs.confluent.io/confluent-cli/current/release-notes.html
-->

Breaking Changes
- PLACEHOLDER

New Features
- PLACEHOLDER

Bug Fixes
- PLACEHOLDER

Checklist
---------
<!-- 
Check each item below to ensure high-quality CLI development practices are followed. PR approval will not be granted until the checklist is fully reviewed.
For detailed instructions, please refer to this Confluence page: https://confluentinc.atlassian.net/wiki/spaces/AEGI/pages/3949592874/
-->
- [ ] I have successfully built and used a custom CLI binary, without linter issues from this PR.
- [ ] I have clearly specified in the `What` section below whether this PR applies to Confluent Cloud, Confluent Platform, or both. 
- [ ] I have verified this PR in Confluent Cloud pre-prod/production environment, if applicable.
- [ ] I have verified this PR in Confluent Platform on-premises environment, if applicable.
- [ ] I have attached manual CLI verification results or screenshots in the `Test & Review` section below.
- [ ] I have added appropriate CLI integration or unit tests for any new or updated commands and functionality.
- [ ] I confirm that this PR introduces no breaking changes or backward compatibility issues.
- [ ] I have indicated the potential customer(s) impact if something goes wrong in the `Blast Radius` section below.
- [ ] I have put checkmark below about the feature associated with this PR is enabled in:
  - [ ] Confluent Cloud prod
  - [ ] Confluent Cloud stag
  - [ ] Confluent Cloud devel
  - [ ] Confluent Platform
  - [ ] Check this box if the feature flag is enabled for certain organization only

What
----
<!--
Briefly describe **what** you have changed and **why** these changes are necessary.
Optionally include: 
- The problem being solved or the feature being added. 
- The implementation strategy or approach taken. 
- Key technical details, design decisions, or any additional context reviewers should be aware of.
-->

Blast Radius
----
<!--
The Blast Radius section should include information on what will be the customer(s) impact if something goes wrong or unexpectedly, 
adding this section will trigger the PR author to think about the impact from product perspective, examples can be:
- Confluent Cloud customers who are using confluent kafka topic any subcommand will be blocked.
- Confluent Cloud customers who are using confluent kafka topic list commands will be blocked.
- Confluent Platform customers who are using --schema flag will be impacted.
- All customers who are using SSO to login will be impacted.
-->

References
----------
<!-- Include links to relevant resources for this PR, such as: 
- Related GitHub issues 
- Tickets (JIRA, etc.) 
- Internal documentation or design specs 
- Other related PRs 
Copy and paste the links below for easy reference.
-->

Test & Review
-------------
<!-- Has this PR been tested? If so, explain **how** it was tested. Include: 
- Steps taken to verify the changes. 
- Links to manual verification documents, logs, or screenshots to save reviewers' time. 
- Any additional notes on testing (e.g., environments used, edge cases tested). 
- Screenshot showing successful resource creation, updates etc.
Example: - [Manual Verification Document](https://docs.google.com/document/d/1GwXz9hNOkub_Br-2nssoYWCf6elZBvwo7TMhCNYinwE/edit?tab=t.0#heading=h.dvbi09ntxjw6)
-->
