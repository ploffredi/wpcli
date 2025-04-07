package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

const (
	wpstoreRepoURL = "https://github.com/ploffredi/wpstore.git"
	defaultBranch  = "main"
)

type RepoManager struct {
	repoPath string
	repo     *git.Repository
}

func NewRepoManager(basePath string) *RepoManager {
	return &RepoManager{
		repoPath: filepath.Join(basePath, "wpstore"),
	}
}

func (rm *RepoManager) Clone() error {
	if _, err := os.Stat(rm.repoPath); err == nil {
		// Repository already exists, try to open it
		repo, err := git.PlainOpen(rm.repoPath)
		if err != nil {
			return fmt.Errorf("failed to open existing repository: %w", err)
		}
		rm.repo = repo
		return nil
	}

	// Clone the repository
	repo, err := git.PlainClone(rm.repoPath, false, &git.CloneOptions{
		URL:      wpstoreRepoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	rm.repo = repo
	return nil
}

func (rm *RepoManager) Pull() error {
	if rm.repo == nil {
		return fmt.Errorf("repository not initialized")
	}

	worktree, err := rm.repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	return nil
}

func (rm *RepoManager) GetRepoPath() string {
	return rm.repoPath
}
