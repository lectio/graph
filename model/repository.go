package model

import (
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/spf13/afero"
	"io"
	"io/ioutil"
	"os"
)

type RepositoryName string

func (t RepositoryName) MarshalGQL(w io.Writer) {
	graphql.MarshalString(string(t)).MarshalGQL(w)
}

func (t *RepositoryName) UnmarshalGQL(v interface{}) error {
	str, err := graphql.UnmarshalString(v)
	if err == nil {
		*t = RepositoryName(str)
	}
	return err
}

// RepositoryManager manages a repository
type RepositoryManager interface {
	FileSystem() afero.Fs
	DirPerm() os.FileMode
	Repository() Repository
	io.Closer
}

// OpenRepositoryName finds a suitable repository manager for the given name
func (r Repositories) OpenRepositoryName(name RepositoryName) (RepositoryManager, error) {
	var searched []string
	for _, repo := range r.All {
		switch castedRepo := repo.(type) {
		case FileRepository:
			searched = append(searched, fmt.Sprintf("%q (%T)", castedRepo.Name, castedRepo))
			if castedRepo.Name == name {
				return r.OpenRepository(repo)
			}
		case TempFileRepository:
			searched = append(searched, fmt.Sprintf("%q (%T)", castedRepo.Name, castedRepo))
			if castedRepo.Name == name {
				return r.OpenRepository(repo)
			}
		case GitHubRepository:
			searched = append(searched, fmt.Sprintf("%q (%T)", castedRepo.Name, castedRepo))
			if castedRepo.Name == name {
				return r.OpenRepository(repo)
			}
		default:
			return nil, fmt.Errorf("Repository type %T is invalid", castedRepo)
		}
	}
	return nil, fmt.Errorf("Repository %q not found (found %+v)", name, searched)
}

// OpenRepository finds a suitable repository manager for the given repo and config
func (r Repositories) OpenRepository(repo Repository) (RepositoryManager, error) {
	result := repositoryManager{repo: repo}
	err := result.open()
	return result, err
}

type repositoryManager struct {
	config *Configuration
	exec   PipelineExecution
	repo   Repository
	fs     afero.Fs
}

func (rm repositoryManager) DirPerm() os.FileMode {
	return os.FileMode(0755)
}

func (rm *repositoryManager) open() error {
	switch castedRepo := rm.repo.(type) {
	case FileRepository:
		rootFs := afero.NewOsFs()
		if castedRepo.CreateRootPath {
			rootFs.MkdirAll(string(castedRepo.RootPath), rm.DirPerm())
		}
		rm.fs = afero.NewBasePathFs(afero.NewOsFs(), castedRepo.RootPath)
	case TempFileRepository:
		rootPath, err := ioutil.TempDir("", castedRepo.Prefix)
		if err != nil {
			return err
		}
		rm.fs = afero.NewBasePathFs(afero.NewOsFs(), rootPath)
	case GitHubRepository:
		rootPath, err := ioutil.TempDir("", "lectio_githubrepo_")
		if err != nil {
			return err
		}
		rm.fs = afero.NewBasePathFs(afero.NewOsFs(), rootPath)
	default:
		return fmt.Errorf("Repository type %T is invalid", castedRepo)
	}
	return nil
}

func (rm repositoryManager) FileSystem() afero.Fs {
	return rm.fs
}

func (rm repositoryManager) Repository() Repository {
	return rm.repo
}

func (rm repositoryManager) Close() error {
	return nil
}
