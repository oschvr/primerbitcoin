package utils

import (
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"regexp"
)

func GetLatestTag() string {
	repoPath := "."

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		Logger.Warn("Unable to open repo to get latest tag", err)
	}

	tagRefs, err := repo.Tags()
	if err != nil {
		Logger.Warn("Unable to get tags", err)
	}

	pattern := `refs/tags/(\d+\.\d+\.\d+)`
	reg := regexp.MustCompile(pattern)

	var latest *plumbing.Reference
	var latestTag string

	err = tagRefs.ForEach(func(tag *plumbing.Reference) error {
		if latest == nil || tag.Hash().String() > latest.Hash().String() {
			matches := reg.FindStringSubmatch(tag.Name().String())
			latestTag = matches[1]
		}
		return nil
	})
	if err != nil {
		Logger.Warn("Unable to get latest tag", err)
	}

	return latestTag
}
