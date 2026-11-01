package kb

import "gopkg.in/yaml.v3"

// Frontmatter is optional metadata supplied with agent writes.
type Frontmatter struct {
	Title   string   `json:"title,omitempty" yaml:"title,omitempty" jsonschema:"Optional display title"`
	Tags    []string `json:"tags,omitempty" yaml:"tags,omitempty" jsonschema:"Optional tags"`
	Aliases []string `json:"aliases,omitempty" yaml:"aliases,omitempty" jsonschema:"Optional link aliases"`
	Created string   `json:"-" yaml:"created,omitempty"`
}

// Content serializes safe YAML frontmatter without interpreting body text.
func Content(body string, metadata *Frontmatter) (string, error) {
	if metadata == nil {
		return body, nil
	}
	raw, err := yaml.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return "---\n" + string(raw) + "---\n" + body, nil
}
