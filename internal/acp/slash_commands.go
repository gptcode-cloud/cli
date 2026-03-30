package acp

// GetSlashCommands returns the available slash commands that GPTCode exposes via ACP.
func GetSlashCommands() []SlashCommand {
	return []SlashCommand{
		{
			Name:        "/plan",
			Description: "Switch to planning mode — analyze and plan before coding",
		},
		{
			Name:        "/review",
			Description: "Switch to code review mode — audit code for bugs, security, and best practices",
		},
		{
			Name:        "/research",
			Description: "Switch to research mode — investigate codebase, docs, and web resources",
		},
		{
			Name:        "/implement",
			Description: "Switch to implementation mode — write and test code changes",
		},
		{
			Name:        "/tdd",
			Description: "Switch to TDD mode — write failing test first, then make it pass",
		},
		{
			Name:        "/security",
			Description: "Run a security scan on the codebase",
		},
	}
}
