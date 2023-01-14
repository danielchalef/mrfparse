/*
Copyright © 2023 Daniel Chalef

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package pipeline

import (
	"github.com/danielchalef/mrfparse/pkg/mrfparse/utils"
)

// A very simple composable pipeline framework. Steps are added to a pipeline and then run in order.
// Each step is timed and logged.

var log = utils.GetLogger()

// Step is an interface that defines a pipeline step. Name() returns the step name.
type Step interface {
	Name() string
	Run()
}

type Pipeline struct {
	Steps []Step
}

func (p *Pipeline) AddStep(step Step) {
	p.Steps = append(p.Steps, step)
}

// Run executes each step in the pipeline in the order of the Steps slice. Each step is timed and logged.
// No effort is made to manage errors or recover from them. Each step is responsible for handling errors.
func (p *Pipeline) Run() {
	var fn func()

	for _, step := range p.Steps {
		log.Infof("Running step: %s", step.Name())

		fn = func() { step.Run() }
		elapsed := utils.Timed(fn)
		log.Infof("Step %s completed in %d seconds", step.Name(), elapsed)
	}
}

// New creates a new pipeline with the provided steps.
func New(steps ...Step) *Pipeline {
	return &Pipeline{Steps: steps}
}
