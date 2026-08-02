package publicnews

const SkillLatestNewsBrief = "latest-news-brief"

func ToolNames() []string {
	return []string{
		ToolLatestNewsLookup,
		ToolLatestNewsExtract,
		ToolLatestNewsGuard,
	}
}

func SkillNames() []string {
	return []string{SkillLatestNewsBrief}
}
