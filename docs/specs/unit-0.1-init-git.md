# Unit 0.1: Initialize Git Repository and GitHub Project

**Status**: ✅ COMPLETE

**Goal**: Set up version control and remote repository

**Prerequisites**: None (fresh start)

**Estimated time**: 5 minutes

---

## Requirements

### 1. Git Repository

- Initialize a git repository in `/Users/erik/Private/code/exp/todu.sh`
- Repository must be properly initialized and functional

### 2. Ignore File

- Create `.gitignore` file appropriate for Go projects
- Must ignore:
  - Compiled binaries (todu, *.exe,*.dll, *.so,*.dylib)
  - Test binaries (*.test)
  - Coverage output (*.out)
  - Dependency directories (vendor/)
  - Go workspace file (go.work)
  - Build directories (bin/, dist/)
  - IDE files (.idea/, .vscode/, *.swp,*.swo, *~)
  - OS files (.DS_Store)
  - Config files that may contain secrets (config.yaml, .env)

### 3. GitHub Repository

- Create a private GitHub repository named `todu.sh`
- Use `gh` CLI to create the repository
- Repository must be accessible and properly configured
- Set up `origin` remote pointing to the GitHub repository

### 4. Initial Commit

- Create and push an initial commit to GitHub
- Commit must include at minimum the `.gitignore` file

---

## Success Criteria

- ✅ Git repository is initialized
- ✅ `.gitignore` file exists with appropriate patterns
- ✅ GitHub repository `todu.sh` exists and is private
- ✅ Initial commit is pushed to GitHub
- ✅ Remote `origin` is configured

---

## Verification

- Running `git status` should show a clean working directory
- Running `gh repo view` should display the repository information
- Repository should be accessible on GitHub

---

## Commit Message

```text
chore: initialize git repository and GitHub project
```
