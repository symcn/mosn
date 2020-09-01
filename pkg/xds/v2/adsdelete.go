package v2

import (
	"strings"

	"mosn.io/mosn/pkg/admin/store"
	v2 "mosn.io/mosn/pkg/config/v2"
	clusterAdapter "mosn.io/mosn/pkg/upstream/cluster"
)

func DeleteEnvoyCluster(clusterNames []string) {
	// store.AddService
	sobj := store.GetMOSNConfig(store.CfgTypeCluster)
	if sobj == nil {
		return
	}
	storeClusters, ok := sobj.(map[string]v2.Cluster)
	if !ok {
		return
	}
	unexpectCluster := make([]string, 0, len(storeClusters))
	isExist := false

	for k := range storeClusters {
		if !strings.HasPrefix(k, "outbound") {
			continue
		}

		isExist = false
		for _, cn := range clusterNames {
			if strings.EqualFold(k, cn) {
				isExist = true
				break
			}
		}
		if !isExist {
			unexpectCluster = append(unexpectCluster, k)
		}
	}
	clusterAdapter.GetClusterMngAdapterInstance().RemovePrimaryCluster(unexpectCluster...)
}
