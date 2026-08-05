package catalog

import (
	toolcontract "github.com/wsnacj/agentx-go/components/tool"
	"github.com/wsnacj/agentx-go/extensions/skills"
)

// ProjectTools projects model-facing Tool declarations into discovery-only
// assets. Parameters, handlers and executors intentionally remain with Tool.
func ProjectTools(sourceRef string, definitions []toolcontract.Definition) []Asset {
	assets := make([]Asset, 0, len(definitions))
	for _, definition := range definitions {
		assets = append(assets, Asset{
			Identity:    Identity{Kind: KindTool, ID: definition.Function.Name},
			Name:        definition.Function.Name,
			Description: definition.Function.Description,
			SourceRef:   sourceRef,
		})
	}
	return assets
}

// ProjectSkills projects loaded Skill metadata into discovery-only assets.
// Prompt content, install instructions, dispatch and resource paths are not
// copied into the catalog envelope.
func ProjectSkills(sourceRef string, items []skills.Skill) []Asset {
	assets := make([]Asset, 0, len(items))
	for _, item := range items {
		assets = append(assets, Asset{
			Identity:    Identity{Kind: KindSkill, ID: skills.ResolveSkillKey(item)},
			Name:        item.Name,
			Description: item.Description,
			SourceRef:   sourceRef,
			Tags:        item.Tags,
			Keywords:    item.Keywords,
		})
	}
	return assets
}
