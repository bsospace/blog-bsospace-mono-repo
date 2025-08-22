# Husky Git Hooks

This directory contains Git hooks that run automatically before commits and pushes to ensure code quality.

## Hooks

### Pre-commit Hook (`.husky/pre-commit`)
Runs before each commit to ensure:
- ✅ **ESLint** - Code quality and style consistency
- ✅ **TypeScript Type Check** - No type errors

### Pre-push Hook (`.husky/pre-push`)
Runs before pushing to remote repository to ensure:
- ✅ **Build Success** - Project builds without errors
- ✅ **No Breaking Changes** - Prevents pushing broken code

## How It Works

1. **Pre-commit**: When you run `git commit`, it automatically:
   - Runs `pnpm lint` to check code quality
   - Runs `pnpm type-check` to verify TypeScript types
   - Only allows commit if all checks pass

2. **Pre-push**: When you run `git push`, it automatically:
   - Runs `pnpm build` to ensure the project builds successfully
   - Only allows push if build passes
   - Prevents pushing broken code to the repository

## Benefits

- 🚫 **Prevents Broken Code** - No more pushing code that doesn't build
- 🔍 **Code Quality** - Ensures consistent code style and no linting errors
- 🛡️ **Type Safety** - Catches TypeScript errors before they reach production
- ⚡ **Early Detection** - Issues are caught locally before affecting the team

## Troubleshooting

If hooks fail:

1. **Linting Errors**: Run `pnpm lint` to see and fix issues
2. **Type Errors**: Run `pnpm type-check` to see type issues
3. **Build Errors**: Run `pnpm build` to see build problems

## Disabling Hooks (Emergency Only)

If you need to bypass hooks in an emergency:

```bash
# Skip pre-commit hook
git commit --no-verify -m "Emergency fix"

# Skip pre-push hook
git push --no-verify
```

⚠️ **Warning**: Only use `--no-verify` in true emergencies!
