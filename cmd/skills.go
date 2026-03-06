package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dynatrace-oss/dtctl/pkg/skills"
)

var (
	skillsForAgent string // --for flag
	skillsGlobal   bool   // --global flag
	skillsYes      bool   // --yes flag (skip confirmation)
	skillsList     bool   // --list flag
)

// skillsCmd is the parent command for AI assistant skill management.
var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage AI coding assistant skill files",
	Long: `Manage dtctl skill files for AI coding assistants.

Skill files teach your AI assistant how to use dtctl effectively.
Supported agents: claude, copilot, cursor, opencode.

Examples:
  # Show available skills and detected agent
  dtctl skills

  # Auto-detect agent and install skill file
  dtctl skills install

  # Install for a specific agent
  dtctl skills install --for claude

  # List all supported agents
  dtctl skills install --list

  # Check what's installed
  dtctl skills status`,
	RunE: runSkills,
}

// skillsInstallCmd installs skill files for an AI coding assistant.
var skillsInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install skill file for an AI coding assistant",
	Long: `Install a dtctl skill file for the specified AI coding assistant.

If no agent is specified with --for, the command auto-detects the current
agent from environment variables. Use --global to install to the user-wide
location instead of the project directory.

Examples:
  # Auto-detect and install
  dtctl skills install

  # Install for Claude Code
  dtctl skills install --for claude

  # Install globally (Claude Code only)
  dtctl skills install --for claude --global

  # Overwrite existing file
  dtctl skills install --for claude --yes

  # List supported agents
  dtctl skills install --list`,
	RunE: runSkillsInstall,
}

// skillsUninstallCmd removes installed skill files.
var skillsUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove installed skill files",
	Long: `Remove dtctl skill files installed for an AI coding assistant.

If no agent is specified with --for, the command auto-detects the current
agent. Removes skill files from both project-local and global locations.

Examples:
  # Auto-detect and uninstall
  dtctl skills uninstall

  # Uninstall for a specific agent
  dtctl skills uninstall --for claude`,
	RunE: runSkillsUninstall,
}

// skillsStatusCmd shows the installation state of skill files.
var skillsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installation status of skill files",
	Long: `Show the current installation status of dtctl skill files.

Checks both project-local and global locations for all supported agents.

Examples:
  # Check all agents
  dtctl skills status

  # Check a specific agent
  dtctl skills status --for claude`,
	RunE: runSkillsStatus,
}

func init() {
	rootCmd.AddCommand(skillsCmd)

	skillsCmd.AddCommand(skillsInstallCmd)
	skillsCmd.AddCommand(skillsUninstallCmd)
	skillsCmd.AddCommand(skillsStatusCmd)

	// Flags for install
	skillsInstallCmd.Flags().StringVar(&skillsForAgent, "for", "", "install for a specific agent (claude, copilot, cursor, opencode)")
	skillsInstallCmd.Flags().BoolVar(&skillsGlobal, "global", false, "install to user-wide location instead of project directory")
	skillsInstallCmd.Flags().BoolVar(&skillsYes, "yes", false, "overwrite existing files without prompting")
	skillsInstallCmd.Flags().BoolVar(&skillsList, "list", false, "list all supported agents")

	// Flags for uninstall
	skillsUninstallCmd.Flags().StringVar(&skillsForAgent, "for", "", "uninstall for a specific agent")

	// Flags for status
	skillsStatusCmd.Flags().StringVar(&skillsForAgent, "for", "", "check status for a specific agent")
}

// runSkills shows available skills and detected agent.
func runSkills(cmd *cobra.Command, _ []string) error {
	agent, detected := skills.DetectAgent()

	fmt.Println("dtctl skill files for AI coding assistants")
	fmt.Println()

	if detected {
		fmt.Printf("Detected agent: %s (via %s env)\n", agent.DisplayName, agent.EnvVar)
	} else {
		fmt.Println("No AI agent detected in current environment.")
	}

	fmt.Println()
	fmt.Println("Supported agents:")
	for _, a := range skills.AllAgents() {
		marker := "  "
		if detected && a.Name == agent.Name {
			marker = "* "
		}
		fmt.Printf("  %s%-10s %s\n", marker, a.Name, a.DisplayName)
	}

	fmt.Println()
	fmt.Println("Run 'dtctl skills install' to install skill files.")
	fmt.Println("Run 'dtctl skills status' to check installation status.")

	return nil
}

