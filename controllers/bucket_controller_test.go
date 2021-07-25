package controllers

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	chartTestPath = "testdata/charts/helmchart"
)

var _ = Describe("BucketReconciler", func() {
	Describe("Calculating checksum", func() {
		var (
			tmpDir string
			bucket *BucketReconciler

			checksum string
		)

		BeforeEach(func() {
			tmpDir = setupTestCharts()
			bucket = &BucketReconciler{}

			var err error
			checksum, err = bucket.checksum(tmpDir)
			Expect(err).Should(Succeed())
		})

		AfterEach(func() {
			err := os.RemoveAll(tmpDir)
			Expect(err).Should(Succeed())
		})

		Context("When no changes to Helm Manifests have been performed", func() {
			It("should generate the same checksum", func() {
				actual, err := bucket.checksum(tmpDir)
				Expect(err).Should(Succeed())
				Expect(actual).Should(Equal(checksum))
			})
		})

		Context("When introducing a change in a Helm manifest", func() {
			BeforeEach(func() {
				err := os.WriteFile(filepath.Join(tmpDir, "some-file.yaml"), []byte("some-content"), 0644)
				Expect(err).Should(Succeed())
			})

			It("should generate a different checksum", func() {
				actual, err := bucket.checksum(tmpDir)
				Expect(err).Should(Succeed())
				Expect(actual).ShouldNot(Equal(checksum))
			})
		})

		Context("When moving Helm manifests to a subdirectory", func() {
			BeforeEach(func() {
				subdir := "duplicate"
				err := os.Mkdir(filepath.Join(tmpDir, subdir), 0755)
				Expect(err).Should(Succeed())

				filename := "duplicate.yaml"
				srcFile := filepath.Join(tmpDir, filename)
				content, err := os.ReadFile(srcFile)
				Expect(err).Should(Succeed())

				err = os.Remove(srcFile)
				Expect(err).Should(Succeed())

				err = os.WriteFile(filepath.Join(tmpDir, subdir, filename), content, 0644)
				Expect(err).Should(Succeed())
			})

			It("should generate a different checksum", func() {
				actual, err := bucket.checksum(tmpDir)
				Expect(err).Should(Succeed())
				Expect(actual).ShouldNot(Equal(checksum))
			})
		})
	})
})

func setupTestCharts() string {
	tmpDir, err := os.MkdirTemp(os.TempDir(), "bucket-controller-tests*")
	Expect(err).Should(Succeed())

	err = filepath.Walk(chartTestPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			err := os.MkdirAll(filepath.Join(tmpDir, withoutTestPath(path)), 0755)
			Expect(err).To(Succeed())
			return nil
		}

		content, err := os.ReadFile(path)
		Expect(err).To(Succeed())

		err = os.WriteFile(filepath.Join(tmpDir, withoutTestPath(path)), content, 0644)
		Expect(err).To(Succeed())

		return nil
	})

	Expect(err).Should(Succeed())

	return tmpDir
}

func withoutTestPath(path string) string {
	return strings.ReplaceAll(path, chartTestPath, "")
}
