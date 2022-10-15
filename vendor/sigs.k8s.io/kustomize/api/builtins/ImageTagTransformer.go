// Code generated by pluginator on ImageTagTransformer; DO NOT EDIT.
// pluginator {unknown  1970-01-01T00:00:00Z  }

package builtins

import (
	"sigs.k8s.io/kustomize/api/filters/imagetag"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

// Find matching image declarations and replace
// the name, tag and/or digest.
type ImageTagTransformerPlugin struct {
	ImageTag   types.Image       `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	FieldSpecs []types.FieldSpec `json:"fieldSpecs,omitempty" yaml:"fieldSpecs,omitempty"`
}

func (p *ImageTagTransformerPlugin) Config(
	_ *resmap.PluginHelpers, c []byte) (err error) {
	p.ImageTag = types.Image{}
	p.FieldSpecs = nil
	return yaml.Unmarshal(c, p)
}

func (p *ImageTagTransformerPlugin) Transform(m resmap.ResMap) error {
	if err := m.ApplyFilter(imagetag.LegacyFilter{
		ImageTag: p.ImageTag,
	}); err != nil {
		return err
	}
	return m.ApplyFilter(imagetag.Filter{
		ImageTag: p.ImageTag,
		FsSlice:  p.FieldSpecs,
	})
}

func NewImageTagTransformerPlugin() resmap.TransformerPlugin {
	return &ImageTagTransformerPlugin{}
}
