# Tech Debt Cleanup

Pay down tech debt accumulated during feature development. General cleanup pass.

## Key Areas

- **Repeated code**: Find and consolidate duplicated patterns across backend and frontend.
- **TypeScript `any`/`unknown` types**: Audit the frontend for uses of `any` and `unknown`. Replace with proper types.
- **DRY violations**: Identify and refactor repeated logic into shared utilities/hooks/helpers.
- **Missing tests**: Add tests for important areas, especially new features (DMs, projects, @refs, memory). Follow the fluent test pattern from CLAUDE.md (`scenario().withUser(alice()).withChannel(help())...run()`).
- **General cleanup**: Dead code removal, consistent naming, proper error handling, any mess left behind by previous crumbs.
