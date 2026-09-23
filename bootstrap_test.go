package autoconfigure

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/larsartmann/go-finding"
	toolsdk "github.com/larsartmann/go-finding/toolsdk"
)

// errRegistryUnavailable is a static test failure for the Generate hook.
var errRegistryUnavailable = errors.New("registry unavailable")

// testBootstrapSpec is a minimal fake domain over a settings map: the project
// is "recognizable" when a package.json marker exists, Generate produces a
// fixed optimal config, and Compare projects onto DiffMaps. Its lifecycle
// contract mirrors the oxlint provider semantics BootstrapProviderFromSpec
// absorbed.
func testBootstrapSpec() BootstrapSpec[map[string]string] {
	return BootstrapSpec[map[string]string]{
		Name:        "fake-auto-configure",
		Description: "generates .fakerc.json for recognizable projects",
		ConfigFile:  ".fakerc.json",
		ConfigFiles: []finding.FilePath{".fakerc.json", ".fakerc.jsonc"},
		MissingRule: "FAKE_CONFIG_MISSING",
		FixCommand:  "fake-auto-configure configure",
		CountLabel:  "rules",
		Recognizable: func(ctx context.Context) (bool, error) {
			_, err := os.Stat(filepath.Join(WorkingDir(ctx), "package.json"))

			return err == nil, nil
		},
		Generate: func(ctx context.Context) (map[string]string, int, error) {
			return map[string]string{"plugins": "react", "rules": "strict"}, 3, nil
		},
		Marshal: func(v map[string]string) ([]byte, *ConfigError) {
			data, err := MarshalJSONIndented(v)
			if err != nil {
				return nil, err
			}

			return append(data, '\n'), nil
		},
		Compare: func(existing, expected map[string]string) []Change {
			return DiffMaps(existing, expected, "")
		},
	}
}

func bootstrapCtx(t *testing.T, dir string) context.Context {
	t.Helper()

	return finding.WithWorkingDir(context.Background(), dir)
}

func writeBootstrapFile(t *testing.T, dir, name, content string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func mustBootstrapSpec(t *testing.T) toolsdk.Spec {
	t.Helper()

	spec, err := BootstrapProviderFromSpec(testBootstrapSpec())
	if err != nil {
		t.Fatalf("BootstrapProviderFromSpec: %v", err)
	}

	return spec
}

func TestBootstrapProviderFromSpec_Validation(t *testing.T) {
	t.Parallel()

	valid := testBootstrapSpec()

	cases := []struct {
		name string
		mut  func(BootstrapSpec[map[string]string]) BootstrapSpec[map[string]string]
		want error
	}{
		{"missing name", func(s BootstrapSpec[map[string]string]) BootstrapSpec[map[string]string] {
			s.Name = ""

			return s
		}, ErrNameRequired},
		{"missing description", func(s BootstrapSpec[map[string]string]) BootstrapSpec[map[string]string] {
			s.Description = ""

			return s
		}, ErrDescriptionRequired},
		{"missing config file", func(s BootstrapSpec[map[string]string]) BootstrapSpec[map[string]string] {
			s.ConfigFile = ""

			return s
		}, ErrConfigFileRequired},
		{"missing generate", func(s BootstrapSpec[map[string]string]) BootstrapSpec[map[string]string] {
			s.Generate = nil

			return s
		}, ErrGenerateRequired},
		{"missing compare", func(s BootstrapSpec[map[string]string]) BootstrapSpec[map[string]string] {
			s.Compare = nil

			return s
		}, ErrCompareRequired},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			_, err := BootstrapProviderFromSpec(testCase.mut(valid))
			if !errors.Is(err, testCase.want) {
				t.Fatalf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestBootstrapProviderFromSpec_InputsDeriveFromConfigFiles(t *testing.T) {
	t.Parallel()

	spec := mustBootstrapSpec(t)

	want := []string{".fakerc.json", ".fakerc.jsonc"}
	if len(spec.Inputs) != len(want) {
		t.Fatalf("Inputs = %v, want %v", spec.Inputs, want)
	}

	for i, input := range want {
		if spec.Inputs[i] != input {
			t.Fatalf("Inputs = %v, want %v", spec.Inputs, want)
		}
	}

	single := testBootstrapSpec()
	single.ConfigFiles = nil

	singleSpec, err := BootstrapProviderFromSpec(single)
	if err != nil {
		t.Fatalf("BootstrapProviderFromSpec: %v", err)
	}

	if len(singleSpec.Inputs) != 1 || singleSpec.Inputs[0] != ".fakerc.json" {
		t.Fatalf("Inputs = %v, want [.fakerc.json]", singleSpec.Inputs)
	}
}

func TestBootstrapDetect_MissingConfigInRecognizableProject(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)

	findings, err := mustBootstrapSpec(t).Detect.Detect(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(findings))
	}

	missing := findings[0]
	if missing.Rule != "FAKE_CONFIG_MISSING" {
		t.Errorf("Rule = %q, want FAKE_CONFIG_MISSING", missing.Rule)
	}

	if missing.ToolName != "fake-auto-configure" {
		t.Errorf("ToolName = %q, want fake-auto-configure", missing.ToolName)
	}

	if missing.Severity != finding.SeverityWarning {
		t.Errorf("Severity = %v, want warning", missing.Severity)
	}

	if missing.Position.File != ".fakerc.json" {
		t.Errorf("Position.File = %q, want .fakerc.json", missing.Position.File)
	}

	if missing.FixStrategy != finding.FixStrategySuggest {
		t.Errorf("FixStrategy = %v, want suggest", missing.FixStrategy)
	}
}

func TestBootstrapDetect_DefaultRuleName(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)

	spec := testBootstrapSpec()
	spec.MissingRule = ""

	provider, err := BootstrapProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("BootstrapProviderFromSpec: %v", err)
	}

	findings, err := provider.Detect.Detect(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 1 || findings[0].Rule != "CONFIG_MISSING" {
		t.Fatalf("rule = %v, want CONFIG_MISSING", findings)
	}
}

func TestBootstrapDetect_ExistingConfigNeverFlagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.json", `{"rules":"off"}`)

	findings, err := mustBootstrapSpec(t).Detect.Detect(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("an existing config must never be flagged for regeneration, got %v", findings)
	}
}

