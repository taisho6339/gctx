package main

import (
	"bytes"
	"errors"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	ActiveItemColor = color.New(color.FgYellow, color.Bold)
)

func runCommand(in io.Reader, name string, args ...string) (*bytes.Buffer, error) {
	var out bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdin = in
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	return &out, nil
}

func projects() (*bytes.Buffer, error) {
	projOut, err := runCommand(os.Stdin, "gcloud", "projects", "list", "--format", "value(projectId)")
	if err != nil {
		return nil, err
	}
	curOut, err := runCommand(os.Stdin, "gcloud", "config", "get-value", "core/project")
	if err != nil {
		return nil, err
	}

	proj := projOut.String()
	curProj := curOut.String()
	projs := strings.Split(proj, "\n")
	ret := make([]string, len(projs))
	for i, p := range projs {
		if strings.TrimSpace(p) == strings.TrimSpace(curProj) {
			ret[i] = ActiveItemColor.Sprint(p)
		} else {
			ret[i] = p
		}
	}
	return bytes.NewBufferString(strings.Join(ret, "\n")), nil
}

func switchProject(projectId string) error {
	if _, err := runCommand(os.Stdin, "gcloud", "config", "set", "project", projectId); err != nil {
		return err
	}
	return nil
}

func fzf(in io.Reader) string {
	out, err := runCommand(in, "fzf", "--ansi", "--no-preview")
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			log.Fatal(err)
		}
		return ""
	}
	return strings.TrimSpace(out.String())
}

func isGCPAuthenticated() bool {
	out, err := runCommand(os.Stdin, "gcloud", "config", "list", "--format", "value(core.account)")
	if err != nil {
		return false
	}
	if strings.TrimSuffix(out.String(), "\n") == "" {
		return false
	}
	return true
}

func isIstalledfzf() bool {
	v, _ := exec.LookPath("fzf")
	if v != "" {
		return true
	}
	return false
}

func isInstallGcloud() bool {
	v, _ := exec.LookPath("gcloud")
	if v != "" {
		return true
	}
	return false
}

var rootCmd = &cobra.Command{
	Use:     "gctx",
	Short:   "gctx is a tool to switch GCP project",
	Version: "1.0.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		// gcloudの環境変数チェック
		if !isIstalledfzf() {
			return errors.New("gctx requires fzf, please install fzf")
		}
		if !isInstallGcloud() {
			return errors.New("gctx requires gcloud, please install gcloud")
		}
		if !isGCPAuthenticated() {
			return errors.New("you haven't logged in yet, please run gcloud auth login ACCOUNT")
		}
		projects, err := projects()
		if err != nil {
			return err
		}
		selectedProj := fzf(projects)
		if selectedProj == "" {
			return nil
		}
		if err := switchProject(selectedProj); err != nil {
			return err
		}
		return nil
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
