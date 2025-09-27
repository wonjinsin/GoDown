package entity

import (
	"cheetah/config"
	"cheetah/internal/domain/value"
)

// LegacyInput represents the old Input struct for backward compatibility
type LegacyInput struct {
	URL       string
	Host      *string
	Origin    *string
	Folder    string
	Separator *string
}

// ToDownloadJob converts legacy input to new DownloadJob
func (li *LegacyInput) ToDownloadJob(cfg *config.Config) (*DownloadJob, error) {
	job := NewDownloadJob(li.URL, li.Folder, cfg.App.RepoDir)

	if li.Host != nil {
		job.SetHost(*li.Host)
	}

	if li.Origin != nil {
		job.SetOrigin(*li.Origin)
	}

	if li.Separator != nil {
		job.SetSeparator(*li.Separator)
	}

	return job, job.Validate()
}

// LegacyFile represents the old File struct for backward compatibility
type LegacyFile struct {
	Repo      string
	URL       string
	Separator *string
	Extension string
	Folder    string
}

// ToDownloadJob converts legacy file to new DownloadJob
func (lf *LegacyFile) ToDownloadJob() (*DownloadJob, error) {
	job := NewDownloadJob(lf.URL, lf.Folder, lf.Repo)

	if lf.Separator != nil {
		job.SetSeparator(*lf.Separator)
	}

	return job, job.Validate()
}

// ToFileSequence converts legacy file to new FileSequence
func (lf *LegacyFile) ToFileSequence() (*FileSequence, error) {
	return NewFileSequence(lf.URL, lf.Extension, lf.Separator), nil
}

// FromDownloadJob creates a legacy file from DownloadJob (for backward compatibility)
func FromDownloadJob(job *DownloadJob) (*LegacyFile, error) {
	sequence, err := job.GetFileSequence()
	if err != nil {
		return nil, err
	}

	return &LegacyFile{
		Repo:      job.RepoDir,
		URL:       job.URL,
		Separator: job.Separator,
		Extension: sequence.Extension,
		Folder:    job.Folder,
	}, nil
}

// CreateLegacyFileFromConfig creates a legacy file structure from input and config
func CreateLegacyFileFromConfig(input *LegacyInput, cfg *config.Config) (*LegacyFile, error) {
	// Extract extension from URL
	urlValue, err := value.NewURL(input.URL)
	if err != nil {
		return nil, err
	}

	extension, err := urlValue.GetFileExtension()
	if err != nil {
		return nil, err
	}

	return &LegacyFile{
		Repo:      cfg.App.RepoDir,
		URL:       input.URL,
		Separator: input.Separator,
		Extension: extension,
		Folder:    input.Folder,
	}, nil
}
