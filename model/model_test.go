package model

func TestModel(t *testing.T) {
	gm.RegisterFailHandler(Fail)
	RunSpecs(t, "Test Model Suite")
}
