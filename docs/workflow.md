# Development Workflow

This document describes the systematic workflow for planning, implementing, and tracking feature development using opencode and revdiff.

## Overview

The workflow consists of three phases:
1. **Planning** - Create and review implementation plans
2. **Implementation** - Execute the plan with code changes
3. **Progress Tracking** - Track completion and handle issues

## 1. Planning Phase

### Creating a Plan

When a new feature or task is requested:

1. **Create plan file** in `plans/` directory with format: `YYYY-MM-DD-<feature-name>.md`
   - Example: `plans/2026-06-11-user-authentication.md`
   - Date prefix ensures chronological sorting

2. **Use the plan template** (see below)

3. **Set status to `DRAFT`**

4. **Inform user** that plan is ready for review

### Plan Template

```markdown
# <Feature Name> Plan

## Status: DRAFT

## Overview
Brief description of what needs to be done.

## Context
- Which parts of source code/documentation should be added/changed/moved
- Other useful information discovered during analysis and plan refinement

## Tasks
- [ ] Task 1: Description
- [ ] Task 2: Description
- [ ] Task 3: Description

## Development Approach
- Testing approach (testify framework, run `make test-<service-name>`)
- Use Air hot-reload for services (changes trigger automatic rebuild/restart)
- Log files: `./docker/volumes/[service-name]/logs/[service-name].log`
- Make small, focused changes
- Complete each task fully before moving to the next
- CRITICAL: All tests must pass before starting next task
- CRITICAL: Update this plan file when scope changes during implementation
- Run tests after each change

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with `➕` prefix
- Document issues/blockers with attention symbol prefix (⚠️, ❗, etc.)

## Acceptance Criteria
- Criterion 1
- Criterion 2

## Post-completion
- Manual verification / end-to-end user-driven tests
- User inspection for database changes / manual verification as needed

## Notes
Additional context or decisions.
```

### Reviewing a Plan

1. **Automatic review**: When plan mode ends, revdiff-plan-review plugin automatically launches revdiff
2. **Manual review**: User can run `/revdiff` to review the plan
3. **Provide annotations**: User adds feedback in revdiff
4. **Revise plan**: AI updates plan based on annotations
5. **Approve plan**: When satisfied, change status to `APPROVED`

### Plan States

- `DRAFT` - Initial state, plan is being created
- `IN_REVIEW` - Plan is being reviewed with revdiff
- `APPROVED` - Plan is approved and ready for implementation
- `IN_PROGRESS` - Implementation has started
- `COMPLETED` - All tasks are done

## 2. Implementation Phase

### Starting Implementation

1. **Read approved plan** from `plans/` directory
2. **Change status to `IN_PROGRESS`**
3. **Inform user** that implementation is starting

### During Implementation

1. **Work through tasks** in order
2. **Mark tasks complete** with `[x]` as you finish them
3. **Add new tasks** with `➕` prefix if discovered
4. **Document issues** with attention symbols (⚠️, ❗, etc.)
5. **Run tests** after each change
6. **Update plan** if scope changes

### Implementation Rules

- Make small, focused changes
- Complete each task fully before moving to the next
- All tests must pass before starting next task
- Update the plan file when scope changes
- Use Air hot-reload for services (automatic rebuild/restart)
- Check logs at `./docker/volumes/[service-name]/logs/[service-name].log`

### Reviewing Implementation

Implementation review has two parts:

#### Plan Adherence Review
1. Compare implementation against plan
2. Verify all tasks are addressed
3. Check that acceptance criteria are met

#### Code Review with revdiff
1. User runs `/revdiff` to review git diff
2. Or use revdiff tool: `revdiff HEAD` or `revdiff main`
3. User provides feedback via annotations
4. AI makes revisions based on feedback

### Completing Implementation

1. **Verify all tasks are complete** in plan
2. **Change status to `COMPLETED`**
3. **Inform user** that implementation is done
4. **Suggest next steps**: manual verification, testing, etc.

## 3. Progress Tracking

### In Plan Files

- Use markdown checkboxes: `- [ ]` for incomplete, `- [x]` for complete
- Status field tracks overall state
- `➕` prefix for newly discovered tasks
- Attention symbols for issues/blockers

### Status Updates

- Automatic status updates with clear user information
- When all tasks checked, inform user and suggest moving to completed
- Update plan file immediately when status changes

### Future Enhancements (Optional)

- A separate tracking file (e.g., `plans/STATUS.md`) can be added later
- CLI scripts or Makefile targets for plan management can be considered
- Enhanced plugin for automatic implementation reviews

## Tools Integration

### opencode

- **Plan agent**: Read-only agent for creating/reviewing plans
- **Build agent**: Full permissions for implementation
- **Session management**: Persistent sessions with export/import

### revdiff

- **Plan review**: Automatic review via revdiff-plan-review plugin
- **Code review**: Manual review with `/revdiff` command or revdiff tool
- **Annotations**: User feedback captured and injected back

### Files and Directories

- `plans/` - Plan files with date prefix
- `docs/workflow.md` - This workflow documentation
- `PROGRESS.md` - Optional session progress tracking
- `docs/compaction/` - Session state preservation

## Examples

### Example: Creating a Plan

```
User: I need to add user profile image upload feature

AI: I'll create a plan for this feature.
[Creates plans/2026-06-11-profile-image-upload.md with template]
[Sets status to DRAFT]
Plan created at plans/2026-06-11-profile-image-upload.md
Status: DRAFT - ready for your review
```

### Example: Reviewing a Plan

```
[revdiff-plan-review plugin triggers automatically]
[User reviews plan in revdiff and adds annotations]

User: The plan looks good, but add error handling for invalid file types

AI: [Updates plan based on feedback]
[Changes status to IN_REVIEW, then APPROVED after final review]
Plan approved - ready for implementation
```

### Example: Implementing a Plan

```
AI: Starting implementation of profile image upload
[Changes status to IN_PROGRESS]
[Works through tasks, marking them complete]
[Runs tests after each change]
[Adds new task with ➕ prefix when discovering edge case]

AI: All tasks complete. Status: COMPLETED
[Informs user to run manual verification]
```

## Notes

- This workflow is flexible - steps can be skipped if needed
- Existing `docs/*-plan.md` files are left as is
- Start simple and enhance as needed
- The workflow can be adapted for different project sizes
