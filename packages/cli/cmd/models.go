package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/hieudoanm/free.router/libs/openrouter"
	"github.com/spf13/cobra"
)

var (
	modelsSearch string
	modelsJSON   bool
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List all free models available on OpenRouter",
	RunE:  runModels,
}

func init() {
	modelsCmd.Flags().StringVarP(&modelsSearch, "search", "s", "", "Filter models by name or ID")
	modelsCmd.Flags().BoolVar(&modelsJSON, "json", false, "Output raw JSON")
}

func runModels(cmd *cobra.Command, args []string) error {
	fmt.Fprint(os.Stderr, color.CyanString("⠿ Fetching free models from OpenRouter...\n"))

	models, err := openrouter.FetchFreeModels()
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	// Filter
	if modelsSearch != "" {
		q := strings.ToLower(modelsSearch)
		filtered := models[:0]
		for _, m := range models {
			if strings.Contains(strings.ToLower(m.ID), q) ||
				strings.Contains(strings.ToLower(m.Name), q) {
				filtered = append(filtered, m)
			}
		}
		models = filtered
	}

	if len(models) == 0 {
		color.Yellow("No models found.")
		return nil
	}

	if modelsJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(models)
	}

	// Group by provider (part before the first "/")
	grouped := map[string][]openrouter.Model{}
	for _, m := range models {
		parts := strings.SplitN(m.ID, "/", 2)
		provider := parts[0]
		grouped[provider] = append(grouped[provider], m)
	}

	// Sort provider names
	providers := make([]string, 0, len(grouped))
	for p := range grouped {
		providers = append(providers, p)
	}
	sort.Strings(providers)

	bold := color.New(color.Bold)
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite)
	dim := color.New(color.Faint)

	green.Printf("\n✨ %d free model(s) on OpenRouter\n\n", len(models))

	for _, provider := range providers {
		cyan.Printf("  %s\n", provider)
		for _, m := range grouped[provider] {
			ctx := ""
			if m.ContextLength > 0 {
				ctx = dim.Sprintf(" [%s ctx]", formatCtx(m.ContextLength))
			}
			bold.Printf("    ")
			white.Printf("%s", m.ID)
			fmt.Println(ctx)

			if m.Description != "" {
				desc := m.Description
				if len(desc) > 72 {
					desc = desc[:71] + "…"
				}
				dim.Printf("      %s\n", desc)
			}
		}
		fmt.Println()
	}

	dim.Printf("  Run: ")
	white.Printf("fr run <model-id>")
	dim.Printf(" to start a local proxy\n\n")

	return nil
}

func formatCtx(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%dM", n/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%dk", n/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
