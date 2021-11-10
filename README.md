# git-ops-update

Replacement for `git describe --tags` that produces [semver](https://semver.org/) compatible versions that follow to semver sorting rules.

## Usage

TODO

### Binary

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/choffmeister/git-ops-update/master/install.sh)"
cd ~/my-git-directory
~/bin/git-ops-update
```

### Docker

```bash
cd my-git-directory
docker pull choffmeister/git-ops-update:latest
docker run --rm -v $PWD:/workdir choffmeister/git-ops-update:latest
```
