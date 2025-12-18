# POST_PHASE Hook

Ensure that:

 - The code base is passing the linting requirements
 - A descriptive commit (using conventional commits with a subject and a description) for the phase was successfully created.

### Execution Monitoring

#### Progress Tracking

Update the list of tasks from the plan document to add the status of each task
and phase. Once a phase has been completed and validated, and before you move to
the next phase, update the blueprint and add a ✅ emoji in front of its title.
Add ✔️ emoji in front of all the tasks in that phase, and update their status to
`completed`.

#### Task Status Updates
Valid status transitions:
- `pending` → `in-progress` (when agent starts)
- `in-progress` → `completed` (successful execution)
- `in-progress` → `failed` (execution error)
- `failed` → `in-progress` (retry attempt)

# Linting Requirements

* Uses conventional commits.
* Uses pre-commit to validate golang tests (`go test`) and commit messages.
* Writes documentation as Markdown files in `docs`, with URL friendly file names in lower case and with dashes, not underscores.
* Checks code for common security issues as described by the OWASP Top 10.

If any checks fail, fix the issues and rerun the validation. Do not proceed until all checks pass.
