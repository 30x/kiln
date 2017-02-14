package registry

import "strings"

func filterReposByRootName(repos []string, name string) (filtered []string) {
	for _, repo := range repos {
		nameEnd := strings.Index(repo, "/")
		if name == repo[:nameEnd] {
			filtered = append(filtered, repo)
		}
	}

	return
}
