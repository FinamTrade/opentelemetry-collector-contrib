// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package aerospikereceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver"

import (
	"testing"

	as "github.com/aerospike/aerospike-client-go/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/cluster"
	cm "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/cluster/mocks"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/aerospikereceiver/mocks"
)

func TestAerospike_Info(t *testing.T) {
	t.Parallel()

	testNode0 := cm.NewNode(t)
	testNode0.On("GetName").Return("BB990C28F270008")
	testNode0.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "node", "statistics").Return(metricsMap{
		"node":       "BB990C28F270008",
		"statistics": "failed_best_practices=true;client_connections=1;client_connections_opened=9",
	}, nil)

	testNode1 := cm.NewNode(t)
	testNode1.On("GetName").Return("BB990C28F270009")
	testNode1.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "node", "statistics").Return(metricsMap{
		"node":       "BB990C28F270009",
		"statistics": "failed_best_practices=true;client_connections=1;client_connections_opened=9",
	}, nil)

	testNodes := []cluster.Node{
		testNode0,
		testNode1,
	}

	testCluster := mocks.NewNodeGetter(t)
	testCluster.On("GetNodes").Return(testNodes)
	testCluster.On("Close").Return()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	clientCfg := clientConfig{
		logger: logger.Sugar(),
	}

	nodeGetterFactoryFunc := func(cfg *clientConfig, policy *as.ClientPolicy, authEnabled bool) (nodeGetter, error) {
		return testCluster, nil
	}

	client, err := newASClient(&clientCfg, nodeGetterFactoryFunc)
	require.NoError(t, err)

	actualResults := client.Info()

	expectedMetrics := clusterInfo{
		"BB990C28F270008": metricsMap{
			"node":                      "BB990C28F270008",
			"failed_best_practices":     "true",
			"client_connections":        "1",
			"client_connections_opened": "9",
		},
		"BB990C28F270009": metricsMap{
			"node":                      "BB990C28F270009",
			"failed_best_practices":     "true",
			"client_connections":        "1",
			"client_connections_opened": "9",
		},
	}
	require.Equal(t, expectedMetrics, actualResults)

	client.Close()
}

func TestAerospike_NamespaceInfo(t *testing.T) {
	t.Parallel()

	testNode0 := cm.NewNode(t)
	testNode0.On("GetName").Return("BB990C28F270008")
	testNode0.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespaces").Return(metricsMap{"namespaces": "test;bar"}, nil)
	testNode0.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespace/test", "namespace/bar").Return(metricsMap{
		"namespace/test": "memory_used_bytes=27721728;memory_used_data_bytes=0;memory_used_index_bytes=0",
		"namespace/bar":  "memory_used_index_bytes=0;memory_used_set_index_bytes=0;memory_used_sindex_bytes=27721728",
	}, nil)

	testNode1 := cm.NewNode(t)
	testNode1.On("GetName").Return("BB990C28F270009")
	testNode1.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespaces").Return(metricsMap{"namespaces": "test;bar"}, nil)
	testNode1.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespace/test", "namespace/bar").Return(metricsMap{
		"namespace/test": "memory_used_bytes=27721728;memory_used_data_bytes=0;memory_used_index_bytes=0",
		"namespace/bar":  "memory_used_index_bytes=0;memory_used_set_index_bytes=0;memory_used_sindex_bytes=27721728",
	}, nil)

	testNodes := []cluster.Node{
		testNode0,
		testNode1,
	}

	testCluster := mocks.NewNodeGetter(t)
	testCluster.On("GetNodes").Return(testNodes)
	testCluster.On("Close").Return()

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	clientCfg := clientConfig{
		logger: logger.Sugar(),
	}

	nodeGetterFactoryFunc := func(cfg *clientConfig, policy *as.ClientPolicy, authEnabled bool) (nodeGetter, error) {
		return testCluster, nil
	}

	client, err := newASClient(&clientCfg, nodeGetterFactoryFunc)
	require.NoError(t, err)

	actualResults := client.NamespaceInfo()

	expectedMetrics := namespaceInfo{
		"BB990C28F270008": map[string]map[string]string{
			"test": metricsMap{
				"memory_used_bytes":       "27721728",
				"memory_used_data_bytes":  "0",
				"memory_used_index_bytes": "0",
			},
			"bar": metricsMap{
				"memory_used_index_bytes":     "0",
				"memory_used_set_index_bytes": "0",
				"memory_used_sindex_bytes":    "27721728",
			},
		},
		"BB990C28F270009": map[string]map[string]string{
			"test": metricsMap{
				"memory_used_bytes":       "27721728",
				"memory_used_data_bytes":  "0",
				"memory_used_index_bytes": "0",
			},
			"bar": metricsMap{
				"memory_used_index_bytes":     "0",
				"memory_used_set_index_bytes": "0",
				"memory_used_sindex_bytes":    "27721728",
			},
		},
	}
	require.Equal(t, expectedMetrics, actualResults)

	client.Close()
}

func TestAerospike_NamespaceInfo_Negative(t *testing.T) {
	t.Parallel()

	// test error in NamespaceInfo
	testNodeNeg := cm.NewNode(t)
	testNodeNeg.On("GetName").Return("BB990C28F270009")
	testNodeNeg.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespaces").Return(metricsMap{"namespaces": "test;bar"}, nil)
	testNodeNeg.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespace/test", "namespace/bar").Return(metricsMap{
		"ERROR:NOT_AUTHENTICATED": "",
	}, nil)

	testNodesNeg := []cluster.Node{
		testNodeNeg,
	}

	testClusterNeg := mocks.NewNodeGetter(t)
	testClusterNeg.On("GetNodes").Return(testNodesNeg)
	testClusterNeg.On("Close").Return()

	nodeGetterFactoryFuncNeg := func(cfg *clientConfig, policy *as.ClientPolicy, authEnabled bool) (nodeGetter, error) {
		return testClusterNeg, nil
	}

	logger, err := zap.NewDevelopment()
	require.NoError(t, err)

	clientCfg := clientConfig{
		logger: logger.Sugar(),
	}

	client, err := newASClient(&clientCfg, nodeGetterFactoryFuncNeg)
	require.NoError(t, err)

	actualResults := client.NamespaceInfo()
	require.Equal(t, namespaceInfo{"BB990C28F270009": map[string]map[string]string{}}, actualResults)

	client.Close()

	// test error in mapNodeInfoFunc
	testNodeNeg = cm.NewNode(t)
	testNodeNeg.On("GetName").Return("BB990C28F270009")
	testNodeNeg.On("RequestInfo", &as.InfoPolicy{Timeout: 0}, "namespaces").Return(nil, as.ErrNotAuthenticated)

	testNodesNeg = []cluster.Node{
		testNodeNeg,
	}

	testClusterNeg = mocks.NewNodeGetter(t)
	testClusterNeg.On("GetNodes").Return(testNodesNeg)
	testClusterNeg.On("Close").Return()

	nodeGetterFactoryFuncNeg = func(cfg *clientConfig, policy *as.ClientPolicy, authEnabled bool) (nodeGetter, error) {
		return testClusterNeg, nil
	}

	client, err = newASClient(&clientCfg, nodeGetterFactoryFuncNeg)
	require.NoError(t, err)

	actualResults = client.NamespaceInfo()
	require.Equal(t, namespaceInfo{"BB990C28F270009": map[string]map[string]string{}}, actualResults)

	client.Close()
}
