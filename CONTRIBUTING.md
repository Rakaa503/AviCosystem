# Contributing

Thank you for contributing to AviCosystem.

## Development Rules

- Follow the existing Go project structure.
- Keep domain logic inside the appropriate module.
- Avoid unnecessary dependencies.
- Keep handlers focused on HTTP concerns.
- Keep credentials and secrets out of source control.
- Add tests for new behavior and bug fixes.

## Before Submitting

Run:

```powershell
gofmt -w .\cmd .\internal
go test ./... -count=1
go vet ./...
go build ./...
```

## Pull Requests

A pull request should explain:

1. What changed.
2. Why the change is necessary.
3. Which modules are affected.
4. How the change was tested.

Keep changes focused and reviewable.
