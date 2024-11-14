package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type PalladiumTestServerModelLogicDrawer struct {
	ID              int    `json:"id"`
	NumberOfDomains int    `json:"NumberOfDomains"`
	Status          string `json:"Status"`
	Domains         []struct {
		ID          int    `json:"id"`
		Owner       string `json:"Owner"`
		Host        string `json:"Host"`
		Pid         int    `json:"pid"`
		Design      string `json:"Design"`
		ElapsedTime string `json:"ElapsedTime"`
	} `json:"Domains"`
}

type PalladiumTestServerModel struct {
	Emulator     string `json:"Emulator"`
	Hardware     string `json:"Hardware"`
	SystemStatus string `json:"SystemStatus"`
	Clusters     []struct {
		ID          int                                   `json:"id"`
		NumOfLD     int                                   `json:"numOfLD"`
		Status      string                                `json:"status"`
		LogicDrawer []PalladiumTestServerModelLogicDrawer `json:"LogicDrawer"`
		Board       []struct {
			ID              int    `json:"id"`
			NumberOfDomains int    `json:"NumberOfDomains"`
			Status          string `json:"Status"`
			Domains         []struct {
				ID          int    `json:"id"`
				Owner       string `json:"Owner"`
				Host        string `json:"Host"`
				Pid         int    `json:"pid"`
				Design      string `json:"Design"`
				ElapsedTime string `json:"ElapsedTime"`
			} `json:"Domains"`
		} `json:"Board"`
	} `json:"Clusters"`
	AllTargetPods []struct {
		RackID  int `json:"RackId"`
		Targets []struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Length int    `json:"length"`
		} `json:"Targets"`
	} `json:"AllTargetPods"`
	AvailableHDSB []struct {
		RackID  int      `json:"RackId"`
		Targets []string `json:"Targets"`
	} `json:"AvailableHDSB"`
	AvailableTpods []struct {
		RackID  int      `json:"RackId"`
		Targets []string `json:"Targets"`
	} `json:"AvailableTpods"`
	UnavailableHDSB []struct {
		RackID  int      `json:"RackId"`
		Targets []string `json:"Targets"`
	} `json:"UnavailableHDSB"`
	UnavailableTpods []struct {
		RackID  int      `json:"RackId"`
		Targets []string `json:"Targets"`
	} `json:"UnavailableTpods"`
}

type ZebuBoardModel struct {
	Unit    string `json:"unit"`
	Module  string `json:"module"`
	System  string `json:"system"`
	Status  string `json:"status"`
	Session string `json:"session"`
	User    string `json:"user"`
	Server  string `json:"server"`
	Type    string `json:"type"`
}

type ZebuStatusModel struct {
	Boards       []ZebuBoardModel `json:"boards"`
	Formatboards []struct {
		Board  string `json:"board"`
		Status string `json:"status"`
	} `json:"formatboards"`
	Boards2 []struct {
		Unit   string `json:"unit"`
		Module string `json:"module"`
		System string `json:"system"`
		Status string `json:"status"`
	} `json:"boards2"`
	Formatboards2 []struct {
		Board  string `json:"board"`
		Status string `json:"status"`
	} `json:"formatboards2"`
}

type ConfigModel struct {
	MetricsInterval int `yaml:"metricsInterval"`
	Palladium       struct {
		Name  string `yaml:"name"`
		Cmd   string `yaml:"cmd"`
		Racks []struct {
			Name       int   `yaml:"name"`
			ClusterIDs []int `yaml:"clusterIDs"`
		} `yaml:"racks"`
	} `yaml:"palladium"`
	Zebu struct {
		Name string `yaml:"name"`
		Cmd  string `yaml:"cmd"`
	} `yaml:"zebu"`
}

func contains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// get status data from txt for testing
func ReadPalladiumTestServerDataTxt() []byte {
	wd, _ := os.Getwd()
	filePath := "palladium-test_server.txt"

	dat, _ := os.ReadFile(filepath.Join(wd, filePath))
	return dat
}

// get status data from palladium command
func ReadPalladiumTestServerData(c ConfigModel) []byte {
	cmd := exec.Command("bash", "-c", "export NO_PLATFORM_REL_CHECK=1;"+c.Palladium.Cmd)
	data, _ := cmd.Output()
	return data
}

func GetConfig() ConfigModel {
	wd, _ := os.Getwd()
	filePath := "config.yaml"

	dat, _ := os.ReadFile(filepath.Join(wd, filePath))
	config := ConfigModel{}
	_ = yaml.Unmarshal(dat, &config)
	return config
}

func GetPalladiumTestServerData(c ConfigModel, isTesting bool) PalladiumTestServerModel {
	m := PalladiumTestServerModel{}

	var data []byte

	if isTesting {
		data = ReadPalladiumTestServerDataTxt()
	} else {
		data = ReadPalladiumTestServerData(c)

	}
	_ = json.Unmarshal(data, &m)

	return ReformatPalladiumTestServerModel(m)
}

