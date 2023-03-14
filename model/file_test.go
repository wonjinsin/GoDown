package model

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Model ts test", func() {
	Describe("Make file", func() {
		var (
			input *Input
			file  *File
		)
		BeforeEach(func() {
			input = &Input{
				VideoType: "ts",
				URL:       "https://www.youtube.com/index_001.ts?token=dfljekf",
				Separator: "index_001.ts",
				Folder:    "Video1",
			}
			file = &File{
				Repo:  "../repo",
				Input: input,
			}
		})
		Context("SetFileSavePath", func() {
			It("should be a same", func() {
				Expect(file.SetFileSavePath()).To(Equal("../repo/Video1"))
			})
		})
		Context("SetURL", func() {
			It("should be a same", func() {
				Expect(file.SetURL(0)).To(Equal("https://www.youtube.com/index_000.ts?token=dfljekf"))
				Expect(file.SetURL(1)).To(Equal("https://www.youtube.com/index_001.ts?token=dfljekf"))
				Expect(file.SetURL(1000)).To(Equal("https://www.youtube.com/index_1000.ts?token=dfljekf"))
			})
		})
		Context("MakeDirectory", func() {
			It("shouldn't be error", func() {
				Expect(file.MakeDirectory()).Should(BeNil())
			})
		})
	})
})
