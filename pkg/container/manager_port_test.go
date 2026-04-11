package container

import (
	"testing"

	"reddock/pkg/config"
)

func TestBuildRunArgsPublishesHostADBPort(t *testing.T) {
	m := &Manager{containerName: "c2", config: &config.Config{}}
	c := &config.Container{
		Name:     "c2",
		Port:     5556,
		ImageURL: "redroid/redroid:13.0.0-latest",
		DataPath: "/tmp/reddock-test-data",
	}
	if c.HostADBPort() != 5556 {
		t.Fatalf("HostADBPort: got %d", c.HostADBPort())
	}
	args := m.buildRunArgs(c)
	var publish string
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-p" {
			publish = args[i+1]
			break
		}
	}
	if publish != "5556:5555" {
		t.Fatalf("docker -p mapping: got %q want 5556:5555 (args=%v)", publish, args)
	}
}
