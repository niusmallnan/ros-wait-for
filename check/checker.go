package check

import (
	"context"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	apiTypes "github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/niusmallnan/ros-wait-for/types"
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

const (
	containerNetwork = "network"
)

type Checker struct {
	timeout  time.Duration
	interval time.Duration

	containersMap map[string]bool
	interfacesMap map[string]bool

	client dockerClient.APIClient
}

func NewChecker(timeout, interval time.Duration, containers, interfaces string) (*Checker, error) {
	client, err := dockerClient.NewClient("unix:///var/run/system-docker.sock", "", nil, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to connect system-docker")
	}

	containersMap := make(map[string]bool)
	interfacesMap := make(map[string]bool)

	for _, name := range strings.Split(containers, ",") {
		containersMap[name] = false
	}
	for _, iface := range strings.Split(interfaces, ",") {
		interfacesMap[iface] = false
	}

	return &Checker{
		timeout:       timeout,
		interval:      interval,
		containersMap: containersMap,
		interfacesMap: interfacesMap,
		client:        client,
	}, nil
}

func (c *Checker) Check() types.Exit {
	var status1, status2 bool
	for {
		status1 = c.checkRunning()
		if _, ok := c.containersMap[containerNetwork]; ok {
			status2 = c.checkNetwork()
		}
		if status1 && status2 {
			return types.Exit{
				Success: true,
				Err:     nil,
			}
		}
		time.Sleep(c.interval)
	}
}

func (c *Checker) checkRunning() bool {
	containers, err := c.client.ContainerList(context.Background(), apiTypes.ContainerListOptions{})
	if err != nil {
		logrus.Errorf("Failed to get container list: %v", err)
		return false
	}
	for _, container := range containers {
		logrus.Debugf("Got container %s status: %s", container.Names, container.Status)
		if container.Status == "running" {
			for _, name := range container.Names {
				c.containersMap[name] = true
			}
		}
	}

	for _, running := range c.containersMap {
		if !running {
			return false
		}
	}
	return true
}

func (c *Checker) checkNetwork() bool {
	links, err := netlink.LinkList()
	if err != nil {
		logrus.Errorf("Failed to get netlinks: %v", err)
		return false
	}
	for _, link := range links {
		linkName := link.Attrs().Name

		addrs, err := netlink.AddrList(link, netlink.FAMILY_V4)
		if err != nil {
			logrus.Errorf("Failed to get ip address from %s: %v", linkName, err)
			return false
		}
		if len(addrs) > 0 {
			c.interfacesMap[linkName] = true
		}
	}

	for _, active := range c.interfacesMap {
		if !active {
			return false
		}
	}
	return true
}

func (c *Checker) ThrowTimeout() types.Exit {
	time.Sleep(c.timeout)
	return types.Exit{
		Success: false,
		Err:     errors.New("Error Timeout in checking"),
	}
}
