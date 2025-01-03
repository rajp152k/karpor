// Copyright The Karpor Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package insight

import (
	"github.com/KusionStack/karpor/pkg/core/entity"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// Global-related
type Statistics struct {
	ClusterCount           int `json:"clusterCount"`
	ResourceCount          int `json:"resourceCount"`
	ResourceGroupRuleCount int `json:"resourceGroupRuleCount"`
}

// Resource-related

type ResourceSummary struct {
	Resource          entity.ResourceGroup `json:"resource"`
	CreationTimestamp metav1.Time          `json:"creationTimestamp"`
	ResourceVersion   string               `json:"resourceVersion"`
	UID               types.UID            `json:"uid"`
}

type ResourceEvents struct {
	Resource       entity.ResourceGroup `json:"resource"`
	Count          int                  `json:"count"`
	Reason         string               `json:"reason"`
	Source         string               `json:"source"`
	Type           string               `json:"type"`
	LastTimestamp  metav1.Time          `json:"lastTimestamp"`
	FirstTimestamp metav1.Time          `json:"firstTimestamp"`
}

type ResourceTopology struct {
	ResourceGroup entity.ResourceGroup `json:"resourceGroup"`
	Parents       []string             `json:"parents"`
	Children      []string             `json:"children"`
}

// Cluster-related

type ClusterTopology struct {
	ResourceGroup entity.ResourceGroup `json:"resourceGroup"`
	Count         int                  `json:"count"`
	Relationship  map[string]string    `json:"relationship"`
}

type MetricPoint struct {
	Timestamp metav1.Time `json:"timestamp"`
	Value     float64     `json:"value"`
}

type ResourceMetrics struct {
	Points []MetricPoint `json:"points"`
}

type ClusterDetail struct {
	NodeCount     int    `json:"nodeCount"`
	ServerVersion string `json:"serverVersion"`
	ReadyNodes    int    `json:"readyNodes"`
	NotReadyNodes int    `json:"notReadyNodes"`

	MemoryCapacity int64   `json:"memoryCapacity"`
	MemoryUsage    float64 `json:"memoryUsage"`
	CPUCapacity    int64   `json:"cpuCapacity"`
	CPUUsage       float64 `json:"cpuUsage"`
	PodsCapacity   int64   `json:"podsCapacity"`
	PodsUsage      int64   `json:"podsUsage"`

	// Latency is the latency of the cluster in milliseconds
	Latency int64 `json:"latency"`

	MetricsEnabled bool `json:"metricsEnabled"`

	CPUMetrics    ResourceMetrics `json:"cpuMetrics"`
	MemoryMetrics ResourceMetrics `json:"memoryMetrics"`
}

// Audit-related

// ScoreData encapsulates the results of scoring an audited manifest. It provides
// a numerical score along with statistics about the total number of issues and
// their severities.
type ScoreData struct {
	// Score represents the calculated score of the audited manifest based on
	// the number and severity of issues. It provides a quantitative measure
	// of the security posture of the resources in the manifest.
	Score float64 `json:"score"`

	// ResourceTotal is the count of unique resources audited during the scan.
	ResourceTotal int `json:"resourceTotal"`

	// IssuesTotal is the total count of all issues found during the audit.
	// This count can be used to understand the overall number of problems
	// that need to be addressed.
	IssuesTotal int `json:"issuesTotal"`

	// SeverityStatistic is a mapping of severity levels to their respective
	// number of occurrences. It allows for a quick overview of the distribution
	// of issues across different severity categories.
	SeverityStatistic map[string]int `json:"severityStatistic"`
}

// GVK-related
type GVKSummary struct {
	Cluster string `json:"cluster"`
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
	Count   int    `json:"count"`
}

// Namespace-related
//
//nolint:tagliatelle
type NamespaceSummary struct {
	Cluster    string         `json:"cluster"`
	Namespace  string         `json:"namespace"`
	CountByGVK map[string]int `json:"countByGVK"`
}

// ResourceGroup-related
//
//nolint:tagliatelle
type ResourceGroupSummary struct {
	*entity.ResourceGroup
	CountByGVK map[string]int `json:"countByGVK"`
}

type KeyValuePair struct {
	key   string
	value int
}
