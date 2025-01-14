package init

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/aerogear/charmil/cli/pkg/factory"
	"github.com/aerogear/charmil/core/utils/color"
	"github.com/go-git/go-git/v5"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// promptContent is a struct to hold the prompt content
type promptContent struct {
	errorMsg string
	label    string
}

// TemplateContext contains the context for rendering templates
type TemplateContext struct {
	Owner   string
	Repo    string
	CliName string
}

// InitCommand initializes the starter project
func InitCommand(f *factory.Factory) *cobra.Command {

	cmd := &cobra.Command{
		Use:           f.Localizer.LocalizeByID("init.cmd.use"),
		Short:         f.Localizer.LocalizeByID("init.cmd.short"),
		Long:          f.Localizer.LocalizeByID("init.cmd.long"),
		Example:       f.Localizer.LocalizeByID("init.cmd.example"),
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			owner := promptGetInput(promptContent{
				label:    "GitHub Organization or Username",
				errorMsg: "Please Provide a username",
			})
			repo := promptGetInput(promptContent{
				label:    "GitHub Repo Name",
				errorMsg: "Please Provide a repo name",
			})
			cli_name := promptGetInput(promptContent{
				label:    "CLI Name",
				errorMsg: "Please Provide a cli name",
			})

			templateContext := TemplateContext{
				Owner:   owner,
				Repo:    repo,
				CliName: cli_name,
			}

			cloneStarter(f)
			f.Logger.Infoln(color.Info("updating starter code with names"))
			if err := renderTemplates(templateContext, f); err != nil {
				f.Logger.Error(err)
				os.Exit(1)
			}
			f.Logger.Infof(color.Success("Your %s CLI has been initialized in this directory.\n"), templateContext.CliName)
		},
	}

	return cmd
}

// promptGetInput returns a string got by prompting the user
func promptGetInput(pc promptContent) string {

	// validate function for validating prompts
	validate := func(input string) error {
		if len(input) == 0 {
			return errors.New(pc.errorMsg)
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    pc.label,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	return result
}

// clone a git repository to a local path using go-git
func cloneStarter(f *factory.Factory) {
	path, pathErr := os.Getwd()
	if pathErr != nil {
		f.Logger.Error(pathErr)
		os.Exit(1)
	}

	_, cloneErr := git.PlainClone(path, false, &git.CloneOptions{
		URL:      "https://github.com/aerogear/charmil-starter",
		Progress: f.IOStreams.Out,
	})
	if cloneErr != nil {
		f.Logger.Error(cloneErr)
		os.Exit(1)
	}
}

// render templates
func renderTemplates(templateContext TemplateContext, f *factory.Factory) error {
	path, pathErr := os.Getwd()
	if pathErr != nil {
		f.Logger.Error(pathErr)
		os.Exit(1)
	}

	// rename cli name as given by user
	oldPath := path + "/cmd/cli"
	newPath := path + "/cmd/" + templateContext.CliName
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}

	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// ignore folders
			ignoreSlice := []string{".git", ".github", ".chglog", ".goreleaser.yml", "CONTRIBUTING.md", "bin"}
			pathSplit := strings.Split(path, string(os.PathSeparator))
			intersected := intersection(pathSplit, ignoreSlice)
			if len(intersected) > 0 {
				return nil
			}

			fi, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to read file info: %w", err)
			}

			if fi.IsDir() {
				return nil
			}

			tmpl, tmplErr := template.ParseFiles(path)
			if tmplErr != nil {
				return fmt.Errorf("failed to parse template: %w", err)
			}

			f, PathErr := os.Create(path)
			if PathErr != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}

			// apply templateContext to cloned repo
			if err := tmpl.Execute(f, templateContext); err != nil {
				return fmt.Errorf("failed to execute template: %w", err)
			}

			return nil
		})
	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	return nil
}

// find intersection between two slices
func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}
	return
}
