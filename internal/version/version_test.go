package version

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = t.Describe("Version", func() {
	tempFileName := "tempVersionFile"
	tempVersion := "1.1.1"
	tempVersion2 := "1.13.1"
	testInfo := Info{
		Version:         "1.0.0",
		GitCommit:       "abcdef123456",
		GitCommitDate:   "2024-08-13T12:34:56Z",
		GitTreeState:    "clean",
		BuildDate:       "2024-08-13T12:34:56Z",
		GoVersion:       "go1.20.5",
		Compiler:        "gc",
		Platform:        "linux/amd64",
		Linkmode:        "external",
		BuildTags:       []string{"tag1", "tag2"},
		LDFlags:         "-X main.Version=1.0.0",
		SeccompEnabled:  true,
		AppArmorEnabled: false,
		Dependencies:    []string{"dep1", "dep2"},
	}
	testInfoStr := `Version:        1.0.0
GitCommit:      abcdef123456
GitCommitDate:  2024-08-13T12:34:56Z
GitTreeState:   clean
BuildDate:      2024-08-13T12:34:56Z
GoVersion:      go1.20.5
Compiler:       gc
Platform:       linux/amd64
Linkmode:       external
BuildTags:
  tag1
  tag2
LDFlags:          -X main.Version=1.0.0
SeccompEnabled:   true
AppArmorEnabled:  false
Dependencies:
  dep1
  dep2`
	testJSONInfoStr := `{
  "version": "1.0.0",
  "gitCommit": "abcdef123456",
  "gitCommitDate": "2024-08-13T12:34:56Z",
  "gitTreeState": "clean",
  "buildDate": "2024-08-13T12:34:56Z",
  "goVersion": "go1.20.5",
  "compiler": "gc",
  "platform": "linux/amd64",
  "linkmode": "external",
  "buildTags": [
    "tag1",
    "tag2"
  ],
  "ldFlags": "-X main.Version=1.0.0",
  "seccompEnabled": true,
  "appArmorEnabled": false,
  "dependencies": [
    "dep1",
    "dep2"
  ]
}`
	t.Describe("test setting version", func() {
		It("should succeed to parse version", func() {
			_, err := parseVersionConstant("1.1.1", "")
			Expect(err).ToNot(HaveOccurred())
			_, err = parseVersionConstant("1.1.1-dev", "")
			Expect(err).ToNot(HaveOccurred())
			_, err = parseVersionConstant("1.1.1-dev", "biglonggitcommit")
			Expect(err).ToNot(HaveOccurred())
		})
		It("should succeed to parse the version with a git commit", func() {
			gitCommit := "\"myfavoritecommit\""
			v, err := parseVersionConstant(tempVersion, gitCommit)
			Expect(err).ToNot(HaveOccurred())
			Expect(v.Build).To(HaveLen(1))
			trimmed := strings.Trim(gitCommit, "\"")
			Expect(v.Build[0]).To(Equal(trimmed))
		})
		It("should ignore empty git commit", func() {
			v, err := parseVersionConstant(tempVersion, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(v.Build).To(BeEmpty())
		})
		It("should fail to set a bad version", func() {
			_, err := parseVersionConstant("badversion", "")
			Expect(err).To(HaveOccurred())
		})
		It("should parse version with current version", func() {
			_, err := parseVersionConstant(Version, "")
			Expect(err).ToNot(HaveOccurred())
		})
		It("should write version for file writes", func() {
			version := tempVersion
			gitCommit := "fakeGitCommit"
			tempFileName := tempFileName
			tempFile := t.MustTempFile(tempFileName)
			Expect(os.WriteFile(tempFile, []byte(""), 0)).NotTo(HaveOccurred())

			err := writeVersionFile(tempFileName, gitCommit, version)
			defer os.Remove(tempFileName)
			Expect(err).ToNot(HaveOccurred())

			versionBytes, err := os.ReadFile(tempFileName)
			Expect(err).ToNot(HaveOccurred())

			versionConstantVersion, err := parseVersionConstant(version, gitCommit)
			Expect(err).ToNot(HaveOccurred())

			versionConstantJSON, err := versionConstantVersion.MarshalJSON()
			Expect(err).ToNot(HaveOccurred())

			Expect(string(versionBytes)).To(Equal(string(versionConstantJSON)))
		})
		It("should create dir for version file", func() {
			filename := "/tmp/crio/temp-testing-file"
			err := writeVersionFile(filename, "", tempVersion)
			Expect(err).ToNot(HaveOccurred())

			_, err = os.ReadFile(filename)
			Expect(err).ToNot(HaveOccurred())
		})
		It("should not wipe with empty version file", func() {
			upgrade, err := shouldCrioWipe("", tempVersion)
			Expect(upgrade).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should fail to upgrade with empty version file", func() {
			tempFileName := tempFileName
			_ = t.MustTempFile(tempFileName)

			upgrade, err := shouldCrioWipe(tempFileName, tempVersion)
			Expect(upgrade).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})
		It("should fail upgrade with faulty version", func() {
			tempFileName := "tempVersionFile"
			tempFile := t.MustTempFile(tempFileName)
			Expect(os.WriteFile(tempFile, []byte("bad version file"), 0o644)).NotTo(HaveOccurred())

			upgrade, err := shouldCrioWipe(tempFileName, tempVersion)
			Expect(upgrade).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})
		It("should fail to upgrade with same version", func() {
			oldVersion := tempVersion
			newVersion := tempVersion

			tempFileName := tempFileName
			_ = t.MustTempFile(tempFileName)

			err := writeVersionFile(tempFileName, "", oldVersion)
			defer os.Remove(tempFileName)
			Expect(err).ToNot(HaveOccurred())

			upgrade, err := shouldCrioWipe(tempFileName, newVersion)
			Expect(upgrade).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should not upgrade with sub minor release", func() {
			oldVersion := tempVersion
			newVersion := "1.1.2"

			tempFileName := tempFileName
			_ = t.MustTempFile(tempFileName)

			err := writeVersionFile(tempFileName, "", oldVersion)
			defer os.Remove(tempFileName)
			Expect(err).ToNot(HaveOccurred())

			upgrade, err := shouldCrioWipe(tempFileName, newVersion)
			Expect(upgrade).To(BeFalse())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should upgrade between versions", func() {
			oldVersion := "1.14.1"
			newVersion := tempVersion2

			tempFileName := tempFileName
			_ = t.MustTempFile(tempFileName)

			err := writeVersionFile(tempFileName, "", oldVersion)
			defer os.Remove(tempFileName)
			Expect(err).ToNot(HaveOccurred())

			upgrade, err := shouldCrioWipe(tempFileName, newVersion)
			Expect(upgrade).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should upgrade with major release", func() {
			oldVersion := tempVersion2
			newVersion := "2.0.0"

			tempFileName := tempFileName
			_ = t.MustTempFile(tempFileName)

			err := writeVersionFile(tempFileName, "", oldVersion)
			defer os.Remove(tempFileName)
			Expect(err).ToNot(HaveOccurred())

			upgrade, err := shouldCrioWipe(tempFileName, newVersion)
			Expect(upgrade).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		})
		It("should fail to upgrade with bad version", func() {
			oldVersion := "bad version format"
			newVersion := tempVersion2

			tempFileName := tempFileName
			_ = t.MustTempFile(tempFileName)

			err := writeVersionFile(tempFileName, "", oldVersion)
			Expect(err).To(HaveOccurred())

			upgrade, err := shouldCrioWipe(tempFileName, newVersion)
			Expect(upgrade).To(BeTrue())
			Expect(err).To(HaveOccurred())
		})
	})
	t.Describe("test generating string from info", func() {
		It("should succeed returning a formatted string", func() {
			infoString := testInfo.String()
			Expect(infoString).To(Equal(testInfoStr))
		})
		It("should succeed returning a JSON document", func() {
			jsonInfoString, err := testInfo.JSONString()
			Expect(jsonInfoString).To(Equal(testJSONInfoStr))
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
