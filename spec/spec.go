// Package spec embeds the impact model specification template and the
// rubrics an LLM-as-a-Judge uses to review an impact model. It makes no model calls:
// a judge, human or agent, reads a bundle built from these and the document.
package spec

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"slices"

	"github.com/plexusone/structured-evaluation/rubric"
	"gopkg.in/yaml.v3"
)

//go:embed toc.md
var template string

//go:embed judge.md
var judgeInstructions string

//go:embed toc.rubric.yaml
var coreYAML []byte

//go:embed toc.development.rubric.yaml
var developmentYAML []byte

// JudgeInstructions returns the instructions a judge reads with the bundle.
func JudgeInstructions() string { return judgeInstructions }

// Rubric profiles.
const (
	// ProfileCore is the sector-neutral rubric.
	ProfileCore = "core"
	// ProfileDevelopment adds the development-sector items.
	ProfileDevelopment = "development"
)

// ErrUnknownProfile is returned for a profile that does not exist.
var ErrUnknownProfile = errors.New("unknown rubric profile")

// Profiles lists the rubric profiles.
func Profiles() []string { return []string{ProfileCore, ProfileDevelopment} }

// Template returns the specification template, in Markdown.
func Template() string { return template }

// Rubric returns the rubric of a profile. The development rubric is the core
// rubric followed by the development categories, with the core weights scaled
// so all weights still sum to 1. Each call returns a new value.
func Rubric(profile string) (*rubric.RubricSet, error) {
	core, err := parse(coreYAML)
	if err != nil {
		return nil, fmt.Errorf("spec: core rubric: %w", err)
	}
	switch profile {
	case ProfileCore:
		return core, nil
	case ProfileDevelopment:
		add, err := parse(developmentYAML)
		if err != nil {
			return nil, fmt.Errorf("spec: development rubric: %w", err)
		}
		return merge(core, add), nil
	default:
		return nil, fmt.Errorf("%w: %q (want one of %v)", ErrUnknownProfile, profile, Profiles())
	}
}

func parse(data []byte) (*rubric.RubricSet, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var rs rubric.RubricSet
	if err := dec.Decode(&rs); err != nil {
		return nil, err
	}
	if issues := rs.Validate(); len(issues) > 0 {
		return nil, fmt.Errorf("invalid rubric: %v", issues)
	}
	return &rs, nil
}

func merge(core, add *rubric.RubricSet) *rubric.RubricSet {
	var addWeight float64
	for _, c := range add.Categories {
		addWeight += c.Weight
	}
	out := *core
	out.ID, out.Name, out.Description, out.Version = add.ID, add.Name, add.Description, add.Version
	out.Categories = make([]rubric.Category, 0, len(core.Categories)+len(add.Categories))
	for _, c := range core.Categories {
		c.Weight *= 1 - addWeight
		out.Categories = append(out.Categories, c)
	}
	out.Categories = append(out.Categories, slices.Clone(add.Categories)...)
	return &out
}
