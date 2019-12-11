package command

import (
	"fmt"
	"github.com/turbonomic/lemur/lemurctl/pkg/influx"
	"github.com/urfave/cli"
	"strings"
)

var (
	headerFormat  = "%-40s%-30s%-25s\n"
	contentFormat = "%-40s%-30s%-25s\n"
)

type cluster struct {
	clusterID   string
	clusterType string
	displayName string
}

var clusterMap = map[string]*cluster{}

func GetCluster(c *cli.Context) error {
	clusterType := c.String("type")
	if clusterType != "vm" && clusterType != "host" {
		return fmt.Errorf("you must specify a valid cluster type (vm, host)")
	}
	db, err := influx.NewDBInstance(c)
	if err != nil {
		return err
	}
	defer db.Close()
	fmt.Printf(headerFormat, "ID", "NAME", "TYPE")
	if clusterType == "vm" {
		return GetVMCluster(c, db)
	}
	return GetHostCluster(c, db)
}

func GetVMCluster(c *cli.Context, db *influx.DBInstance) error {
	row, err := db.Query(influx.NewDBQuery(c).
		WithQueryType("schema").
		WithColumns("VM_CLUSTER").
		WithName("commodity_sold").
		WithConditions("entity_type='VIRTUAL_MACHINE'"))
	if err != nil {
		return err
	}
	for _, value := range row.Values {
		addCluster(value[1].(string), "vm")
	}
	for _ , cluster := range clusterMap {
		fmt.Printf(
			contentFormat,
			cluster.clusterID,
			cluster.displayName,
			cluster.clusterType)
	}
	return nil
}

func GetHostCluster(c *cli.Context, db *influx.DBInstance) error {
	row, err := db.Query(influx.NewDBQuery(c).
		WithQueryType("schema").
		WithColumns("HOST_CLUSTER").
		WithName("commodity_sold").
		WithConditions("entity_type='PHYSICAL_MACHINE'"))
	if err != nil {
		return err
	}
	for _, value := range row.Values {
		addCluster(value[1].(string), "host")
	}
	for _ , cluster := range clusterMap {
		fmt.Printf(
			contentFormat,
			cluster.clusterID,
			cluster.displayName,
			cluster.clusterType)
	}
	return nil
}

func addCluster(rowValue string, clusterType string) {
	parts := strings.SplitN(rowValue, "::", 2)
	clusterID := parts[0]
	var displayName string
	if len(parts) > 1 {
		displayName = parts[1]
	}
	if _, ok := clusterMap[clusterID]; !ok {
		clusterMap[clusterID] = &cluster{
			clusterID:   clusterID,
			clusterType: clusterType,
			displayName: displayName,
		}
	} else {
		if clusterMap[clusterID].displayName == "" {
			clusterMap[clusterID].displayName = displayName
		}
	}
}