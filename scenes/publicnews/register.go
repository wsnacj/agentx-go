package publicnews

import agentxpack "github.com/wsnacj/agentx-go/extensions/pack"

// RegisterPacksIntoRegistry installs the reusable public-news packs into a pack
// registry without requiring access to the host runner.
func RegisterPacksIntoRegistry(reg agentxpack.Registry) error {
	if reg == nil {
		return nil
	}
	return RegisterInto(reg)
}
