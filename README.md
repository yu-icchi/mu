# mu

## Usage

```yaml
name: mu
on:
  pull_request:
    types: ["opened", "synchronize", "reopened", "closed"]
  issue_comment:
    types: ["created"]

jobs:
  mu:
    name: terraform
    runs-on: ubuntu-latest
    timeout-minutes: 180
    permissions:
      contents: write
      pull-requests: write
      issues: write
      statuses: write # commit status
      actions: write # artifact download and delete
    steps:
      - name: "Checkout pull_request"
        if: github.event_name == 'pull_request'
        uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
      - name: "Checkout issue_comment"
        if: github.event.issue.pull_request
        uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
        with:
          ref: refs/pull/${{ github.event.issue.number }}/merge
      - name: "mu"
        uses: yu-icchi/mu@v0
        with:
          config_path: '.github/mu.yaml'
```
