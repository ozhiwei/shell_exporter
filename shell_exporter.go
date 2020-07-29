package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strconv"
	"sync"

	"gopkg.in/yaml.v2"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
)

var (
	showVersion   = flag.Bool("version", false, "Print version information.")
	configFile    = flag.String("config.file", "config.yml", "shell exporter configuration file.")
	listenAddress = flag.String("web.listen-address", ":9191", "The address to listen on for HTTP requests.")
	metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
)

type Config struct {
	Shells []*Shell `yaml:"shells"`
}

type Shell struct {
	Name         string            `yaml:"name"`
	Help         string            `yaml:"help"`
	Cmd          string            `yaml:"cmd"`
	Line         bool              `yaml:"line`
	ConstLabels  map[string]string `yaml:"const_labels"`
	LabelsRegexp string            `yaml:"labels_regexp"`
	Bin          string            `yaml:"bin"`

	VariableLabels []string

	Metrics []prometheus.Metric
	Output    string
	MatchMaps []map[string]string

	Desc *prometheus.Desc
}

type ShellManger struct {
	Config Config
	Shells []*Shell
}

func findStringSubmatchMaps(re *regexp.Regexp, s string) (matchMaps []map[string]string) {
	matchMaps = make([]map[string]string, 0)

	matchs := re.FindAllStringSubmatch(s, -1)
	labels := re.SubexpNames()[1:]

	for _, match := range matchs {
		matchMap := make(map[string]string)
		for index, name := range labels {
			matchMap[name] = match[index+1]
		}

		matchMaps = append(matchMaps, matchMap)
	}

	return
}

func NewShellManger() *ShellManger {

	yamlFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatalln("read config fail", err, *configFile)
	}

	config := Config{}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalln("parse yaml fail", err, *configFile)
	}

	shellManger := &ShellManger{Config: config}

	return shellManger
}

func (s *ShellManger) initShellManger() {
	s.Shells = s.Config.Shells

	for _, shell := range s.Shells {
		shell.init()
	}
}

func (s *ShellManger) Describe(ch chan<- *prometheus.Desc) {
	for _, shell := range s.Shells {
		ch <- shell.Desc
	}
}

func (s *ShellManger) Collect(ch chan<- prometheus.Metric) {
	s.runShells(ch)
}

func (s *ShellManger) runShells(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup


	for _, shell := range s.Shells {
		shell.Metrics = make([]prometheus.Metric, 0)

		wg.Add(1)
		go func(shell *Shell) {
			shell.run()
			shell.match()
			shell.collect()

			for _, metric := range shell.Metrics {
				ch <- metric
			}
			wg.Done()
		}(shell)
	}

	wg.Wait()
}

func (s *Shell) init() {
	lRe := regexp.MustCompile(s.LabelsRegexp)
	labels := lRe.SubexpNames()

	s.VariableLabels = make([]string, 0)

	for _, v := range labels[1:] {
		if v == "" {
			log.Fatalf("ERROR labels_regexp: '%s', '%s'", s.LabelsRegexp, labels)
		}

		if v != "value" {
			s.VariableLabels = append(s.VariableLabels, v)
		}
	}

	desc := prometheus.NewDesc(
		s.Name,
		s.Help,
		s.VariableLabels,
		s.ConstLabels,
	)
	s.Desc = desc
}

func (s *Shell) run() {
	var stdout, stderr bytes.Buffer

	cmd := exec.Command(s.Bin, "-c", s.Cmd)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Run()

	s.Output = stdout.String()
}

func (s *Shell) match() {
	re := regexp.MustCompile(s.LabelsRegexp)
	s.MatchMaps = findStringSubmatchMaps(re, s.Output)

	log.Debugf("Shell Run: '%s', '%s', '%s'", s.Cmd, s.Output, s.MatchMaps)
}

func (s *Shell) collect() {
	for _, matchMap := range s.MatchMaps {

		labelValues := make([]string, 0)
		for _, name := range s.VariableLabels {
			labelValues = append(labelValues, matchMap[name])
		}

		valueStr := matchMap["value"]
		value, _ := strconv.ParseFloat(valueStr, 64)


		metric := prometheus.MustNewConstMetric(
			s.Desc,
			prometheus.GaugeValue,
			value,
			labelValues...,
		)

		s.Metrics = append(s.Metrics, metric)
	}
}

func main() {
	flag.Parse()

	newShellManger := NewShellManger()
	newShellManger.initShellManger()
	prometheus.MustRegister(newShellManger)

	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Shell Exporter</title></head>
			<body>
			<h1>Shell Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Start Server and Listening on", *listenAddress)
	http.ListenAndServe(*listenAddress, nil)
}
