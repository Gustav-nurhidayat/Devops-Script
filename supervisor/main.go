package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const (
	appPath    = "/app"
	newAppPath = "/app.new"
)

func startApp() (*exec.Cmd, error) {
	cmd := exec.Command(appPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	fmt.Printf("started application, pid=%d\n", cmd.Process.Pid)

	return cmd, nil
}

func restartApp(app **exec.Cmd) error {
	if _, err := os.Stat(newAppPath); err == nil {
		if err := os.Rename(newAppPath, appPath); err != nil {
			return fmt.Errorf("failed to replace application: %w", err)
		}

		fmt.Println("replaced application binary")
	}

	if (*app).Process != nil {
		_ = (*app).Process.Kill()
	}

	_ = (*app).Wait()

	newApp, err := startApp()
	if err != nil {
		return err
	}

	*app = newApp
	return nil
}

func main() {
	app, err := startApp()
	if err != nil {
		fmt.Println("failed to start application:", err)
		os.Exit(1)
	}

	signals := make(chan os.Signal, 1)

	signal.Notify(
		signals,
		syscall.SIGHUP,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	for {
		switch sig := <-signals; sig {
		case syscall.SIGHUP:
			fmt.Println("received SIGHUP: restarting application")

			if err := restartApp(&app); err != nil {
				fmt.Println("failed to restart application:", err)
				os.Exit(1)
			}

		case syscall.SIGTERM, syscall.SIGINT:
			fmt.Println("received shutdown signal")

			if app.Process != nil {
				_ = app.Process.Signal(syscall.SIGTERM)
			}

			_ = app.Wait()
			os.Exit(0)
		}
	}
}