// runSkillsInstall installs skill files.
func runSkillsInstall(_ *cobra.Command, _ []string) error {
	if skillsList {
		fmt.Println("Supported agents:")
		for _, a := range skills.AllAgents() {
			globalNote := ""
			if a.GlobalPath != "" {
				globalNote = " (supports --global)"
			}
			fmt.Printf("  %-10s %s%s\n", a.Name, a.DisplayName, globalNote)
			fmt.Printf("             Project path: %s\n", a.ProjectPath)
		}
		return nil
	}

	agent, err := resolveAgent(skillsForAgent)
	if err != nil {
		return err
	}

	baseDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine working directory: %w", err)
	}

	result, err := skills.Install(agent, baseDir, skillsGlobal, skillsYes)
	if err != nil {
		return err
	}

	if result.Replaced {
		fmt.Printf("Updated %s skill file: %s\n", result.Agent.DisplayName, result.Path)
	} else {
		fmt.Printf("Installed %s skill file: %s\n", result.Agent.DisplayName, result.Path)
	}

	scope := "project"
	if result.Global {
		scope = "global"
	}
	fmt.Printf("Scope: %s\n", scope)

	return nil
}

// runSkillsUninstall removes installed skill files.
func runSkillsUninstall(_ *cobra.Command, _ []string) error {
	agent, err := resolveAgent(skillsForAgent)
	if err != nil {
		return err
	}

	baseDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine working directory: %w", err)
	}

	removed, err := skills.Uninstall(agent, baseDir)
	if err != nil {
		return err
	}

	if len(removed) == 0 {
		fmt.Printf("No %s skill files found to remove.\n", agent.DisplayName)
		return nil
	}

	for _, path := range removed {
		fmt.Printf("Removed: %s\n", path)
	}

	return nil
}

// runSkillsStatus shows installation status.
func runSkillsStatus(_ *cobra.Command, _ []string) error {
	baseDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to determine working directory: %w", err)
	}

	if skillsForAgent != "" {
		agent, err := resolveAgent(skillsForAgent)
		if err != nil {
			return err
		}

		result := skills.Status(agent, baseDir)
		printStatus(result)
		return nil
	}

	// Show all agents
	results := skills.StatusAll(baseDir)
	anyInstalled := false
	for _, r := range results {
		if r.Installed {
			anyInstalled = true
			printStatus(r)
			fmt.Println()
		}
	}

	if !anyInstalled {
		fmt.Println("No skill files installed.")
		fmt.Println("Run 'dtctl skills install' to get started.")
	}

	return nil
}

// printStatus prints a single agent's status.
func printStatus(r *skills.StatusResult) {
	detectedAgent, detected := skills.DetectAgent()

	fmt.Printf("Agent:     %s", r.Agent.DisplayName)
	if detected && detectedAgent.Name == r.Agent.Name {
		fmt.Printf(" (detected via %s env)", r.Agent.EnvVar)
	}
	fmt.Println()

	if r.Installed {
		scope := "project"
		if r.Global {
			scope = "global"
		}
		fmt.Printf("Installed: %s (%s)\n", r.Path, scope)
	} else {
		fmt.Println("Installed: no")
	}
}

// resolveAgent resolves the target agent from --for flag or auto-detection.
func resolveAgent(forFlag string) (skills.Agent, error) {
	if forFlag != "" {
		agent, ok := skills.FindAgent(forFlag)
		if !ok {
			return skills.Agent{}, fmt.Errorf(
				"unknown agent %q\nSupported agents: %s",
				forFlag, strings.Join(skills.SupportedAgents(), ", "),
			)
		}
		return agent, nil
	}

	// Auto-detect
	agent, detected := skills.DetectAgent()
	if !detected {
		return skills.Agent{}, fmt.Errorf(
			"no AI agent detected\nUse --for to specify an agent: %s",
			strings.Join(skills.SupportedAgents(), ", "),
		)
	}

	return agent, nil
}
