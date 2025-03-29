# Version History

This document provides a detailed history of releases for the gh-stars project, based on the versioning system used in the Makefile.

## Overview

The gh-stars project uses semantic versioning (MAJOR.MINOR.PATCH) for its releases. The Makefile includes commands for creating major, minor, and patch releases, which automatically update the version number in the VERSION file.

## Release Types

- **Major Release**: Increments the first number (X.0.0). Used for significant changes or breaking updates.
- **Minor Release**: Increments the second number (0.X.0). Used for new features that don't break existing functionality.
- **Patch Release**: Increments the third number (0.0.X). Used for bug fixes and small improvements.

## Release Process

The release process is automated using the Makefile. To create a new release:

1. Choose the appropriate release type (major, minor, or patch).
2. Run the corresponding make command:
   - For a major release: `make release-major`
   - For a minor release: `make release-minor`
   - For a patch release: `make release-patch`
3. The Makefile will:
   - Update the VERSION file
   - Commit the change with a signed commit (-S flag)
   - Create a new git tag
   - Push changes to the master branch
   - Run goreleaser to create the release

## Release Commands

```makefile
# Do a major release
.PHONY: release-major
release-major:
	@echo $(shell $(call bump_major)) > VERSION
	@git add VERSION
	@git commit -S -m "Release $(shell $(call bump_major))"
	$(MAKE) release

# Do a minor release
.PHONY: release-minor
release-minor:
	@echo $(shell $(call bump_minor)) > VERSION
	@git add VERSION
	@git commit -S -m "Release $(shell $(call bump_minor))"
	$(MAKE) release

# Do a patch release
.PHONY: release-patch
release-patch:
	@echo $(shell $(call bump_patch)) > VERSION
	@git add VERSION
	@git commit -S -m "Release $(shell $(call bump_patch))"
	$(MAKE) release
```

## Version History

Below is a chronological list of releases for the gh-stars project. Each release includes the version number and a brief description of the changes introduced.

**Note**: As the actual version history is not provided in the given context, we recommend maintaining this section by adding entries for each release, following this format:

### vX.Y.Z (YYYY-MM-DD)

- Brief description of major changes or new features
- List of bug fixes or minor improvements
- Any breaking changes or deprecations

Example:

### v1.0.0 (2023-04-15)

- Initial release of gh-stars
- Implemented core functionality for managing GitHub stars
- Added commands for saving, showing, and cleaning up starred repositories

### v1.1.0 (2023-05-01)

- Added support for filtering stars by programming language
- Improved performance of star fetching process
- Fixed bug in cleanup command for archived repositories

### v1.1.1 (2023-05-10)

- Fixed issue with authentication using personal access tokens
- Updated documentation for installation process

(Continue adding entries for each release as they occur)

## Updating This Document

To keep this version history up-to-date, make sure to add a new entry to the Version History section each time a release is made. Include the version number, release date, and a concise summary of the changes introduced in that release.