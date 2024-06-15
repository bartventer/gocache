# Versioning Strategy

This repository uses a versioning strategy that is a mix of [Semantic Versioning](https://semver.org/) and [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/).

- The main module's git tag is always incremented for any changes in the root files or submodules.
- Each submodule is individually tagged for changes within it, allowing for different version tags across submodules.
- In case of a breaking change, all tags (main module and submodules) are upgraded to the next major version.