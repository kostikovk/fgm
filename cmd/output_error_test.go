package cmd

import (
	"context"
	"fmt"
	"testing"

	"github.com/kostikovk/fgm/internal/app"
	"github.com/spf13/viper"
)

type failAfterWriter struct {
	writes int
	failAt int
}

func (w *failAfterWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes >= w.failAt {
		return 0, fmt.Errorf("write boom")
	}
	return len(p), nil
}

func TestCurrentCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newCurrentCmd(&app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7"}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 1})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestCurrentCmd_PropagatesResolverError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newCurrentCmd(&app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{}, fmt.Errorf("resolver boom")
			},
		},
	}, v)

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "resolver boom" {
		t.Fatalf("err = %v, want resolver boom", err)
	}
}

func TestCurrentCmd_ReturnsSecondWriteErrorForLintLine(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newCurrentCmd(&app.App{
		Resolver: stubSelectionResolver{
			currentFn: func(ctx context.Context, workDir string) (app.Selection, error) {
				return app.Selection{GoVersion: "1.25.7", LintVersion: "v2.11.2"}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 2})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestDoctorCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newDoctorCmd(&app.App{
		Doctor: stubDoctor{
			diagnoseFn: func(ctx context.Context, workDir string) ([]app.DoctorFinding, error) {
				return []app.DoctorFinding{{Severity: "OK", Message: "line"}}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 1})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestEnvCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	cmd := newEnvCmd(&app.App{
		EnvRenderer: stubEnvRenderer{
			renderFn: func(ctx context.Context, shell string) ([]string, error) {
				return []string{"export PATH=/tmp/fgm/shims:$PATH"}, nil
			},
		},
	})
	cmd.SetOut(&failAfterWriter{failAt: 1})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestImportAutoCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	cmd := newImportAutoCmd(&app.App{
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return []app.ImportedGo{{Version: "1.25.7", Path: "/usr/local/go"}}, nil
			},
		},
	})
	cmd.SetOut(&failAfterWriter{failAt: 1})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestImportAutoCmd_ReturnsWriteErrorForEmptyResultMessage(t *testing.T) {
	t.Parallel()

	cmd := newImportAutoCmd(&app.App{
		GoImporter: stubGoImporter{
			importAutoFn: func(ctx context.Context) ([]app.ImportedGo, error) {
				return nil, nil
			},
		},
	})
	cmd.SetOut(&failAfterWriter{failAt: 1})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestUpgradeGoCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newUpgradeGoCmd(&app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{Version: "1.26.1"}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 1})
	if err := cmd.Flags().Set("global", "true"); err != nil {
		t.Fatalf("set global flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestUpgradeGoCmd_ReturnsSecondWriteErrorForLintInstall(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newUpgradeGoCmd(&app.App{
		GoUpgrader: stubGoUpgrader{
			upgradeGlobalFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{Version: "1.26.1", LintVersion: "v2.11.2"}, nil
			},
			upgradeProjectFn: func(ctx context.Context, options app.GoUpgradeOptions) (app.GoUpgradeResult, error) {
				return app.GoUpgradeResult{}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 2})
	if err := cmd.Flags().Set("global", "true"); err != nil {
		t.Fatalf("set global flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestVersionsGoCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newVersionsGoCmd(&app.App{
		GoStore: stubGoStore{
			listLocalGoVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"1.25.7"}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 1})
	if err := cmd.Flags().Set("local", "true"); err != nil {
		t.Fatalf("set local flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestVersionsLintCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newVersionsLintCmd(&app.App{
		LintRemoteProvider: stubLintRemoteProvider{
			listRemoteLintVersionsFn: func(ctx context.Context, goVersion string) ([]app.LintVersion, error) {
				return []app.LintVersion{{Version: "v2.11.2", Recommended: true}}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 1})
	if err := cmd.Flags().Set("remote", "true"); err != nil {
		t.Fatalf("set remote flag: %v", err)
	}
	if err := cmd.Flags().Set("go", "1.25.7"); err != nil {
		t.Fatalf("set go flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestVersionsLintCmd_LocalWriteError(t *testing.T) {
	t.Parallel()

	v := viper.New()
	v.Set(flagChdir, ".")

	cmd := newVersionsLintCmd(&app.App{
		LintStore: stubLintStore{
			listLocalLintVersionsFn: func(ctx context.Context) ([]string, error) {
				return []string{"v2.11.2"}, nil
			},
		},
	}, v)
	cmd.SetOut(&failAfterWriter{failAt: 1})
	if err := cmd.Flags().Set("local", "true"); err != nil {
		t.Fatalf("set local flag: %v", err)
	}

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}

func TestVersionCmd_ReturnsWriteError(t *testing.T) {
	t.Parallel()

	cmd := newVersionCmd(&app.App{
		BuildInfo: app.BuildInfo{
			Version: "v0.1.0",
			Commit:  "abc1234",
			Date:    "2026-03-11T10:00:00Z",
		},
	})
	cmd.SetOut(&failAfterWriter{failAt: 1})

	err := cmd.RunE(cmd, nil)
	if err == nil || err.Error() != "write boom" {
		t.Fatalf("err = %v, want write boom", err)
	}
}
