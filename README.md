# git-ops-update

## Usage

### Define files

Files define where to look for potential updates. Excludes overrule includes.

```yaml
# .git-ops-update.yaml
files:
  includes:
    - '\.yaml$'
  excludes:
    - '\.generated\.yaml$'
    - '\/system\/.*\.yaml$'
```

### Define registries

Registries define sources where you can lookup version numbers for individual resources.

#### Docker

```yaml
# .git-ops-update.yaml
registries:
  my-docker-registry:
    interval: 1h
    docker:
      url: https://registry-1.docker.io
      credentials:
        username: user
        password: pass
```

#### Helm

```yaml
# .git-ops-update.yaml
registries:
  my-helm-registry:
    interval: 1h
    helm:
      url: https://helm.nginx.com/stable
      credentials:
        username: user
        password: pass
```

### Define policies

Policies define how you would select and compare different potential new versions of your resources.

```yaml
# .git-ops-update.yaml
policies:
  my-semver-policy:
    extracts:
      - type: semver
  my-ubuntu-specific-policy:
    pattern: '^(?P<year>\d+)\.(?P<month>\d+)$'
    extracts:
      - type: numeric
        value: '<year>'
      - type: numeric
        value: '<month>'
```

### Annotate your files

In order for this tool to know where to update version numbers you have to annotate the relevant places

```yaml
# deployment.yaml
apiVersion: v1
kind: Pod
metadata:
  name: ubuntu
spec:
    containers:
      - name: ubuntu
        image: ubuntu:18.04 # git-ops-update {"registry":"my-docker-registry","resource":"library/ubuntu","policy":"my-ubuntu-policy","format":"docker-image","action":"push"}
```

### Provide configuration via environment variables

Every value in your configuration can be overwritten by an environment variable, that resembles the path to the value in uppercase letters and with an `_` instead of `.` or `-`. For example:

```yaml
# .git-ops-update.yaml
registries:
  my-docker-policy:
    type: docker
    interval: 1h
    url: https://registry-1.docker.io
    credentials:
      username: my-user
      password: ""
```

```bash
export GIT_OPS_UPDATE_REGISTRIES_MY_DOCKER_POLICY_CREDENTIALS_PASSWORD=my-pass
git-ops-update
```

### GitHub action

```yaml
# .github/workflows/update.yml
name: update
on:
  schedule:
    - cron: '0 2 * * *'
jobs:
  update:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 0
    - uses: docker://ghcr.io/choffmeister/git-ops-update
```

## Installation

### Docker

```bash
cd my-git-directory
docker pull ghcr.io/choffmeister/git-ops-update:latest
docker run --rm -v $PWD:/workdir ghcr.io/choffmeister/git-ops-update:latest
```