func TestBootstrapDetect_ExistingShadowConfigNeverFlagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.jsonc", "// curated config\n{}")

	findings, err := mustBootstrapSpec(t).Detect.Detect(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf(
			"an existing .fakerc.jsonc is a config too: flagging it would make repair generate a shadowing .fakerc.json; got %v",
			findings,
		)
	}
}

func TestBootstrapDetect_UnrecognizableProjectSkipped(t *testing.T) {
	t.Parallel()

	findings, err := mustBootstrapSpec(t).Detect.Detect(bootstrapCtx(t, t.TempDir()))
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if len(findings) != 0 {
		t.Fatalf("a directory without any project marker must not be flagged, got %v", findings)
	}
}

func TestBootstrapRepair_DryRunHoldsBackWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	ctx := toolsdk.WithDryRun(bootstrapCtx(t, dir), true)

	result, err := mustBootstrapSpec(t).Repair.Repair(ctx)
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}

	if !strings.Contains(result.Description, "dry-run") {
		t.Errorf("description = %q, want it to mention dry-run", result.Description)
	}

	if _, err := os.Stat(filepath.Join(dir, ".fakerc.json")); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not write the config")
	}
}

func TestBootstrapRepair_WritesConfig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)

	result, err := mustBootstrapSpec(t).Repair.Repair(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}

	if !strings.Contains(result.Description, "wrote") {
		t.Errorf("description = %q, want it to mention wrote", result.Description)
	}

	if !strings.Contains(result.Description, "(3 rules)") {
		t.Errorf("description = %q, want the count label (3 rules)", result.Description)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".fakerc.json"))
	if err != nil {
		t.Fatalf("read written config: %v", err)
	}

	if len(data) == 0 || data[len(data)-1] != '\n' {
		t.Errorf("written config must carry the Marshal hook's trailing newline")
	}

	parsed, configErr := ParseJSON[map[string]string](data)
	if configErr != nil {
		t.Fatalf("written config must round-trip: %v", configErr)
	}

	if len(*parsed) != 2 {
		t.Errorf("parsed = %v, want the 2 generated settings", parsed)
	}
}

