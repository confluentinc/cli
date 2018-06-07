# This exists so the user can create a short URL for private repos
# API requires Accept header to download a file. We use the API to pass netrc auth via `curl -n`.
curl -ns https://api.github.com/repos/confluentinc/cli/contents/install.sh?ref=master -H Accept:application/vnd.github.raw | bash -s -- $@
