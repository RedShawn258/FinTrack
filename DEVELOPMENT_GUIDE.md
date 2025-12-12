# Development Guide

## Branch Strategy

You're now working on the **`development`** branch, which is separate from `main`. This keeps the original group project code intact.

### Current Setup

- **`main` branch**: Original group project code (untouched)
- **`development` branch**: Your personal development branch (current)

### Working with Branches

#### Switch to your development branch (if not already there)
```bash
git checkout development
```

#### Switch back to main (to view original code)
```bash
git checkout main
```

#### Create a new feature branch from development
```bash
git checkout development
git checkout -b feature/your-feature-name
# Make your changes
git add .
git commit -m "feat: your feature description"
git checkout development
git merge feature/your-feature-name
```

#### Push your development branch to remote
```bash
git push -u origin development
```

#### Keep development branch updated with main (if main gets updates)
```bash
git checkout main
git pull origin main
git checkout development
git merge main  # Merge any updates from main into your development branch
```

### Important Notes

1. **Never commit directly to `main`** - Always work on `development` or feature branches
2. **The `main` branch remains unchanged** - Your cleanup and improvements are on `development`
3. **All your future work** should be done on `development` or feature branches

### Current Changes on Development Branch

The following improvements have been committed to `development`:

✅ Backend cleanup and improvements:
- Initialized `go.mod` file
- Removed duplicate middleware setup
- Fixed logger error handling
- Made CORS configurable
- Added backend structure documentation

### Next Steps

1. Continue developing on the `development` branch
2. Create feature branches for specific features
3. Keep `main` branch as reference to original group project

### Useful Git Commands

```bash
# See what branch you're on
git branch

# See all branches (local and remote)
git branch -a

# See commit history
git log --oneline --graph --all

# See what files changed between branches
git diff main..development

# Discard local changes (be careful!)
git restore <file>
```