func TestBootstrapRepair_ExistingConfigUntouched(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.json", `{"rules":"off"}`)

	result, err := mustBootstrapSpec(t).Repair.Repair(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}

	if !strings.Contains(result.Description, "already exists") {
		t.Errorf("description = %q, want it to mention already exists", result.Description)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".fakerc.json"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	if string(data) != `{"rules":"off"}` {
		t.Fatalf("repair must never overwrite an existing config, got %q", data)
	}
}

func TestBootstrapRepair_ExistingShadowConfigNotShadowed(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.jsonc", "// curated config\n{}")

	result, err := mustBootstrapSpec(t).Repair.Repair(bootstrapCtx(t, dir))
	if err != nil {
		t.Fatalf("Repair: %v", err)
	}

	if !strings.Contains(result.Description, "already exists") {
		t.Errorf("description = %q, want it to mention already exists", result.Description)
	}

	if _, err := os.Stat(filepath.Join(dir, ".fakerc.json")); !os.IsNotExist(err) {
		t.Fatalf("generating .fakerc.json next to a curated .fakerc.jsonc would shadow it")
	}
}

func TestBootstrapRepair_GenerateErrorPropagates(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)

	spec := testBootstrapSpec()
	spec.Generate = func(ctx context.Context) (map[string]string, int, error) {
		return nil, 0, errRegistryUnavailable
	}

	provider, err := BootstrapProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("BootstrapProviderFromSpec: %v", err)
	}

	if _, err := provider.Repair.Repair(bootstrapCtx(t, dir)); err == nil ||
		!strings.Contains(err.Error(), "registry unavailable") {
		t.Fatalf("Repair error = %v, want the Generate failure to propagate", err)
	}
}

func TestBootstrapHealth_NoConfigIsHealthy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)

	if err := mustBootstrapSpec(t).HealthCheck(bootstrapCtx(t, dir)); err != nil {
		t.Fatalf("a missing config is Detect's finding, not a health failure: %v", err)
	}
}

func TestBootstrapHealth_ShadowOnlyConfigIsHealthy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.jsonc", `{"rules":"off"}`)

	if err := mustBootstrapSpec(t).HealthCheck(bootstrapCtx(t, dir)); err != nil {
		t.Fatalf("a curated .fakerc.jsonc is out of scope: %v", err)
	}
}

func TestBootstrapHealth_FreshRepairIsHealthy(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	ctx := bootstrapCtx(t, dir)
	provider := mustBootstrapSpec(t)

	if _, err := provider.Repair.Repair(ctx); err != nil {
		t.Fatalf("Repair: %v", err)
	}

	if err := provider.HealthCheck(ctx); err != nil {
		t.Fatalf("a config this tool just wrote must match what it would generate: %v", err)
	}
}

func TestBootstrapHealth_DriftReportedAdvisory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.json", `{"plugins":"react","rules":"off"}`)

	err := mustBootstrapSpec(t).HealthCheck(bootstrapCtx(t, dir))
	if err == nil {
		t.Fatalf("a drifted config must be reported")
	}

	if !errors.Is(err, errConfigDrift) {
		t.Errorf("error must wrap the drift sentinel, got %v", err)
	}

	for _, want := range []string{"drifted", "fake-auto-configure configure", "advisory only"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must contain %q", err.Error(), want)
		}
	}

	data, readErr := os.ReadFile(filepath.Join(dir, ".fakerc.json"))
	if readErr != nil {
		t.Fatalf("read config: %v", readErr)
	}

	if string(data) != `{"plugins":"react","rules":"off"}` {
		t.Fatalf("the health check is report-only and must never touch the config, got %q", data)
	}
}

func TestBootstrapHealth_MalformedConfigReported(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	writeBootstrapFile(t, dir, ".fakerc.json", `{"rules":`)

	err := mustBootstrapSpec(t).HealthCheck(bootstrapCtx(t, dir))
	if err == nil {
		t.Fatalf("a malformed config must be reported")
	}

	for _, want := range []string{"malformed", "fake-auto-configure configure"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must contain %q", err.Error(), want)
		}
	}
}

func TestBootstrapHealth_NormalizeExpectedSuppressesDrift(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeBootstrapFile(t, dir, "package.json", `{"name":"app"}`)
	// The "overrides" key is user policy: regeneration must treat it as
	// preserved, not as drift.
	writeBootstrapFile(t, dir, ".fakerc.json", `{"plugins":"react","rules":"strict","overrides":"keep"}`)

	spec := testBootstrapSpec()
	spec.NormalizeExpected = func(existing, expected map[string]string) map[string]string {
		if policy, ok := existing["overrides"]; ok {
			expected["overrides"] = policy
		}

		return expected
	}

	provider, err := BootstrapProviderFromSpec(spec)
	if err != nil {
		t.Fatalf("BootstrapProviderFromSpec: %v", err)
	}

	if err := provider.HealthCheck(bootstrapCtx(t, dir)); err != nil {
		t.Fatalf("preserved user policy must not read as drift: %v", err)
	}
}