func ReformatPalladiumTestServerModel(m PalladiumTestServerModel) PalladiumTestServerModel {
	if m.Clusters[0].LogicDrawer == nil {
		m.Clusters[0].LogicDrawer = []PalladiumTestServerModelLogicDrawer{}
		for i, _ := range m.Clusters {
			for j, _ := range m.Clusters[i].Board {
				m.Clusters[0].LogicDrawer = append(m.Clusters[0].LogicDrawer, m.Clusters[i].Board[j])
			}
		}
	}

	return m
}

// get status data from txt for testing
func ReadZebuStatusDataTxt() []byte {
	wd, _ := os.Getwd()
	filePath := "zebu_zRscManager.txt"

	dat, _ := os.ReadFile(filepath.Join(wd, filePath))
	return dat
}

// get status data from zebu command
func ReadZebuStatusDataData(c ConfigModel) []byte {
	cmd := exec.Command(c.Zebu.Cmd)
	data, _ := cmd.Output()
	return data
}

func GetZebuStatusData(c ConfigModel, isTesting bool) ([]string, []string) {
	var data []byte

	if isTesting {
		data = ReadZebuStatusDataTxt()
	} else {
		data = ReadZebuStatusDataData(c)

	}
	lines := strings.Split(string(data), "\n")

	var (
		statusList          []string
		statusSmartZICEList []string
	)

	re := regexp.MustCompile(`[A-Z][0-9]+\.[A-Z][0-9]+\.[A-Z][0-9]+\s.*`)
	re_smartZICE := regexp.MustCompile(`[A-Z][0-9]+\.smartZICE\.[A-Z][0-9]+\s.*`)
	for _, line := range lines {
		if re.MatchString(line) {
			statusList = append(statusList, line)
		}
		if re_smartZICE.MatchString(line) {
			statusSmartZICEList = append(statusSmartZICEList, line)
		}
	}
	return statusList, statusSmartZICEList
}

func GetZebuStatusObjData(statusList []string) ZebuStatusModel {
	m := ZebuStatusModel{}
	for _, status := range statusList {
		var (
			session    string
			user       string
			server     string
			statusType string
		)
		statusCols := strings.Split(status, " ")
		boardCols := strings.Split(statusCols[0], ".")
		if len(statusCols) > 2 {
			session = statusCols[2]
			user = statusCols[3]
			server = statusCols[4]
			statusType = statusCols[5]
		}
		b := ZebuBoardModel{
			Unit:    boardCols[0],
			Module:  boardCols[1],
			System:  boardCols[2],
			Status:  statusCols[1],
			Session: session,
			User:    user,
			Server:  server,
			Type:    statusType,
		}
		m.Boards = append(m.Boards, b)
	}

	return m
}

func GetPalladiumAvailableBoards(m PalladiumTestServerModel, requestNum int, c ConfigModel) ([]int, error) {
	var boards []int
	for clusterIdx, cluster := range m.Clusters {
		for _, rack := range c.Palladium.Racks {
			if clusterIdx == rack.ClusterIDs[0] {
				boards = []int{}
			}
		}
		for _, board := range cluster.LogicDrawer {
			for domainIdx, domain := range board.Domains {
				// domain free
				if domainIdx == 0 && domain.Pid == 0 {
					boards = append(boards, board.ID)
					break
				} else {
					boards = []int{}
				}
			}
			if len(boards) == requestNum {
				return boards, nil
			}
		}

	}
	return []int{}, errors.New("boards not enough")
}

