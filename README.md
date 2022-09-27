# kubectl-nplist

kubectl-nplist does things.

## Installation

```shell
# homebrew
brew install stenic/tap/kubectl-nplist

# gofish
gofish rig add https://github.com/stenic/fish-food
gofish install github.com/stenic/fish-food/kubectl-nplist

# scoop
scoop bucket add kubectl-nplist https://github.com/stenic/scoop-bucket.git
scoop install kubectl-nplist

# go
go install github.com/stenic/kubectl-nplist@latest

# docker 
docker pull ghcr.io/stenic/kubectl-nplist:latest

# dockerfile
COPY --from=ghcr.io/stenic/kubectl-nplist:latest /kubectl-nplist /usr/local/bin/
```

> For even more options, check the [releases page](https://github.com/stenic/kubectl-nplist/releases).


## Run

```shell
# Installed
kubectl-nplist -h

# Docker
docker run -ti ghcr.io/stenic/kubectl-nplist:latest -h

# Kubernetes
kubectl run kubectl-nplist --image=ghcr.io/stenic/kubectl-nplist:latest --restart=Never -ti --rm -- -h
```

## Documentation

```shell
kubectl-nplist -h
```

## Badges

[![Release](https://img.shields.io/github/release/stenic/kubectl-nplist.svg?style=for-the-badge)](https://github.com/stenic/kubectl-nplist/releases/latest)
[![Software License](https://img.shields.io/github/license/stenic/kubectl-nplist?style=for-the-badge)](./LICENSE)
[![Build status](https://img.shields.io/github/workflow/status/stenic/kubectl-nplist/Release?style=for-the-badge)](https://github.com/stenic/kubectl-nplist/actions?workflow=build)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)

## License

[License](./LICENSE)
