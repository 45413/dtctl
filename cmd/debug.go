package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/dynatrace-oss/dtctl/pkg/resources/livedebugger"
	"github.com/dynatrace-oss/dtctl/pkg/safety"
	"github.com/spf13/cobra"
)

var (
	debugFilters    string
	debugBreakpoint string
)

var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Manage Live Debugger workspace filters and breakpoints",
	Long: `POC command for Live Debugger GraphQL integration.

Examples:
  dtctl debug --filters k8s.namespace.name=prod
  dtctl debug --filters k8s.namespace.name=prod,dt.entity.host=HOST-123
  dtctl debug --breakpoint OrderController.java:306`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if strings.TrimSpace(debugFilters) == "" && strings.TrimSpace(debugBreakpoint) == "" {
			return fmt.Errorf("at least one of --filters or --breakpoint is required")
		}

		cfg, err := LoadConfig()
		if err != nil {
			return err
		}

		ctx, err := cfg.CurrentContextObj()
		if err != nil {
			return err
		}

		c, err := NewClientFromConfig(cfg)
		if err != nil {
			return err
		}

		handler, err := livedebugger.NewHandler(c, ctx.Environment)
		if err != nil {
			return err
		}

		workspaceResp, workspaceID, err := handler.GetOrCreateWorkspace(currentProjectPath())
		if err != nil {
			_ = printGraphQLResponse("getOrCreateWorkspaceV2", workspaceResp)
			return err
		}
		if err := printGraphQLResponse("getOrCreateWorkspaceV2", workspaceResp); err != nil {
			return err
		}

		if strings.TrimSpace(debugFilters) != "" {
			checker, err := NewSafetyChecker(cfg)
			if err != nil {
				return err
			}
			if err := checker.CheckError(safety.OperationUpdate, safety.OwnershipUnknown); err != nil {
				return err
			}

			parsedFilters, err := parseFilters(debugFilters)
			if err != nil {
				return err
			}

			updateResp, err := handler.UpdateWorkspaceFilters(workspaceID, livedebugger.BuildFilterSets(parsedFilters))
			if err != nil {
				_ = printGraphQLResponse("updateWorkspaceV2", updateResp)
				return err
			}
			if err := printGraphQLResponse("updateWorkspaceV2", updateResp); err != nil {
				return err
			}
		}

		if strings.TrimSpace(debugBreakpoint) != "" {
			checker, err := NewSafetyChecker(cfg)
			if err != nil {
				return err
			}
			if err := checker.CheckError(safety.OperationCreate, safety.OwnershipUnknown); err != nil {
				return err
			}

			fileName, lineNumber, err := parseBreakpoint(debugBreakpoint)
			if err != nil {
				return err
			}

			createResp, err := handler.CreateBreakpoint(workspaceID, fileName, lineNumber)
			if err != nil {
				_ = printGraphQLResponse("createRuleV2", createResp)
				return err
			}
			if err := printGraphQLResponse("createRuleV2", createResp); err != nil {
				return err
			}
		}

		return nil
	},
}

func currentProjectPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "no-project"
	}
	project := filepath.Base(cwd)
	if project == "" || project == "." || project == string(filepath.Separator) {
		return "no-project"
	}
	return project
}

func parseFilters(input string) (map[string][]string, error) {
	filters := map[string][]string{}
	parts := strings.Split(input, ",")
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}

		key, value, found := strings.Cut(trimmed, "=")
		if !found {
			return nil, fmt.Errorf("invalid filter %q: expected key=value", trimmed)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return nil, fmt.Errorf("invalid filter %q: key and value are required", trimmed)
		}

		filters[key] = append(filters[key], value)
	}

	if len(filters) == 0 {
		return nil, fmt.Errorf("no valid filters provided")
	}

	for key := range filters {
		sort.Strings(filters[key])
	}

	return filters, nil
}

func parseBreakpoint(input string) (string, int, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", 0, fmt.Errorf("breakpoint cannot be empty")
	}

	fileName, lineString, found := strings.Cut(trimmed, ":")
	if !found {
		return "", 0, fmt.Errorf("invalid breakpoint %q: expected File.java:line", trimmed)
	}

	fileName = strings.TrimSpace(fileName)
	lineString = strings.TrimSpace(lineString)
	if fileName == "" || lineString == "" {
		return "", 0, fmt.Errorf("invalid breakpoint %q: file and line are required", trimmed)
	}

	lineNumber, err := strconv.Atoi(lineString)
	if err != nil || lineNumber <= 0 {
		return "", 0, fmt.Errorf("invalid breakpoint line %q: must be a positive integer", lineString)
	}

	return fileName, lineNumber, nil
}

func printGraphQLResponse(operation string, payload map[string]interface{}) error {
	if payload == nil {
		return nil
	}

	wrapper := map[string]interface{}{
		"operation": operation,
		"response":  payload,
	}

	encoded, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode %s response: %w", operation, err)
	}

	fmt.Println(string(encoded))
	return nil
}

func init() {
	debugCmd.Flags().StringVar(&debugFilters, "filters", "", "filters to apply (comma-separated key=value pairs)")
	debugCmd.Flags().StringVar(&debugBreakpoint, "breakpoint", "", "breakpoint location in File.java:line format")
	rootCmd.AddCommand(debugCmd)
}
