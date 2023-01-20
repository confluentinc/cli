# confluent login headless-sso

Use a headless browser to automatically authenticate to Confluent Cloud with SSO.
This plugin is best suited for CI jobs which cannot rely on human interaction.

Supported providers:
* Okta

## Usage

Once the command starts running, no interaction is necessary.

```
$ go install github.com/confluentinc/cli/cmd/plugins/confluent-login-headless_sso

$ confluent plugin list
          Plugin Name          |               File Path                
-------------------------------+----------------------------------------
  confluent login headless-sso | ~/go/bin/confluent-login-headless_sso  

$ confluent login headless-sso --provider okta --email example@confluent.io --password $(cat password.txt)
Enter your Confluent Cloud credentials:
Email: example@confluent.io
Navigate to the following link in your browser to authenticate:
https://login.confluent.io/authorize

After authenticating in your browser, paste the code here:
00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
Logged in as "example@confluent.io" for organization "00000000-0000-0000-0000-000000000000" ("Confluent").
```
