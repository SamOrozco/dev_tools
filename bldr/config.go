package main

type Config struct {
	RootDir    string       `yaml:"root_dir"` // the build will move to this dir to start the build
	Components []*Component `yaml:"components"`
}

type Component struct {
	Name     string `yaml:"name"`
	Location string `yaml:"location"`
	Build    *Build `yaml:"build"`
}

type Build struct {
	Commands  []*Command  `yaml:"commands"`
	Artifacts []*Artifact `yaml:"artifacts"`
}

type Command struct {
	Exec    string `yaml:"exec"`
	Windows string `yaml:"windows"`
	Mac     string `yaml:"mac"`
	Linux   string `yaml:"linux"`
}

type Artifact struct {
	Name string `yaml:"name"`
}
