package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"
)

var (
	o = flag.String("o", "", "")
	v = flag.Bool("v", false, "")
	q = flag.Bool("q", false, "")
)

const usage = `Usage: app [OPTIONS]

Options:
  -o <file>     Write the output to a file.
  -v            Enable verbose logging.
  -q            Disable all logging. Useful in CI/CD pipelines.
  -h            Show this help message.
`

type config struct {
	quiet   bool
	verbose bool

	outputFile string
}

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	cfg, err := flagsToConfig()
	if err != nil {
		usageAndExit(err.Error())
	}

	log := logger(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		log("\nInterrupted. Exiting...")
		log("\nThe stored data may be incomplete...")
		cancel()
	}()

	if err := run(ctx, cfg, log); err != nil {
		errAndExit(err.Error())
	}

	log("Done.")
}

func run(ctx context.Context, cfg config, log logF) error {
	var dst io.Writer
	if cfg.outputFile != "" {
		dst, err := os.OpenFile(cfg.outputFile, os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open file: %v", err)
		}
		defer dst.Close()
	} else {
		dst = os.Stdout
	}

	writer := csv.NewWriter(dst)
	defer func() {
		log("Flushing writer...")
		writer.Flush()
	}()

	header := []string{"Email", "Status", "Balance"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header row: %v", err)
	}

	outChan := make(chan []string)
	go func() {
		records := [][]string{
			{"john@example.com", "active", "$100.00"},
			{"jane@example.com", "inactive", "$50.00"},
			{"bob@example.com", "active", "$75.00"},
		}
		for _, record := range records {
			outChan <- record
		}
		close(outChan)
	}()

	for {
		timeout := time.After(5 * time.Second)
		select {
		case record, more := <-outChan:
			if !more {
				return nil
			}
			if err := writer.Write(record); err != nil {
				return fmt.Errorf("failed to write header row: %v", err)
			}
		case <-timeout:
			log("\nGiving up waiting...")
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func flagsToConfig() (config, error) {
	flag.Parse()
	cfg := config{
		outputFile: *o,
		verbose:    *v,
		quiet:      *q,
	}

	return cfg, nil
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprint(os.Stderr, msg)
		fmt.Fprint(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

func errAndExit(msg string) {
	fmt.Fprint(os.Stderr, msg)
	fmt.Fprint(os.Stderr, "\n")
	os.Exit(1)
}

type logF func(string, ...interface{})

func logger(cfg config) logF {
	return func(format string, args ...interface{}) {
		if cfg.quiet {
			return
		}

		if cfg.verbose {
			fmt.Fprintf(os.Stderr, format, args...)
			fmt.Fprint(os.Stderr, "\n")
		}

	}
}
