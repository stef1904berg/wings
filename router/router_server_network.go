package router

import (
	"github.com/docker/docker/api/types"
	"github.com/gin-gonic/gin"
	"github.com/pelican-dev/wings/environment"
	"github.com/pelican-dev/wings/router/middleware"
	"net/http"
)

// Shows all networks the server is connected to
func getNetworks(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil { /* error here */
	}

	server := middleware.ExtractServer(c)

	container, err := cli.ContainerInspect(c, server.ID())
	if err != nil {
	}

	networks := make([]APIResponse, len(container.NetworkSettings.Networks))
	i := 0 // .Networks doesn't have an index, so make our own.
	for _, v := range container.NetworkSettings.Networks {

		// Gets basic information about the network, as the container does not contain
		// network information like the driver.
		network, err := cli.NetworkInspect(c, v.NetworkID, types.NetworkInspectOptions{})
		if err != nil {
			continue
		}

		networks[i] = APIResponse{
			Name:      network.Name,
			Driver:    network.Driver,
			NetworkID: network.ID,
		}
		i++
	}

	c.JSON(http.StatusOK, networks)
}

// Makes a server join a network
func postJoinNetwork(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil {
	}

	server := ExtractServer(c)

	container, err := cli.ContainerInspect(c, server.ID())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "container does not exist"})
		return
	}

	var networkInfo Network
	if err := c.BindJSON(&networkInfo); err != nil {
	}

	network, err := cli.NetworkInspect(c, networkInfo.NetworkID, types.NetworkInspectOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "network does not exist: " + err.Error()})
		return
	}

	err = cli.NetworkConnect(c, network.ID, container.ID, nil)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "error adding container to network: " + err.Error()})
	}
}

// Makes a server leave a network
func postLeaveNetwork(c *gin.Context) {
	cli, err := environment.Docker()
	if err != nil {
	}

	server := ExtractServer(c)

	container, err := cli.ContainerInspect(c, server.ID())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "container doesnt exist"})
		return
	}

	var networkInfo Network
	if err := c.BindJSON(&networkInfo); err != nil {
	}

	network, err := cli.NetworkInspect(c, networkInfo.NetworkID, types.NetworkInspectOptions{})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "network does not exist"})
		return
	}

	err = cli.NetworkDisconnect(c, network.ID, container.ID, false)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "error removing container from network: " + err.Error()})
	}
}
