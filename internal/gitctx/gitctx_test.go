package gitctx

import "testing"

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		wantWorkspace string
		wantRepo      string
		wantErr       bool
	}{
		{
			name:          "ssh shorthand",
			raw:           "git@bitbucket.org:myworkspace/myrepo.git",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "ssh shorthand without .git suffix",
			raw:           "git@bitbucket.org:myworkspace/myrepo",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "ssh url form",
			raw:           "ssh://git@bitbucket.org/myworkspace/myrepo.git",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "https",
			raw:           "https://bitbucket.org/myworkspace/myrepo.git",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "https without .git suffix",
			raw:           "https://bitbucket.org/myworkspace/myrepo",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "https with userinfo",
			raw:           "https://someuser@bitbucket.org/myworkspace/myrepo.git",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "http scheme",
			raw:           "http://bitbucket.org/myworkspace/myrepo.git",
			wantWorkspace: "myworkspace",
			wantRepo:      "myrepo",
		},
		{
			name:          "workspace and repo with dashes and dots",
			raw:           "git@bitbucket.org:my-workspace/my.repo-name.git",
			wantWorkspace: "my-workspace",
			wantRepo:      "my.repo-name",
		},
		{
			name:    "non-bitbucket ssh host",
			raw:     "git@github.com:myworkspace/myrepo.git",
			wantErr: true,
		},
		{
			name:    "non-bitbucket https host",
			raw:     "https://github.com/myworkspace/myrepo.git",
			wantErr: true,
		},
		{
			name:    "missing repo segment",
			raw:     "git@bitbucket.org:myworkspace",
			wantErr: true,
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: true,
		},
		{
			name:    "garbage",
			raw:     "not a url at all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace, repo, err := ParseRemoteURL(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseRemoteURL(%q) error = nil, want error", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRemoteURL(%q) unexpected error: %v", tt.raw, err)
			}
			if workspace != tt.wantWorkspace || repo != tt.wantRepo {
				t.Errorf("ParseRemoteURL(%q) = (%q, %q), want (%q, %q)",
					tt.raw, workspace, repo, tt.wantWorkspace, tt.wantRepo)
			}
		})
	}
}
