# Commit Message Style

## Format

```
<action> <what>

- <detail 1>
- <detail 2>
- <detail 3>
```

## Rules

1. **lowercase everything** - no capitals except proper nouns (MongoDB, IoC, HTTP, CCPA)
2. **be direct** - get to the point immediately
3. **no fluff words** - avoid: comprehensive, ensure, enhance, improve, robust, leverage
4. **use simple verbs** - add, fix, remove, update, refactor, create
5. **bulletpoints** - use `-` for multi-line details
6. **under 72 chars** - first line should fit in terminal

## Good Examples

```
add billing architecture docs

- 65+ operation types for all monetizable actions
- two-phase pattern: validate before, execute after
- integrated with BaseUseCase for DRY
- security: always use authenticated context UserID
```

```
fix mongodb test connection in docker

- use host.docker.internal instead of 127.0.0.1
- add MONGO_URI env variable support
- update TESTING.md with docker instructions
```

```
refactor PlayerProfileController to use usecases

- inject usecases in constructor instead of resolving in handlers
- remove direct repository access
- add billing validation to delete operation
```

## Bad Examples

❌ **Too verbose**
```
Comprehensively enhance the billing architecture to ensure robust monetization

This commit enhances the billing system by adding comprehensive documentation
that leverages best practices and ensures secure implementation across all
monetizable operations in the platform.
```

❌ **Too vague**
```
update stuff

- made some changes
- fixed things
```

❌ **Unnecessary details**
```
Add billing documentation (BILLING.md) to the project repository

Created a new file called BILLING.md which contains documentation about
the billing system. This file explains how billing works and provides
examples of how to implement billing in usecases.
```

## Action Verbs

- `add` - new files, features, functionality
- `remove` - delete files, code, dependencies
- `fix` - bug fixes, corrections
- `refactor` - code restructuring without behavior change
- `update` - modify existing features
- `create` - similar to add, use interchangeably
- `implement` - when adding new patterns/architecture
- `integrate` - connecting systems together

## Details Format

- start with **what**, not **why** (unless bug fix)
- be specific: "add BaseUseCase.RequireAuth()" not "add auth"
- mention affected files when relevant
- include breaking changes if any
- max 5 bulletpoints

## Examples by Type

### New Feature
```
add visibility middleware for CCPA compliance

- protects GET endpoints with ownership checks
- supports ShareToken validation
- no admin bypass (AdminBypassDisabled enum)
- checks visibility levels: public/restricted/private
```

### Bug Fix
```
fix squad member ownership check bypass

- add RequireOwnership() in add/remove member usecases
- verify caller owns or admins squad before modifying
- prevents unauthorized member management
```

### Refactoring
```
create BaseUseCase for DRY across usecases

- RequireAuthentication, RequireOwnership helpers
- ValidateBilling, ExecuteBilling methods
- ExecuteOperation template for common pattern
- reduces duplication in 15+ usecases
```

### Documentation
```
add billing architecture docs

- operation types reference (65+ operations)
- two-phase billing pattern examples
- security best practices
- testing guidelines
```

### Configuration
```
update mongodb uri for docker environments

- use host.docker.internal by default
- add MONGO_URI env variable support
- update test documentation
```

## For AI Assistants (Claude, Copilot)

When generating commits:
1. read this file first
2. match the style exactly
3. no marketing language
4. no unnecessary adjectives
5. github user style: direct, lowercase, factual
6. if unsure, look at recent commits in `git log --oneline -20`
