/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tdimitrov/tranqap/internal/tqlog"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type configParams struct {
	Targets []target
}

type target struct {
	Name        *string
	Host        *string
	Port        *int
	User        *string
	Key         *string
	Destination *string
	FilePattern *string `yaml:"file_pattern"`
	RotationCnt *int    `yaml:"file_rotation_count"`
	UseSudo     *bool   `yaml:"use_sudo"`
	FilterPort  *int    `yaml:"filter_port"`
}

func checkForDuplicates(config configParams) error {
	nameSet := make(map[string]struct{})

	for _, t := range config.Targets {
		_, exists := nameSet[*t.Name]

		if exists == true {
			return fmt.Errorf("target %s is defined more than once", *t.Name)
		}

		nameSet[*t.Name] = struct{}{}
	}

	return nil
}

func readConfigFromFile(fname string) (configParams, error) {
	confFile, err := ioutil.ReadFile(fname)
	if err != nil {
		return configParams{}, fmt.Errorf("%s. Run init subcommand to generate empty config or provide path to existing config with -c", err.Error())
	}

	conf, err := parseConfig(confFile)
	if err != nil {
		return configParams{}, fmt.Errorf("Error parsing %s: %s", fname, err.Error())
	}

	return conf, nil
}

func parseConfig(confFile []byte) (configParams, error) {
	var conf configParams
	var err error

	err = yaml.Unmarshal(confFile, &conf)
	if err != nil {
		return conf, err
	}

	// Basic validation
	if len(conf.Targets) == 0 {
		return conf, fmt.Errorf("No targets defined in config")
	}

	if err := checkForDuplicates(conf); err != nil {
		return conf, err
	}

	return conf, nil
}

func getClientConfig(t *target) (*ssh.ClientConfig, *string, error) {
	var clientConfig ssh.ClientConfig

	clientConfig.Auth = make([]ssh.AuthMethod, 0, 2)

	if t.Name == nil {
		return nil, nil, errors.New("Missing Name in configuration")
	}

	if t.User == nil {
		return nil, nil, fmt.Errorf("Missing user for target <%s> in configuration", *t.Name)
	}

	if t.Key == nil {
		return nil, nil, fmt.Errorf("Missing Key path for target <%s> in configuration", *t.Name)
	}

	if t.Host == nil {
		return nil, nil, fmt.Errorf("Missing Host for target <%s> in configuration", *t.Name)
	}

	if t.Port == nil {
		tqlog.Info("Port not set for target <%s>. Setting to 22.\n", *t.Name)
		t.Port = new(int)
		*t.Port = 22
	}

	if t.Destination == nil {
		return nil, nil, fmt.Errorf("Missing destination for target <%s> in configuration", *t.Name)
	}

	if t.FilePattern == nil {
		return nil, nil, fmt.Errorf("Missing File Pattern for target <%s>", *t.Name)
	}

	if t.RotationCnt == nil {
		tqlog.Info("File Rotation Count not set for target <%s>. Setting to 10.\n", *t.Name)
		t.RotationCnt = new(int)
		*t.RotationCnt = 10
	}

	if *t.RotationCnt < 0 {
		return nil, nil, fmt.Errorf("Invalid rotation count for target <%s> (%d)", *t.Name, *t.RotationCnt)
	}

	if t.UseSudo == nil {
		t.UseSudo = new(bool)
		*t.UseSudo = false
	}

	if t.FilterPort != nil {
		if *t.FilterPort < 1 || *t.FilterPort > 65535 {
			return nil, nil, fmt.Errorf("Invalid port number for Filter port parameter: %d. Expected value between 1 and 65535", *t.FilterPort)
		}
	}

	dest := fmt.Sprintf("%s:%d", *t.Host, *t.Port)

	clientConfig.User = *t.User
	clientConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	if t.Key != nil {
		key, err := ioutil.ReadFile(*t.Key)
		if err != nil {
			msg := fmt.Sprintf("unable to read private key: %v", err)
			return nil, nil, errors.New(msg)
		}

		// Create the Signer for this private key.
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			msg := fmt.Sprintf("unable to parse private key: %v", err)
			return nil, nil, errors.New(msg)
		}

		clientConfig.Auth = append(clientConfig.Auth, ssh.PublicKeys(signer))
	}

	return &clientConfig, &dest, nil
}

func generateSampleConfig(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists. Will not overwrite existing config", path)
	}

	name := "Target name. Informational identification only."
	host := "Hostname/IP address of the target."
	port := 22
	login := "SSH login."
	key := "Path to private key, used for authentication."
	dest := "Path to destination dir for the PCAP files."
	pattern := "Filename pattern for each pcap file. Index and file extension will be added to this string."
	rotCnt := 5
	useSudo := true
	filterPort := 22

	t := make([]target, 1, 1)
	t[0] = target{&name, &host, &port, &login, &key, &dest, &pattern, &rotCnt, &useSudo, &filterPort}
	conf := make(map[string][]target)
	conf["targets"] = t

	// And finally create the new file
	confYAML, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("Error serializing sample configuration: %s", err)
	}

	err = ioutil.WriteFile(path, confYAML, 0644)
	if err != nil {
		return fmt.Errorf("Error writing sample configuration to file: %s", err)
	}

	return nil
}

func (cp *configParams) getTargetsList() []string {
	targets := make([]string, 0, len(cp.Targets))

	for _, t := range cp.Targets {
		targets = append(targets, *t.Name)
	}

	return targets
}
