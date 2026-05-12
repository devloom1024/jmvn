package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectJDKVersion_PrefersJavaVersionFile(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, ".java-version"), []byte("17.0.8\n"), 0o644); err != nil {
		t.Fatalf("write .java-version: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(`<project></project>`), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "17" {
		t.Fatalf("expected 17, got %q", got)
	}
}

func TestDetectJDKVersion_FromPomCompilerRelease(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(`<project><properties><maven.compiler.release>21</maven.compiler.release></properties></project>`), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "21" {
		t.Fatalf("expected 21, got %q", got)
	}
}

func TestDetectJDKVersion_FromPomCompilerSource(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(`<project><properties><maven.compiler.source>11</maven.compiler.source></properties></project>`), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "11" {
		t.Fatalf("expected 11, got %q", got)
	}
}

func TestDetectJDKVersion_FromPomJavaVersion(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(`<project><properties><java.version>17</java.version></properties></project>`), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "17" {
		t.Fatalf("expected 17, got %q", got)
	}
}

func TestDetectJDKVersion_FromCompilerPluginRelease(t *testing.T) {
	projectDir := t.TempDir()
	pom := `<project><build><plugins><plugin><artifactId>maven-compiler-plugin</artifactId><configuration><release>21</release></configuration></plugin></plugins></build></project>`
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "21" {
		t.Fatalf("expected 21, got %q", got)
	}
}

func TestDetectJDKVersion_FromCompilerPluginSource(t *testing.T) {
	projectDir := t.TempDir()
	pom := `<project><build><plugins><plugin><artifactId>maven-compiler-plugin</artifactId><configuration><source>1.8</source></configuration></plugin></plugins></build></project>`
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "8" {
		t.Fatalf("expected 8, got %q", got)
	}
}

func TestDetectJDKVersion_FromMvnJdkConfig(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectDir, ".mvn"), 0o755); err != nil {
		t.Fatalf("mkdir .mvn: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".mvn", "jdk.config"), []byte(`/opt/jdks/jdk-21.0.2`), 0o644); err != nil {
		t.Fatalf("write jdk.config: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "21" {
		t.Fatalf("expected 21, got %q", got)
	}
}

func TestDetectJDKVersion_ResolvePlaceholderFromProperties(t *testing.T) {
	projectDir := t.TempDir()
	pom := `<project><properties><java>17</java><java.version>${java}</java.version></properties></project>`
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "17" {
		t.Fatalf("expected 17 from resolving ${java}, got %q", got)
	}
}

func TestDetectJDKVersion_PlaceholderNotFoundSkips(t *testing.T) {
	projectDir := t.TempDir()
	pom := `<project><properties><java.version>${unknown}</java.version></properties></project>`
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "" {
		t.Fatalf("expected empty when placeholder cannot be resolved, got %q", got)
	}
}

func TestDetectJDKVersion_NormalPropertyStillWorks(t *testing.T) {
	projectDir := t.TempDir()
	pom := `<project><properties><foo>bar</foo><java.version>21</java.version></properties></project>`
	if err := os.WriteFile(filepath.Join(projectDir, "pom.xml"), []byte(pom), 0o644); err != nil {
		t.Fatalf("write pom.xml: %v", err)
	}

	got := DetectJDKVersion(projectDir)
	if got != "21" {
		t.Fatalf("expected 21 from normal property, got %q", got)
	}
}
