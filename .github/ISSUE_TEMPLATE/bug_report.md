---
name: Bug report
about: Create a report to help us improve
title: ''
labels: ''
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**Relevant Config**
Any specific information about regex used. Specific information about
HTTP query would be appreciated as well!

**Expected behavior**
A clear and concise description of what you expected to happen.


**Actual behavior**
A clear and concise description of what actually happened.

**Logs**
Any logs from your Treafik logs would be greatly appreciated!!

You can add the following to your Traefik config:
```yaml
log:
  level: DEBUG
```

Then look for logs that end with:
```console
module=github.com/packruler/rewrite-body plugin=plugin-rewritebody
```

**Server (please complete the following information):**
 - OS: [e.g. Docker, Ubuntu]
 - Traefik Version [e.g. 2.6]
 - Plugin Version [e.g. 0.5.0]

**Additional context**
Add any other context about the problem here.
