package openmc

import (
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var testModel = Model{
	Nuclides: []Nuclide{{"U235", 1}},
	Density:  19.1, Radius: 8.7,
	Particles: 1000, Batches: 50, Inactive: 10,
}

func TestWriteXML(t *testing.T) {
	dir := t.TempDir()
	if err := testModel.WriteXML(dir); err != nil {
		t.Fatal(err)
	}
	for file, want := range map[string]string{
		"materials.xml": `<nuclide name="U235" ao="1">`,
		"geometry.xml":  `type="sphere" coeffs="0 0 0 8.7" boundary="vacuum"`,
		"settings.xml":  `<run_mode>eigenvalue</run_mode>`,
	} {
		b, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(b), want) {
			t.Errorf("%s missing %q:\n%s", file, want, b)
		}
		if err := xml.Unmarshal(b, new(any)); err != nil {
			t.Errorf("%s is not well-formed: %v", file, err)
		}
	}
}

func TestValidate(t *testing.T) {
	for _, m := range []Model{{}, {Nuclides: testModel.Nuclides, Density: 1, Radius: 0, Particles: 1, Batches: 2}, {Nuclides: testModel.Nuclides, Density: 1, Radius: 1, Particles: 1, Batches: 5, Inactive: 5}} {
		if m.WriteXML(t.TempDir()) == nil {
			t.Errorf("expected error for %+v", m)
		}
	}
}

func TestParseKEff(t *testing.T) {
	k, sd, err := ParseKEff(" Combined k-effective        = 1.00003 +/- 0.00091\n")
	if err != nil || k != 1.00003 || sd != 0.00091 {
		t.Fatalf("got %v %v %v", k, sd, err)
	}
	if _, _, err := ParseKEff("nothing"); err == nil {
		t.Fatal("want error")
	}
}

func TestRunWithFakeBinary(t *testing.T) {
	dir := t.TempDir()
	bin := filepath.Join(dir, "fake-openmc")
	script := "#!/bin/sh\ntest -f settings.xml || exit 1\necho ' Combined k-effective        = 0.99500 +/- 0.00200'\n"
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	res, err := testModel.Run(context.Background(), bin, dir)
	if err != nil || res.KEff != 0.995 || res.StdDev != 0.002 {
		t.Fatalf("got %+v, %v", res, err)
	}
}
