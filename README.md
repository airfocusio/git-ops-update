# git-ops-update

## Usage

### Getting started

Take a lock at the example in the [example](example) folder.

### Provide secrets via environment variables

Every value in your configuration can be overwritten by an environment variable, that resembles the path to the value in uppercase letters and with an `_` as path separator. For example:

```yaml
# .git-ops-update.yaml
registries:
  docker:
    interval: 1h
    docker:
      url: https://registry-1.docker.io
      credentials:
        username: my-user
        password: ""
```

```bash
export REGISTRIES_DOCKER_DOCKER_CREDENTIALS_PASSWORD=my-pass
git-ops-update
```

## Installation

### Binary

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/choffmeister/git-ops-update/main/install.sh)"
cd ~/my-git-directory
~/bin/git-ops-update
```

### Docker

```bash
cd my-git-directory
docker pull choffmeister/git-ops-update:latest
docker run --rm -v $PWD:/workdir choffmeister/git-ops-update:latest
```
