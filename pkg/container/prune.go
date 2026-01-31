package container

import (
	"fmt"
)

type Pruner struct {
	runtime Runtime
}

func NewPruner() *Pruner {
	return &Pruner{
		runtime: NewRuntime(),
	}
}

func (p *Pruner) Prune() error {
	if err := CheckRoot(); err != nil {
		return err
	}

	fmt.Printf("Pruning unused images using %s...\n", p.runtime.Name())
	if err := p.runtime.PruneImages(); err != nil {
		return fmt.Errorf("failed to prune images: %v", err)
	}

	fmt.Println("Unused images have been pruned successfully.")
	return nil
}