func ExportPalladiumMetrics(c ConfigModel, isTesting bool) {
	m := GetPalladiumTestServerData(c, isTesting)

	palladium_status_gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "palladium_status",
		Help: "palladium_status",
	},
		[]string{
			"emulator",
		},
	)

	if m.SystemStatus == "ONLINE" {
		palladium_status_gauge.With(
			prometheus.Labels{"emulator": m.Emulator},
		).Set(1)
	} else {
		palladium_status_gauge.With(
			prometheus.Labels{"emulator": m.Emulator},
		).Set(0)
	}

	palladium_cluster_status_gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "palladium_cluster_status",
		Help: "palladium_cluster_status",
	},
		[]string{
			"emulator", "rack", "cluster", "id",
		},
	)

	palladium_cluster_board_status_gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "palladium_cluster_board_status",
		Help: "palladium_cluster_board_status",
	},
		[]string{
			"emulator", "rack", "cluster", "board", "id",
		},
	)

	palladium_cluster_board_domain_status_gauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "palladium_cluster_board_domain_status",
		Help: "palladium_cluster_board_domain_status",
	},
		[]string{
			"emulator", "rack", "cluster", "board", "domain", "owner", "host", "pid", "design", "id",
		},
	)

	var rackId int

	go func() {
		for clusterIdx, cluster := range m.Clusters {
			for _, rack := range c.Palladium.Racks {
				if contains(rack.ClusterIDs, clusterIdx) {
					rackId = rack.Name
				}
			}
			if cluster.Status == "ONLINE" {
				palladium_cluster_status_gauge.With(
					prometheus.Labels{"emulator": m.Emulator, "rack": strconv.Itoa(rackId), "cluster": strconv.Itoa(clusterIdx), "id": m.Emulator + "-rack" + strconv.Itoa(rackId) + "-cluster" + strconv.Itoa(clusterIdx)},
				).Set(1)
			} else {
				palladium_cluster_status_gauge.With(
					prometheus.Labels{"emulator": m.Emulator, "rack": strconv.Itoa(rackId), "cluster": strconv.Itoa(clusterIdx), "id": m.Emulator + "-rack" + strconv.Itoa(rackId) + "-cluster" + strconv.Itoa(clusterIdx)},
				).Set(0)
			}

			for _, board := range cluster.LogicDrawer {
				if board.Status == "ONLINE" {
					palladium_cluster_board_status_gauge.With(
						prometheus.Labels{"emulator": m.Emulator, "rack": strconv.Itoa(rackId), "cluster": strconv.Itoa(clusterIdx), "board": strconv.Itoa(board.ID), "id": m.Emulator + "-rack" + strconv.Itoa(rackId) + "-cluster" + strconv.Itoa(clusterIdx) + "-board" + strconv.Itoa(board.ID)},
					).Set(1)
				} else {
					palladium_cluster_board_status_gauge.With(
						prometheus.Labels{"emulator": m.Emulator, "rack": strconv.Itoa(rackId), "cluster": strconv.Itoa(clusterIdx), "board": strconv.Itoa(board.ID), "id": m.Emulator + "-rack" + strconv.Itoa(rackId) + "-cluster" + strconv.Itoa(clusterIdx) + "-board" + strconv.Itoa(board.ID)},
					).Set(0)
				}

				for domainIdx, domain := range board.Domains {
					// domain free
					if domain.Pid == 0 {
						palladium_cluster_board_domain_status_gauge.With(
							prometheus.Labels{
								"emulator": m.Emulator,
								"rack":     strconv.Itoa(rackId),
								"cluster":  strconv.Itoa(clusterIdx),
								"board":    strconv.Itoa(board.ID),
								"domain":   strconv.Itoa(domainIdx),
								"id":       m.Emulator + "-rack" + strconv.Itoa(rackId) + "-cluster" + strconv.Itoa(clusterIdx) + "-board" + strconv.Itoa(board.ID) + "-domain" + strconv.Itoa(domainIdx),
								"owner":    "NONE",
								"host":     "UNKNOWN",
								"pid":      strconv.Itoa(domain.Pid),
								"design":   "",
							},
						).Set(0)
					} else {
						elapsedTime := strings.Split(domain.ElapsedTime, "")
						elapsedTime[2] = "h"
						elapsedTime[5] = "m"
						elapsedTimeStr := strings.Join(elapsedTime, "") + "s"
						comp, _ := time.ParseDuration(elapsedTimeStr)
						palladium_cluster_board_domain_status_gauge.With(
							prometheus.Labels{
								"emulator": m.Emulator,
								"rack":     strconv.Itoa(rackId),
								"cluster":  strconv.Itoa(clusterIdx),
								"board":    strconv.Itoa(board.ID),
								"domain":   strconv.Itoa(domainIdx),
								"id":       m.Emulator + "-rack" + strconv.Itoa(rackId) + "-cluster" + strconv.Itoa(clusterIdx) + "-board" + strconv.Itoa(board.ID) + "-domain" + strconv.Itoa(domainIdx),
								"owner":    domain.Owner,
								"host":     domain.Host,
								"pid":      strconv.Itoa(domain.Pid),
								"design":   domain.Design,
							},
						).Set(comp.Seconds())
					}
				}

			}

		}

		time.Sleep(time.Duration(c.MetricsInterval) * time.Second)
	}()

}

func main() {
	metricsBoolArg := flag.Bool("metrics", false, "start prometheus metrics exporter server")
	testFormatDataBoolArg := flag.Bool("test", false, "test format data")
	emuTypeArg := flag.String("emu", "", "emu type, palladium or zebu")
	emuBoardRequestNumArg := flag.Int("num", 1, "emu board request num")

	flag.Parse()

	c := GetConfig()

	if *metricsBoolArg {
		if *emuTypeArg == "palladium" {
			ExportPalladiumMetrics(c, *testFormatDataBoolArg)
		}

		if *emuTypeArg == "zebu" {
		}

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":27777", nil)

	} else {
		if *emuTypeArg == "palladium" {
			m := GetPalladiumTestServerData(c, *testFormatDataBoolArg)
			boards, _ := GetPalladiumAvailableBoards(m, *emuBoardRequestNumArg, c)
			// boards enough
			bb, _ := json.Marshal(map[string]interface{}{"boards": boards})
			fmt.Println(string(bb))
		}

		if *emuTypeArg == "zebu" {
			l1, _ := GetZebuStatusData(c, *testFormatDataBoolArg)
			m := GetZebuStatusObjData(l1)
			bb, _ := json.Marshal(map[string]interface{}{"boards": m})
			fmt.Println(string(bb))
		}
	}

}
