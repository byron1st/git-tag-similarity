package internal

import (
	"errors"
	"testing"

	"github.com/byron1st/git-tag-similarity/mocks"
	"github.com/go-git/go-git/v5/plumbing"
	"go.uber.org/mock/gomock"
)

// TestConfigValidate tests the Validate method of Config
func TestConfigValidate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		config    Config
		wantError error
	}{
		{
			name: "Valid configuration",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			wantError: nil,
		},
		{
			name: "Missing repository path",
			config: Config{
				RepoPath: "",
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			wantError: ErrMissingRepo,
		},
		{
			name: "Missing tag1 name",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "",
				Tag2Name: "v2.0.0",
			},
			wantError: ErrMissingTag1,
		},
		{
			name: "Missing tag2 name",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "",
			},
			wantError: ErrMissingTag2,
		},
		{
			name: "Non-existent repository path",
			config: Config{
				RepoPath: "/non/existent/path",
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			wantError: ErrInvalidRepo,
		},
		{
			name: "All required fields missing",
			config: Config{
				RepoPath: "",
				Tag1Name: "",
				Tag2Name: "",
			},
			wantError: ErrMissingRepo, // Should fail on first check
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantError == nil {
				if err != nil {
					t.Errorf("Validate() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("Validate() error = nil, want %v", tt.wantError)
				} else if !errors.Is(err, tt.wantError) {
					t.Errorf("Validate() error = %v, want %v", err, tt.wantError)
				}
			}
		})
	}
}

// TestConfigValidateWithRepository tests the ValidateWithRepository method
func TestConfigValidateWithRepository(t *testing.T) {
	tempDir := t.TempDir()

	// Create mock tags
	tag1 := plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "0000000000000000000000000000000000000001")
	tag2 := plumbing.NewReferenceFromStrings("refs/tags/v2.0.0", "0000000000000000000000000000000000000002")
	tags := []*plumbing.Reference{tag1, tag2}

	tests := []struct {
		name      string
		config    Config
		wantError error
	}{
		{
			name: "Both tags exist",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			wantError: nil,
		},
		{
			name: "Tag1 does not exist",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v3.0.0",
				Tag2Name: "v2.0.0",
			},
			wantError: ErrTag1NotFound,
		},
		{
			name: "Tag2 does not exist",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "v3.0.0",
			},
			wantError: ErrTag2NotFound,
		},
		{
			name: "Both tags do not exist",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v3.0.0",
				Tag2Name: "v4.0.0",
			},
			wantError: ErrTag1NotFound, // Should fail on first check
		},
		{
			name: "Invalid repository path",
			config: Config{
				RepoPath: "/non/existent/path",
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			wantError: ErrInvalidRepo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().FetchAllTags().Return(tags, nil).AnyTimes()

			err := tt.config.ValidateWithRepository(mockRepo)
			if tt.wantError == nil {
				if err != nil {
					t.Errorf("ValidateWithRepository() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("ValidateWithRepository() error = nil, want %v", tt.wantError)
				} else if !errors.Is(err, tt.wantError) {
					t.Errorf("ValidateWithRepository() error = %v, want %v", err, tt.wantError)
				}
			}
		})
	}
}

// TestConfigGetTagReference tests the GetTagReference method
func TestConfigGetTagReference(t *testing.T) {
	tempDir := t.TempDir()

	tag1 := plumbing.NewReferenceFromStrings("refs/tags/v1.0.0", "0000000000000000000000000000000000000001")
	tag2 := plumbing.NewReferenceFromStrings("refs/tags/v2.0.0", "0000000000000000000000000000000000000002")
	tags := []*plumbing.Reference{tag1, tag2}

	tests := []struct {
		name      string
		config    Config
		tagName   string
		wantTag   string
		wantError bool
	}{
		{
			name: "Find existing tag v1.0.0",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			tagName:   "v1.0.0",
			wantTag:   "v1.0.0",
			wantError: false,
		},
		{
			name: "Find existing tag v2.0.0",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			tagName:   "v2.0.0",
			wantTag:   "v2.0.0",
			wantError: false,
		},
		{
			name: "Tag not found",
			config: Config{
				RepoPath: tempDir,
				Tag1Name: "v1.0.0",
				Tag2Name: "v2.0.0",
			},
			tagName:   "v3.0.0",
			wantTag:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			mockRepo.EXPECT().FetchAllTags().Return(tags, nil).AnyTimes()

			ref, err := tt.config.GetTagReference(mockRepo, tt.tagName)
			if tt.wantError {
				if err == nil {
					t.Errorf("GetTagReference() error = nil, want error")
				}
			} else {
				if err != nil {
					t.Errorf("GetTagReference() error = %v, want nil", err)
				}
				if ref == nil {
					t.Errorf("GetTagReference() returned nil reference")
				} else if ref.Name().Short() != tt.wantTag {
					t.Errorf("GetTagReference() tag = %v, want %v", ref.Name().Short(), tt.wantTag)
				}
			}
		})
	}
}
