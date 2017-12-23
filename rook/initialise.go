package rook

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"gopkg.in/AlecAivazis/survey.v1"

	"github.com/Southclaws/sampctl/types"
	"github.com/Southclaws/sampctl/versioning"
)

// Init prompts the user to initialise a package
func Init(dir string) (err error) {
	var (
		pwnFiles []string
		incFiles []string
	)

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) (innerErr error) {
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		rel, innerErr := filepath.Rel(dir, path)
		if innerErr != nil {
			return
		}

		if ext == ".pwn" {
			pwnFiles = append(pwnFiles, rel)
		} else if ext == ".inc" {
			incFiles = append(incFiles, rel)
		}

		return
	})
	if err != nil {
		return
	}

	color.Green("Found %d pwn files and %d inc files.", len(pwnFiles), len(incFiles))

	var questions = []*survey.Question{
		{
			Name: "Format",
			Prompt: &survey.Select{
				Message: "Preferred package format",
				Options: []string{"json", "yaml"},
			},
			Validate: survey.Required,
		},
		{
			Name:     "User",
			Prompt:   &survey.Input{Message: "Your Name - If you plan to release, must be your GitHub username."},
			Validate: survey.Required,
		},
		{
			Name:     "Repo",
			Prompt:   &survey.Input{Message: "Package Name - If you plan to release, must be the GitHub project name."},
			Validate: survey.Required,
		},
	}

	if len(pwnFiles) > 0 {
		questions = append(questions, &survey.Question{
			Name: "Entry",
			Prompt: &survey.Select{
				Message: "Choose an entry point - this is the file that is passed to the compiler.",
				Options: pwnFiles,
			},
			Validate: survey.Required,
		})
	} else {
		if len(incFiles) > 0 {
			questions = append(questions, &survey.Question{
				Name: "EntryGenerate",
				Prompt: &survey.MultiSelect{
					Message: "No .pwn found but .inc found - create .pwn file that includes .inc?",
					Options: incFiles,
				},
				Validate: survey.Required,
			})
		} else {
			questions = append(questions, &survey.Question{
				Name:   "Entry",
				Prompt: &survey.Input{Message: "No .pwn or .inc files - enter name for new script"},
			})
		}
	}

	answers := struct {
		Format        string
		User          string
		Repo          string
		EntryGenerate []string
		Entry         string
	}{}

	err = survey.Ask(questions, &answers)
	if err != nil {
		return
	}

	pkg := types.Package{
		Parent: true,
		Local:  dir,
		Format: answers.Format,
		DependencyMeta: versioning.DependencyMeta{
			User: answers.User,
			Repo: answers.Repo,
		},
	}

	if answers.Entry != "" {
		pkg.Entry = answers.Entry
		pkg.Output = strings.TrimSuffix(answers.Entry, filepath.Ext(answers.Entry)) + ".amx"
	} else {
		if len(answers.EntryGenerate) > 0 {
			buf := bytes.Buffer{}

			buf.WriteString(`// generated by "sampctl package generate"`)
			buf.WriteString("\n\n")
			for _, inc := range answers.EntryGenerate {
				buf.WriteString(fmt.Sprintf(`#include "%s"%s`, filepath.Base(inc), "\n"))
			}
			buf.WriteString("\nmain() {\n")
			buf.WriteString(`	// write tests for libraries here and run "sampctl package run"`)
			buf.WriteString("\n}\n")
			err = ioutil.WriteFile(filepath.Join(dir, "test.pwn"), buf.Bytes(), 0755)
			if err != nil {
				color.Red("failed to write generated tests.pwn file: %v", err)
			}
		}
		pkg.Entry = "test.pwn"
		pkg.Output = "test.amx"
	}

	err = pkg.WriteDefinition()

	return
}
