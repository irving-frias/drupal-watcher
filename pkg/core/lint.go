package core

type LintResult struct {
	File  string
	Error string
}

type LintChecker interface {
	Lint(filePath string) *LintResult
}
