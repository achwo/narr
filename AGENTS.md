# Agent Guidelines for narr

## Build/Test Commands
- Run all tests: `make test` or `go test -v -race -buildvcs ./...`
- Run single test: `go test -v -race -buildvcs ./m4b -run TestName`
- Run tests with coverage: `make test/cover`
- Build: `make build` (outputs to /tmp/bin/narr)
- Format code: `make tidy` (runs go mod tidy and go fmt)
- Full audit: `make audit` (runs tests, vet, staticcheck, govulncheck)

## Code Style
- **Imports**: Use standard library first, then external packages (spf13/cobra, gopkg.in/yaml.v3), then internal packages (github.com/achwo/narr/*)
- **Formatting**: Use `go fmt` - tabs for indentation, standard Go formatting
- **Naming**: Exported names start with capital (Project, NewProject), unexported lowercase (audioFileProvider, trackFactory). Use descriptive names.
- **Error handling**: Always wrap errors with context using `fmt.Errorf("description: %w", err)`. Check errors immediately after operations.
- **Interfaces**: Define small, focused interfaces (audioFileProvider, audioProcessor). Place interface definitions near usage, not implementation.
- **Comments**: Add doc comments for all exported types/functions. Format: "FunctionName does X and returns Y."
- **Testing**: Use testify/assert for assertions. Name tests TestFunctionName_Condition_ExpectedResult.
- **Types**: Prefer explicit types. Use pointers for large structs and when mutation is needed.
- **Dependencies**: Use dependency injection via struct fields (see ProjectDependencies pattern).
