package v1

type Pipeline struct {
	Name string
}

func (p Pipeline) CreatePipeline() Pipeline {
	p.Name = "dsadsad"

	return p
}
