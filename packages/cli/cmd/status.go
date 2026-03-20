package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/hieudoanm/free.router/libs/config"
	"github.com/hieudoanm/free.router/libs/openrouter"
	"github.com/spf13/cobra"
)

var (
	statusSearch  string
	statusWorkers int
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Probe free models and show which are available, rate-limited, or restricted",
	Long: `Sends a minimal request to each free model in parallel and reports:
  ✔ OK          — model is reachable (shows latency)
  ⚡ RATE-LIMIT  — hit upstream rate limit, try again later
  🔒 RESTRICTED  — blocked by provider privacy/guardrail settings
  ✖ ERROR        — unexpected failure

Examples:
  freerouter status                  # probe all free models
  freerouter status --search llama   # probe only matching models`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().StringVarP(&statusSearch, "search", "s", "", "Filter models by name or ID before probing")
	statusCmd.Flags().IntVarP(&statusWorkers, "workers", "w", 6, "Parallel probe workers")
}

func runStatus(cmd *cobra.Command, args []string) error {
	apiKey := config.LoadAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, color.RedString("✖ No OpenRouter API key found."))
		fmt.Fprintln(os.Stderr, color.New(color.Faint).Sprint(
			"  Set OPENROUTER_API_KEY or add it to ~/.freerouter"))
		os.Exit(1)
	}

	fmt.Fprint(os.Stderr, color.CyanString("⠿ Fetching free models from OpenRouter...\n"))

	models, err := openrouter.FetchFreeModels()
	if err != nil {
		return fmt.Errorf("failed to fetch models: %w", err)
	}

	// Optional filter
	if statusSearch != "" {
		q := strings.ToLower(statusSearch)
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

	fmt.Fprintf(os.Stderr, color.CyanString("⠿ Probing %d model(s) with %d workers...\n"),
		len(models), statusWorkers)

	// Fan-out probe jobs
	jobs := make(chan openrouter.Model, len(models))
	for _, m := range models {
		jobs <- m
	}
	close(jobs)

	results := make([]openrouter.ProbeResult, len(models))
	var mu sync.Mutex
	var wg sync.WaitGroup
	idx := 0

	for i := 0; i < statusWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range jobs {
				r := openrouter.ProbeModel(m, apiKey)
				mu.Lock()
				results[idx] = r
				idx++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	// Sort: OK first (by latency), then rate-limited, restricted, errors
	sort.Slice(results, func(i, j int) bool {
		if results[i].Status != results[j].Status {
			return results[i].Status < results[j].Status
		}
		if results[i].Status == openrouter.StatusOK {
			return results[i].Latency < results[j].Latency
		}
		return results[i].Model.ID < results[j].Model.ID
	})

	// Tally
	counts := map[openrouter.ProbeStatus]int{}
	for _, r := range results {
		counts[r.Status]++
	}

	fmt.Println()

	green := color.New(color.FgGreen, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	red := color.New(color.FgRed, color.Bold)
	dim := color.New(color.Faint)
	white := color.New(color.FgWhite)

	green.Printf("  ✔ %d OK", counts[openrouter.StatusOK])
	fmt.Print("  ")
	yellow.Printf("⚡ %d rate-limited", counts[openrouter.StatusRateLimited])
	fmt.Print("  ")
	red.Printf("🔒 %d restricted", counts[openrouter.StatusRestricted])
	fmt.Print("  ")
	red.Printf("✖ %d error", counts[openrouter.StatusError])
	fmt.Println()
	fmt.Println()

	for _, r := range results {
		switch r.Status {
		case openrouter.StatusOK:
			green.Print("  ✔ OK          ")
			white.Printf("%-52s", r.Model.ID)
			dim.Printf(" %dms\n", r.Latency)

		case openrouter.StatusRateLimited:
			yellow.Print("  ⚡ RATE-LIMIT  ")
			white.Printf("%-52s", r.Model.ID)
			if r.Message != "" {
				dim.Printf(" %s\n", truncate(r.Message, 60))
			} else {
				fmt.Println()
			}

		case openrouter.StatusRestricted:
			red.Print("  🔒 RESTRICTED  ")
			white.Printf("%-52s", r.Model.ID)
			dim.Printf(" privacy/guardrail\n")

		case openrouter.StatusError:
			red.Print("  ✖ ERROR        ")
			white.Printf("%-52s", r.Model.ID)
			if r.Message != "" {
				dim.Printf(" %s\n", truncate(r.Message, 60))
			} else {
				fmt.Println()
			}
		}
	}

	fmt.Println()
	dim.Print("  Tip: ")
	fmt.Print("rate-limited models recover in ~1 min. ")
	dim.Println("Restricted models need openrouter.ai/settings/privacy")
	fmt.Println()

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
